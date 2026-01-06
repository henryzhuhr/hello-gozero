#!/usr/bin/env python3
"""测试密码更新功能"""

import sys

sys.path.insert(0, "/root/hello-gozero")

from test.helpers import ApiClient
from test.user.helpers import create_mock_user

# 创建客户端
client = ApiClient(base_url="http://localhost:8888")

# 创建用户
user = create_mock_user()
print(f"创建用户: {user.username}, 密码: {user.password}")

create_resp = client.post("/api/v1/users/register", data=user.model_dump())
print(f"创建响应: {create_resp.status_code}")
if create_resp.status_code != 200:
    print(f"创建失败: {create_resp.data}")
    sys.exit(1)

# 尝试更新密码
old_password = user.password
new_password = "NewPassword@123"

print(f"\n尝试更新密码:")
print(f"  旧密码: {old_password}")
print(f"  新密码: {new_password}")

update_resp = client.put(
    f"/api/v1/users/{user.username}/password",
    data={
        "old_password": old_password,
        "new_password": new_password,
    },
)

print(f"\n更新响应: {update_resp.status_code}")
print(f"响应数据: {update_resp.data}")

# 关闭客户端
client.close()
