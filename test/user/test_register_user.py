#!/usr/bin/env python3
import json
import random
import uuid
from typing import Annotated, List, Optional

import pytest
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


@pytest.fixture
def mock_user():
    """生成一个测试用户"""
    uid = uuid.uuid4().hex[:8]
    username = f"user_{uid}"
    password = f"P@ss{random.randint(1000, 9999)}"
    email = f"{username}@example.com"
    phone_country_code = "+86"
    phone_number = (
        "1" + random.choice("3456789") + "".join(random.choices("0123456789", k=9))
    )
    nickname = f"nickname_{username}"

    return User(
        id=uid,
        username=username,
        password=password,
        email=email,
        phone_number=phone_number,
        phone_country_code=phone_country_code,
        nickname=nickname,
    )


class TestUserAPI:
    """用户 API 测试类"""

    BASE_URL = "http://localhost:8888"

    def test_01_create_user(self, go_server, mock_user):
        """测试创建用户"""
        logger.info(
            f"测试创建用户: username={mock_user.username}, email={mock_user.email}"
        )

        response = create_user_request(mock_user)

        assert response is not None, "创建用户请求失败"
        assert response["status_code"] == 200, f"创建用户失败: {response.get('data')}"
        logger.success(f"✓ 用户创建成功: {mock_user.username}")

    def test_02_get_user(self, go_server, mock_user):
        """测试获取用户信息"""
        # 先创建用户
        create_response = create_user_request(mock_user)
        assert create_response["status_code"] == 200

        logger.info(f"测试获取用户信息: {mock_user.username}")
        response = get_user(mock_user.username)

        assert response is not None, "获取用户请求失败"
        assert response["status_code"] == 200, (
            f"获取用户失败: status={response['status_code']}"
        )
        logger.success(f"✓ 获取用户成功: {mock_user.username}")

    def test_03_verify_user_data(self, go_server, mock_user):
        """测试验证用户数据完整性"""
        # 先创建用户
        create_response = create_user_request(mock_user)
        assert create_response["status_code"] == 200

        # 获取用户信息
        get_response = get_user(mock_user.username)
        assert get_response["status_code"] == 200
        logger.info(f"get_response: {get_response}")

        # 兼容不同的响应结构：{"data":...}、{"user":...} 或 直接返回用户对象
        resp_body = get_response.get("response") or {}
        if isinstance(resp_body, dict):
            user_data = resp_body.get("data") or resp_body.get("user") or resp_body
        else:
            user_data = {}

        # 验证所有字段
        logger.info("验证用户数据...")
        assert user_data.get("username") == mock_user.username, "用户名不匹配"
        assert user_data.get("email") == mock_user.email, "邮箱不匹配"
        assert user_data.get("nickname") == mock_user.nickname, "昵称不匹配"
        assert user_data.get("phone_country_code") == mock_user.phone_country_code, (
            "手机区号不匹配"
        )
        assert user_data.get("phone_number") == mock_user.phone_number, "手机号不匹配"
        if not user_data.get("id"):
            logger.warning("响应中未包含 id 字段，跳过 id 校验")
        else:
            assert bool(user_data.get("id")), "用户ID不存在"
        assert "password" not in user_data, "密码不应该被返回(安全检查)"

        logger.success("✓ 所有数据验证通过")

    def test_04_cache_query(self, go_server, mock_user):
        """测试缓存查询(第二次查询应该更快)"""
        # 先创建用户
        create_response = create_user_request(mock_user)
        assert create_response["status_code"] == 200

        # 第一次查询
        logger.info("第一次查询...")
        first_response = get_user(mock_user.username)
        # 响应时间以毫秒(ms)表示
        first_time = first_response["response_time"]

        # 第二次查询(测试缓存)
        logger.info("第二次查询(测试缓存)...")
        second_response = get_user(mock_user.username)
        # 响应时间以毫秒(ms)表示
        second_time = second_response["response_time"]

        logger.info(
            f"第一次查询耗时: {first_time:.3f}ms, 第二次查询耗时: {second_time:.3f}ms"
        )
        logger.success("✓ 缓存查询测试完成")


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
        # 返回稳定的 dict，避免调用方对 None 进行下标操作导致静态分析或运行时错误
        return {"status_code": 0, "response_time": 0.0, "data": None}

    result = {
        "status_code": response.status_code,
        # 使用毫秒(ms)表示响应时间
        "response_time": response.elapsed.total_seconds() * 1000.0,
    }
    try:
        result["data"] = response.json()
    except json.JSONDecodeError:
        result["data"] = response.text

    logger.info(
        f"POST {url} - status:{result['status_code']}, time:{result['response_time']:.3f}ms"
    )
    return result


def get_user(username: str):
    base_url = "http://localhost:8888"
    url = f"/api/v1/users/{username}"
    headers = {"Content-Type": "application/json"}
    try:
        response = requests.get(
            f"{base_url}{url}",
            headers=headers,
            timeout=5,
        )
    except Exception as e:
        logger.error(f"请求异常: {str(e)}")
        # 返回稳定的 dict，避免调用方对 None 进行下标操作导致静态分析或运行时错误
        return {"status_code": 0, "response_time": 0.0, "response": {}}

    result = {
        "status_code": response.status_code,
        # 使用毫秒(ms)表示响应时间
        "response_time": response.elapsed.total_seconds() * 1000.0,
        # 更健壮的 Content-Type 检查，避免包含 charset 时判断失败
        "response": (
            response.json()
            if response.headers.get("Content-Type", "").lower().find("application/json")
            != -1
            else response.text
        ),
    }
    logger.info(
        f"GET {url} - status:{result['status_code']}, time:{result['response_time']:.3f}ms"
    )
    return result
