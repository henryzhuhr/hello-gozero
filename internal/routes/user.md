# 用户接口文档

## 概述

本文档描述了用户管理系统的 RESTful API 接口规范。

认证相关

- `POST /api/v1/users/login` - 用户登录
- `POST /api/v1/users/logout` - 用户登出
- `POST /api/v1/users/refresh-token` - 刷新认证令牌
- `GET /api/v1/users/me` - 获取当前登录用户信息

用户信息管理

- `PUT /api/v1/users/:username` - 更新用户信息（完整更新）
- `PATCH /api/v1/users/:username` - 部分更新用户信息
- `GET /api/v1/users/:username/profile` - 获取用户详细资料

密码管理

- `PUT /api/v1/users/:username/password` - 修改密码
- `POST /api/v1/users/password/reset` - 重置密码（忘记密码）
- `POST /api/v1/users/password/reset/verify` - 验证重置密码令牌

账户验证

- `POST /api/v1/users/verify-email` - 邮箱验证
- `POST /api/v1/users/resend-verification` - 重新发送验证邮件
- `POST /api/v1/users/verify-phone` - 手机号验证

账户状态管理

- `PUT /api/v1/users/:username/status` - 更新用户状态（启用/禁用/锁定）
- `PUT /api/v1/users/:username/activate` - 激活用户
- `PUT /api/v1/users/:username/deactivate` - 停用用户

权限和角色

- `GET /api/v1/users/:username/roles` - 获取用户角色
- `PUT /api/v1/users/:username/roles` - 更新用户角色
- `GET /api/v1/users/:username/permissions` - 获取用户权限

用户关系

- `GET /api/v1/users/:username/followers` - 获取关注者列表
- `GET /api/v1/users/:username/following` - 获取关注列表
- `POST /api/v1/users/:username/follow` - 关注用户
- `DELETE /api/v1/users/:username/follow` - 取消关注

用户活动

- `GET /api/v1/users/:username/activities` - 获取用户活动记录
- `GET /api/v1/users/:username/login-history` - 获取登录历史

## 基础信息

- **基础路径**: `/api/v1`
- **请求格式**: `application/json`
- **响应格式**: `application/json`

---

## 已实现的接口

### 1. 用户注册

- **端点**: `POST /api/v1/users/register`
- **描述**: 注册新用户
- **请求体**:

```json
{
  "username": "string",
  "email": "string",
  "password": "string"
}
```

- **响应**:

```json
{
  "code": 0,
  "msg": "success",
  "data": {
    "user_id": "string",
    "username": "string",
    "email": "string"
  }
}
```

### 2. 获取单个用户

- **端点**: `GET /api/v1/users/:username`
- **描述**: 根据用户名获取用户信息
- **路径参数**:
  - `username`: 用户名
- **响应**:

```json
{
  "code": 0,
  "msg": "success",
  "data": {
    "username": "string",
    "email": "string",
    "created_at": "timestamp",
    "updated_at": "timestamp"
  }
}
```

### 3. 获取用户列表

- **端点**: `GET /api/v1/users`
- **描述**: 获取用户列表（支持分页）
- **查询参数**:
  - `page`: 页码（默认：1）
  - `page_size`: 每页数量（默认：10）
  - `keyword`: 搜索关键词（可选）
- **响应**:

```json
{
  "code": 0,
  "msg": "success",
  "data": {
    "users": [
      {
        "username": "string",
        "email": "string",
        "created_at": "timestamp"
      }
    ],
    "total": 100,
    "page": 1,
    "page_size": 10
  }
}
```

### 4. 删除用户

- **端点**: `DELETE /api/v1/users/:username`
- **描述**: 删除指定用户
- **路径参数**:
  - `username`: 用户名
- **响应**:

```json
{
  "code": 0,
  "msg": "success",
  "data": null
}
```

---

## 推荐实现的接口

### 认证相关

#### 5. 用户登录

- **端点**: `POST /api/v1/users/login`
- **描述**: 用户登录获取认证令牌
- **请求体**:

