#!/usr/bin/env python3
"""
pytest 配置文件
定义全局 fixtures 供所有测试使用
"""

import io
import os
import signal
import subprocess
import sys
import time
from datetime import datetime
from pathlib import Path
from typing import TextIO, cast

import pytest
import requests
from loguru import logger

from test.helpers import ApiClient
from test.performance_models import TestConfig


class TeeFile(io.TextIOBase):
    """同时写入文件和终端的类文件对象（兼容 Python 标准 I/O 接口）"""

    def __init__(self, file_path: Path):
        super().__init__()  # 调用父类初始化（可选，但规范）
        # 注意：用 buffering=1（行缓冲）可能更合适，但默认也行
        self.file = open(file_path, "w", encoding="utf-8")  # 指定编码更健壮
        self.terminal = sys.stdout

    def write(self, message: str) -> int:
        """实现标准 write 方法，同时写文件和终端"""
        if not message:
            return 0
        n1 = self.file.write(message)
        n2 = self.terminal.write(message)
        self.terminal.flush()
        self.file.flush()
        # 返回写入的字符数（符合 TextIOBase 协议）
        return max(n1, n2)  # return n1 或 max(n1, n2)，通常按主输出为准

    def flush(self):
        """实现标准 flush 方法，刷新缓冲区"""
        self.file.flush()
        self.terminal.flush()

    def close(self):
        """实现标准 close 方法，关闭文件"""
        if not self.file.closed:
            self.file.close()
        # 注意：不要关闭 sys.stdout！

    def fileno(self):
        """实现标准 fileno 方法，返回文件描述符（解决 subprocess 报错的核心）"""
        #  ⚠️ 注意：这个方法可能引发问题！
        # 因为 fileno() 应该返回底层文件描述符，但你有两个输出流。
        # subprocess 有时会调用它（比如重定向时），但你不能同时返回两个 fd。
        # 如果 subprocess 不实际使用 fileno（只是检查是否存在），可以保留；
        # 否则建议：**不要实现 fileno()，或让它抛出异常**。
        #
        # 实际上，subprocess 在 text=True + 自定义 IO 时通常不会用 fileno。
        # 但为了安全，你可以选择：
        #   - 删除 fileno() 方法，或
        #   - 返回文件的 fileno（但终端可能不同）
        return self.file.fileno()

    def __del__(self):
        """析构函数，确保对象销毁时关闭文件（兜底）"""
        self.close()


"""
测试套
Pytest Fixture 机制

@pytest.fixture(scope="session") 将 go_server 定义为一个会话级别的 fixture
scope="session" 表示这个 fixture 在整个测试会话中只创建一次，多个测试可以共享
"""


@pytest.fixture(scope="session")
def go_server(test_config: TestConfig):
    """启动 Go 服务器并在测试结束后停止"""
    # 获取项目根目录
    project_root = Path(__file__).parent.parent  # 这里假设 conftest.py 在 test/ 目录下

    logger.info("启动 Go 服务器...")

    # 将 Go 日志写入文件，避免 PIPE 填满导致子进程阻塞
    # - 根本原因：子进程（Go 服务器）往未读取的 PIPE 写日志，缓冲区填满后阻塞了服务器的正常运行，导致 HTTP 请求超时；
    # - 解决方案：将子进程的 stdout/stderr 重定向到文件，避开 PIPE 缓冲区限制，让服务器能持续输出日志且不阻塞；
    # - 现象差异：单测日志量小，缓冲区没满所以没问题；全量测试日志多，触发了缓冲区阻塞，问题才显现。
    log_dir = project_root / "logs"
    log_dir.mkdir(exist_ok=True)
    timestamp = datetime.now().strftime("%Y%m%d_%H%M%S")
    log_file = TeeFile(log_dir / f"pytest.{timestamp}.log")

    # 启动服务器
    process = subprocess.Popen(
        ["go", "run", "app/main.go"],
        cwd=str(project_root),
        stdout=cast(TextIO, log_file),  # 👈 关键：类型断言，显式告诉类型检查器
        stderr=subprocess.STDOUT,
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
                log_file.close()
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
        except Exception:
            pass
    logger.success("Go 服务器已停止")
    log_file.close()


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


@pytest.fixture(scope="function")
def db_helper(test_config: TestConfig):
    """数据库辅助工具

    提供数据库连接和清理功能
    每个测试函数结束后自动关闭连接

    示例用法：
        def test_something(db_helper):
            # 清空用户表
            db_helper.clear_user_table()

            # 查询数据
            users = db_helper.query("SELECT * FROM t_user")

            # 获取用户数量
            count = db_helper.get_user_count()
    """
    from test.helpers import DatabaseHelper

    db_config = test_config.database
    helper = DatabaseHelper(
        host=db_config.host,
        port=db_config.port,
        user=db_config.user,
        password=db_config.password,
        database=db_config.database,
        charset=db_config.charset,
    )

    yield helper

    # 测试结束后关闭连接
    helper.close()
