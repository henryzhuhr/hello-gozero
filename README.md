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

```bash
# 生成 api
goctl api go --style gozero \
    --api api/main.api \
    --dir . --test
```

## 技术栈

- json 校验库 [gojsonschema](https://github.com/xeipuuv/gojsonschema) ，Json Schema 教程 [JSON Schema 规范](https://json-schema.apifox.cn)
- [Go-Zero中Validator库参数校验指南](https://www.bytezonex.com/archives/mPk2tcIv.html)
- [Validator 校验规则](https://juejin.cn/post/6847902214279659533)

## 更新日志

[CHANGELOG.md](./CHANGELOG.md)
