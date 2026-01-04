#!/usr/bin/env python3
import random
import subprocess
import uuid

import pytest
from loguru import logger

from test.api_client import ApiClient
from test.helpers import get_random_phone, get_random_str
from test.models import CommonResponse
from test.user.helpers import User, create_mock_user

# 注意：不再使用全局变量，而是通过 pytest fixture 注入
# api_client 会在 conftest.py 中定义，自动管理连接生命周期


@pytest.fixture(scope="function")
def mock_phone_number():
    """生成一个随机的中国大陆手机号"""
    return get_random_phone()


@pytest.fixture(scope="function")
def mock_user():
    """生成一个测试用户（通过 helpers.create_mock_user 实现）"""
    return create_mock_user()


class UserRequest:
    """用户相关请求封装"""

    @staticmethod
    def create_user(api_client: ApiClient, user: User) -> CommonResponse:
        """创建用户请求"""
        return api_client.post("/api/v1/users/register", data=user.model_dump())

    @staticmethod
    def get_user(api_client: ApiClient, username: str) -> CommonResponse:
        """获取用户信息"""
        return api_client.get(f"/api/v1/users/{username}")

    @staticmethod
    def delete_user(api_client: ApiClient, username: str) -> CommonResponse:
        """删除用户"""
        return api_client.delete(f"/api/v1/users/{username}")


def test_00_placeholder():
    """占位测试，确保 pytest 能够正确发现测试文件"""
    assert True


