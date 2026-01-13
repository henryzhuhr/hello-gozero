#!/usr/bin/env python3
import json

import requests
from loguru import logger

from test.user.helpers import User


def main():
    pass
    # 随机创建 n 个用户
    # for user in range(10):
    #     mock_user = create_mock_user()

    #     create_user_request(mock_user)

    get_user_list(page=1, page_size=10)


def get_user_list(page: int, page_size: int):
    base_url = "http://localhost:8888"
    url = "/api/v1/users/list"
    params = {"page": page, "page_size": page_size}
    headers = {"Content-Type": "application/json"}

    try:
        response = requests.get(
            f"{base_url}{url}",
            params=params,
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
        f"GET {url} - status:{result['status_code']}, time:{result['response_time']:.3f}s"
    )
    logger.info(f"Response Data: {result['data']}")
    return result


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


if __name__ == "__main__":
    main()
