#!/usr/bin/env python3
"""
测试辅助工具包

提供所有测试模块共享的公共方法和基类

导出说明：
- 从 generators 导出随机数据生成器
- 从 base 导出测试基类
- 从 client 导出 API 客户端
- 保持向后兼容，原有的 import 语句无需修改
"""

# 从子模块导出
from test.helpers.base import BaseTestWithCleanup
from test.helpers.client import ApiClient
from test.helpers.generators import get_random_phone, get_random_str

# 定义公共接口
__all__ = [
    # 随机数据生成器
    "get_random_str",
    "get_random_phone",
    # 测试基类
    "BaseTestWithCleanup",
    # API 客户端
    "ApiClient",
]