```json
{
  "username": "string",
  "password": "string"
}
```

- **响应**:

```json
{
  "code": 0,
  "msg": "success",
  "data": {
    "access_token": "string",
    "refresh_token": "string",
    "expires_in": 3600,
    "user": {
      "username": "string",
      "email": "string"
    }
  }
}
```

#### 6. 用户登出

- **端点**: `POST /api/v1/users/logout`
- **描述**: 用户登出，清除认证令牌
- **请求头**: `Authorization: Bearer <token>`
- **响应**:

```json
{
  "code": 0,
  "msg": "success",
  "data": null
}
```

#### 7. 刷新令牌

- **端点**: `POST /api/v1/users/refresh-token`
- **描述**: 使用刷新令牌获取新的访问令牌
- **请求体**:

```json
{
  "refresh_token": "string"
}
```

- **响应**:

```json
{
  "code": 0,
  "msg": "success",
  "data": {
    "access_token": "string",
    "expires_in": 3600
  }
}
```

#### 8. 获取当前用户信息

- **端点**: `GET /api/v1/users/me`
- **描述**: 获取当前登录用户的信息
- **请求头**: `Authorization: Bearer <token>`
- **响应**:

```json
{
  "code": 0,
  "msg": "success",
  "data": {
    "username": "string",
    "email": "string",
    "created_at": "timestamp",
    "last_login": "timestamp"
  }
}
```

### 用户信息管理

#### 9. 更新用户信息（完整）

- **端点**: `PUT /api/v1/users/:username`
- **描述**: 完整更新用户信息
- **请求头**: `Authorization: Bearer <token>`
- **请求体**:

```json
{
  "email": "string",
  "phone": "string",
  "nickname": "string",
  "avatar": "string"
}
```

- **响应**:

```json
{
  "code": 0,
  "msg": "success",
  "data": {
    "username": "string",
    "email": "string",
    "phone": "string",
    "nickname": "string"
  }
}
```

#### 10. 部分更新用户信息

- **端点**: `PATCH /api/v1/users/:username`
- **描述**: 部分更新用户信息
- **请求头**: `Authorization: Bearer <token>`
- **请求体**:

```json
{
  "nickname": "string"
}
```

- **响应**:

```json
{
  "code": 0,
  "msg": "success",
  "data": {
    "username": "string",
    "nickname": "string"
  }
}
```

#### 11. 获取用户详细资料

- **端点**: `GET /api/v1/users/:username/profile`
- **描述**: 获取用户的详细资料信息
- **响应**:

```json
{
  "code": 0,
  "msg": "success",
  "data": {
    "username": "string",
    "email": "string",
    "phone": "string",
    "nickname": "string",
    "avatar": "string",
    "bio": "string",
    "created_at": "timestamp",
    "updated_at": "timestamp"
  }
}
```

### 密码管理

#### 12. 修改密码

- **端点**: `PUT /api/v1/users/:username/password`
- **描述**: 修改用户密码
- **请求头**: `Authorization: Bearer <token>`
- **请求体**:

```json
{
  "old_password": "string",
  "new_password": "string"
}
```

- **响应**:

```json
{
  "code": 0,
  "msg": "success",
  "data": null
}
```

#### 13. 重置密码请求

- **端点**: `POST /api/v1/users/password/reset`
- **描述**: 请求重置密码（忘记密码）
- **请求体**:

```json
{
  "email": "string"
}
```

- **响应**:

```json
{
  "code": 0,
  "msg": "success",
  "data": {
    "message": "重置密码邮件已发送"
  }
}
```

#### 14. 验证重置密码令牌

- **端点**: `POST /api/v1/users/password/reset/verify`
- **描述**: 验证重置密码令牌并设置新密码
- **请求体**:

```json
{
  "token": "string",
  "new_password": "string"
}
```

- **响应**:

```json
{
  "code": 0,
  "msg": "success",
  "data": null
}
```

### 账户验证

