# 项目架构文档

## 1. 概述

本项目基于 [Go-Zero](https://go-zero.dev/) 框架开发，是一个遵循**分层架构**和**依赖倒置原则**的 RESTful API 服务。采用标准的 Go 工程结构，旨在提供清晰的职责划分、高可测试性和良好的扩展性。

### 1.1 技术栈

- **框架**: Go-Zero 1.9.2
- **数据库**: MySQL (GORM)
- **缓存**: Redis (go-redis v9)
- **消息队列**: Kafka
- **日志**: go-zero logx
- **配置管理**: go-zero conf

### 1.2 核心设计原则

- **分层架构**: 清晰的职责分离（Handler → Service → Repository → Infrastructure）
- **依赖倒置**: Service 依赖 Repository 接口而非具体实现
- **依赖注入**: 通过 ServiceContext 统一管理依赖
- **关注点分离**: 业务逻辑、数据访问、基础设施相互独立

---

## 2. 项目结构

```bash
hello-gozero/
├── api/                    # API 定义文件（go-zero api 语法）
│   ├── main.api           # 主 API 文件
│   ├── hello/             # hello 模块 API
│   └── user/              # 用户模块 API
├── app/                    # 应用入口
│   └── main.go            # 主程序入口
├── etc/                    # 配置文件
│   └── hellogozero.yaml   # 应用配置
├── infra/                  # 基础设施层（Infrastructure）
│   ├── cache/             # Redis 缓存封装
│   ├── database/          # MySQL 数据库连接
│   ├── executor/          # 批量执行器
│   └── queue/             # Kafka 消息队列
├── internal/               # 内部核心业务代码
│   ├── config/            # 配置结构体
│   ├── constant/          # 常量定义
│   ├── dto/               # 数据传输对象（Data Transfer Object）
│   ├── entity/            # 实体（Entity/Domain Model）
│   ├── handler/           # HTTP 处理器（Handler/Controller）
│   ├── repository/        # 数据访问层（Repository/DAO）
│   ├── routes/            # 路由注册
│   ├── service/           # 业务逻辑层（Service/Logic）
│   └── svc/               # 服务上下文（全局依赖注入容器）
├── sql/                    # SQL 脚本
├── scripts/                # 脚本文件
├── dockerfiles/            # Docker 相关文件
├── debug/                  # 调试脚本（Python）
├── docker-compose.yml      # Docker Compose 配置
├── go.mod                  # Go 模块依赖
└── README.md              # 项目说明
```

---

## 3. 分层架构

本项目采用**经典四层架构**：

```bash
┌────────────────────────────────────────────────────────┐
│                     Handler Layer                       │
│              (HTTP 请求处理、参数解析与校验)              │
└───────────────────────┬────────────────────────────────┘
                        │
┌────────────────────────────────────────────────────────┐
│                     Service Layer                       │
│              (核心业务逻辑、事务编排)                     │
└───────────────────────┬────────────────────────────────┘
                        │
┌────────────────────────────────────────────────────────┐
│                   Repository Layer                      │
│              (数据访问抽象、数据持久化)                   │
└───────────────────────┬────────────────────────────────┘
                        │
┌────────────────────────────────────────────────────────┐
│                Infrastructure Layer                     │
│         (数据库、缓存、消息队列等基础设施)                │
└────────────────────────────────────────────────────────┘
```

### 3.1 Handler Layer（处理器层）

**路径**: `internal/handler/`

**职责**:

- 接收 HTTP 请求并解析参数
- 调用 DTO 进行参数校验
- 调用 Service 层执行业务逻辑
- 处理响应和错误信息
- **不包含业务逻辑**

**特点**: 薄层设计，仅负责 HTTP 协议相关的处理

---

### 3.2 Service Layer（服务层）

**路径**: `internal/service/`

**职责**:

- 实现核心业务逻辑
- 编排多个 Repository 操作
- 处理事务管理
- 业务规则验证
- **通过 Repository 接口访问数据，不直接操作数据库**

**特点**: 包含所有业务逻辑，是系统的核心

---

### 3.3 Repository Layer（仓储层）

**路径**: `internal/repository/`

**职责**:

- 定义数据访问接口（Repository Interface）
- 实现具体的数据库操作
- 封装缓存逻辑（通过装饰器模式）
- **与业务逻辑解耦**

**设计模式**: Repository Pattern（仓储模式）

**特点**:

- 接口定义清晰，便于测试和替换实现
- 支持缓存装饰器（`CachedUserRepository`）用于特殊场景

---

### 3.4 Infrastructure Layer（基础设施层）

**路径**: `infra/`

**职责**:

- 封装外部依赖（MySQL、Redis、Kafka）
- 提供统一的连接管理和配置
- **不包含业务逻辑**

**组件**:

- `database/`: MySQL 连接池管理
- `cache/`: Redis 客户端封装
- `queue/`: Kafka 生产者/消费者
- `executor/`: 批量执行器

---

## 4. 核心组件

### 4.1 Entity（实体）

**路径**: `internal/entity/`

**定义**: 数据库表的 Go 结构体映射，使用 GORM 标签定义表结构

**特点**:

- 使用 UUID v7 作为主键（BINARY(16)）
- 支持软删除（`DeletedAt`）
- 包含 GORM Hooks（如 `BeforeCreate` 自动生成 UUID）

---

### 4.2 DTO（数据传输对象）

**路径**: `internal/dto/`

**定义**: API 请求/响应的数据结构

**特点**:

- 包含自定义校验逻辑（`Validate()` 方法）
- 与 Entity 分离，避免暴露数据库结构
- 支持结构化的校验错误响应

---

### 4.3 ServiceContext（服务上下文）

**路径**: `internal/svc/servicecontext.go`

**定义**: 全局依赖注入容器

**包含**:

- 配置对象（`Config`）
- 全局日志（`Logger`）
- 基础设施连接（`Infra`：MySQL、Redis、Kafka）
- 仓储接口（`Repository`）

**特点**:

- 在应用启动时初始化一次
- 通过构造函数注入到各层
- 支持优雅关闭（`Close()` 方法）

---

### 4.4 路由注册

**路径**: `internal/routes/routes.go`

**职责**:

- 注册所有 HTTP 路由
- 配置路由前缀和中间件
- 将路由与 Handler 绑定

**特点**: 集中管理，便于维护和版本控制

---

## 5. 数据流示例（用户注册）

```bash
客户端
  ↓ POST /api/v1/users/register
Handler Layer
  ├─ 解析请求参数（RegisterUserReq）
  ├─ DTO 校验（Validate）
  └─ 调用 Service
       ↓
Service Layer
  ├─ 检查用户名是否存在（调用 Repository）
  ├─ 密码加密（bcrypt）
  ├─ 创建 Entity 对象
  └─ 保存到数据库（调用 Repository）
       ↓
Repository Layer
  └─ 执行 GORM 操作（Create）
       ↓
Infrastructure Layer
  └─ 执行 SQL（INSERT INTO t_user ...）
       ↓
数据库（MySQL）
```

**关键点**:

- 每一层只关注自己的职责
- 层与层之间通过接口通信
- 数据在各层之间通过 DTO/Entity 传递

---

## 6. 配置管理

**配置文件**: `etc/hellogozero.yaml`

**配置结构**:

```bash
├─ 服务配置（Name, Host, Port）
└─ 基础设施配置（Infra）
   ├─ MySQL（连接信息、连接池参数）
   ├─ Redis（地址、超时、连接池）
   └─ Kafka（Brokers、Topic、Consumer Group）
```

**特点**:

- 分层配置结构，清晰易维护
- 支持默认值和校验
- 通过 `config.Config` 结构体加载

---

## 7. 依赖关系

### 7.1 依赖方向

```bash
Handler → Service → Repository → Infrastructure
   ↓         ↓          ↓              ↓
  DTO      Entity   Interface      实现
```

**原则**: 只能从上到下依赖，**禁止逆向依赖**

### 7.2 依赖注入

- 所有依赖通过 `ServiceContext` 管理
- 使用构造函数注入（`NewXxxService(ctx, svcCtx)`）
- Repository 使用接口定义，便于 Mock 测试

---

## 8. 错误处理策略

### 8.1 分层错误处理

| 层级 | 错误类型 | 处理方式 |
| ---- | ------- | ------- |
| **Handler** | HTTP 错误 | 返回 HTTP 状态码和结构化错误响应 |
| **Service** | 业务错误 | 定义业务错误类型（如 `ErrUsernameExists`） |
| **Repository** | 数据访问错误 | 返回原始数据库错误（如 `gorm.ErrRecordNotFound`） |
| **Infrastructure** | 基础设施错误 | 连接错误、超时等 |

### 8.2 错误传播

- Repository 返回原始错误
- Service 包装并添加上下文信息
- Handler 根据错误类型返回不同 HTTP 状态码

---

## 9. 最佳实践

### 9.1 命名规范

- **Handler**: `xxxHandler` (如 `RegisterUserHandler`)
- **Service**: `xxxService` (如 `RegisterUserService`)
- **Repository**: `xxxRepository` 接口 + `xxxRepositoryImpl` 实现
- **Entity**: 与表名对应（如 `User` → `t_user`）
- **DTO**: `xxxReq`/`xxxResp` (如 `RegisterUserReq`)

### 9.2 目录组织

- 按**业务模块**分包（如 `user/`, `order/`）
- 相同层级的代码放在同一目录下
- 避免循环依赖

### 9.3 测试策略

- **单元测试**: 使用 Mock Repository 测试 Service 层
- **集成测试**: 测试完整的 Handler → Service → Repository 流程
- **接口测试**: 测试 HTTP API

### 9.4 事务处理

- 在 Service 层使用 GORM 事务
- 通过 `db.Transaction()` 确保数据一致性
- 事务范围应尽可能小

---

## 10. 扩展指南

### 10.1 添加新功能

1. 定义 API 接口（可选）
2. 创建 DTO（`internal/dto/`）
3. 实现 Service 业务逻辑（`internal/service/`）
4. 创建 Handler（`internal/handler/`）
5. 注册路由（`internal/routes/`）
6. 如需要，扩展 Repository（`internal/repository/`）

### 10.2 添加新的基础设施

1. 在 `infra/` 创建新包（如 `elasticsearch/`）
2. 在 `config.Config.Infra` 添加配置
3. 在 `svc.ServiceContext.Infra` 添加连接字段
4. 在 `NewServiceContext()` 中初始化
5. 在 `Close()` 中清理资源

### 10.3 性能优化

- **缓存**: 使用 `CachedRepository` 装饰器添加缓存层
- **批量操作**: 使用 `infra/executor/batchexecutor.go`
- **连接池**: 调整 MySQL/Redis 连接池参数
- **异步处理**: 使用 Kafka 进行异步任务

---

## 11. 常见问题

### Q1: 为什么要分离 DTO 和 Entity？

**A**:

- **DTO** 面向 API，可灵活调整字段和校验规则
- **Entity** 面向数据库，保持稳定的表结构
- 避免数据库字段直接暴露给外部

### Q2: Repository 接口的优势是什么？

**A**:

- 便于单元测试（使用 Mock）
- 支持多种实现（MySQL、PostgreSQL、MongoDB）
- 可通过装饰器模式添加缓存、日志等功能

### Q3: 何时使用 CachedRepository？

**A**:

- 防止重复提交（分布式锁）
- API 限流（基于 Redis）
- 短期缓存热点数据
- **不要用于所有查询**，避免缓存一致性问题

### Q4: 事务应该在哪一层处理？

**A**: 在 **Service 层**处理事务，因为：

- Service 包含业务逻辑，知道哪些操作需要原子性
- Repository 只负责单一数据操作
- 避免事务跨层传播

---

## 12. 架构优势

### 12.1 可测试性

- 各层职责清晰，易于单元测试
- 通过接口隔离，便于 Mock
- 依赖注入使测试更灵活

### 12.2 可维护性

- 分层架构降低耦合
- 业务逻辑集中在 Service 层
- 统一的错误处理和日志

### 12.3 可扩展性

- 易于添加新功能模块
- 支持水平扩展（无状态设计）
- 基础设施可独立升级

### 12.4 性能

- 连接池管理优化资源使用
- 支持缓存和异步处理
- 可按需优化特定层级

---

## 13. 参考资料

- [Go-Zero 官方文档](https://go-zero.dev/)
- [GORM 文档](https://gorm.io/)
- [Clean Architecture](https://blog.cleancoder.com/uncle-bob/2012/08/13/the-clean-architecture.html)
- [Repository Pattern](https://martinfowler.com/eaaCatalog/repository.html)
- [依赖倒置原则（DIP）](https://en.wikipedia.org/wiki/Dependency_inversion_principle)

---

## 14. 变更日志

详见 [CHANGELOG.md](./CHANGELOG.md)