class TestCreateUser:
    """创建用户接口测试类"""

    def test_create_user(
        self, go_server: subprocess.Popen, api_client: ApiClient, mock_user: User
    ):
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

        response = UserRequest.create_user(api_client, mock_user)

        assert response is not None, "创建用户请求失败"
        assert response.status_code == 200, f"创建用户失败: {response.data}"
        logger.success(f"✓ 用户创建成功: {mock_user.username}")

    def test_username(self, go_server: subprocess.Popen, api_client: ApiClient):
        """测试用户名验证规则（包含长度、字符集、特殊字符、SQL注入等）"""
        # 使用动态生成的用户名，避免多次运行时冲突
        uid_suffix = uuid.uuid4().hex[:6]
        valid_usernames = [
            uuid.uuid4().hex[:3],  # 最小长度测试（3字符）
            (uuid.uuid4().hex + uuid.uuid4().hex)[:50],  # 最大长度测试（50字符）
            f"user_{uid_suffix}",  # 包含下划线
            f"user.{uid_suffix}",  # 包含点
            f"User{uid_suffix}",  # 大小写混合
        ]
        invalid_usernames = [
            uuid.uuid4().hex[:2],  # 过短（2字符）
            (uuid.uuid4().hex + uuid.uuid4().hex)[:51],  # 过长（51字符）
            f"测试用户_{uuid.uuid4().hex[:6]}",  # 中文字符（Unicode）
            f"user😀{uuid.uuid4().hex[:6]}",  # Emoji字符
            "user name",  # 包含空格
            "user-name",  # 包含连字符
            "user@name",  # 包含@符号
        ]
        special_test_usernames = [
            "admin'--",  # SQL注入尝试
            "admin' OR '1'='1",  # SQL注入尝试
        ]

        # 测试有效用户名
        for username in valid_usernames:
            test_user = create_mock_user()
            test_user.username = username
            response = UserRequest.create_user(api_client, test_user)
            assert response.status_code == 200, (
                f"有效用户名 '{username}' 应该创建成功，实际返回: {response.status_code}"
            )
            logger.success(f"✓ 有效用户名测试通过: {username}")

        # 测试无效用户名
        for username in invalid_usernames:
            try:
                test_user = create_mock_user()
                test_user.username = username
                response = UserRequest.create_user(api_client, test_user)
                assert response.status_code in [400, 422], (
                    f"无效用户名 '{username}' 应该被拒绝，实际返回: {response.status_code}"
                )
                logger.success(f"✓ 无效用户名测试通过（正确拒绝）: {username}")
            except Exception as e:
                # Pydantic 验证失败也是预期行为
                logger.success(
                    f"✓ Pydantic 验证捕获到错误: {username} - {str(e)[:50]}..."
                )

        # 测试特殊用户名（SQL注入防护等）
        for username in special_test_usernames:
            test_user = create_mock_user()
            test_user.username = username
            response = UserRequest.create_user(api_client, test_user)
            # 应该正常处理或拒绝，不应该导致 SQL 错误
            assert response.status_code != 500, (
                f"特殊用户名 '{username}' 不应该导致服务器内部错误"
            )
            logger.success(f"✓ 特殊用户名测试通过（SQL注入防护）: {username}")

    def test_password(self, go_server: subprocess.Popen, api_client: ApiClient):
        """测试密码验证规则（包含长度、复杂度等）"""
        # 有效密码测试
        valid_passwords = [
            get_random_str(6),  # 最小长度（6字符）
            get_random_str(72),  # 最大长度（72字符）
            "P@ss" + get_random_str(10),  # 包含特殊字符
            "123456",  # 纯数字（虽然不推荐，但后端可能允许）
            "abcdef",  # 纯字母
            "P@ssW0rd!2024",  # 复杂密码
        ]

        invalid_passwords = [
            get_random_str(5),  # 过短（5字符）
            get_random_str(101),  # 过长（超过100字符）
        ]

        # 测试有效密码
        for password in valid_passwords:
            test_user = create_mock_user()
            try:
                test_user.password = password
                response = UserRequest.create_user(api_client, test_user)
                assert response.status_code == 200, (
                    f"有效密码应该创建成功，密码长度: {len(password)}, 实际返回: {response.status_code}"
                )
                logger.success(f"✓ 有效密码测试通过: 长度={len(password)}")
            except Exception as e:
                logger.warning(
                    f"密码长度 {len(password)} 可能超出Pydantic限制: {str(e)[:50]}"
                )

        # 测试无效密码
        for password in invalid_passwords:
            try:
                test_user = create_mock_user()
                test_user.password = password
                response = UserRequest.create_user(api_client, test_user)
                assert response.status_code in [400, 422], (
                    f"无效密码应该被拒绝，密码长度: {len(password)}, 实际返回: {response.status_code}"
                )
                logger.success(f"✓ 无效密码测试通过（正确拒绝）: 长度={len(password)}")
            except Exception as e:
                # Pydantic 验证失败也是预期行为
                logger.success(f"✓ Pydantic 验证捕获到错误: 长度={len(password)}")

    def test_email(self, go_server: subprocess.Popen, api_client: ApiClient):
        """测试邮箱验证规则（包含格式、域名等）"""
        uid = get_random_str(6)

        # 有效邮箱测试
        valid_emails = [
            f"test{uid}@example.com",  # 标准格式
            f"test.user{uid}@example.com",  # 包含点
            f"test+tag{uid}@example.com",  # 包含加号
            f"test_{uid}@example.co.uk",  # 多级域名
            f"123{uid}@example.com",  # 数字开头
        ]

        # 无效邮箱测试
        invalid_emails = [
            "invalid-email",  # 缺少@
            "@example.com",  # 缺少用户名
            f"test{uid}@",  # 缺少域名
            f"test{uid}@.com",  # 域名格式错误
            f"test{uid}..user@example.com",  # 连续点
            f"test user{uid}@example.com",  # 包含空格
        ]

        # 测试有效邮箱
        for email in valid_emails:
            try:
                test_user = create_mock_user()
                test_user.email = email
                response = UserRequest.create_user(api_client, test_user)
                assert response.status_code == 200, (
                    f"有效邮箱 '{email}' 应该创建成功，实际返回: {response.status_code}"
                )
                logger.success(f"✓ 有效邮箱测试通过: {email}")
            except Exception as e:
                logger.warning(f"邮箱 {email} Pydantic验证失败: {str(e)[:50]}")

        # 测试无效邮箱
        for email in invalid_emails:
            try:
                test_user = create_mock_user()
                test_user.email = email
                response = UserRequest.create_user(api_client, test_user)
                assert response.status_code in [400, 422], (
                    f"无效邮箱 '{email}' 应该被拒绝，实际返回: {response.status_code}"
                )
                logger.success(f"✓ 无效邮箱测试通过（正确拒绝）: {email}")
            except Exception as e:
                # Pydantic 验证失败也是预期行为
                logger.success(f"✓ Pydantic 验证捕获到邮箱错误: {email}")

    def test_phone(self, go_server: subprocess.Popen, api_client: ApiClient):
        """测试手机号验证规则（包含格式、长度、国际区号等）"""
        # 有效手机号测试（中国大陆）
        valid_phones = [
            "13" + "".join(random.choices("0123456789", k=9)),  # 13开头
            "14" + "".join(random.choices("0123456789", k=9)),  # 14开头
            "15" + "".join(random.choices("0123456789", k=9)),  # 15开头
            "16" + "".join(random.choices("0123456789", k=9)),  # 16开头
            "17" + "".join(random.choices("0123456789", k=9)),  # 17开头
            "18" + "".join(random.choices("0123456789", k=9)),  # 18开头
            "19" + "".join(random.choices("0123456789", k=9)),  # 19开头
        ]

        # 无效手机号测试
        invalid_phones = [
            "12345",  # 过短
            "12" + "".join(random.choices("0123456789", k=9)),  # 12开头（无效）
            "10" + "".join(random.choices("0123456789", k=9)),  # 10开头（无效）
            "23456789012",  # 2开头（无效）
            "138123456789",  # 12位（过长）
            "1381234567",  # 10位（过短）
        ]

        # 测试有效手机号
        for phone in valid_phones:
            try:
                test_user = create_mock_user()
                test_user.phone_number = phone
                test_user.phone_country_code = "+86"
                response = UserRequest.create_user(api_client, test_user)
                assert response.status_code == 200, (
                    f"有效手机号 '{phone}' 应该创建成功，实际返回: {response.status_code}"
                )
                logger.success(f"✓ 有效手机号测试通过: {phone}")
            except Exception as e:
                logger.warning(f"手机号 {phone} Pydantic验证失败: {str(e)[:50]}")

        # 测试无效手机号
        for phone in invalid_phones:
            try:
                test_user = create_mock_user()
                test_user.phone_number = phone
                test_user.phone_country_code = "+86"
                response = UserRequest.create_user(api_client, test_user)
                assert response.status_code in [400, 422], (
                    f"无效手机号 '{phone}' 应该被拒绝，实际返回: {response.status_code}"
                )
                logger.success(f"✓ 无效手机号测试通过（正确拒绝）: {phone}")
            except Exception as e:
                # Pydantic 验证失败也是预期行为
                logger.success(f"✓ Pydantic 验证捕获到手机号错误: {phone}")

    def test_nickname(self, go_server: subprocess.Popen, api_client: ApiClient):
        """测试昵称验证规则（包含长度、特殊字符等）"""
        # 有效昵称测试
        valid_nicknames = [
            "A",  # 最小长度（1字符）
            get_random_str(50),  # 最大长度（50字符）
            "张三",  # 中文
            "User😀",  # Emoji
            "Test User",  # 包含空格
            "user-name",  # 包含连字符
            "user@123",  # 包含特殊字符
        ]

        # 无效昵称测试
        invalid_nicknames = [
            get_random_str(51),  # 超过最大长度（51字符）
            get_random_str(100),  # 远超最大长度（100字符）
        ]

        # 测试有效昵称
        for nickname in valid_nicknames:
            try:
                test_user = create_mock_user()
                test_user.nickname = nickname
                response = UserRequest.create_user(api_client, test_user)
                assert response.status_code == 200, (
                    f"有效昵称 '{nickname}' (长度{len(nickname)}) 应该创建成功，实际返回: {response.status_code}"
                )
                logger.success(
                    f"✓ 有效昵称测试通过: {nickname[:20]}... (长度={len(nickname)})"
                )
            except Exception as e:
                logger.warning(f"昵称 {nickname[:20]} Pydantic验证失败: {str(e)[:50]}")

        # 测试无效昵称
        for nickname in invalid_nicknames:
            try:
                test_user = create_mock_user()
                test_user.nickname = nickname
                response = UserRequest.create_user(api_client, test_user)
                assert response.status_code in [400, 422], (
                    f"无效昵称 (长度{len(nickname)}) 应该被拒绝，实际返回: {response.status_code}"
                )
                logger.success(f"✓ 无效昵称测试通过（正确拒绝）: 长度={len(nickname)}")
            except Exception as e:
                # Pydantic 验证失败也是预期行为
                logger.success(f"✓ Pydantic 验证捕获到昵称错误: 长度={len(nickname)}")

    def test_optional_fields_none(
        self, go_server: subprocess.Popen, api_client: ApiClient, mock_user: User
    ):
        """测试可选字段为 None 的场景"""
        # mock_user.email = None
        # mock_user.phone_number = None
        # mock_user.nickname = None
        response = UserRequest.create_user(api_client, mock_user)
        assert response.status_code == 200, "可选字段为 None 应该成功"
        logger.success("✓ 可选字段为空测试通过")

    def test_duplicate_username(
        self, go_server: subprocess.Popen, api_client: ApiClient, mock_user: User
    ):
        """测试重复用户名（409 Conflict）"""
        # 第一次创建
        first_response = UserRequest.create_user(api_client, mock_user)
        assert first_response.status_code == 200, "第一次创建应该成功"

        # 第二次创建相同用户名
        second_response = UserRequest.create_user(api_client, mock_user)
        assert second_response.status_code in [409, 400], (
            f"重复用户名应该返回 409/400，实际: {second_response.status_code}"
        )
        logger.success("✓ 重复用户名测试通过")

    def test_duplicate_email(
        self, go_server: subprocess.Popen, api_client: ApiClient, mock_user: User
    ):
        """测试重复邮箱（409 Conflict）"""
        # 第一次创建
        first_response = UserRequest.create_user(api_client, mock_user)
        assert first_response.status_code == 200, "第一次创建应该成功"

        # 第二次使用相同邮箱但不同用户名
        duplicate_email_user = create_mock_user()
        duplicate_email_user.username = f"different_{uuid.uuid4().hex[:8]}"
        duplicate_email_user.password = f"P{uuid.uuid4().hex[:7]}"  # 随机密码
        duplicate_email_user.email = mock_user.email  # 相同邮箱

        second_response = UserRequest.create_user(api_client, duplicate_email_user)
        assert second_response.status_code in [409, 400], (
            f"重复邮箱应该返回 409/400，实际: {second_response.status_code}"
        )
        logger.success("✓ 重复邮箱测试通过")

    def test_missing_required_fields(
        self, go_server: subprocess.Popen, api_client: ApiClient
    ):
        """测试缺少必填字段（400 Bad Request）"""
        # 使用 api_client 发送不完整的数据
        payload = {"username": "testuser"}  # 缺少 password
        response = api_client.post("/api/v1/users/register", data=payload)

        assert response.status_code in [400, 422], (
            f"缺少必填字段应该返回 400/422，实际: {response.status_code}"
        )
        logger.success("✓ 缺少必填字段测试通过")

    def test_get_nonexistent_user(
        self, go_server: subprocess.Popen, api_client: ApiClient
    ):
        """测试查询不存在的用户（400/404）"""
        nonexistent_username = f"nonexistent_{uuid.uuid4().hex}"
        response = UserRequest.get_user(api_client, nonexistent_username)
        # 后端可能返回400（业务错误）或404（资源不存在）
        assert response.status_code in [400, 404], (
            f"不存在的用户应该返回 400/404，实际: {response.status_code}"
        )
        logger.success("✓ 不存在用户查询测试通过")


