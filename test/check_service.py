#!/usr/bin/env python3
"""
快速验证脚本 - 在不启动完整测试套件的情况下验证服务是否正常
"""
import requests
from loguru import logger


def check_service():
    """检查服务是否运行"""
    try:
        response = requests.get("http://localhost:8888/api/v1/health", timeout=2)
        if response.status_code == 200:
            logger.success("✓ 服务正常运行在 http://localhost:8888")
            return True
        else:
            logger.error(f"✗ 服务返回异常状态码: {response.status_code}")
            return False
    except requests.RequestException as e:
        logger.error(f"✗ 无法连接到服务: {e}")
        logger.info("请先启动服务: go run app/main.go")
        return False


if __name__ == "__main__":
    check_service()
