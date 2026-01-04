# 测试文档

## 目录结构

```
test/
├── __init__.py
├── conftest.py              # pytest fixtures（全局共享）
├── helpers.py               # 通用测试工具（跨模块共享）⭐
├── models.py                # 测试数据模型（统一数据结构）
├── api_client.py            # HTTP 客户端封装（统一接口调用）⭐
├── README.md                # 本文档
└── user/                    # 用户模块测试
    ├── __init__.py
    ├── helpers.py           # 用户模块专用工具（User、create_mock_user）⭐
    └── test_register_user.py  # 用户相关测试
```

## 架构设计原则

### 分层设计

- **test/helpers.py**: 全局通用工具（`get_random_str`、`get_random_phone`）- 所有模块可用
- **test/user/helpers.py**: 用户模块专用工具（`User` 模型、`create_mock_user`）- 仅用户模块
- **test/api_client.py**: HTTP 客户端封装 - 所有模块共享
- **test/models.py**: 通用数据模型（`CommonResponse`）- 所有模块共享

### 模块化扩展

未来可能的目录结构：

```
test/
├── helpers.py               # 通用工具
├── api_client.py            # HTTP 客户端
├── models.py                # 通用模型
├── user/                    # 用户模块
│   ├── helpers.py           # User、create_mock_user
│   └── test_*.py
├── product/                 # 商品模块（未来）
│   ├── helpers.py           # Product、create_mock_product
│   └── test_*.py
└── payment/                 # 支付模块（未来）
    ├── helpers.py           # Payment、create_mock_payment
    └── test_*.py
```

## 核心测试工具

### 1. API 客户端 (api_client.py) ⭐

**统一的 HTTP 请求封装**，消除重复代码，提供一致的错误处理和日志记录。

#### 连接管理设计

ApiClient 使用 `requests.Session()` 维护 HTTP 连接池，并通过 pytest fixture 管理生命周期：

```python
# test/conftest.py 中定义的 session 级 fixture
@pytest.fixture(scope="session")
def api_client():
    """创建 API 客户端，自动管理连接生命周期"""
    client = ApiClient(base_url="http://localhost:8888")
    yield client
    client.close()  # 测试结束后自动关闭连接池
```

**优势**：
- ✅ **复用连接**: Session 自动管理连接池，提高性能
- ✅ **无泄漏**: 测试结束后自动调用 `close()` 释放资源
- ✅ **线程安全**: Session 对象在整个测试会话共享

#### 基本使用

```python
# 测试方法通过 fixture 注入，无需手动管理
def test_example(api_client: ApiClient):
    # 直接使用，无需担心连接管理
# 直接使用，无需担心连接管理
    response = api_client.get("/api/v1/users/testuser")
    assert response.status_code == 200
    print(f"响应时间: {response.response_time}ms")

# POST 请求
def test_create_user(api_client: ApiClient):
    user_data = {"username": "newuser", "password": "pass123", "email": "test@example.com"}
# DELETE/PUT/PATCH 请求
def test_operations(api_client: ApiClient):
    api_client.delete("/api/v1/users/testuser")
    api_client.put("/api/v1/users/testuser", data={"nickname": "新昵称"})
    api_client.patch("/api/v1/users/testuser", data={"email": "new@example.com"})
response: CommonResponse = api_client.put("/api/v1/users/testuser", data={"nickname": "新昵称"})
response: CommonResponse = api_client.patch("/api/v1/users/testuser", data={"email": "new@example.com"})
```

#### 特性

- ✅ **自动响应解析**: 根据 Content-Type 自动解析 JSON 或文本
- ✅ **连接池复用**: 使用 `requests.Session()` 自动管理连接池，减少 TCP 握手开销
- ✅ **自动资源释放**: pytest fixture 自动管理生命周期，测试结束时调用 `close()`
- ✅ **自动响应解析**: 根据 Content-Type 自动解析 JSON 或文本
- ✅ **统一错误处理**: 网络异常、超时等自动转换为 `CommonResponse(status_code=0)`
- ✅ **请求日志**: 自动记录 HTTP 方法、URL、状态码、响应时间
- ✅ **响应时间统计**: 每个请求自动计算响应时间（毫秒）
- ✅ **可配置超时**: 默认 5 秒，可自定义
- ✅ **上下文管理器支持**: 支持 `with` 语句（可选）

#### 手动管理（可选，不推荐）

如需手动管理连接（例如在非 pytest 环境）：

```python
# 方式1：使用 with 语句（推荐）
with ApiClient(base_url="http://localhost:8888") as client:
    response = client.get("/api/v1/users")
    # 退出时自动调用 close()

# 方式2：手动调用 close()
client = ApiClient(base_url="http://localhost:8888")
try:
    response = client.get("/api/v1/users")
finally:
    client.close()  # 确保连接被关闭
```
#### 自定义请求头

```python
def test_with_auth(api_client: ApiClient):
    # 添加自定义请求头
    custom_headers = {"Authorization": "Bearer token123"}
    response = api_client.get("/api/v1/protected", headers=custom_headers)
```

#### 日志输出示例

```22:52.686 | INFO | test.api_client:_make_request:97 - POST /api/v1/users/register - status:200, time:71.480ms
2026-01-04 23:22:52.947 | DEBUG | test.api_client:close:206 - ApiClient Session 已关闭
2026-01-04 23:04:22.792 | INFO | test.api_client:_make_request:92 - POST /api/v1/users/register - status:200, time:89.862ms
```

