# Pprof 性能分析使用指南

## 简介

pprof 是 Go 语言内置的性能分析工具，可以帮助开发者分析程序的 CPU 使用、内存分配、goroutine 状态、阻塞情况等，找出性能瓶颈。

## 配置启用

在 [etc/hellogozero.yaml](../etc/hellogozero.yaml) 中配置：

```yaml
Pprof:
  Enabled: true   # 是否启用 pprof
  Port: 6060      # pprof 服务端口
```

**⚠️ 安全提示**：

- 生产环境建议设置 `Enabled: false` 或通过防火墙限制访问
- pprof 端点会暴露程序内部信息，不应该公开访问

## 访问方式

启动服务后，pprof 会在独立端口提供 HTTP 服务：

```bash
# 查看所有可用的性能分析类型
http://localhost:6060/debug/pprof/
```

## 主要分析类型

### 1. CPU Profiling（CPU 性能分析）

分析 CPU 时间消耗，找出最耗时的函数。

```bash
# 采集 30 秒的 CPU profile
go tool pprof http://localhost:6060/debug/pprof/profile?seconds=30

# 采集后进入交互式命令行
# 常用命令：
# - top: 显示 CPU 占用最高的函数
# - list <函数名>: 查看函数的源代码和 CPU 占用
# - web: 生成调用图（需要安装 graphviz）
```

**示例输出**：

```bash
(pprof) top10
Showing nodes accounting for 1200ms, 80% of 1500ms total
      flat  flat%   sum%        cum   cum%
     500ms 33.33% 33.33%      500ms 33.33%  runtime.mallocgc
     300ms 20.00% 53.33%      300ms 20.00%  crypto/sha256.block
     200ms 13.33% 66.67%      400ms 26.67%  encoding/json.Marshal
```

### 2. Heap Profiling（堆内存分析）

分析内存分配情况，找出内存泄漏或过度分配。

```bash
# 查看当前堆内存使用情况
go tool pprof http://localhost:6060/debug/pprof/heap

# 进入交互式命令行后
(pprof) top
(pprof) list <函数名>
(pprof) web

# 或者直接保存为文件分析
curl http://localhost:6060/debug/pprof/heap > heap.prof
go tool pprof heap.prof
```

**常用选项**：

```bash
# 按分配对象数量排序
go tool pprof -alloc_objects http://localhost:6060/debug/pprof/heap

# 按分配空间大小排序（默认）
go tool pprof -alloc_space http://localhost:6060/debug/pprof/heap

# 按当前使用的对象数量
go tool pprof -inuse_objects http://localhost:6060/debug/pprof/heap

# 按当前使用的空间大小
go tool pprof -inuse_space http://localhost:6060/debug/pprof/heap
```

### 3. Goroutine Profiling（协程分析）

查看当前所有 goroutine 的状态。

```bash
# 查看 goroutine 信息
go tool pprof http://localhost:6060/debug/pprof/goroutine

# 或者直接在浏览器查看
curl http://localhost:6060/debug/pprof/goroutine?debug=1

# 查看完整的 goroutine 堆栈
curl http://localhost:6060/debug/pprof/goroutine?debug=2
```

### 4. Block Profiling（阻塞分析）

分析导致阻塞的操作，如 channel 操作、锁竞争等。

```bash
# 需要先在代码中启用 block profiling
# runtime.SetBlockProfileRate(1)

go tool pprof http://localhost:6060/debug/pprof/block
```

### 5. Mutex Profiling（互斥锁分析）

分析锁竞争情况。

```bash
# 需要先在代码中启用 mutex profiling
# runtime.SetMutexProfileFraction(1)

go tool pprof http://localhost:6060/debug/pprof/mutex
```

### 6. Allocs（内存分配采样）

类似 heap，但包含所有历史分配信息。

```bash
go tool pprof http://localhost:6060/debug/pprof/allocs
```

### 7. Threadcreate（线程创建分析）

查看导致创建新操作系统线程的堆栈。

```bash
go tool pprof http://localhost:6060/debug/pprof/threadcreate
```

## 实战示例

### 场景 1：查找 CPU 热点

```bash
# 1. 采集 CPU profile（30秒）
go tool pprof -http=:8080 http://localhost:6060/debug/pprof/profile?seconds=30

# 这会自动在浏览器打开一个 Web UI，展示：
# - 火焰图 (Flame Graph)
# - 调用图 (Graph)
# - Top 函数列表
# - 源代码视图
```

### 场景 2：诊断内存泄漏

```bash
# 1. 服务运行一段时间后，采集第一次内存快照
curl http://localhost:6060/debug/pprof/heap > heap1.prof

# 2. 继续运行一段时间，采集第二次快照
curl http://localhost:6060/debug/pprof/heap > heap2.prof

# 3. 对比两次快照，查看内存增长
go tool pprof -base heap1.prof heap2.prof

# 4. 在交互模式中查看内存增长最多的函数
(pprof) top
(pprof) list <函数名>
```

### 场景 3：排查 goroutine 泄漏

```bash
# 查看当前 goroutine 数量和状态
curl http://localhost:6060/debug/pprof/goroutine?debug=1

# 或使用 pprof 交互模式
go tool pprof http://localhost:6060/debug/pprof/goroutine
(pprof) top
(pprof) traces  # 显示完整的调用栈
```

### 场景 4：分析锁竞争

```bash
# 在程序启动时添加（可以在 main.go 中）
# runtime.SetMutexProfileFraction(1)

# 采集锁竞争信息
go tool pprof -http=:8080 http://localhost:6060/debug/pprof/mutex
```

