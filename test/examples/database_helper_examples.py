"""
数据库辅助工具快速使用示例
展示常见使用场景
"""

import pytest

from test.helpers import DatabaseHelper


# ============================================================================
# 场景 1: 清空数据库开始测试（最常用）
# ============================================================================
def test_with_clean_database(db_helper, api_client):
    """测试前清空数据库"""
    # 清空用户表
    db_helper.clear_user_table()

    # 现在可以确保数据库是干净的
    assert db_helper.get_active_user_count() == 0

    # 执行测试...


# ============================================================================
# 场景 2: 验证数据库状态
# ============================================================================
def test_verify_database_state(db_helper, api_client):
    """验证操作后的数据库状态"""
    db_helper.clear_user_table()

    # 通过 API 创建用户
    # ... create user via API ...

    # 验证数据库中确实有记录
    count = db_helper.get_active_user_count()
    assert count == 1

    # 查询具体数据
    users = db_helper.query("SELECT * FROM t_user WHERE deleted_at IS NULL")
    assert len(users) == 1


# ============================================================================
# 场景 3: 直接准备测试数据（跳过 API）
# ============================================================================
def test_with_preset_data(db_helper, api_client):
    """直接在数据库中准备测试数据"""
    db_helper.clear_user_table()

    # 直接插入测试数据
    db_helper.execute(
        """
        INSERT INTO t_user (username, password, email, status)
        VALUES (%s, %s, %s, 1)
        """,
        ("test_user", "hashed_password", "test@example.com"),
    )

    # 验证
    assert db_helper.get_active_user_count() == 1


# ============================================================================
# 场景 4: 批量准备数据
# ============================================================================
def test_with_bulk_data(db_helper):
    """批量准备测试数据"""
    db_helper.clear_user_table()

    # 批量插入
    users = [
        ("user1", "pass1", "user1@example.com"),
        ("user2", "pass2", "user2@example.com"),
        ("user3", "pass3", "user3@example.com"),
    ]

    db_helper.execute_many(
        """
        INSERT INTO t_user (username, password, email, status)
        VALUES (%s, %s, %s, 1)
        """,
        users,
    )

    assert db_helper.get_active_user_count() == 3


# ============================================================================
# 场景 5: 测试软删除
# ============================================================================
def test_soft_delete(db_helper, api_client):
    """测试软删除功能"""
    db_helper.clear_user_table()

    # 创建用户
    # ... create user ...

    # 删除用户（通过 API，软删除）
    # ... delete user via API ...

    # 验证：总数不变，活跃数为 0
    assert db_helper.get_user_count() == 1  # 包含软删除
    assert db_helper.get_active_user_count() == 0  # 不包含软删除

    # 查询软删除的记录
    deleted = db_helper.query_one(
        "SELECT * FROM t_user WHERE username = %s", ("test_user",)
    )
    assert deleted["deleted_at"] is not None


# ============================================================================
# 场景 6: 复杂查询和统计
# ============================================================================
def test_statistics(db_helper):
    """测试复杂查询和统计"""
    db_helper.clear_user_table()

    # 准备不同状态的数据
    for i in range(10):
        status = 1 if i < 7 else 0
        db_helper.execute(
            """
            INSERT INTO t_user (username, password, email, status)
            VALUES (%s, %s, %s, %s)
            """,
            (f"user{i}", "pass", f"user{i}@example.com", status),
        )

    # 复杂统计查询
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

    assert stats["total"] == 10
    assert stats["active"] == 7
    assert stats["inactive"] == 3


# ============================================================================
# 场景 7: 测试缓存失效
# ============================================================================
def test_cache_invalidation(db_helper, api_client):
    """测试缓存失效逻辑"""
    db_helper.clear_user_table()

    # 通过 API 创建用户（会写入缓存）
    # ... create user via API ...

    # 通过 API 查询（命中缓存）
    # ... get user via API ...

    # 直接从数据库删除（绕过业务层和缓存）
    db_helper.execute(
        "UPDATE t_user SET deleted_at = NOW() WHERE username = %s", ("test_user",)
    )

    # 再次通过 API 查询
    # 如果缓存正确失效，应该查不到
    # 如果缓存未失效，会查到（这是 bug）
    # ... verify via API ...


# ============================================================================
# 场景 8: 自定义 Fixture（自动清理）
# ============================================================================
@pytest.fixture(autouse=True)
def auto_clean_db(db_helper):
    """每个测试自动清空数据库"""
    db_helper.clear_user_table()
    yield
    # 可以在这里做测试后的清理


# ============================================================================
# 场景 9: 清空所有表
# ============================================================================
def test_clean_all_tables(db_helper):
    """清空所有表"""
    # 清空所有表
    db_helper.truncate_all_tables()

    # 或排除某些系统表
    db_helper.truncate_all_tables(exclude_tables=["t_config"])


# ============================================================================
# 场景 10: 使用上下文管理器
# ============================================================================
def test_with_context_manager():
    """使用上下文管理器（不依赖 fixture）"""
    with DatabaseHelper(
        host="mysql",
        port=3306,
        user="root",
        password="rootpassword",
        database="hello_gozero_db",
    ) as db:
        db.clear_user_table()

        # 执行测试...

        # 自动关闭连接


# ============================================================================
# 常用方法速查表
# ============================================================================
"""
清理方法：
- db_helper.clear_user_table()                     # 清空用户表
- db_helper.truncate_table("table_name")           # 清空指定表
- db_helper.delete_all("table_name", "where ...")  # 删除数据（可带条件）
- db_helper.truncate_all_tables()                  # 清空所有表

查询方法：
- db_helper.query(sql, params)                     # 查询多条
- db_helper.query_one(sql, params)                 # 查询单条
- db_helper.get_user_count()                       # 用户总数（含软删除）
- db_helper.get_active_user_count()                # 活跃用户数

执行方法：
- db_helper.execute(sql, params)                   # 执行单条 SQL
- db_helper.execute_many(sql, params_list)         # 批量执行
"""
