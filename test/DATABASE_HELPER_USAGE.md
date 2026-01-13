# 数据库辅助工具使用指南

## 简介

`DatabaseHelper` 是一个为测试环境设计的数据库辅助类，提供了便捷的数据库连接、查询和清理功能。

## 基本用法

### 1. 作为 Fixture 使用（推荐）

在测试函数中直接注入 `db_helper` fixture：

```python
def test_something(db_helper):
    """使用 db_helper fixture"""
    # 清空用户表
    db_helper.clear_user_table()
    
    # 执行测试...
    # 连接会在测试结束后自动关闭
```

### 2. 直接实例化

```python
from test.helpers import DatabaseHelper

# 创建实例
db = DatabaseHelper(
    host="mysql",
    port=3306,
    user="root",
    password="rootpassword",
    database="hello_gozero",
)

# 使用完毕后记得关闭
db.close()
```

### 3. 使用上下文管理器

```python
from test.helpers import DatabaseHelper

with DatabaseHelper(
    host="mysql",
    port=3306,
    user="root",
    password="rootpassword",
    database="hello_gozero",
) as db:
    # 执行数据库操作
    db.clear_user_table()
    # 自动关闭连接
```

## API 参考

### 数据清理方法

#### `clear_user_table()`

清空用户表（专用方法）

```python
def test_user_list(db_helper):
    # 清空用户表，确保测试环境干净
    db_helper.clear_user_table()
```

#### `truncate_table(table_name: str)`

清空指定表的数据（保留表结构）

```python
db_helper.truncate_table("t_user")
db_helper.truncate_table("t_order")
```

#### `delete_all(table_name: str, where: Optional[str] = None)`

删除表中的所有数据（支持 WHERE 条件）

```python
# 删除所有记录
db_helper.delete_all("t_user")

# 删除符合条件的记录
db_helper.delete_all("t_user", "status = 0")
db_helper.delete_all("t_user", "deleted_at IS NULL")
```

#### `truncate_all_tables(exclude_tables: Optional[list] = None)`

清空数据库中所有表的数据

```python
# 清空所有表
db_helper.truncate_all_tables()

# 保留某些表的数据
db_helper.truncate_all_tables(exclude_tables=["t_config", "t_system"])
```

### 查询方法

#### `query(sql: str, params=None)`

查询多条数据，返回字典列表

```python
# 查询所有用户
users = db_helper.query("SELECT * FROM t_user")

# 带参数查询
users = db_helper.query(
    "SELECT * FROM t_user WHERE status = %s",
    (1,)
)

for user in users:
    print(user['username'], user['email'])
```

#### `query_one(sql: str, params=None)`

查询单条数据，返回字典

```python
# 查询单个用户
user = db_helper.query_one(
    "SELECT * FROM t_user WHERE username = %s",
    ("alice",)
)

if user:
    print(user['email'])
```

#### `get_user_count()`

获取用户表的记录数（包含软删除）

```python
total = db_helper.get_user_count()
print(f"用户总数: {total}")
```

#### `get_active_user_count()`

获取用户表的活跃记录数（不包含软删除）

```python
active = db_helper.get_active_user_count()
print(f"活跃用户数: {active}")
```

### 执行方法

#### `execute(sql: str, params=None)`

执行单条 SQL 语句，返回影响的行数

```python
# 插入数据
affected = db_helper.execute(
    "INSERT INTO t_user (username, password) VALUES (%s, %s)",
    ("alice", "hashed_password")
)

# 更新数据
affected = db_helper.execute(
    "UPDATE t_user SET status = %s WHERE username = %s",
    (1, "alice")
)
```

#### `execute_many(sql: str, params_list)`

批量执行 SQL 语句，返回影响的行数

```python
# 批量插入
users = [
    ("alice", "pass1"),
    ("bob", "pass2"),
    ("charlie", "pass3"),
]
affected = db_helper.execute_many(
    "INSERT INTO t_user (username, password) VALUES (%s, %s)",
    users
)
```

## 完整测试示例

### 示例 1：测试用户列表分页

```python
class TestGetUserList(BaseTestWithCleanup):
    """获取用户列表接口测试类"""

    def test_get_user_list_pagination(
        self, go_server, api_client, db_helper
    ):
        """测试用户列表的分页功能"""
        # 清空用户表，确保测试环境干净
        db_helper.clear_user_table()
        
        # 创建测试数据
        total_users = 25
        for i in range(total_users):
            user = create_mock_user()
            UserRequest.create_user(api_client, user)
        
        # 验证数据已创建
        assert db_helper.get_active_user_count() == total_users
        
        # 测试分页查询
        response = UserRequest.get_user_list(api_client, page=1, page_size=10)
        assert response.status_code == 200
```

