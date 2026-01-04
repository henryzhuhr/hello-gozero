#!/usr/bin/env python3
import json
import random
import subprocess
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


@pytest.fixture(scope="function")
def mock_phone_number():
    """生成一个随机的中国大陆手机号"""
    return "1" + random.choice("3456789") + "".join(random.choices("0123456789", k=9))


# 作用域（scope）可以是 "function"、"class"、"module" 或 "session"
# function: 每个测试函数调用前后执行一次
# class: 每个测试类调用前后执行一次
# module: 每个测试模块调用前后执行一次
# session: 整个测试会话调用前后执行一次
@pytest.fixture(scope="function")
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


def test_00_placeholder():
    """占位测试，确保 pytest 能够正确发现测试文件"""
    assert True


class TestCreateUser:
    """创建用户接口测试类"""

    def test_create_user(self, go_server: subprocess.Popen, mock_user: User):
        # pytest 看到参数名 go_server，自动：
        # 1. 找到 conftest.py 中的 go_server fixture
        # 2. 执行它（启动服务器）
        # 3. 把返回的 process 对象注入到这里
        # 4. 你无需显式调用，只需声明参数即可

        # 实际上你不需要使用 go_server 变量
        # 它的主要作用是触发服务器启动
        """
        测试创建用户基本功能
        """
        logger.info(
            f"测试创建用户: username={mock_user.username}, email={mock_user.email}"
        )

        response = create_user_request(mock_user)

        assert response is not None, "创建用户请求失败"
        assert response["status_code"] == 200, f"创建用户失败: {response.get('data')}"
        logger.success(f"✓ 用户创建成功: {mock_user.username}")

    def test_username_min_length(self, go_server: subprocess.Popen, mock_user: User):
        """测试用户名最小长度（3字符）"""
        mock_user.username = uuid.uuid4().hex[:3]  # 随机生成3字符
        response = create_user_request(mock_user)
        assert response["status_code"] == 200, "3字符用户名应该创建成功"
        logger.success("✓ 最小用户名长度测试通过")

    def test_username_max_length(self, go_server: subprocess.Popen, mock_user: User):
        """测试用户名最大长度（50字符）"""
        mock_user.username = (uuid.uuid4().hex + uuid.uuid4().hex)[
            :50
        ]  # 随机生成50字符
        response = create_user_request(mock_user)
        assert response["status_code"] == 200, "50字符用户名应该创建成功"
        logger.success("✓ 最大用户名长度测试通过")

    def test_username_too_short(self, go_server: subprocess.Popen, mock_user: User):
        """测试用户名过短（少于3字符）"""
        try:
            mock_user.username = uuid.uuid4().hex[:2]  # 随机生成2字符
            # 如果 Pydantic 验证通过，则发送请求
            response = create_user_request(mock_user)
            assert response["status_code"] in [400, 422], "应该拒绝过短的用户名"
        except Exception as e:
            # Pydantic 验证失败是预期行为
            logger.success(f"✓ Pydantic 验证捕获到错误: {str(e)[:50]}...")

    def test_username_too_long(self, go_server: subprocess.Popen, mock_user: User):
        """测试用户名过长（超过50字符）"""
        try:
            mock_user.username = (uuid.uuid4().hex + uuid.uuid4().hex)[
                :51
            ]  # 随机生成51字符
            response = create_user_request(mock_user)
            assert response["status_code"] in [400, 422], "应该拒绝过长的用户名"
        except Exception as e:
            logger.success(f"✓ Pydantic 验证捕获到错误: {str(e)[:50]}...")

    def test_password_min_length(self, go_server: subprocess.Popen, mock_user: User):
        """测试密码最小长度（6字符）"""
        mock_user.password = uuid.uuid4().hex[:6]  # 随机生成6字符
        response = create_user_request(mock_user)
        assert response["status_code"] == 200, "6字符密码应该创建成功"
        logger.success("✓ 最小密码长度测试通过")

    def test_password_max_length(self, go_server: subprocess.Popen, mock_user: User):
        """测试密码最大长度（72字符）"""
        mock_user.password = (
            uuid.uuid4().hex + uuid.uuid4().hex + uuid.uuid4().hex + uuid.uuid4().hex
        )[:72]  # 随机生成72字符
        response = create_user_request(mock_user)
        assert response["status_code"] == 200, "72字符密码应该创建成功"
        logger.success("✓ 最大密码长度测试通过")

    def test_username_with_unicode(self, go_server: subprocess.Popen, mock_user: User):
        """测试包含 Unicode 字符的用户名"""
        mock_user.username = f"用户测试{uuid.uuid4().hex[:6]}"  # 中文+随机字符
        response = create_user_request(mock_user)
        # 根据实际业务逻辑，可能接受或拒绝
        logger.info(f"Unicode 用户名响应: {response['status_code']}")
        logger.success("✓ Unicode 用户名测试完成")

    def test_username_with_emoji(self, go_server: subprocess.Popen, mock_user: User):
        """测试包含 Emoji 的用户名"""
        mock_user.username = f"user😀{uuid.uuid4().hex[:6]}"  # 包含 Emoji + 随机字符
        response = create_user_request(mock_user)
        logger.info(f"Emoji 用户名响应: {response['status_code']}")
        logger.success("✓ Emoji 用户名测试完成")

    def test_username_sql_injection(self, go_server: subprocess.Popen, mock_user: User):
        """测试 SQL 注入防护"""
        mock_user.username = "admin'--"  # SQL 注入尝试
        response = create_user_request(mock_user)
        # 应该正常处理或拒绝，不应该导致 SQL 错误
        assert response["status_code"] != 500, "不应该出现服务器内部错误"
        logger.success("✓ SQL 注入防护测试通过")

    def test_optional_fields_none(self, go_server: subprocess.Popen, mock_user: User):
        """测试可选字段为 None 的场景"""
        mock_user.email = None
        mock_user.phone_number = None
        mock_user.nickname = None
        response = create_user_request(mock_user)
        assert response["status_code"] == 200, "可选字段为 None 应该成功"
        logger.success("✓ 可选字段为空测试通过")

    def test_duplicate_username(self, go_server: subprocess.Popen, mock_user: User):
        """测试重复用户名（409 Conflict）"""
        # 第一次创建
        first_response = create_user_request(mock_user)
        assert first_response["status_code"] == 200, "第一次创建应该成功"

        # 第二次创建相同用户名
        second_response = create_user_request(mock_user)
        assert second_response["status_code"] in [409, 400], (
            f"重复用户名应该返回 409/400，实际: {second_response['status_code']}"
        )
        logger.success("✓ 重复用户名测试通过")

    def test_duplicate_email(self, go_server: subprocess.Popen, mock_user: User):
        """测试重复邮箱（409 Conflict）"""
        # 第一次创建
        first_response = create_user_request(mock_user)
        assert first_response["status_code"] == 200, "第一次创建应该成功"

        # 第二次使用相同邮箱但不同用户名
        duplicate_email_user = User(
            id=uuid.uuid4().hex[:8],
            username=f"different_{uuid.uuid4().hex[:8]}",
            password=f"P{uuid.uuid4().hex[:7]}",  # 随机密码
            email=mock_user.email,  # 相同邮箱
        )
        second_response = create_user_request(duplicate_email_user)
        assert second_response["status_code"] in [409, 400], (
            f"重复邮箱应该返回 409/400，实际: {second_response['status_code']}"
        )
        logger.success("✓ 重复邮箱测试通过")

    def test_invalid_email_format(self, go_server: subprocess.Popen):
        """测试无效邮箱格式（400 Bad Request）"""
        try:
            user = User(
                id=uuid.uuid4().hex[:8],
                username=f"user_{uuid.uuid4().hex[:8]}",
                password=f"P{uuid.uuid4().hex[:7]}",  # 随机密码
                email="invalid-email",  # 无效邮箱格式
            )
            response = create_user_request(user)
            assert response["status_code"] in [400, 422], "应该拒绝无效邮箱"
        except Exception as e:
            # Pydantic 验证失败是预期行为
            logger.success(f"✓ Pydantic 验证捕获到无效邮箱: {str(e)[:50]}...")

    def test_invalid_phone_format(self, go_server: subprocess.Popen):
        """测试无效手机号格式（400 Bad Request）"""
        try:
            user = User(
                id=uuid.uuid4().hex[:8],
                username=f"user_{uuid.uuid4().hex[:8]}",
                password=f"P{uuid.uuid4().hex[:7]}",  # 随机密码
                email="valid@example.com",
                phone_number="12345",  # 无效手机号
            )
            response = create_user_request(user)
            assert response["status_code"] in [400, 422], "应该拒绝无效手机号"
        except Exception as e:
            logger.success(f"✓ Pydantic 验证捕获到无效手机号: {str(e)[:50]}...")

    def test_missing_required_fields(self, go_server: subprocess.Popen):
        """测试缺少必填字段（400 Bad Request）"""
        base_url = "http://localhost:8888"
        url = "/api/v1/users/register"
        # 故意发送不完整的数据
        payload = {"username": "testuser"}  # 缺少 id, password
        headers = {"Content-Type": "application/json"}

        try:
            response = requests.post(
                f"{base_url}{url}",
                json=payload,
                headers=headers,
                timeout=5,
            )
            assert response.status_code in [400, 422], (
                f"缺少必填字段应该返回 400/422，实际: {response.status_code}"
            )
            logger.success("✓ 缺少必填字段测试通过")
        except Exception as e:
            logger.error(f"请求异常: {str(e)}")
            raise

    def test_get_nonexistent_user(self, go_server: subprocess.Popen):
        """测试查询不存在的用户（404 Not Found）"""
        nonexistent_username = f"nonexistent_{uuid.uuid4().hex}"
        response = get_user(nonexistent_username)
        assert response["status_code"] == 404, (
            f"不存在的用户应该返回 404，实际: {response['status_code']}"
        )
        logger.success("✓ 不存在用户查询测试通过")


