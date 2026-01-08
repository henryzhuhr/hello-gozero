---
applyTo: '**/*.sql'
---
# SQL 的一些规范

## 字段对齐

在 `CREATE TABLE` 的时候确保字段和类型分别对齐，增强可读性，例如

```sql
CREATE TABLE IF NOT EXISTS user (
    `id`        BINARY(16) NOT NULL PRIMARY KEY,
    `username`  VARCHAR(50) NOT NULL 
)
```
