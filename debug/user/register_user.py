from loguru import logger

from test.helpers.client import ApiClient
from test.user.test_user import UserRequest, create_mock_user


def main():
    user = create_mock_user()
    logger.info(f"创建用户: user{user.model_dump()}")

    api_client = ApiClient(base_url="http://localhost:8888")
    resp = UserRequest.create_user(api_client, user)
    print(resp.model_dump())

    resp = UserRequest.create_user(api_client, user)
    print(resp.model_dump())

    return

    # for user in random_generate_mock_user(1):
    #     # 1. 创建用户
    #     logger.info(f"创建用户: username={user.username}, email={user.email}")
    #     create_response = create_user_request(user)

    #     if not create_response:
    #         logger.error("创建用户失败，跳过后续验证")
    #         continue

    #     # 2. 验证创建响应
    #     if create_response["status_code"] == 200:
    #         logger.success(f"✓ 用户创建成功: {user.username}")
    #     else:
    #         logger.error(
    #             f"✗ 用户创建失败: status={create_response['status_code']}, data={create_response.get('data')}"
    #         )
    #         continue

    #     # 3. 获取用户信息
    #     logger.info(f"获取用户信息: {user.username}")
    #     get_response = get_user(user.username)

    #     if not get_response:
    #         logger.error("获取用户失败")
    #         continue

    #     # 4. 验证用户数据
    #     if get_response["status_code"] == 200:
    #         user_data = get_response.get("response", {}).get("data", {})

    #         # 验证关键字段
    #         checks = []
    #         checks.append(("用户名", user_data.get("username") == user.username))
    #         checks.append(("邮箱", user_data.get("email") == user.email))
    #         checks.append(("昵称", user_data.get("nickname") == user.nickname))
    #         checks.append(
    #             (
    #                 "手机区号",
    #                 user_data.get("phone_country_code") == user.phone_country_code,
    #             )
    #         )
    #         checks.append(
    #             ("手机号", user_data.get("phone_number") == user.phone_number)
    #         )
    #         checks.append(("用户ID存在", bool(user_data.get("id"))))
    #         checks.append(
    #             ("密码未返回", "password" not in user_data)
    #         )  # 安全检查：密码不应该被返回

    #         logger.info("=" * 60)
    #         logger.info("数据验证结果:")
    #         all_passed = True
    #         for check_name, passed in checks:
    #             status = "✓" if passed else "✗"
    #             logger.info(f"  {status} {check_name}: {'通过' if passed else '失败'}")
    #             if not passed:
    #                 all_passed = False
    #                 if check_name != "密码未返回":
    #                     expected = getattr(
    #                         user, check_name.lower().replace(" ", "_"), "N/A"
    #                     )
    #                     actual = user_data.get(check_name.lower().replace(" ", "_"))
    #                     logger.warning(f"    期望: {expected}, 实际: {actual}")

    #         if all_passed:
    #             logger.success("✓ 所有验证通过！")
    #         else:
    #             logger.error("✗ 部分验证失败")
    #         logger.info("=" * 60)
    #     else:
    #         logger.error(f"✗ 获取用户失败: status={get_response['status_code']}")

    #     # 5. 再次查询，测试缓存
    #     logger.info("第二次查询（测试缓存）...")
    #     get_user(user.username)


if __name__ == "__main__":
    main()
