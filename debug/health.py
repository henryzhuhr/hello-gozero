#!/usr/bin/env python3
"""
健康检查接口调用脚本
调用go-zero服务的health接口
"""

import json

import requests


def main():
    response = requests.get("http://localhost:8888/api/v1/health", timeout=5)

    result = {
        "status_code": response.status_code,
        "response_time": response.elapsed.total_seconds(),
    }
    try:
        result["data"] = response.json()
    except json.JSONDecodeError:
        result["data"] = response.text

    print(result)

    # response = requests.get("http://localhost:8888/api/v1/hello", timeout=5)

    # result = {
    #     "status_code": response.status_code,
    #     "response_time": response.elapsed.total_seconds(),
    # }
    # try:
    #     result["data"] = response.json()
    # except json.JSONDecodeError:
    #     result["data"] = response.text

    # print(result)


if __name__ == "__main__":
    main()
