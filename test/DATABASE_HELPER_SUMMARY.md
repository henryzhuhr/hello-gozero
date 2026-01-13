# 数据库辅助工具总结

## 🎯 已完成的工作

### 1. 创建了 `DatabaseHelper` 类

**文件：** `test/helpers/database.py`

提供的功能：

- ✅ 数据库连接管理（自动连接、关闭）
- ✅ 数据清理（清空表、删除数据）
- ✅ 数据查询（单条、多条、统计）
- ✅ SQL 执行（单条、批量）
- ✅ 上下文管理器支持（`with` 语句）

### 2. 添加了 pytest fixture

**文件：** `test/conftest.py`

- ✅ 创建了 `db_helper` fixture
- ✅ 自动管理连接生命周期
- ✅ 从配置文件读取数据库连接信息

### 3. 更新了依赖

**文件：** `pyproject.toml`

- ✅ 添加了 `pymysql>=1.1.0` 依赖
- ✅ 已安装依赖

### 4. 更新了测试用例

**文件：** `test/user/test_user.py`

- ✅ 修复了分页测试，使用 `db_helper.clear_user_table()` 清空数据
- ✅ 测试已通过

### 5. 创建了文档

- ✅ `test/DATABASE_HELPER_USAGE.md` - 完整使用指南
- ✅ `test/examples/database_helper_examples.py` - 快速示例代码

### 6. 修复了测试配置

**文件：** `test/test_config.yaml`

- ✅ 更正了数据库名称从 `hello_gozero` 到 `hello_gozero_db`

---

## 📚 使用方法

### 最简单的用法（推荐）

```python
def test_something(db_helper, api_client):
    """测试某个功能"""
    # 清空数据库
    db_helper.clear_user_table()
    
    # 执行测试
    # ...
    
    # 验证数据库状态
    count = db_helper.get_active_user_count()
    assert count == expected
```

### 快速参考

```python
# 清理
db_helper.clear_user_table()                      # 清空用户表
db_helper.truncate_table("t_order")               # 清空其他表
db_helper.truncate_all_tables()                   # 清空所有表

# 查询
users = db_helper.query("SELECT * FROM t_user")   # 查询多条
user = db_helper.query_one("SELECT * FROM ...")   # 查询单条
count = db_helper.get_active_user_count()         # 活跃用户数

# 执行
db_helper.execute("INSERT INTO ...", params)      # 执行 SQL
db_helper.execute_many("INSERT ...", params_list) # 批量执行
```

---

## 🔧 配置位置

数据库连接配置在 `test/test_config.yaml`：

```yaml
database:
  host: "mysql"
  port: 3306
  user: "root"
  password: "rootpassword"
  database: "hello_gozero_db"
  charset: "utf8mb4"
```

---

## ✅ 测试验证

运行测试验证功能：

```bash
# 运行单个测试
uv run pytest test/user/test_user.py::TestGetUserList::test_get_user_list_pagination -v

# 运行所有用户测试
uv run pytest test/user/test_user.py -v
```

**测试结果：** ✅ PASSED

---

## 📖 详细文档

1. **完整使用指南：** `test/DATABASE_HELPER_USAGE.md`
   - API 参考
   - 完整示例
   - 常见问题
   - 注意事项

2. **快速示例：** `test/examples/database_helper_examples.py`
   - 10+ 实用场景
   - 即用即查
   - 带注释说明

---

## 🎯 典型使用场景

### 场景 1：清空数据库开始测试

```python
def test_user_list(db_helper, api_client):
    db_helper.clear_user_table()
    # 确保测试环境干净
```

### 场景 2：验证数据库状态

```python
def test_create_user(db_helper, api_client):
    db_helper.clear_user_table()
    # 创建用户
    assert db_helper.get_active_user_count() == 1
```

### 场景 3：直接准备测试数据

```python
def test_with_data(db_helper):
    db_helper.clear_user_table()
    db_helper.execute(
        "INSERT INTO t_user (...) VALUES (...)",
        params
    )
```

### 场景 4：测试软删除

```python
def test_soft_delete(db_helper):
    # 删除后验证
    assert db_helper.get_user_count() == 1        # 含软删除
    assert db_helper.get_active_user_count() == 0 # 不含软删除
```

---

## ⚠️ 注意事项

1. **仅用于测试环境** - 不要在生产环境使用
2. **自动提交** - 每个操作立即提交，无需手动 commit
3. **外键处理** - `truncate_table` 自动处理外键约束
4. **参数化查询** - 始终使用参数化查询防止 SQL 注入

---

## 🚀 下一步

现在你可以：

1. ✅ 在任何测试中注入 `db_helper` 参数
2. ✅ 使用 `db_helper.clear_user_table()` 清空数据
3. ✅ 使用 `db_helper.query()` 查询数据库
4. ✅ 使用 `db_helper.execute()` 执行 SQL
5. ✅ 查看文档了解更多功能

---

## 📦 文件清单

```
test/
├── helpers/
│   ├── __init__.py           # 导出 DatabaseHelper
│   └── database.py           # DatabaseHelper 实现 ⭐
├── examples/
│   └── database_helper_examples.py  # 快速示例 ⭐
├── conftest.py               # pytest fixtures（含 db_helper）
├── test_config.yaml          # 数据库配置
└── DATABASE_HELPER_USAGE.md  # 完整文档 ⭐
```

---

## 🎉 总结

通过 `DatabaseHelper`，你现在可以：

- 轻松清空数据库，确保测试环境干净
- 直接查询数据库验证测试结果
- 准备复杂的测试数据
- 测试缓存失效等高级场景

**使用方法超级简单：**

```python
def test_something(db_helper):
    db_helper.clear_user_table()  # 就这么简单！
```

Happy Testing! 🚀
