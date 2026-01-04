#!/usr/bin/env python3
"""
用户模块测试辅助工具

提供用户相关的测试数据生成和辅助函数
"""

from typing import Annotated, Optional

from pydantic import BaseModel, ConfigDict, EmailStr, Field
from pydantic.types import StringConstraints

from test.helpers import get_random_phone, get_random_str


class User(BaseModel):
    """用户模型"""

    username: Annotated[str, StringConstraints(min_length=3, max_length=50)] = Field(
        ..., description="用户名，长度 3-50"
    )
    password: Annotated[str, StringConstraints(min_length=6, max_length=100)] = Field(
        ..., description="密码，长度 6-100"
    )
    email: Optional[EmailStr] = Field(None, description="邮箱，格式校验")

    phone_country_code: Optional[
        Annotated[str, StringConstraints(pattern=r"^\+[1-9]\d{0,3}$")]
    ] = Field("+86", description="手机号国际区号，例如 +86、+1、+44")

    phone_number: Optional[
        Annotated[str, StringConstraints(pattern=r"^1[3-9]\d{9}$")]
    ] = Field(None, description="手机号，11 位")

    nickname: Optional[Annotated[str, StringConstraints(max_length=50)]] = Field(
        None, description="昵称，最多 50 字符"
    )

    model_config = ConfigDict(extra="forbid")


def create_mock_user() -> User:
    """
    创建一个随机测试用户，避免数据冲突
    """
    uid = get_random_str(8)
    return User(
        username=f"testuser_{uid}",
        password="Test@123456",
        email=f"test_{uid}@example.com",
        phone_country_code="+86",
        phone_number=get_random_phone(),
        nickname=f"nickname_{uid}",
    )