### 2. 公共测试工具 (helpers.py) ⭐

**通用测试工具**，提供跨模块共享的基础函数。

### 核心功能
` fixture、全局 `helpers.py` 和模块专用 `helpers.py` 编写简洁的测试代码。

```python
from test.helpers import get_random_str  # 全局通用工具
from test.user.helpers import create_mock_user  # 用户模块专用工具

def test_create_user(api_client: ApiClient):  # 通过 fixture 注入
    # 生成随机用户（使用用户模块的工厂函数）
    user = create_mock_user()
    
    # 发送请求（api_client 由 pytest 自动管理）
    # 发送请求
    response = api_client.post("/api/v1/users/register", data=user.model_dump())
    
    # 验证响应
    assert response.status_code == 200
    assert response.data["username"] == user.username
```

#### 1. `get_random_str(length: int) -> str`

生成指定长度的随机十六进制字符串。

```python
from test.helpers import get_random_str

uid = get_random_str(8)  # 例如: "a3f8b2c1"
```

#### 2. `get_random_phone() -> str`

生成符合中国大陆格式的随机手机号。

```python
from test.helpers import get_random_phone

phone = get_random_phone()  # 例如: "13812345678"
```

### 3. 用户模块专用工具 (user/helpers.py) ⭐

**用户模块特定的测试工具**，包含用户相关的模型和工厂函数。

#### 1. `User` - 用户数据模型

用户测试数据的 Pydantic 模型，包含完整的字段验证。

```python
from test.user.helpers import User

user = User(
    username="testuser",
    password="password123",
    email="test@example.com",
    phone_country_code="+86",
    phone_number="13812345678",
    nickname="测试用户"
)
```

#### 2. `create_mock_user(**kwargs) -> User`

创建随机测试用户，支持覆盖特定字段，避免数据冲突。

```python
from test.user.helpers import create_mock_user

# 生成完全随机的用户
user = create_mock_user()

# 覆盖特定字段
custom_user = create_mock_user(
    username="myuser",
    password="MyP@ssw0rd"
)
# 其他字段（email/phone）仍然随机生成，避免冲突
```

### 使用示例

```python
from test.helpers import get_random_str  # 全局工具
from test.user.helpers import create_mock_user  # 用户模块工具

def test_example():
    # 在循环中生成不同用户，避免 email/phone 冲突
    for i in range(3):
        test_user = create_mock_user()
        # 每个用户的 email/phone/username 都是唯一的
        
    # 使用全局工具生成随机字符串
    unique_id = get_random_str(8)
```

## 测试数据模型 (models.py)

所有测试用例使用统一的数据模型，定义在 `test/models.py` 中：

### CommonResponse - 统一的HTTP响应格式

**由 `ApiClient` 自动返回**，封装所有 HTTP 响应信息。

```python
from test.models import CommonResponse

response: CommonResponse = api_client.get("/api/v1/users/testuser")

# 检查状态码
assert response.status_code == 200

# 检查响应时间（毫秒）
assert response.response_time < 1000

# 访问响应数据（自动解析 JSON）
user_data = response.data
assert user_data["username"] == "testuser"
```

#### 字段说明

- `status_code: int` - HTTP 状态码（网络异常时为 0）
- `response_time: float` - 响应时间，单位毫秒
- `data: Any` - 响应数据（JSON 自动解析为 dict，其他为原始文本）

详细说明请参考 `test/models.py`、`test/helpers.py` 和 `test/user/helpers.py` 中的文档字符串。

## 最佳实践

### 编写新测试的步骤

1. **初始化 API 客户端**（测试文件顶部）

```python
from 声明 api_client fixture**（pytest 自动注入）

```python
from test.user.helpers import create_mock_user

def test_example(api_client: ApiClient):  # 自动注入，无需手动创建
    # api_client 已经配置好，可以直接使用
    pass

```python
def test_example():
    user = create_mock_user()
    # user.email、user.username、user.phone_number 都是唯一的
```

1. **发送请求**（使用 `api_client`，自动处理日志和异常）

```python
    response = api_client.post("/api/v1/users/register", data=user.model_dump())
```

1. **验证响应**（使用 `CommonResponse` 的结构化字段）

```python
    assert response.status_code == 200
    assert response.response_time < 500
    assert response.data["username"] == user.username
```

### 代码对比：重构前 vs 重构后

#### 重构前（~40 行）

```python
def create_user_request(user):
    url = "http://localhost:8888/api/v1/users/register"
    headers = {"Content-Type": "application/json"}
    
    try:
        start_time = time.time()
        resp = requests.post(url, json=user.model_dump(), headers=headers, timeout=5)
        end_time = time.time()
        response_time = (end_time - start_time) * 1000
        
        logger.info(f"POST {url} - status:{resp.status_code}, time:{response_time:.3f}ms")
        
        if "application/json" in resp.headers.get("Content-Type", ""):
            data = resp.json()
        else:
            data = resp.text
        
        return CommonResponse(
            status_code=resp.status_code,
            response_time=response_time,
            data=data
        )
    except requests.exceptions.RequestException as e:
        logger.error(f"请求失败: {e}")
        return CommonResponse(status_code=0, response_time=0, data={"error": str(e)})
```

#### 重构后（1 行）

```python
def create_user_request(user):
    return api_client.post("/api/v1/users/register", data=user.model_dump())
```

**效果**：

- 减少 ~110 行重复代码
- 统一错误处理和日志格式
- 更易维护和扩展

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
