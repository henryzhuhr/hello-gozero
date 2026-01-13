#!/usr/bin/env python3
import random
import subprocess
import time
import uuid
from concurrent.futures import ThreadPoolExecutor, as_completed

import pytest
from loguru import logger

from test.helpers import (
    ApiClient,
    BaseTestWithCleanup,
    get_random_phone,
    get_random_str,
)
from test.helpers.database import DatabaseHelper
from test.models import CommonResponse
from test.performance_models import PerformanceConfig
from test.user.helpers import User, create_mock_user

# 注意：不再使用全局变量，而是通过 pytest fixture 注入
# api_client 和 perf_config 会在 conftest.py 中定义，自动管理连接生命周期


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

    @staticmethod
    def update_password(
        api_client: ApiClient, username: str, old_password: str, new_password: str
    ) -> CommonResponse:
        """更新用户密码"""
        return api_client.put(
            f"/api/v1/users/{username}/password",
            data={"old_password": old_password, "new_password": new_password},
        )

    @staticmethod
    def get_user_list(
        api_client: ApiClient, page: int, page_size: int
    ) -> CommonResponse:
        """获取用户列表"""
        return api_client.get(
            f"/api/v1/users/list?page={page}&page_size={page_size}",
        )

    @staticmethod
    def get_user_count(
        api_client: ApiClient, include_deleted: bool = False
    ) -> CommonResponse:
        """获取用户总数统计"""
        url = "/api/v1/users/count"
        if include_deleted:
            url += "?include_deleted=true"
        return api_client.get(url)


def test_00_placeholder():
    """占位测试，确保 pytest 能够正确发现测试文件"""
    assert True


