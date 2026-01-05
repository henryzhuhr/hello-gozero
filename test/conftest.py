#!/usr/bin/env python3
"""
pytest 配置文件
定义全局 fixtures 供所有测试使用
"""

import os
import signal
import subprocess
import time
from pathlib import Path

import pytest
import requests
from loguru import logger

from test.api_client import ApiClient
from test.performance_models import TestConfig

"""测试套
Pytest Fixture 机制

@pytest.fixture(scope="session") 将 go_server 定义为一个会话级别的 fixture
scope="session" 表示这个 fixture 在整个测试会话中只创建一次，多个测试可以共享
"""


@pytest.fixture(scope="session")
def go_server(test_config: TestConfig):
    """启动 Go 服务器并在测试结束后停止"""
    # 获取项目根目录
    project_root = Path(__file__).parent.parent

    logger.info("启动 Go 服务器...")

    # 启动服务器
    process = subprocess.Popen(
        ["go", "run", "app/main.go"],
        cwd=str(project_root),
        stdout=subprocess.PIPE,
        stderr=subprocess.PIPE,
        text=True,
        preexec_fn=os.setsid,  # 创建新进程组
    )

    # 等待服务器启动
    max_retries = test_config.service.startup_timeout
    for i in range(max_retries):
        try:
            response = requests.get(test_config.service.health_check_url, timeout=1)
            if response.status_code == 200:
                logger.success("Go 服务器启动成功")
                break
        except requests.RequestException:
            if i < max_retries - 1:
                time.sleep(1)
            else:
                logger.error("服务器启动超时")
                os.killpg(os.getpgid(process.pid), signal.SIGTERM)
                raise RuntimeError("服务器启动失败")

    yield process

    # 测试结束后停止服务器
    logger.info("停止 Go 服务器...")
    try:
        os.killpg(os.getpgid(process.pid), signal.SIGTERM)
        process.wait(timeout=test_config.service.shutdown_timeout)
    except Exception as e:
        logger.warning(f"停止服务器时出错: {e}")
        try:
            os.killpg(os.getpgid(process.pid), signal.SIGKILL)
        except:
            pass
    logger.success("Go 服务器已停止")


@pytest.fixture(scope="session")
def api_client(test_config: TestConfig):
    """创建 API 客户端 fixture，自动管理连接生命周期

    使用 session 级别，整个测试会话共享同一个客户端实例，
    测试结束后自动关闭连接池
    """
    client = ApiClient(base_url=test_config.service.base_url)
    logger.debug("创建 ApiClient 实例")

    yield client

    # 测试结束后关闭连接
    client.close()
    logger.debug("ApiClient 连接已关闭")


@pytest.fixture(scope="session")
def test_config():
    """加载并验证测试配置

    使用 Pydantic 进行配置格式校验，确保配置文件格式正确
    包含：服务配置、数据库配置、Redis配置、性能测试配置等
    """
    config_path = Path(__file__).parent / "test_config.yaml"
    try:
        config = TestConfig.load_from_yaml(config_path)
        logger.debug(f"测试配置加载成功: {config_path}")
        return config
    except Exception as e:
        logger.error(f"测试配置加载失败: {e}")
        raise


@pytest.fixture(scope="session")
def perf_config(test_config: TestConfig):
    """性能测试配置（向后兼容）

    从 test_config 中提取性能配置，方便现有测试使用
    """
    return test_config.performance