#### 15. 邮箱验证

- **端点**: `POST /api/v1/users/verify-email`
- **描述**: 验证用户邮箱
- **请求体**:

```json
{
  "token": "string"
}
```

- **响应**:

```json
{
  "code": 0,
  "msg": "success",
  "data": {
    "message": "邮箱验证成功"
  }
}
```

#### 16. 重新发送验证邮件

- **端点**: `POST /api/v1/users/resend-verification`
- **描述**: 重新发送验证邮件
- **请求头**: `Authorization: Bearer <token>`
- **响应**:

```json
{
  "code": 0,
  "msg": "success",
  "data": {
    "message": "验证邮件已发送"
  }
}
```

#### 17. 手机号验证

- **端点**: `POST /api/v1/users/verify-phone`
- **描述**: 验证用户手机号
- **请求体**:

```json
{
  "phone": "string",
  "code": "string"
}
```

- **响应**:

```json
{
  "code": 0,
  "msg": "success",
  "data": {
    "message": "手机号验证成功"
  }
}
```

### 账户状态管理

#### 18. 更新用户状态

- **端点**: `PUT /api/v1/users/:username/status`
- **描述**: 更新用户状态（管理员权限）
- **请求头**: `Authorization: Bearer <token>`
- **请求体**:

```json
{
  "status": "active|inactive|suspended|locked"
}
```

- **响应**:

```json
{
  "code": 0,
  "msg": "success",
  "data": {
    "username": "string",
    "status": "string"
  }
}
```

#### 19. 激活用户

- **端点**: `PUT /api/v1/users/:username/activate`
- **描述**: 激活用户账户（管理员权限）
- **请求头**: `Authorization: Bearer <token>`
- **响应**:

```json
{
  "code": 0,
  "msg": "success",
  "data": null
}
```

#### 20. 停用用户

- **端点**: `PUT /api/v1/users/:username/deactivate`
- **描述**: 停用用户账户（管理员权限）
- **请求头**: `Authorization: Bearer <token>`
- **响应**:

```json
{
  "code": 0,
  "msg": "success",
  "data": null
}
```

### 权限和角色

#### 21. 获取用户角色

- **端点**: `GET /api/v1/users/:username/roles`
- **描述**: 获取用户的角色列表
- **请求头**: `Authorization: Bearer <token>`
- **响应**:

```json
{
  "code": 0,
  "msg": "success",
  "data": {
    "roles": ["admin", "user"]
  }
}
```

#### 22. 更新用户角色

- **端点**: `PUT /api/v1/users/:username/roles`
- **描述**: 更新用户角色（管理员权限）
- **请求头**: `Authorization: Bearer <token>`
- **请求体**:

```json
{
  "roles": ["admin", "user"]
}
```

- **响应**:

```json
{
  "code": 0,
  "msg": "success",
  "data": {
    "username": "string",
    "roles": ["admin", "user"]
  }
}
```

#### 23. 获取用户权限

- **端点**: `GET /api/v1/users/:username/permissions`
- **描述**: 获取用户的权限列表
- **请求头**: `Authorization: Bearer <token>`
- **响应**:

```json
{
  "code": 0,
  "msg": "success",
  "data": {
    "permissions": [
      "user:read",
      "user:write",
      "admin:all"
    ]
  }
}
```

### 用户关系

#### 24. 获取关注者列表

- **端点**: `GET /api/v1/users/:username/followers`
- **描述**: 获取用户的关注者列表
- **查询参数**:
  - `page`: 页码
  - `page_size`: 每页数量
- **响应**:

```json
{
  "code": 0,
  "msg": "success",
  "data": {
    "followers": [
      {
        "username": "string",
        "avatar": "string",
        "followed_at": "timestamp"
      }
    ],
    "total": 100
  }
}
```

#### 25. 获取关注列表

- **端点**: `GET /api/v1/users/:username/following`
- **描述**: 获取用户关注的人列表
- **查询参数**:
  - `page`: 页码
  - `page_size`: 每页数量
