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
        # user.username="admin"
        # user.email="adasdmi1212n"
        create_user_request(user)
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
        print({"error": str(e)})
        return

    result = {
        "status_code": response.status_code,
        "response_time": response.elapsed.total_seconds(),
    }
    try:
        result["data"] = response.json()
    except json.JSONDecodeError:
        result["data"] = response.text

    logger.info(f"url:{url}, {result}")


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
    except requests.RequestException as e:
        print({"error": str(e)})
        return
    result = {
        "status_code": response.status_code,
        "response_time": response.elapsed.total_seconds(),
        "response": response.json()
        if response.headers.get("Content-Type") == "application/json"
        else response.text,
    }
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
