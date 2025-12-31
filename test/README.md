# 测试文档

## 环境准备

1. 安装 Python 依赖:

```bash
uv sync
```

1. 确保 Docker 环境运行（MySQL、Redis、Kafka）:

```bash
docker-compose up -d
```

## 运行测试

### 使用 pytest 自动化测试

pytest 会自动启动和停止 Go 服务器:

```bash
# 运行所有测试
pytest

# 运行指定测试文件
pytest test/user/test_register_user.py

# 运行指定测试用例
pytest test/user/test_register_user.py::TestUserAPI::test_01_create_user

# 显示详细输出
pytest -v

# 显示打印信息
pytest -s

# 运行测试并显示覆盖率
pytest --cov
```

### 手动运行测试脚本

如果需要手动控制服务器:

```bash
# 1. 启动 Go 服务器
go run app/main.go

# 2. 在另一个终端运行测试脚本
python test/user/test_register_user.py
```

## 测试说明

### test_register_user.py

用户注册和查询的集成测试，包含以下测试用例:

1. **test_01_create_user**: 测试创建用户
2. **test_02_get_user**: 测试获取用户信息
3. **test_03_verify_user_data**: 测试验证用户数据完整性
4. **test_04_cache_query**: 测试缓存查询功能

### Fixtures

- `go_server` (session 级别): 自动启动和停止 Go 服务器
- `mock_user`: 为每个测试生成一个随机测试用户

## 测试特点

- ✅ 自动启动/停止服务器
- ✅ 每个测试用例独立的测试数据
- ✅ 完整的数据验证
- ✅ 安全性检查（密码不返回）
- ✅ 缓存功能测试

## 注意事项

1. 测试会在 `localhost:8888` 端口启动服务器
2. 需要确保 MySQL、Redis 等基础设施已就绪
3. 每个测试会创建新的用户数据
4. 测试结束后服务器会自动清理
