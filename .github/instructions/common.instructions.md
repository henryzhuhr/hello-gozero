
# 项目通用规范

## 项目运行环境

### 容器运行环境

当前的项目已经通过 `docker compose up -d` 在容器中启动开发，完全隔离宿主机环境，所有的开发均已经在容器中进行，不会影响宿主机环境。

### Python 环境

当前项目使用 `uv` 进行包管理，如果需要执行 Python 相关命令，可以执行

```bash
# 
uv run <your_python_script>.py
```

如果要运行测试框架 pytest，可以执行

```bash
uv run pytest
uv run pytest <your_test_script>.py
```

## 连接中间件

### 连接 MySQL

当前的环境已经内置了 MySQL 数据库容器，可以通过以下命令连接到数据库容器内，并执行 SQL 语句

```bash
mysql -h mysql -P 3306 -u root -prootpassword
```
