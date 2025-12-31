# pytest 自动化测试改造说明

## 改动概览

已将 `test/user/test_register_user.py` 从手动测试脚本改造为 pytest 自动化测试框架。

## 主要改动

### 1. 添加 pytest 依赖

**文件**: [pyproject.toml](pyproject.toml)

```toml
dependencies = [
    "loguru>=0.7.3",
    "pydantic[email]>=2.12.5",
    "requests>=2.32.5",
    "pytest>=8.3.4",  # 新增
]
```

### 2. 测试文件改造

**文件**: [test/user/test_register_user.py](test/user/test_register_user.py)

#### 新增 Fixtures

```python
@pytest.fixture(scope="session")
def go_server():
    """启动 Go 服务器并在测试结束后停止"""
    # 自动执行: go run app/main.go
    # 等待服务就绪
    # 测试结束后自动清理
    ...

@pytest.fixture
def mock_user():
    """为每个测试生成独立的测试用户"""
    ...
```

#### 测试类结构

```python
class TestUserAPI:
    """用户 API 测试类"""
    
    def test_01_create_user(self, go_server, mock_user):
        """测试创建用户"""
        ...
    
    def test_02_get_user(self, go_server, mock_user):
        """测试获取用户信息"""
        ...
    
    def test_03_verify_user_data(self, go_server, mock_user):
        """测试验证用户数据完整性"""
        ...
    
    def test_04_cache_query(self, go_server, mock_user):
        """测试缓存查询"""
        ...
```

### 3. 配置文件

**文件**: [pytest.ini](pytest.ini)

配置 pytest 行为:

- 测试发现规则
- 日志格式
- 命令行选项

### 4. 新增辅助工具

#### 服务检查工具

**文件**: [test/check_service.py](test/check_service.py)

快速检查服务是否运行:

```bash
python test/check_service.py
```

#### Makefile

**文件**: [Makefile](Makefile)

简化常用操作:

```bash
make help          # 查看所有命令
make install       # 安装依赖
make test          # 运行测试
make docker-up     # 启动 Docker
```

### 5. 文档

- [test/README.md](test/README.md) - 完整测试文档
- [QUICKSTART.md](QUICKSTART.md) - 快速开始指南

## 使用方法

### 自动化测试 (推荐)

pytest 会自动管理服务器生命周期:

```bash
# 1. 确保 Docker 环境运行
docker-compose up -d
# 或
make docker-up

# 2. 运行测试 (pytest 自动启动/停止 Go 服务器)
pytest
# 或
make test
```

### 手动测试

如果需要手动控制服务器:

```bash
# Terminal 1: 启动服务器
go run app/main.go

# Terminal 2: 运行测试脚本
python test/user/test_register_user.py
```

## 测试特性

✅ **自动化**: pytest 自动启动和停止 Go 服务器  
✅ **隔离性**: 每个测试使用独立的测试数据  
✅ **完整性**: 覆盖创建、查询、验证、缓存等功能  
✅ **安全性**: 验证密码不被返回  
✅ **可读性**: 清晰的测试用例命名和文档  

## 常用命令

```bash
# 运行所有测试
pytest

# 显示详细输出
pytest -v

# 显示打印信息
pytest -s

# 运行指定测试
pytest test/user/test_register_user.py::TestUserAPI::test_01_create_user

# 停止在第一个失败
pytest -x

# 只运行失败的测试
pytest --lf

# 使用 Makefile
make test          # 运行测试
make test-v        # 详细输出
make dev-setup     # 完整环境设置
```

## 测试输出示例

```
test/user/test_register_user.py::TestUserAPI::test_01_create_user PASSED [25%]
test/user/test_register_user.py::TestUserAPI::test_02_get_user PASSED [50%]
test/user/test_register_user.py::TestUserAPI::test_03_verify_user_data PASSED [75%]
test/user/test_register_user.py::TestUserAPI::test_04_cache_query PASSED [100%]

============================== 4 passed in 5.23s ===============================
```

## 故障排查

### 服务启动失败

如果测试报告 "服务器启动失败":

1. 检查 8888 端口是否被占用
2. 确保 MySQL/Redis/Kafka 正常运行
3. 查看 Go 代码是否有编译错误

```bash
# 检查端口
lsof -i :8888

# 检查 Docker 服务
docker-compose ps

# 手动测试服务启动
go run app/main.go
```

### 测试数据冲突

测试会生成随机用户名，不应该出现冲突。如果出现:

1. 检查数据库是否正常
2. 考虑在测试前清理测试数据

## 下一步

可以考虑的改进:

- [ ] 添加更多测试用例 (边界情况、错误处理)
- [ ] 集成 pytest-cov 生成覆盖率报告
- [ ] 使用 pytest-xdist 并行运行测试
- [ ] 添加 CI/CD 集成
- [ ] 添加性能测试