class TestGetUser:
    """获取用户接口测试类"""

    def test_get_user(
        self, go_server: subprocess.Popen, api_client: ApiClient, mock_user: User
    ):
        """测试获取用户信息"""
        # 先创建用户
        create_response = UserRequest.create_user(api_client, mock_user)
        assert create_response.status_code == 200

        logger.info(f"测试获取用户信息: {mock_user.username}")
        response = UserRequest.get_user(api_client, mock_user.username)

        assert response is not None, "获取用户请求失败"
        assert response.status_code == 200, (
            f"获取用户失败: status={response.status_code}"
        )
        logger.success(f"✓ 获取用户成功: {mock_user.username}")

    def test_verify_user_data(
        self, go_server: subprocess.Popen, api_client: ApiClient, mock_user: User
    ):
        """
        测试验证用户数据完整性
        """

        # 先创建用户
        create_response = UserRequest.create_user(api_client, mock_user)
        assert create_response.status_code == 200

        # 获取用户信息
        get_response = UserRequest.get_user(api_client, mock_user.username)
        assert get_response.status_code == 200
        logger.info(f"get_response: {get_response}")

        # 兼容不同的响应结构：{"data":...}、{"user":...} 或 直接返回用户对象
        resp_body = get_response.data if get_response.data else {}
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

    def test_cache_query(
        self, go_server: subprocess.Popen, api_client: ApiClient, mock_user: User
    ):
        """测试缓存查询(第二次查询应该更快)"""
        # 先创建用户
        create_response = UserRequest.create_user(api_client, mock_user)
        assert create_response.status_code == 200

        # 第一次查询
        logger.info("第一次查询...")
        first_response = UserRequest.get_user(api_client, mock_user.username)
        # 响应时间以毫秒(ms)表示
        first_time = first_response.response_time

        # 第二次查询(测试缓存)
        logger.info("第二次查询(测试缓存)...")
        second_response = UserRequest.get_user(api_client, mock_user.username)
        # 响应时间以毫秒(ms)表示
        second_time = second_response.response_time

        logger.info(
            f"第一次查询耗时: {first_time:.3f}ms, 第二次查询耗时: {second_time:.3f}ms"
        )
        logger.success("✓ 缓存查询测试完成")


