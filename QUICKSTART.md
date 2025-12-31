# 快速开始指南

## 自动化测试 (推荐)

使用 pytest 自动化测试，会自动管理服务器启停:

```bash
# 确保 Docker 服务运行
docker-compose up -d

# 运行所有测试 (pytest 会自动启动和停止 go 服务器)
pytest

# 或者运行指定测试
pytest test/user/test_register_user.py -v
```

## 手动测试

如果你想手动控制服务器:

```bash
# Terminal 1: 启动服务器
go run app/main.go

# Terminal 2: 运行测试脚本
python test/user/test_register_user.py
```

## 检查服务状态

```bash
# 快速检查服务是否正常
python test/check_service.py
```

## 测试输出示例

```
test/user/test_register_user.py::TestUserAPI::test_01_create_user PASSED
test/user/test_register_user.py::TestUserAPI::test_02_get_user PASSED
test/user/test_register_user.py::TestUserAPI::test_03_verify_user_data PASSED
test/user/test_register_user.py::TestUserAPI::test_04_cache_query PASSED
```

## 更多选项

```bash
# 显示详细日志
pytest -s

# 只运行失败的测试
pytest --lf

# 停止在第一个失败
pytest -x

# 并行运行 (需要安装 pytest-xdist)
pytest -n auto
```
