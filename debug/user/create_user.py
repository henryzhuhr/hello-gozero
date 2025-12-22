#!/usr/bin/env python3


import json

import requests


def main():
    base_url = "http://localhost:8888"
    url = "/api/v1/user"
    payload = {
        "username": "bob",
        "password": "secret123",
        "email": "bob@example.com",
        "phone": "1234567890",
        "nickname": "Bobby",
    }
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

    print(result)


if __name__ == "__main__":
    main()
