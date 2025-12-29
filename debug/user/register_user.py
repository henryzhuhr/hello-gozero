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


def main():
    for user in random_generate_mock_user(1):
        # 1. 创建用户
        logger.info(f"创建用户: username={user.username}, email={user.email}")
        create_response = create_user_request(user)

        if not create_response:
            logger.error("创建用户失败，跳过后续验证")
            continue

        # 2. 验证创建响应
        if create_response["status_code"] == 200:
            logger.success(f"✓ 用户创建成功: {user.username}")
        else:
            logger.error(
                f"✗ 用户创建失败: status={create_response['status_code']}, data={create_response.get('data')}"
            )
            continue

        # 3. 获取用户信息
        logger.info(f"获取用户信息: {user.username}")
        get_response = get_user(user.username)

        if not get_response:
            logger.error("获取用户失败")
            continue

        # 4. 验证用户数据
        if get_response["status_code"] == 200:
            user_data = get_response.get("response", {}).get("data", {})

            # 验证关键字段
            checks = []
            checks.append(("用户名", user_data.get("username") == user.username))
            checks.append(("邮箱", user_data.get("email") == user.email))
            checks.append(("昵称", user_data.get("nickname") == user.nickname))
            checks.append(
                (
                    "手机区号",
                    user_data.get("phone_country_code") == user.phone_country_code,
                )
            )
            checks.append(
                ("手机号", user_data.get("phone_number") == user.phone_number)
            )
            checks.append(("用户ID存在", bool(user_data.get("id"))))
            checks.append(
                ("密码未返回", "password" not in user_data)
            )  # 安全检查：密码不应该被返回

            logger.info("=" * 60)
            logger.info("数据验证结果:")
            all_passed = True
            for check_name, passed in checks:
                status = "✓" if passed else "✗"
                logger.info(f"  {status} {check_name}: {'通过' if passed else '失败'}")
                if not passed:
                    all_passed = False
                    if check_name != "密码未返回":
                        expected = getattr(
                            user, check_name.lower().replace(" ", "_"), "N/A"
                        )
                        actual = user_data.get(check_name.lower().replace(" ", "_"))
                        logger.warning(f"    期望: {expected}, 实际: {actual}")

            if all_passed:
                logger.success("✓ 所有验证通过！")
            else:
                logger.error("✗ 部分验证失败")
            logger.info("=" * 60)
        else:
            logger.error(f"✗ 获取用户失败: status={get_response['status_code']}")

        # 5. 再次查询，测试缓存
        logger.info("第二次查询（测试缓存）...")
        get_user(user.username)


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
        return None

    result = {
        "status_code": response.status_code,
        "response_time": response.elapsed.total_seconds(),
        "response": response.json()
        if response.headers.get("Content-Type") == "application/json"
        else response.text,
    }
    logger.info(
        f"GET {url} - status:{result['status_code']}, time:{result['response_time']:.3f}s"
    )
    return result
    logger.info(f"url:{url}, {result}")


def random_generate_mock_user(num: int) -> List[User]:
    created_users: List[User] = []

    for _ in range(num):
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
        created_users.append(user)
    return created_users


if __name__ == "__main__":
    main()