## 可视化工具

### 1. 命令行交互模式

```bash
go tool pprof http://localhost:6060/debug/pprof/profile
```

常用命令：

- `top [N]`: 显示前 N 个函数（默认 10）
- `list <函数名>`: 显示函数源代码和性能数据
- `web`: 生成调用图并在浏览器打开（需要 graphviz）
- `pdf`: 生成 PDF 格式的调用图
- `png`: 生成 PNG 格式的调用图
- `traces`: 显示调用链
- `help`: 查看帮助

### 2. Web UI 模式（推荐）

```bash
# 启动 Web UI，自动在浏览器打开
go tool pprof -http=:8080 http://localhost:6060/debug/pprof/profile?seconds=30
```

Web UI 提供：

- **Graph**: 调用关系图
- **Flame Graph**: 火焰图（最直观）
- **Peek**: 函数详情
- **Source**: 源代码视图
- **Disassemble**: 汇编代码

### 3. 火焰图分析

火焰图是最直观的性能分析方式：

- X 轴：函数的执行时间占比
- Y 轴：调用栈深度
- 颜色：不同的函数
- 宽度：CPU 时间占用

```bash
# 生成火焰图
go tool pprof -http=:8080 http://localhost:6060/debug/pprof/profile?seconds=30
# 在浏览器中选择 "VIEW" -> "Flame Graph"
```

## 性能指标对比

### 采集基准数据

```bash
# 1. 压测前采集 CPU profile
go tool pprof -http=:8080 http://localhost:6060/debug/pprof/profile?seconds=60 > before.prof

# 2. 运行压测
# 例如使用 wrk 或 ab 工具

# 3. 压测中采集 CPU profile
go tool pprof -http=:8080 http://localhost:6060/debug/pprof/profile?seconds=60 > during.prof

# 4. 对比分析
go tool pprof -base before.prof during.prof
```

## 与压测工具结合

### 使用 wrk 进行压测

```bash
# 终端 1: 启动服务
./bin/server

# 终端 2: 开始压测
wrk -t4 -c100 -d30s http://localhost:8888/api/health

# 终端 3: 在压测期间采集 CPU profile
go tool pprof -http=:8080 http://localhost:6060/debug/pprof/profile?seconds=30
```

### 使用 ab 进行压测

```bash
# 压测同时采集性能数据
ab -n 10000 -c 100 http://localhost:8888/api/health &
go tool pprof http://localhost:6060/debug/pprof/profile?seconds=30
```

## 常见问题

### 1. graphviz 未安装

**错误**: `Failed to execute dot. Is Graphviz installed?`

**解决**:

```bash
# Ubuntu/Debian
apt-get install graphviz

# macOS
brew install graphviz

# Alpine (Docker)
apk add graphviz
```

### 2. pprof 端口无法访问

**检查**:

- 配置文件中 `Pprof.Enabled` 是否为 `true`
- 端口是否被占用
- 防火墙或 Docker 网络配置

### 3. 采集时间过长

**建议**:

- CPU profile：30-60 秒足够
- 如果程序 QPS 很高，可以缩短到 10-30 秒
- 内存 profile：实时快照，无需等待

### 4. profile 数据为空

**可能原因**:

- 采集时间内程序没有负载
- 某些 profile 类型需要手动启用（如 block, mutex）

## 生产环境建议

### 启用方式

1. **按需启用**: 默认关闭，需要时通过配置文件或环境变量临时启用
2. **访问限制**:
   - 通过防火墙限制只能从特定 IP 访问
   - 使用 VPN 或跳板机访问
   - 不要暴露到公网

3. **性能影响**:
   - pprof 本身对性能影响很小（<5%）
   - CPU profiling 会有 1-3% 的额外开销
   - Heap profiling 几乎无开销

### 配置示例

```yaml
# 开发环境
Pprof:
  Enabled: true
  Port: 6060

# 生产环境（默认关闭）
Pprof:
  Enabled: false
  Port: 6060
```

### 通过环境变量控制

可以扩展代码支持环境变量覆盖：

```go
// 示例：支持环境变量
if os.Getenv("ENABLE_PPROF") == "true" {
    startPprofServer(config.PprofConfig{
        Enabled: true,
        Port:    6060,
    })
}
```

## 参考资料

- [Go 官方 pprof 文档](https://pkg.go.dev/net/http/pprof)
- [Go 性能优化实战](https://go.dev/blog/pprof)
- [Profiling Go Programs](https://go.dev/blog/profiling-go-programs)
- [理解 Go pprof](https://github.com/google/pprof/blob/main/doc/README.md)
- [为什么golang pprof检测出的内存占用远小于top命令查看到的内存占用量？ - 小徐先生的回答 - 知乎](https://www.zhihu.com/question/453744293/answer/1899379237477146894)

## 快速命令参考

```bash
# CPU 分析（Web UI）
go tool pprof -http=:8080 http://localhost:6060/debug/pprof/profile?seconds=30

# 内存分析（Web UI）
go tool pprof -http=:8080 http://localhost:6060/debug/pprof/heap

# Goroutine 数量
curl http://localhost:6060/debug/pprof/goroutine?debug=1 | head -n 1

# 保存快照用于后续分析
curl http://localhost:6060/debug/pprof/profile?seconds=30 > cpu.prof
curl http://localhost:6060/debug/pprof/heap > heap.prof

# 离线分析
go tool pprof -http=:8080 cpu.prof
go tool pprof -http=:8080 heap.prof

# 对比两个 profile
go tool pprof -base=old.prof -http=:8080 new.prof
```