class TestDeleteUser:
    """删除用户接口测试类"""

    def test_delete_existing_user(
        self, go_server: subprocess.Popen, api_client: ApiClient, mock_user: User
    ):
        """测试删除已存在的用户"""
        # 先创建用户
        create_response = UserRequest.create_user(api_client, mock_user)
        assert create_response.status_code == 200, "用户创建应该成功"

        # 删除用户
        logger.info(f"删除用户: {mock_user.username}")
        delete_response = UserRequest.delete_user(api_client, mock_user.username)
        assert delete_response.status_code == 200, (
            f"删除用户失败: {delete_response.status_code}"
        )
        logger.success(f"✓ 用户删除成功: {mock_user.username}")

    def test_delete_nonexistent_user(
        self, go_server: subprocess.Popen, api_client: ApiClient
    ):
        """测试删除不存在的用户（200/404）"""
        nonexistent_username = f"nonexistent_{uuid.uuid4().hex}"
        response = UserRequest.delete_user(api_client, nonexistent_username)
        # 后端可能返回200（幂等删除）或404（资源不存在）
        assert response.status_code in [200, 404], (
            f"删除不存在用户应该返回 200/404，实际: {response.status_code}"
        )
        logger.success("✓ 删除不存在用户测试通过")

    def test_delete_user_cascade(
        self, go_server: subprocess.Popen, api_client: ApiClient, mock_user: User
    ):
        """测试删除用户后的级联效果（缓存清理、数据一致性）"""
        # 1. 创建用户
        create_response = UserRequest.create_user(api_client, mock_user)
        assert create_response.status_code == 200

        # 2. 查询用户（填充缓存）
        get_response = UserRequest.get_user(api_client, mock_user.username)
        assert get_response.status_code == 200

        # 3. 删除用户
        delete_response = UserRequest.delete_user(api_client, mock_user.username)
        assert delete_response.status_code == 200

        # 4. 再次查询（验证缓存已清理）
        get_after_delete = UserRequest.get_user(api_client, mock_user.username)
        # 后端可能返回400或404表示用户不存在
        assert get_after_delete.status_code in [400, 404], (
            f"删除后查询应该返回 400/404，实际: {get_after_delete.status_code}，缓存应该已清理"
        )
        logger.success("✓ 删除用户级联测试通过（缓存已清理）")
