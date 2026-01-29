#!/usr/bin/env python3
import json
import random
import uuid
from typing import Annotated, List, Optional

import requests
from loguru import logger
from pydantic import BaseModel, ConfigDict, EmailStr, Field, StringConstraints


class User(BaseModel):
    id: Annotated[str, StringConstraints(min_length=1)] = Field(
        ..., description="用户 ID"
    )
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
    ] = Field(None, description="中国大陆手机号，11 位")

    nickname: Optional[Annotated[str, StringConstraints(max_length=50)]] = Field(
        None, description="昵称，最多 50 字符"
    )

    model_config = ConfigDict(extra="forbid")


class LoginRequest(BaseModel):
    username: Annotated[str, StringConstraints(min_length=3, max_length=50)] = Field(
        ..., description="用户名，长度 3-50"
    )
    password: Annotated[str, StringConstraints(min_length=6, max_length=100)] = Field(
        ..., description="密码，长度 6-100"
    )

    model_config = ConfigDict(extra="forbid")


def main():
    user = generate_mock_user()
    # 1. 创建用户
    logger.info(f"创建用户: username={user.username}, email={user.email}")
    create_response = create_user_request(user)

    if not create_response:
        logger.error("创建用户失败，跳过后续验证")
        return

    # 2. 验证创建响应
    if create_response["status_code"] == 200:
        logger.success(f"✓ 用户创建成功: {user.username}")
    else:
        logger.error(
            f"✗ 用户创建失败: status={create_response['status_code']}, data={create_response.get('data')}"
        )
        return

    # 3. 登录后，获取用户信息进行验证
    logger.info(f"用户登录: username={user.username}")
    user.password = user.password + "12"
    login_response = login_request(user.username, user.password)
    if not login_response:
        logger.error("用户登录请求失败")
        return

    if login_response["status_code"] == 200:
        logger.success(f"✓ 用户登录成功: {user.username}")
    else:
        logger.error(
            f"✗ 用户登录失败: status={login_response['status_code']}, data={login_response.get('data')}"
        )

    print(login_response)


def create_user_request(user: User):
    base_url = "http://localhost:8888"
    url = "/api/v1/users/register"
    payload = user.model_dump()
    headers = {"Content-Type": "application/json"}

    try:
        response = requests.post(
            f"{base_url}{url}",
            json=payload,
            headers=headers,
            timeout=5,
        )
    except requests.RequestException as e:
        logger.error(f"请求异常: {str(e)}")
        return None

    result = {
        "status_code": response.status_code,
        "response_time": response.elapsed.total_seconds(),
    }
    try:
        result["data"] = response.json()
    except json.JSONDecodeError:
        result["data"] = response.text

    logger.info(
        f"POST {url} - status:{result['status_code']}, time:{result['response_time']:.3f}s"
    )
    return result


def login_request(username: str, password: str):
    base_url = "http://localhost:8888"
    url = "/api/v1/users/auth/login"
    payload = LoginRequest(username=username, password=password).model_dump()
    headers = {"Content-Type": "application/json"}

    try:
        response = requests.post(
            f"{base_url}{url}",
            json=payload,
            headers=headers,
            timeout=5,
        )
    except requests.RequestException as e:
        logger.error(f"请求异常: {str(e)}")
        return None

    result = {
        "status_code": response.status_code,
        "response_time": response.elapsed.total_seconds(),
    }
    try:
        result["data"] = response.json()
    except json.JSONDecodeError:
        result["data"] = response.text

    logger.info(
        f"POST {url} - status:{result['status_code']}, time:{result['response_time']:.3f}s"
    )
    return result


def generate_mock_user():
    uid = uuid.uuid4().hex[:8]
    username = f"user_{uid}"
    password = f"P@ss{random.randint(1000, 9999)}"
    email = f"{username}@example.com"
    # 中国大陆手机号：11位，1开头，第二位 3-9
    phone_country_code = "+86"
    phone_number = (
        "1" + random.choice("3456789") + "".join(random.choices("0123456789", k=9))
    )
    nickname = f"nickname_{username}"

    user = User(
        id=uid,
        username=username,
        password=password,
        email=email,
        phone_number=phone_number,
        phone_country_code=phone_country_code,
        nickname=nickname,
    )
    return user


if __name__ == "__main__":
    main()