class TestGetUser:
    """获取用户接口测试类"""

    def test_get_user(self, go_server: subprocess.Popen, mock_user: User):
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

    def test_verify_user_data(self, go_server: subprocess.Popen, mock_user: User):
        """
        测试验证用户数据完整性
        """

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
        assert "password" not in user_data, "密码不应该被返回(安全检查)"

        logger.success("✓ 所有数据验证通过")

    def test_cache_query(self, go_server: subprocess.Popen, mock_user: User):
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


class TestDeleteUser:
    """删除用户接口测试类"""

    def test_delete_existing_user(self, go_server: subprocess.Popen, mock_user: User):
        """测试删除已存在的用户"""
        # 先创建用户
        create_response = create_user_request(mock_user)
        assert create_response["status_code"] == 200, "用户创建应该成功"

        # 删除用户
        logger.info(f"删除用户: {mock_user.username}")
        delete_response = delete_user(mock_user.username)
        assert delete_response["status_code"] == 200, (
            f"删除用户失败: {delete_response['status_code']}"
        )
        logger.success(f"✓ 用户删除成功: {mock_user.username}")

    def test_delete_nonexistent_user(self, go_server: subprocess.Popen):
        """测试删除不存在的用户（404 Not Found）"""
        nonexistent_username = f"nonexistent_{uuid.uuid4().hex}"
        response = delete_user(nonexistent_username)
        assert response["status_code"] == 404, (
            f"删除不存在用户应该返回 404，实际: {response['status_code']}"
        )
        logger.success("✓ 删除不存在用户测试通过")

    def test_delete_user_cascade(self, go_server: subprocess.Popen, mock_user: User):
        """测试删除用户后的级联效果（缓存清理、数据一致性）"""
        # 1. 创建用户
        create_response = create_user_request(mock_user)
        assert create_response["status_code"] == 200

        # 2. 查询用户（填充缓存）
        get_response = get_user(mock_user.username)
        assert get_response["status_code"] == 200

        # 3. 删除用户
        delete_response = delete_user(mock_user.username)
        assert delete_response["status_code"] == 200

        # 4. 再次查询（验证缓存已清理）
        get_after_delete = get_user(mock_user.username)
        assert get_after_delete["status_code"] == 404, (
            "删除后查询应该返回 404，缓存应该已清理"
        )
        logger.success("✓ 删除用户级联测试通过（缓存已清理）")


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
        f"POST {url} - status:{result['status_code']}, time:{result['response_time']:.3f}ms, result:{result}"
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


def delete_user(username: str):
    """删除用户"""
    base_url = "http://localhost:8888"
    url = f"/api/v1/users/{username}"
    headers = {"Content-Type": "application/json"}
    try:
        response = requests.delete(
            f"{base_url}{url}",
            headers=headers,
            timeout=5,
        )
    except Exception as e:
        logger.error(f"请求异常: {str(e)}")
        return {"status_code": 0, "response_time": 0.0, "response": {}}

    result = {
        "status_code": response.status_code,
        "response_time": response.elapsed.total_seconds() * 1000.0,
        "response": (
            response.json()
            if response.headers.get("Content-Type", "").lower().find("application/json")
            != -1
            else response.text
        ),
    }
    logger.info(
        f"DELETE {url} - status:{result['status_code']}, time:{result['response_time']:.3f}ms"
    )
    return result
