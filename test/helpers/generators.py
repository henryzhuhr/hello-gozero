#!/usr/bin/env python3
"""
随机数据生成器模块

提供各种测试数据的随机生成功能
"""

import random


def get_random_str(length: int = 8) -> str:
    """
    生成指定长度的随机十六进制字符串

    Args:
        length: 字符串长度，默认 8

    Returns:
        str: 随机生成的十六进制字符串

    Example:
        >>> uid = get_random_str(8)
        >>> print(uid)  # 例如: "a3f8b2c1"
    """
    return "".join(random.choices("0123456789abcdef", k=length))


def get_random_phone() -> str:
    """
    生成一个随机的中国大陆手机号

    格式: 1[3-9]xxxxxxxxx (11位数字)

    Returns:
        str: 随机生成的手机号

    Example:
        >>> phone = get_random_phone()
        >>> print(phone)  # 例如: "13812345678"
    """
    second_digit = random.choice("3456789")
    remaining = "".join(random.choices("0123456789", k=9))
    return f"1{second_digit}{remaining}"
