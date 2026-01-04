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

"""测试套
Pytest Fixture 机制

@pytest.fixture(scope="session") 将 go_server 定义为一个会话级别的 fixture
scope="session" 表示这个 fixture 在整个测试会话中只创建一次，多个测试可以共享
"""


@pytest.fixture(scope="session")
def go_server():
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
    max_retries = 30
    for i in range(max_retries):
        try:
            response = requests.get("http://localhost:8888/api/health", timeout=1)
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
        process.wait(timeout=5)
    except Exception as e:
        logger.warning(f"停止服务器时出错: {e}")
        try:
            os.killpg(os.getpgid(process.pid), signal.SIGKILL)
        except:
            pass
    logger.success("Go 服务器已停止")


@pytest.fixture(scope="session")
def api_client():
    """创建 API 客户端 fixture，自动管理连接生命周期

    使用 session 级别，整个测试会话共享同一个客户端实例，
    测试结束后自动关闭连接池
    """
    client = ApiClient(base_url="http://localhost:8888")
    logger.debug("创建 ApiClient 实例")

    yield client

    # 测试结束后关闭连接
    client.close()
    logger.debug("ApiClient 连接已关闭")