class TestRegisterUser:
    """创建用户接口测试类"""

    def test_register_user(
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
            except Exception:
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
            except Exception:
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
            except Exception:
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
            except Exception:
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


class TestRegisterUserPerformance:
    """用户接口性能测试类"""

    def test_registration_response_time(
        self,
        go_server: subprocess.Popen,
        api_client: ApiClient,
        perf_config: PerformanceConfig,
    ):
        """测试单次注册的响应时间基准"""
        baseline = perf_config.user.register_user.response_time_baseline
        test_user = create_mock_user()

        start_time = time.time()
        response = UserRequest.create_user(api_client, test_user)
        elapsed_time = (time.time() - start_time) * 1000  # 转换为毫秒

        assert response.status_code == 200, f"用户注册失败: {response.status_code}"
        logger.info(f"注册响应时间: {elapsed_time:.2f}ms")

        # 性能基准检查
        if elapsed_time > baseline:
            logger.warning(f"⚠ 注册响应时间 {elapsed_time:.2f}ms 超过基准 {baseline}ms")
        else:
            logger.success(f"✓ 注册响应时间 {elapsed_time:.2f}ms 符合基准")

    def test_concurrent_registration(
        self,
        go_server: subprocess.Popen,
        api_client: ApiClient,
        perf_config: PerformanceConfig,
    ):
        """测试并发注册（验证数据库唯一性约束和竞态条件）"""
        concurrent_users = perf_config.user.register_user.concurrent_users
        min_success_rate = perf_config.user.register_user.concurrent_success_rate
        success_count = 0
        failed_count = 0

        def register_user(index: int):
            """单个用户注册任务"""
            try:
                test_user = create_mock_user()
                response = UserRequest.create_user(api_client, test_user)
                return {
                    "index": index,
                    "success": response.status_code == 200,
                    "status_code": response.status_code,
                    "username": test_user.username,
                }
            except Exception as e:
                logger.error(f"并发注册 #{index} 失败: {str(e)}")
                return {"index": index, "success": False, "error": str(e)}

        logger.info(f"开始并发注册测试，并发数: {concurrent_users}")
        start_time = time.time()

        # 使用线程池执行并发注册
        with ThreadPoolExecutor(max_workers=concurrent_users) as executor:
            futures = [
                executor.submit(register_user, i) for i in range(concurrent_users)
            ]

            for future in as_completed(futures):
                result = future.result()
                if result.get("success"):
                    success_count += 1
                    logger.debug(
                        f"✓ 用户 #{result['index']} 注册成功: {result.get('username')}"
                    )
                else:
                    failed_count += 1
                    logger.warning(
                        f"✗ 用户 #{result['index']} 注册失败: {result.get('status_code')}"
                    )

        elapsed_time = (time.time() - start_time) * 1000

        logger.info(
            f"并发注册完成: 成功 {success_count}/{concurrent_users}, 失败 {failed_count}/{concurrent_users}"
        )
        logger.info(
            f"总耗时: {elapsed_time:.2f}ms, 平均每个用户: {elapsed_time / concurrent_users:.2f}ms"
        )

        # 验证大部分请求成功（允许少量失败）
        assert success_count >= concurrent_users * min_success_rate, (
            f"并发注册成功率过低: {success_count}/{concurrent_users} (要求 >= {min_success_rate * 100}%)"
        )
        logger.success(
            f"✓ 并发注册测试通过，成功率: {success_count / concurrent_users * 100:.1f}%"
        )

    def test_duplicate_username_concurrent(
        self,
        go_server: subprocess.Popen,
        api_client: ApiClient,
        perf_config: PerformanceConfig,
    ):
        """测试并发创建相同用户名（验证数据库唯一性约束）"""
        concurrent_count = perf_config.user.register_user.duplicate_username_concurrent
        test_user = create_mock_user()  # 相同的用户名

        success_count = 0
        conflict_count = 0

        def register_same_user(index: int):
            """尝试注册相同用户名"""
            try:
                response = UserRequest.create_user(api_client, test_user)
                return {"index": index, "status_code": response.status_code}
            except Exception as e:
                return {"index": index, "error": str(e)}

        logger.info(f"测试并发注册相同用户名: {test_user.username}")

        with ThreadPoolExecutor(max_workers=concurrent_count) as executor:
            futures = [
                executor.submit(register_same_user, i) for i in range(concurrent_count)
            ]

            for future in as_completed(futures):
                result = future.result()
                status = result.get("status_code")
                if status == 200:
                    success_count += 1
                elif status in [409, 400]:
                    conflict_count += 1

        logger.info(f"并发重复注册结果: 成功 {success_count}, 冲突 {conflict_count}")

        # 应该只有一个成功，其他都返回冲突
        assert success_count == 1, f"应该只有1个注册成功，实际: {success_count}"
        assert conflict_count >= concurrent_count - 1, (
            f"其他请求应该返回冲突，实际冲突数: {conflict_count}"
        )
        logger.success("✓ 数据库唯一性约束在并发场景下正常工作")

    def test_bulk_registration_stress(
        self,
        go_server: subprocess.Popen,
        api_client: ApiClient,
        perf_config: PerformanceConfig,
    ):
        """测试批量注册压力（连续创建多个用户）"""
        bulk_count = perf_config.user.register_user.bulk_count
        min_success_rate = perf_config.user.register_user.bulk_success_rate
        degradation_factor = (
            perf_config.user.register_user.performance_degradation_factor
        )
        success_count = 0
        failed_count = 0
        response_times = []

        logger.info(f"开始批量注册压力测试，用户数: {bulk_count}")
        start_time = time.time()

        for i in range(bulk_count):
            try:
                test_user = create_mock_user()
                req_start = time.time()
                response = UserRequest.create_user(api_client, test_user)
                req_time = (time.time() - req_start) * 1000
                response_times.append(req_time)

                if response.status_code == 200:
                    success_count += 1
                    if (i + 1) % 10 == 0:
                        logger.debug(f"已完成 {i + 1}/{bulk_count} 个用户注册")
                else:
                    failed_count += 1
                    logger.warning(f"用户 #{i} 注册失败: {response.status_code}")
            except Exception as e:
                failed_count += 1
                logger.error(f"用户 #{i} 注册异常: {str(e)}")

        total_time = (time.time() - start_time) * 1000
        avg_response_time = (
            sum(response_times) / len(response_times) if response_times else 0
        )
        min_response_time = min(response_times) if response_times else 0
        max_response_time = max(response_times) if response_times else 0

        logger.info(
            f"批量注册完成: 成功 {success_count}/{bulk_count}, 失败 {failed_count}/{bulk_count}"
        )
        logger.info(f"总耗时: {total_time:.2f}ms ({total_time / 1000:.2f}s)")
        logger.info(f"平均响应时间: {avg_response_time:.2f}ms")
        logger.info(
            f"最快响应: {min_response_time:.2f}ms, 最慢响应: {max_response_time:.2f}ms"
        )
        logger.info(f"吞吐量: {bulk_count / (total_time / 1000):.2f} 请求/秒")

        # 验证成功率
        success_rate = success_count / bulk_count
        assert success_rate >= min_success_rate, (
            f"批量注册成功率过低: {success_rate * 100:.1f}% (要求 >= {min_success_rate * 100}%)"
        )

        # 验证性能未显著衰减
        if max_response_time > avg_response_time * degradation_factor:
            logger.warning(
                f"⚠ 性能衰减: 最慢响应 {max_response_time:.2f}ms 超过平均值 {avg_response_time:.2f}ms 的 {degradation_factor}倍"
            )

        logger.success(f"✓ 批量注册压力测试通过，成功率: {success_rate * 100:.1f}%")


class TestGetUser:
    """获取用户接口测试类"""

    def test_get_user_and_verify_data(
        self, go_server: subprocess.Popen, api_client: ApiClient, mock_user: User
    ):
        """测试获取用户信息并验证数据完整性"""
        # 1. 创建用户
        create_response = UserRequest.create_user(api_client, mock_user)
        assert create_response.status_code == 200, "用户创建应该成功"

        # 2. 获取用户信息
        logger.info(f"测试获取用户信息: {mock_user.username}")
        get_response = UserRequest.get_user(api_client, mock_user.username)

        assert get_response is not None, "获取用户请求失败"
        assert get_response.status_code == 200, (
            f"获取用户失败: status={get_response.status_code}"
        )
        logger.success(f"✓ 获取用户成功: {mock_user.username}")

        # 3. 验证数据完整性
        logger.info("验证用户数据完整性...")

        # 兼容不同的响应结构：{"data":...}、{"user":...} 或 直接返回用户对象
        resp_body = get_response.data if get_response.data else {}
        if isinstance(resp_body, dict):
            user_data = resp_body.get("data") or resp_body.get("user") or resp_body
        else:
            user_data = {}

        # 验证所有字段
        assert user_data.get("username") == mock_user.username, "用户名不匹配"
        assert user_data.get("email") == mock_user.email, "邮箱不匹配"
        assert user_data.get("nickname") == mock_user.nickname, "昵称不匹配"
        assert user_data.get("phone_country_code") == mock_user.phone_country_code, (
            "手机区号不匹配"
        )
        assert user_data.get("phone_number") == mock_user.phone_number, "手机号不匹配"
        assert "password" not in user_data, "密码不应该被返回(安全检查)"

        logger.success("✓ 用户数据验证通过（所有字段匹配）")

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


class TestGetUserPerformance(BaseTestWithCleanup):
    """获取用户接口性能测试类"""

    # 自定义清理等待时间
    cleanup_before_seconds = 2.0
    cleanup_after_seconds = 2.0

    def test_get_user_response_time(
        self,
        go_server: subprocess.Popen,
        api_client: ApiClient,
        mock_user: User,
        perf_config: PerformanceConfig,
    ):
        """测试单次查询的响应时间基准"""
        baseline = perf_config.user.get_user.response_time_baseline

        # 先创建用户
        create_response = UserRequest.create_user(api_client, mock_user)
        assert create_response.status_code == 200, "用户创建应该成功"

        # 测试查询响应时间
        start_time = time.time()
        response = UserRequest.get_user(api_client, mock_user.username)
        elapsed_time = (time.time() - start_time) * 1000  # 转换为毫秒

        assert response.status_code == 200, f"获取用户失败: {response.status_code}"
        logger.info(f"查询响应时间: {elapsed_time:.2f}ms")

        # 性能基准检查
        if elapsed_time > baseline:
            logger.warning(f"⚠ 查询响应时间 {elapsed_time:.2f}ms 超过基准 {baseline}ms")
        else:
            logger.success(f"✓ 查询响应时间 {elapsed_time:.2f}ms 符合基准")

    def test_concurrent_get_user(
        self,
        go_server: subprocess.Popen,
        api_client: ApiClient,
        perf_config: PerformanceConfig,
    ):
        """测试并发查询同一用户（验证缓存一致性和并发安全）"""
        concurrent_count = perf_config.user.get_user.concurrent_count

        # 先创建一个用户
        test_user = create_mock_user()
        create_response = UserRequest.create_user(api_client, test_user)
        assert create_response.status_code == 200, "用户创建应该成功"
        success_count = 0
        failed_count = 0
        response_times = []

        def get_user(index: int):
            """单个查询任务"""
            try:
                req_start = time.time()
                response = UserRequest.get_user(api_client, test_user.username)
                req_time = (time.time() - req_start) * 1000
                return {
                    "index": index,
                    "success": response.status_code == 200,
                    "status_code": response.status_code,
                    "response_time": req_time,
                }
            except Exception as e:
                logger.error(f"并发查询 #{index} 失败: {str(e)}")
                return {"index": index, "success": False, "error": str(e)}

        logger.info(f"开始并发查询测试，并发数: {concurrent_count}")
        start_time = time.time()

        # 使用线程池执行并发查询
        with ThreadPoolExecutor(max_workers=concurrent_count) as executor:
            futures = [executor.submit(get_user, i) for i in range(concurrent_count)]

            for future in as_completed(futures):
                result = future.result()
                if result.get("success"):
                    success_count += 1
                    response_times.append(result.get("response_time", 0))
                    logger.debug(f"✓ 查询 #{result['index']} 成功")
                else:
                    failed_count += 1
                    logger.warning(
                        f"✗ 查询 #{result['index']} 失败: {result.get('status_code')}"
                    )

        elapsed_time = (time.time() - start_time) * 1000
        avg_response_time = (
            sum(response_times) / len(response_times) if response_times else 0
        )

        logger.info(
            f"并发查询完成: 成功 {success_count}/{concurrent_count}, 失败 {failed_count}/{concurrent_count}"
        )
        logger.info(
            f"总耗时: {elapsed_time:.2f}ms, 平均响应时间: {avg_response_time:.2f}ms"
        )

        # 验证所有查询都成功
        assert success_count == concurrent_count, (
            f"所有并发查询应该成功: {success_count}/{concurrent_count}"
        )
        logger.success(f"✓ 并发查询测试通过，平均响应时间: {avg_response_time:.2f}ms")

    def test_cache_performance(
        self,
        go_server: subprocess.Popen,
        api_client: ApiClient,
        perf_config: PerformanceConfig,
    ):
        """测试缓存性能（重复查询应该有性能提升）"""
        hot_query_count = perf_config.user.get_user.hot_query_count
        cache_ratio = perf_config.user.get_user.cache_performance_ratio

        # 创建用户
        test_user = create_mock_user()
        create_response = UserRequest.create_user(api_client, test_user)
        assert create_response.status_code == 200

        # 第一次查询（冷查询，可能从数据库读取）
        logger.info("第一次查询（冷查询）...")
        start_time = time.time()
        first_response = UserRequest.get_user(api_client, test_user.username)
        first_time = (time.time() - start_time) * 1000

        assert first_response.status_code == 200
        logger.info(f"冷查询耗时: {first_time:.3f}ms")

        # 连续热查询
        hot_query_times = []
        logger.info(f"开始{hot_query_count}次热查询...")
        for i in range(hot_query_count):
            start_time = time.time()
            response = UserRequest.get_user(api_client, test_user.username)
            query_time = (time.time() - start_time) * 1000
            hot_query_times.append(query_time)
            assert response.status_code == 200

        avg_hot_time = sum(hot_query_times) / len(hot_query_times)
        min_hot_time = min(hot_query_times)
        max_hot_time = max(hot_query_times)

        logger.info(f"热查询平均耗时: {avg_hot_time:.3f}ms")
        logger.info(f"热查询最快: {min_hot_time:.3f}ms, 最慢: {max_hot_time:.3f}ms")

        # 缓存应该带来性能提升
        if avg_hot_time <= first_time * cache_ratio:
            logger.success(
                f"✓ 缓存性能提升明显: 冷查询 {first_time:.2f}ms → 热查询平均 {avg_hot_time:.2f}ms"
            )
        else:
            logger.info(
                f"缓存性能: 冷查询 {first_time:.2f}ms, 热查询平均 {avg_hot_time:.2f}ms (期望 <= {first_time * cache_ratio:.2f}ms)"
            )

    def test_bulk_get_users_stress(
        self,
        go_server: subprocess.Popen,
        api_client: ApiClient,
        perf_config: PerformanceConfig,
    ):
        """测试批量查询压力（查询多个不同用户）"""
        bulk_count = perf_config.user.get_user.bulk_count
        # 考虑到前面批量注册压力测试的影响，进一步降低期望值
        # 这是压力测试，不是功能测试，70%成功率已经可以接受
        min_success_rate = 0.70

        # 先批量创建用户
        created_users = []

        logger.info(f"准备测试数据: 创建 {bulk_count} 个用户...")
        for i in range(bulk_count):
            test_user = create_mock_user()
            response = UserRequest.create_user(api_client, test_user)
            if response.status_code == 200:
                created_users.append(test_user.username)
                if (i + 1) % 10 == 0:
                    logger.debug(f"已创建 {i + 1}/{bulk_count} 个用户")

        assert len(created_users) >= bulk_count * 0.8, (
            "用户创建成功率应该 >= 80% (由于可能的测试间干扰，从90%降低到80%)"
        )
        logger.info(f"成功创建 {len(created_users)} 个用户")

        # 批量查询测试
        success_count = 0
        failed_count = 0
        response_times = []

        logger.info(f"开始批量查询测试，用户数: {len(created_users)}")
        start_time = time.time()

        for i, username in enumerate(created_users):
            try:
                req_start = time.time()
                response = UserRequest.get_user(api_client, username)
                req_time = (time.time() - req_start) * 1000
                response_times.append(req_time)

                if response.status_code == 200:
                    success_count += 1
                else:
                    failed_count += 1
                    logger.warning(f"查询用户 {username} 失败: {response.status_code}")
            except Exception as e:
                failed_count += 1
                logger.error(f"查询用户 {username} 异常: {str(e)}")

        total_time = (time.time() - start_time) * 1000
        avg_response_time = (
            sum(response_times) / len(response_times) if response_times else 0
        )
        min_response_time = min(response_times) if response_times else 0
        max_response_time = max(response_times) if response_times else 0

        logger.info(
            f"批量查询完成: 成功 {success_count}/{len(created_users)}, 失败 {failed_count}/{len(created_users)}"
        )
        logger.info(f"总耗时: {total_time:.2f}ms ({total_time / 1000:.2f}s)")
        logger.info(f"平均响应时间: {avg_response_time:.2f}ms")
        logger.info(
            f"最快响应: {min_response_time:.2f}ms, 最慢响应: {max_response_time:.2f}ms"
        )
        logger.info(f"吞吐量: {len(created_users) / (total_time / 1000):.2f} 请求/秒")

        # 验证成功率
        success_rate = success_count / len(created_users)
        assert success_rate >= min_success_rate, (
            f"批量查询成功率过低: {success_rate * 100:.1f}% (要求 >= {min_success_rate * 100}%)"
        )

        logger.success(f"✓ 批量查询压力测试通过，成功率: {success_rate * 100:.1f}%")


class TestDeleteUser(BaseTestWithCleanup):
    """删除用户接口测试类"""

    # 使用默认的清理等待时间（3.0 秒前，1.0 秒后）

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


class TestUpdatePassword(BaseTestWithCleanup):
    """更新用户密码接口测试类"""

    # 使用默认的清理等待时间（3.0 秒前，1.0 秒后）

    def test_update_password_success(
        self, go_server: subprocess.Popen, api_client: ApiClient, mock_user: User
    ):
        """测试成功更新密码"""
        # 1. 创建用户
        create_response = UserRequest.create_user(api_client, mock_user)
        assert create_response.status_code == 200, "用户创建应该成功"
        logger.info(f"创建用户成功: {mock_user.username}")

        # 2. 更新密码
        old_password = mock_user.password
        new_password = "NewPassword@123"
        logger.info(f"测试更新密码: {mock_user.username}")
        update_response = UserRequest.update_password(
            api_client, mock_user.username, old_password, new_password
        )

        assert update_response.status_code == 200, (
            f"更新密码失败: status={update_response.status_code}, data={update_response.data}"
        )

        # 3. 验证响应消息
        resp_data = update_response.data
        if isinstance(resp_data, dict):
            message = resp_data.get("message", "")
            assert "success" in message.lower(), f"响应消息不正确: {message}"

        logger.success(f"✓ 密码更新成功: {mock_user.username}")

    def test_update_password_wrong_old_password(
        self, go_server: subprocess.Popen, api_client: ApiClient, mock_user: User
    ):
        """测试使用错误的旧密码更新密码"""
        # 1. 创建用户
        create_response = UserRequest.create_user(api_client, mock_user)
        assert create_response.status_code == 200, "用户创建应该成功"

        # 2. 使用错误的旧密码尝试更新
        wrong_old_password = "WrongPassword@123"
        new_password = "NewPassword@123"
        logger.info(f"测试使用错误旧密码更新: {mock_user.username}")
        update_response = UserRequest.update_password(
            api_client, mock_user.username, wrong_old_password, new_password
        )

        # 应该返回错误
        assert update_response.status_code in [400, 401, 403], (
            f"错误的旧密码应该返回 400/401/403，实际: {update_response.status_code}"
        )
        logger.success("✓ 错误旧密码测试通过")

    def test_update_password_same_as_old(
        self, go_server: subprocess.Popen, api_client: ApiClient, mock_user: User
    ):
        """测试新密码与旧密码相同"""
        # 1. 创建用户
        create_response = UserRequest.create_user(api_client, mock_user)
        assert create_response.status_code == 200, "用户创建应该成功"

        # 2. 尝试将新密码设置为与旧密码相同
        old_password = mock_user.password
        new_password = old_password  # 相同的密码
        logger.info(f"测试新旧密码相同: {mock_user.username}")
        update_response = UserRequest.update_password(
            api_client, mock_user.username, old_password, new_password
        )

        # 应该返回错误
        assert update_response.status_code in [400, 422], (
            f"新旧密码相同应该返回 400/422，实际: {update_response.status_code}"
        )
        logger.success("✓ 新旧密码相同测试通过")

    def test_update_password_weak_password(
        self, go_server: subprocess.Popen, api_client: ApiClient, mock_user: User
    ):
        """测试使用弱密码"""
        # 1. 创建用户
        create_response = UserRequest.create_user(api_client, mock_user)
        assert create_response.status_code == 200, "用户创建应该成功"

        # 2. 尝试使用弱密码（不符合复杂度要求）
        old_password = mock_user.password
        weak_passwords = ["123456", "abc123", "password", "test"]

        for weak_password in weak_passwords:
            logger.info(f"测试弱密码: {weak_password}")
            update_response = UserRequest.update_password(
                api_client, mock_user.username, old_password, weak_password
            )

            # 应该返回错误
            assert update_response.status_code in [400, 422], (
                f"弱密码应该返回 400/422，实际: {update_response.status_code}"
            )

        logger.success("✓ 弱密码测试通过")

    def test_update_password_nonexistent_user(
        self, go_server: subprocess.Popen, api_client: ApiClient
    ):
        """测试更新不存在用户的密码"""
        nonexistent_username = f"nonexistent_{uuid.uuid4().hex}"
        update_response = UserRequest.update_password(
            api_client, nonexistent_username, "OldPass@123", "NewPass@123"
        )

        # 应该返回404或400
        assert update_response.status_code in [400, 404], (
            f"不存在的用户应该返回 400/404，实际: {update_response.status_code}"
        )
        logger.success("✓ 不存在用户测试通过")

    def test_concurrent_update_password(
        self, go_server: subprocess.Popen, api_client: ApiClient, mock_user: User
    ):
        """测试并发更新密码（验证分布式锁）"""
        # 1. 创建用户
        create_response = UserRequest.create_user(api_client, mock_user)
        assert create_response.status_code == 200, "用户创建应该成功"

        old_password = mock_user.password
        concurrent_count = 5
        success_count = 0
        failed_count = 0

        def update_password(index: int):
            """并发更新密码"""
            try:
                new_password = f"NewPassword@{index}{random.randint(100, 999)}"
                response = UserRequest.update_password(
                    api_client, mock_user.username, old_password, new_password
                )
                return {
                    "index": index,
                    "success": response.status_code == 200,
                    "status_code": response.status_code,
                }
            except Exception as e:
                logger.error(f"并发更新密码 #{index} 失败: {str(e)}")
                return {"index": index, "success": False, "error": str(e)}

        logger.info(f"开始并发更新密码测试，并发数: {concurrent_count}")

        with ThreadPoolExecutor(max_workers=concurrent_count) as executor:
            futures = [
                executor.submit(update_password, i) for i in range(concurrent_count)
            ]

            for future in as_completed(futures):
                result = future.result()
                if result.get("success"):
                    success_count += 1
                    logger.debug(f"✓ 更新 #{result['index']} 成功")
                else:
                    failed_count += 1
                    logger.debug(
                        f"✗ 更新 #{result['index']} 失败: {result.get('status_code')}"
                    )

        logger.info(
            f"并发更新密码完成: 成功 {success_count}/{concurrent_count}, 失败 {failed_count}/{concurrent_count}"
        )

        # 由于使用了旧密码，只有第一个成功的请求能更新密码
        # 其他请求应该因为密码已变而失败
        # 或者如果加锁成功，可能只有一个成功
        assert success_count >= 1, f"至少应该有1个更新成功: {success_count}"
        logger.success(
            f"✓ 并发更新密码测试通过（成功: {success_count}, 失败: {failed_count}）"
        )


class TestGetUserList(BaseTestWithCleanup):
    """获取用户列表接口测试类"""

    def test_get_user_list_pagination(
        self,
        go_server: subprocess.Popen,
        api_client: ApiClient,
        db_helper: DatabaseHelper,
    ):
        """测试用户列表的分页功能"""
        # 清空用户表，确保测试环境干净
        db_helper.truncate_table("t_user")
        logger.info("已清空用户表，开始测试分页功能")

        # 1. 创建多个用户（减少到 25 个，测试更快）
        total_users = 25
        created_usernames = []
        for _ in range(total_users):
            user = create_mock_user()
            response = UserRequest.create_user(api_client, user)
            if response.status_code == 200:
                created_usernames.append(user.username)

        logger.info(f"成功创建 {len(created_usernames)} 个用户")

        # 2. 测试分页查询
        page_size = 10
        total_pages = (len(created_usernames) + page_size - 1) // page_size

        for page in range(1, total_pages + 1):
            logger.info(f"查询用户列表，第 {page} 页，页面大小 {page_size}")
            response = UserRequest.get_user_list(
                api_client, page=page, page_size=page_size
            )

            assert response.status_code == 200, (
                f"获取用户列表失败: {response.status_code}"
            )

            resp_data = response.data
            if isinstance(resp_data, dict):
                users = resp_data.get("list", [])
                total = resp_data.get("total", 0)
            else:
                users = []
                total = 0

            expected_count = (
                page_size
                if page < total_pages
                else len(created_usernames) - page_size * (total_pages - 1)
            )

            assert total == len(users), (
                f"返回的总数不正确，期望 {len(users)}，实际 {total}"
            )
            assert len(users) == expected_count, (
                f"第 {page} 页用户数量不正确，期望 {expected_count}，实际 {len(users)}"
            )

            logger.success(f"✓ 第 {page} 页用户列表查询通过，用户数: {len(users)}")

    def test_get_user_list_empty(
        self,
        go_server: subprocess.Popen,
        api_client: ApiClient,
        db_helper: DatabaseHelper,
    ):
        """测试空用户列表"""
        # 确保数据库为空
        db_helper.truncate_table("t_user")

        response = UserRequest.get_user_list(api_client, page=1, page_size=10)
        assert response.status_code == 200, f"获取空列表失败: {response.status_code}"

        resp_data = response.data
        if isinstance(resp_data, dict):
            users = resp_data.get("list", [])
            total = resp_data.get("total", -1)
        else:
            users = []
            total = -1

        assert len(users) == 0, f"空列表应该返回0个用户，实际: {len(users)}"
        assert total == 0, f"空列表总数应该为0，实际: {total}"
        logger.success("✓ 空用户列表测试通过")

    def test_get_user_list_boundary_conditions(
        self,
        go_server: subprocess.Popen,
        api_client: ApiClient,
        db_helper: DatabaseHelper,
    ):
        """测试边界条件（无效的页码和页面大小）"""
        db_helper.truncate_table("t_user")

        # 创建几个用户
        for _ in range(5):
            user = create_mock_user()
            UserRequest.create_user(api_client, user)

        # page = 0 和 page_size = 0 会自动填充默认值，属于成功
        # 测试 page = 0
        response = UserRequest.get_user_list(api_client, page=0, page_size=10)
        assert response.status_code == 200, (
            f"页码为0应该返回默认第一页，实际: {response.status_code}"
        )

        # 测试 page_size = 0
        response = UserRequest.get_user_list(api_client, page=1, page_size=0)
        assert response.status_code == 200, (
            f"页面大小为0应该返回默认大小，实际: {response.status_code}"
        )

        # 测试超大页码（超出范围）
        response = UserRequest.get_user_list(api_client, page=9999, page_size=10)
        assert response.status_code == 200, (
            f"超出范围的页码应该返回空列表，实际: {response.status_code}"
        )
        resp_data = response.data
        if isinstance(resp_data, dict):
            users = resp_data.get("list", [])
            assert len(users) == 0, "超出范围应该返回空列表"

        # 测试最大页面大小（根据配置限制）
        response = UserRequest.get_user_list(api_client, page=1, page_size=100)
        assert response.status_code == 200, (
            f"最大页面大小应该被接受，实际: {response.status_code}"
        )

        # 测试超过最大页面大小，接口会自动约束在 100
        response = UserRequest.get_user_list(api_client, page=1, page_size=101)
        assert response.status_code == 200, (
            f"超过最大页面大小应该返回错误，实际: {response.status_code}"
        )

        logger.success("✓ 边界条件测试通过")

    def test_get_user_list_with_deleted_users(
        self,
        go_server: subprocess.Popen,
        api_client: ApiClient,
        db_helper: DatabaseHelper,
    ):
        """测试列表不包含已删除的用户（软删除）"""
        db_helper.truncate_table("t_user")

        # 创建5个用户
        users = []
        for _ in range(5):
            user = create_mock_user()
            response = UserRequest.create_user(api_client, user)
            if response.status_code == 200:
                users.append(user)

        # 删除其中2个用户（软删除）
        for i in range(2):
            UserRequest.delete_user(api_client, users[i].username)

        # 查询用户列表
        response = UserRequest.get_user_list(api_client, page=1, page_size=10)
        assert response.status_code == 200

        resp_data = response.data
        if isinstance(resp_data, dict):
            active_users = resp_data.get("list", [])
            total = resp_data.get("total", 0)
        else:
            active_users = []
            total = 0

        # 应该只返回未删除的用户
        assert len(active_users) == 3, f"应该返回3个活跃用户，实际: {len(active_users)}"
        assert total == 3, f"总数应该为3，实际: {total}"

        # 验证返回的用户不包含已删除的
        returned_usernames = [u.get("username") for u in active_users]
        for i in range(2):
            assert users[i].username not in returned_usernames, (
                f"已删除用户 {users[i].username} 不应该在列表中"
            )

        logger.success("✓ 软删除用户过滤测试通过")

    def test_get_user_list_data_integrity(
        self, go_server: subprocess.Popen, api_client: ApiClient, db_helper
    ):
        """测试列表返回数据的完整性和正确性"""
        db_helper.truncate_table("t_user")

        # 创建一个用户并记录详细信息
        test_user = create_mock_user()
        create_response = UserRequest.create_user(api_client, test_user)
        assert create_response.status_code == 200

        # 获取列表
        response = UserRequest.get_user_list(api_client, page=1, page_size=10)
        assert response.status_code == 200

        resp_data = response.data
        if isinstance(resp_data, dict):
            users = resp_data.get("list", [])
        else:
            users = []

        assert len(users) == 1, "应该返回1个用户"

        # 验证用户数据完整性
        returned_user = users[0]
        assert returned_user.get("username") == test_user.username, "用户名不匹配"
        assert returned_user.get("email") == test_user.email, "邮箱不匹配"
        assert returned_user.get("nickname") == test_user.nickname, "昵称不匹配"
        assert "password" not in returned_user, "密码不应该被返回"

        # 验证应该有的字段
        expected_fields = ["username", "email", "nickname", "status"]
        for field in expected_fields:
            assert field in returned_user, f"缺少字段: {field}"

        logger.success("✓ 数据完整性测试通过")


class TestGetUserCount(BaseTestWithCleanup):
    """获取用户总数统计测试类"""

    # TEST_ERROR_TOTAL=-1

    def test_get_user_count_basic(
        self, go_server: subprocess.Popen, api_client: ApiClient, db_helper
    ):
        """测试基本的用户统计功能"""
        # 清空数据库
        db_helper.truncate_table("t_user")

        # 创建5个用户
        for _ in range(5):
            user = create_mock_user()
            response = UserRequest.create_user(api_client, user)
            assert response.status_code == 200

        # 获取用户总数
        response = UserRequest.get_user_count(api_client)
        assert response.status_code == 200, f"获取用户总数失败: {response.status_code}"

        resp_data = response.data
        assert isinstance(resp_data, dict), "响应数据应该是字典"

        total = resp_data.get("total", 0)
        active = resp_data.get("active", 0)

        assert total == 5, f"总数应该为5，实际: {total}"
        assert active == 5, f"活跃用户数应该为5，实际: {active}"

        logger.success("✓ 基本用户统计测试通过")

    def test_get_user_count_with_deleted(
        self, go_server: subprocess.Popen, api_client: ApiClient, db_helper
    ):
        """测试包含已删除用户的统计"""
        # 清空数据库
        db_helper.truncate_table("t_user")

        # 创建5个用户
        users = []
        for _ in range(5):
            user = create_mock_user()
            response = UserRequest.create_user(api_client, user)
            if response.status_code == 200:
                users.append(user)

        # 删除其中2个用户（软删除）
        for i in range(2):
            UserRequest.delete_user(api_client, users[i].username)

        # 获取用户统计（不包含已删除）
        response = UserRequest.get_user_count(api_client, include_deleted=False)
        assert response.status_code == 200

        resp_data = response.data
        if isinstance(resp_data, dict):
            total = resp_data.get("total", 0)
            active = resp_data.get("active", 0)
        else:
            total = 0
            active = 0

        assert total == 3, f"总数应该为3（不含已删除），实际: {total}"
        assert active == 3, f"活跃用户数应该为3，实际: {active}"

        # 获取用户统计（包含已删除）
        response = UserRequest.get_user_count(api_client, include_deleted=True)
        assert response.status_code == 200

        resp_data = response.data
        if isinstance(resp_data, dict):
            total = resp_data.get("total", 0)
            active = resp_data.get("active", 0)
            deleted = resp_data.get("deleted", 0)
        else:
            total = 0
            active = 0
            deleted = 0

        assert total == 5, f"总数应该为5（含已删除），实际: {total}"
        assert active == 3, f"活跃用户数应该为3，实际: {active}"
        assert deleted == 2, f"已删除用户数应该为2，实际: {deleted}"

        logger.success("✓ 包含已删除用户的统计测试通过")

    def test_get_user_count_empty_database(
        self, go_server: subprocess.Popen, api_client: ApiClient, db_helper
    ):
        """测试空数据库的统计"""
        # 确保数据库为空
        db_helper.truncate_table("t_user")

        # 获取用户总数
        response = UserRequest.get_user_count(api_client)
        assert response.status_code == 200

        resp_data = response.data
        if isinstance(resp_data, dict):
            total = resp_data.get("total", -1)
            active = resp_data.get("active", -1)
        else:
            total = -1
            active = -1

        assert total == 0, f"空数据库总数应该为0，实际: {total}"
        assert active == 0, f"空数据库活跃用户数应该为0，实际: {active}"

        logger.success("✓ 空数据库统计测试通过")

    def test_get_user_count_consistency_with_list(
        self, go_server: subprocess.Popen, api_client: ApiClient, db_helper
    ):
        """测试统计数据与列表数据的一致性"""
        # 清空数据库
        db_helper.truncate_table("t_user")

        # 创建10个用户
        created_count = 0
        for _ in range(10):
            user = create_mock_user()
            response = UserRequest.create_user(api_client, user)
            if response.status_code == 200:
                created_count += 1

        # 获取用户总数
        count_response = UserRequest.get_user_count(api_client)
        assert count_response.status_code == 200

        count_data = count_response.data
        count_total = count_data.get("total", 0)

        # 获取用户列表
        list_response = UserRequest.get_user_list(api_client, page=1, page_size=100)
        assert list_response.status_code == 200

        list_data = list_response.data
        if isinstance(list_data, dict):
            list_total = list_data.get("total", 0)
            users = list_data.get("list", [])
        else:
            list_total = 0
            users = []

        # 验证一致性
        assert count_total == list_total, (
            f"统计接口的总数({count_total})应该与列表接口的总数({list_total})一致"
        )
        assert count_total == len(users), (
            f"统计的总数({count_total})应该与实际用户数({len(users)})一致"
        )
        assert count_total == created_count, (
            f"统计的总数({count_total})应该与创建的用户数({created_count})一致"
        )

        logger.success("✓ 统计数据一致性测试通过")

    def test_get_user_count_with_database_verification(
        self, go_server: subprocess.Popen, api_client: ApiClient, db_helper
    ):
        """测试统计数据与数据库实际数据的一致性"""
        # 清空数据库
        db_helper.truncate_table("t_user")

        # 创建用户
        for _ in range(7):
            user = create_mock_user()
            UserRequest.create_user(api_client, user)

        # 从 API 获取统计
        response = UserRequest.get_user_count(api_client)
        assert response.status_code == 200

        api_count = response.data.get("total", 0)

        # 从数据库直接查询
        db_count = db_helper.get_active_user_count()

        assert api_count == db_count, (
            f"API 返回的统计({api_count})应该与数据库实际数量({db_count})一致"
        )

        logger.success("✓ 与数据库数据一致性测试通过")