### 示例 2：测试数据库状态验证

```python
def test_user_deletion_cleanup(db_helper, api_client):
    """测试用户删除后的数据库状态"""
    # 清空并创建测试用户
    db_helper.clear_user_table()
    
    user = create_mock_user()
    UserRequest.create_user(api_client, user)
    
    # 验证用户存在
    assert db_helper.get_active_user_count() == 1
    
    # 删除用户（软删除）
    UserRequest.delete_user(api_client, user.username)
    
    # 验证软删除：总数不变，活跃数为0
    assert db_helper.get_user_count() == 1
    assert db_helper.get_active_user_count() == 0
    
    # 查询数据库验证 deleted_at 字段
    deleted_user = db_helper.query_one(
        "SELECT * FROM t_user WHERE username = %s",
        (user.username,)
    )
    assert deleted_user['deleted_at'] is not None
```

### 示例 3：批量数据准备

```python
def test_with_preset_data(db_helper, api_client):
    """测试需要预设数据的场景"""
    # 清空表
    db_helper.clear_user_table()
    
    # 直接在数据库中插入测试数据（跳过 API）
    test_users = [
        ("alice", "hashed_pass", "alice@example.com"),
        ("bob", "hashed_pass", "bob@example.com"),
        ("charlie", "hashed_pass", "charlie@example.com"),
    ]
    
    for username, password, email in test_users:
        db_helper.execute(
            """
            INSERT INTO t_user (username, password, email, status)
            VALUES (%s, %s, %s, 1)
            """,
            (username, password, email)
        )
    
    # 验证数据已准备好
    assert db_helper.get_active_user_count() == 3
    
    # 执行测试...
```

### 示例 4：复杂查询验证

```python
def test_user_statistics(db_helper, api_client):
    """测试用户统计功能"""
    db_helper.clear_user_table()
    
    # 创建不同状态的用户
    for i in range(10):
        user = create_mock_user()
        UserRequest.create_user(api_client, user)
        
        # 设置部分用户为禁用状态
        if i < 3:
            db_helper.execute(
                "UPDATE t_user SET status = 0 WHERE username = %s",
                (user.username,)
            )
    
    # 查询统计
    stats = db_helper.query_one(
        """
        SELECT 
            COUNT(*) as total,
            SUM(CASE WHEN status = 1 THEN 1 ELSE 0 END) as active,
            SUM(CASE WHEN status = 0 THEN 1 ELSE 0 END) as inactive
        FROM t_user
        WHERE deleted_at IS NULL
        """
    )
    
    assert stats['total'] == 10
    assert stats['active'] == 7
    assert stats['inactive'] == 3
```

## 注意事项

### 1. 安全性

- ⚠️ **仅用于测试环境**：不要在生产环境使用此工具
- 始终使用参数化查询防止 SQL 注入

### 2. 事务

- 连接设置了 `autocommit=True`，每个操作立即提交
- 如需事务控制，需要手动管理

### 3. 外键约束

- `truncate_table` 和 `truncate_all_tables` 会自动处理外键约束
- 临时禁用外键检查，清空后恢复

### 4. 软删除

- `truncate_table` 会物理删除所有记录（包括软删除的）
- `delete_all` 可以通过 WHERE 条件只删除活跃记录
- 使用 `get_active_user_count()` 获取不包含软删除的记录数

## 配置

数据库配置在 `test/test_config.yaml` 中：

```yaml
database:
  host: "mysql"
  port: 3306
  user: "root"
  password: "rootpassword"
  database: "hello_gozero"
  charset: "utf8mb4"
```

## 常见问题

### Q: 如何在每个测试前自动清空数据库？

A: 可以创建自定义 fixture：

```python
@pytest.fixture(autouse=True)
def clean_database(db_helper):
    """每个测试前自动清空数据库"""
    db_helper.clear_user_table()
    yield
```

### Q: 如何验证缓存是否正确清除？

A: 结合数据库查询和 API 调用：

```python
def test_cache_invalidation(db_helper, api_client):
    # 创建用户
    user = create_mock_user()
    UserRequest.create_user(api_client, user)
    
    # API 查询（填充缓存）
    response = UserRequest.get_user(api_client, user.username)
    assert response.status_code == 200
    
    # 直接从数据库删除（绕过业务层）
    db_helper.execute(
        "UPDATE t_user SET deleted_at = NOW() WHERE username = %s",
        (user.username,)
    )
    
    # API 查询应该失败（如果缓存未正确处理）
    response = UserRequest.get_user(api_client, user.username)
    # 根据你的业务逻辑验证
```

## 更多资源

- 查看 `test/helpers/database.py` 了解完整实现
- 查看 `test/user/test_user.py` 了解实际使用案例