- **响应**:

```json
{
  "code": 0,
  "msg": "success",
  "data": {
    "following": [
      {
        "username": "string",
        "avatar": "string",
        "followed_at": "timestamp"
      }
    ],
    "total": 100
  }
}
```

#### 26. 关注用户

- **端点**: `POST /api/v1/users/:username/follow`
- **描述**: 关注指定用户
- **请求头**: `Authorization: Bearer <token>`
- **响应**:

```json
{
  "code": 0,
  "msg": "success",
  "data": {
    "message": "关注成功"
  }
}
```

#### 27. 取消关注

- **端点**: `DELETE /api/v1/users/:username/follow`
- **描述**: 取消关注指定用户
- **请求头**: `Authorization: Bearer <token>`
- **响应**:

```json
{
  "code": 0,
  "msg": "success",
  "data": {
    "message": "已取消关注"
  }
}
```

### 用户活动

#### 28. 获取用户活动记录

- **端点**: `GET /api/v1/users/:username/activities`
- **描述**: 获取用户的活动记录
- **请求头**: `Authorization: Bearer <token>`
- **查询参数**:
  - `page`: 页码
  - `page_size`: 每页数量
  - `type`: 活动类型（可选）
- **响应**:

```json
{
  "code": 0,
  "msg": "success",
  "data": {
    "activities": [
      {
        "id": "string",
        "type": "login|update|delete",
        "description": "string",
        "timestamp": "timestamp"
      }
    ],
    "total": 100
  }
}
```

#### 29. 获取登录历史

- **端点**: `GET /api/v1/users/:username/login-history`
- **描述**: 获取用户的登录历史记录
- **请求头**: `Authorization: Bearer <token>`
- **查询参数**:
  - `page`: 页码
  - `page_size`: 每页数量
- **响应**:

```json
{
  "code": 0,
  "msg": "success",
  "data": {
    "history": [
      {
        "login_time": "timestamp",
        "ip_address": "string",
        "user_agent": "string",
        "location": "string"
      }
    ],
    "total": 100
  }
}
```

---

## 错误代码

| 错误码 | 说明 |
|--------|------|
| 0 | 成功 |
| 400 | 请求参数错误 |
| 401 | 未授权 |
| 403 | 禁止访问 |
| 404 | 资源不存在 |
| 409 | 资源冲突（如用户名已存在） |
| 422 | 验证失败 |
| 429 | 请求过于频繁 |
| 500 | 服务器内部错误 |

---

## 实施建议

### 优先级 P0（核心功能）

- 用户登录 (#5)
- 用户登出 (#6)
- 获取当前用户信息 (#8)
- 修改密码 (#12)

### 优先级 P1（重要功能）

- 刷新令牌 (#7)
- 更新用户信息 (#9, #10)
- 邮箱验证 (#15)
- 重置密码 (#13, #14)

### 优先级 P2（增强功能）

- 用户状态管理 (#18-20)
- 角色和权限管理 (#21-23)
- 用户活动记录 (#28-29)

### 优先级 P3（扩展功能）

- 用户关系（关注/粉丝）(#24-27)
- 手机号验证 (#17)

---

## 注意事项

1. **安全性**
   - 所有需要认证的接口必须验证 JWT token
   - 密码必须使用安全的哈希算法（如 bcrypt）
   - 实施请求频率限制防止暴力攻击
   - 敏感操作需要二次验证

2. **性能优化**
   - 对频繁访问的用户信息使用 Redis 缓存
   - 列表接口必须支持分页
   - 合理使用数据库索引

3. **数据验证**
   - 用户名：3-20 个字符，仅支持字母、数字、下划线
   - 邮箱：符合标准邮箱格式
   - 密码：至少 8 个字符，包含大小写字母和数字

4. **日志记录**
   - 记录所有用户操作日志
   - 记录登录失败尝试
   - 记录敏感操作（密码修改、权限变更等）

---

**文档版本**: v1.0  
**最后更新**: 2026-01-06
