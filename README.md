# Go 故障演练 + 找 Bug 题库

面向 Java 转 Go 的技术组长：既能 Review，也能定位和解决问题。
所有题目在 Go 1.25.0（linux/arm64）上实际运行过，现象与答案中的数据均来自真实采样。

```
go-drills/
├── fault/       13 道故障演练题（带 bug，要用工具定位）   ← 先做这里
├── bugs/        10 道"找 bug"题（代码审查型，读代码 + 跑一下）
├── solutions/   fault 题的修复版（已验证）
├── 01-故障演练-答案.md
└── 02-找bug-答案.md
```

零外部依赖，只需要 Go 1.22+（题目按 1.25 验证）。装好 Go 后：

```bash
cd go-drills
go run ./fault/f01_leak        # 运行某一题
go run -race ./fault/f06_mapfatal   # 带竞态检测
```

## 做题方法（每题 20~40 分钟）

1. 先**只运行**，观察现象，写下"我看到了什么"。
2. 用工具缩小范围，**别急着读代码**。pprof、trace、delve 就是你的 jstack、jmap、Arthas。
3. 写下根因（一句话）和修复方案，再看答案对照。
4. 修复后用同样的方法**验证现象消失**。没有验证的修复不算修复。
5. 每题写一条"能否变成 Review 清单或 lint 规则"。

## 题目索引

### A. 故障演练（fault/）

| 题号 | 现象 | 主要工具 | 难度 |
|---|---|---|---|
| f01_leak | goroutine 数持续上涨 | pprof goroutine | ★ |
| f02_httpbody | goroutine 和文件描述符一起涨 | pprof goroutine、`/proc/self/fd` | ★★ |
| f03_dbpool | 请求越来越慢直至全部超时 | `db.Stats()`、pprof goroutine | ★★ |
| f04_lock | 50 个协程，吞吐只有几十 ops/s | pprof mutex/block、trace | ★★ |
| f05_deadlock | 程序不崩，但转账数永远是 0 | pprof goroutine?debug=2 | ★★ |
| f06_mapfatal | 进程直接挂，recover 也没用 | `-race`、崩溃堆栈 | ★ |
| f07_heap | 内存只增不减 | pprof heap（inuse_space） | ★★ |
| f08_cpu | CPU 满载但吞吐很低 | pprof CPU、benchmark | ★★ |
| f09_gc | GC 极其频繁，吞吐低 | pprof allocs、gctrace、trace | ★★★ |
| f10_shadow | 报"导入成功"，但 5 行只入库 3 行 | delve | ★★ |
| f11_typednil | 校验通过的数据提交时 panic | delve | ★★ |
| f12_zerovalue | 前端把分数改成 0，保存后没变 | 单步 / 打印 | ★ |
| f13_gopanic | 服务有 recover 中间件，仍被打死 | 崩溃堆栈 | ★★ |

f01~f09 启动后自带流量发生器，并在 `localhost:6060` 暴露 pprof，**不需要另外压测**。进程 2~5 分钟后自动退出。

### B. 找 Bug（bugs/）

b01 切片 / b02 defer / b03 nil map / b04 JSON 精度 / b05 range 副本 / b06 时间格式 / b07 WaitGroup / b08 Mutex 拷贝 / b09 UTF-8 / b10 错误包装。
这组题**先不运行**，只读代码，写下你认为的 bug，再运行验证，最后看答案。其中 b07、b08 的 `go vet` 能直接报出来，其余 8 题只能靠人眼。

## 工具速查（对照 JVM）

| 想做什么 | Java | Go |
|---|---|---|
| 看所有线程 / 协程在干什么 | `jstack` | `curl localhost:6060/debug/pprof/goroutine?debug=2` |
| 同类协程聚合计数 | 无 | `...goroutine?debug=1`（最先看这个） |
| CPU 热点 | async-profiler | `go tool pprof -top -cum "http://localhost:6060/debug/pprof/profile?seconds=10"` |
| 堆里谁占着内存 | `jmap -histo` / MAT | `go tool pprof -sample_index=inuse_space -top .../heap` |
| 谁在疯狂分配 | JFR allocation | `go tool pprof -sample_index=alloc_space -top .../allocs` |
| 锁竞争 | JFR / Arthas | 先 `runtime.SetMutexProfileFraction(1)`，再看 `.../mutex`、`.../block` |
| 调度、GC、阻塞的时间线 | JFR | `curl -o t.out ".../trace?seconds=3"`，`go tool trace t.out`（浏览器） |
| GC 日志 | `-Xlog:gc` | `GODEBUG=gctrace=1 ./app` |
| 数据竞争 | 无（靠经验） | `go run -race` / `go test -race` |
| 断点调试 | IDEA / jdb | `dlv debug`、`dlv attach <pid>`、`dlv test` |
| 对比两次堆快照 | MAT 对比 | `go tool pprof -base a.pb.gz b.pb.gz` |

**pprof 交互常用命令：** `top`、`top -cum`、`list 函数名`（看到源码行）、`web`（调用图，需安装 graphviz）。
**把 pprof 开在内网端口**，不要直接暴露到公网。

### delve 脚本化（沙箱或 CI 中没有终端时）

```bash
cat > /tmp/dlv.init <<'EOF'
break fault/f10_shadow/main.go:34
continue
print r
print err
next
exit
EOF
dlv debug ./fault/f10_shadow --allow-non-terminal-interactive=true --init /tmp/dlv.init
```

日常交互调试只需 `dlv debug ./路径`，然后用 `b`（断点）、`c`（继续）、`n`（下一行）、`s`（进入函数）、`p 变量`、`locals`、`goroutines`、`goroutine N bt` 即可。

## 给 Review 的启示

做完整套题，你应该能把下面这些写进团队 Review 清单：

- 任何 `go func()` 都要问：**它怎么退出？panic 谁来兜？**
- 任何 `Body`、`rows`、`tx`、`file` 都要问：**哪一行关闭？所有分支都能走到吗？**
- 持锁区间里出现 IO、sleep、网络调用，直接打回。
- 正则、模板、客户端等"重对象"出现在函数体或循环里，直接打回。
- `if x, err := ...; err != nil` 里面再给外层 err 赋值，一律怀疑。
- 返回 `error` 的函数里出现 `var xxx *SomeErr`，直接打回。
- 请求结构体里 `int/bool/string + omitempty` 并且需要"部分更新"语义，直接打回。
