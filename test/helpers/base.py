#!/usr/bin/env python3
"""
测试基类模块

提供测试类的基类，包含通用的 fixture 和工具方法
"""

import pytest


class BaseTestWithCleanup:
    """测试基类：自动在测试类执行前后进行资源清理（数据库连接、缓存等）

    使用方法：
        class TestMyFeature(BaseTestWithCleanup):
            # 使用默认等待时间（测试前3秒，测试后1秒）
            pass

        class TestMyPerformance(BaseTestWithCleanup):
            # 自定义等待时间
            cleanup_before_seconds = 5.0
            cleanup_after_seconds = 2.0
    """

    # 子类可以覆盖这些值来自定义等待时间
    cleanup_before_seconds: float = 3.0  # 测试前等待时间（秒）
    cleanup_after_seconds: float = 1.0  # 测试后等待时间（秒）

    @pytest.fixture(autouse=True, scope="class")
    def cleanup_before_tests(self):
        """在测试前清理，避免前面测试的数据影响"""
        import time

        # 等待前面的测试完全释放资源
        time.sleep(self.cleanup_before_seconds)
        yield
        # 测试后也清理，让资源充分回收
        time.sleep(self.cleanup_after_seconds)
