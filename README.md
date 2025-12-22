# hello-gozero

## 项目结构

```bash
/internal
  /model      → Entity
  /repository → DAO
  /service    → 业务逻辑（使用 repository 和 model）
  /handler    → HTTP 处理（使用 DTO 和 service）
  /dto        → DTO
```

## 生成代码

### 生成 api 代码

参考：[*api demo 代码生成*](https://go-zero.dev/docs/tasks/cli/api-demo)

> 不推荐生成代码，不认可 goctl 代码结构的管理方式，特别是 `internal/types` 目录下的代码，建议手动编写 DTO 结构体。

```bash
# 生成 api
goctl api go --style gozero \
    --api api/main.api --type-group \
    --dir . #--test
```

## 连接中间件

### 连接 MySQL

```bash
# 从主机
mysql -h 127.0.0.1 -P 35068 -u root -prootpassword

# 从容器内
mysql -h mysql -P 3306 -u root -prootpassword
```

## 技术栈

- json 校验库 [gojsonschema](https://github.com/xeipuuv/gojsonschema) ，Json Schema 教程 [JSON Schema 规范](https://json-schema.apifox.cn)
- [Go-Zero中Validator库参数校验指南](https://www.bytezonex.com/archives/mPk2tcIv.html)
- [Validator 校验规则](https://juejin.cn/post/6847902214279659533)

## 更新日志

[CHANGELOG.md](./CHANGELOG.md)
