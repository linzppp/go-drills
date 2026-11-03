# "找 Bug"题集 参考答案（bugs/ b01~b10）

> 先读代码、写下你认为的 bug，再 `go run ./bugs/bXX_xxx` 验证，最后看这里。
> 所有输出都来自 Go 1.25.0 的真实运行。

| 题 | 一句话 | go vet 能发现？ |
|---|---|---|
| b01 | append 到子切片，污染了原数组 | 否 |
| b02 | 循环里 defer，连接到函数结束才关 | 否 |
| b03 | 向 nil map 写入 panic | 否 |
| b04 | int64 / float64 精度丢失 | 否 |
| b05 | range 拿到的是副本，改了不生效 | 否 |
| b06 | Java 的时间格式 + 默认 RFC3339 + 无时区 | 否 |
| b07 | `wg.Add` 写在协程里面 | **是** |
| b08 | 值接收者拷贝了 Mutex | **是** |
| b09 | `len`/`s[:n]` 是字节，不是字符 | 否 |
| b10 | `%v` 包装丢了错误链，`==`/类型断言失效 | 否 |

---

## b01 切片共享底层数组

**输出：**
```
重点校: [一中 二中 试点校]
全部学校: [一中 二中 试点校 四中 五中]     ← 原数据被改了：三中 消失
```
**原因：** `all[:2]` 与 `all` 共享同一个底层数组，容量是 5。`append(key, "试点校")` 容量够用，直接**写进了原数组的下标 2**，覆盖了"三中"。
**修复：** 截取后需要追加时先拷贝，用 `slices.Clone(all[:n])`，或限制容量：`all[:n:n]`（三下标切片，容量 n，append 时会重新分配）。
**Review 要点：** 函数接收 slice 参数并 `append`，要想清楚调用方是否还在用它。返回新切片的函数，要么明确"会修改入参"，要么先拷贝。

---

## b02 循环中的 defer

**输出：** 三个连接全部打开、处理完、打印"全部处理完毕"之后，才依次 `关闭 C库 / B库 / A库`。
**原因：** `defer` 在**函数返回时**才执行，不是循环迭代结束。处理 1 万个数据源就会同时持有 1 万个连接/文件句柄，并且按后进先出关闭。
**修复：** 把循环体抽成函数 `processOne(name)`，里面 `defer c.Close()`；或者在循环体末尾显式调用 `Close()`。
**Review 要点：** `for` 里的 `defer` 基本都是 bug，除非明确想"延迟到函数结束"。

---

## b03 nil map

**输出：**
```
读 nil map: 0
panic: assignment to entry in nil map
```
**原因：** map 的零值是 nil。读取安全（返回零值），**写入 panic**。`NewSummary` 没有初始化 `Scores`。
**修复：** `NewSummary` 里 `Scores: make(map[string]int)`；或在 `Add` 里做懒初始化 `if s.Scores == nil { s.Scores = ... }`。
**Review 要点：** 结构体里有 map 字段，就要检查构造函数；JSON 反序列化出来的 map 也可能为 nil（字段缺失时）。slice 的零值（nil）可以直接 `append`，这是它与 map 的区别。

---

## b04 JSON 与大整数

**输出：**
```
服务端输出: {"id":1849302847562301441,"title":"期中填报"}
map 接收: 1.8493028475623014e+18 类型 float64
还原成整数: 1849302847562301440          ← 末位 1 变成了 0
```
**原因：**
1. 用 `map[string]any` 接收时，JSON 数字一律变成 `float64`，超过 2^53 的整数**精度丢失**。
2. 同样的问题发生在浏览器：JS 的 `Number` 也是 float64，雪花 ID 返回给前端后，前端读到的已经是另一个数（`...2301400`），再提交回来就变成了别的记录。注意这个在 Go 输出里看不到，**要在前后端联调时才会发现**。

**修复：**
- 返回给前端的 int64 ID 加 `json:"id,string"`，或统一使用字符串类型。
- 服务端解析使用定义好的 struct，不用 `map[string]any`；必须用 map 时，`decoder.UseNumber()`。
**Review 要点：** 看到雪花 ID、订单号、学籍号这类 int64，立刻想到前端精度。看到 `map[string]any` + 数字，问一句"会不会超过 2^53"。

---

## b05 range 是副本

**输出：**
```
[{张三 55} {李四 80} {王五 58}]      ← 没有加分
map[张三:{张三 55}]
```
**原因：** `for _, s := range students` 里的 `s` 是元素的**副本**，改的是副本。
另外，`byName["张三"].Score = 65` 会直接**编译失败**（`cannot assign to struct field ... in map`），因为 map 元素不可寻址。

**修复：**
```go
for i := range students {          // 用下标
    if students[i].Score < 60 { students[i].Score += 10 }
}
// 或：切片元素类型用 []*Student
// map：先取出整个 struct，改完再放回去；或 value 用指针 map[string]*Student
```
**Review 要点：** 对 range 变量的赋值和字段修改，是 Go 新手第一周必踩的坑。

---

## b06 时间格式

**输出：**
```
用 Java 习惯格式化: yyyy-MM-dd HH:mm:ss                  ← 原样输出，毫无意义
默认 JSON 输出: {"createAt":"2026-10-15T09:30:05Z"}       ← 前端期望 "2026-10-15 09:30:05"
解析前端时间: parsing time "2026-10-15 09:30:05" as "2006-01-02T15:04:05Z07:00": cannot parse ...
time.Parse 得到的时区: UTC                                ← 不是本地时区
```
**原因：**
1. Go 用**参考时间**做模板：`2006-01-02 15:04:05`（1 月 2 日下午 3 点 4 分 5 秒，2006 年），不是 `yyyy-MM-dd`。写错时不会报错，只是原样输出。
2. `time.Time` 的 JSON 默认是 RFC3339，与国内常用的 `yyyy-MM-dd HH:mm:ss` 不同；反序列化也只认 RFC3339。
3. `time.Parse` 解析不带时区的字符串时按 **UTC**，不是本机时区。需要北京时间要用 `time.ParseInLocation(layout, s, loc)`。
4. MySQL 驱动的 DSN 需要 `parseTime=true&loc=Local`，否则读出来的时间也会错乱。

**修复：** 定义统一的 `LocalTime` 类型，实现 `MarshalJSON` / `UnmarshalJSON`，全项目复用；固定使用 `time.LoadLocation("Asia/Shanghai")`。
**Review 要点：** 搜索 `Format("yyyy` 或 `"yyyy`、`time.Parse(`、`time.Now()` 在单测中的使用（建议注入 `clock`）。

---

## b07 WaitGroup.Add 位置错误

**输出（每次不同）：** `期望 100 条结果，实际 70 条`
**go vet：** `WaitGroup.Add called from inside new goroutine`

**原因：** `wg.Add(1)` 写在新协程内部，主协程执行到 `wg.Wait()` 时，很多协程还没来得及 `Add`，`Wait` 看到计数为 0 就直接返回，结果不完整。同时 `results` 被读取时可能仍有协程在写。
**修复：** `Add` 必须在 `go` 语句**之前**调用；更推荐用 `errgroup.Group`（Go 1.25 的 `sync.WaitGroup` 也新增了 `Go` 方法，请以官方文档为准）。
**Review 要点：** 这类 bug 本地几乎看不出，生产高负载才暴露。CI 必须跑 `go vet`。

---

## b08 值接收者拷贝了 Mutex

**输出：** `期望 1000，实际 0`
**go vet：** `Inc passes lock by value: Counter contains sync.Mutex`

**原因：** `func (c Counter) Inc()` 是值接收者，每次调用拷贝整个 `Counter`（包括锁和 `n`）。`Inc` 修改的是副本，原对象的 `n` 永远是 0，锁也锁在副本上，毫无保护作用。
**修复：** 所有方法改为**指针接收者** `func (c *Counter)`。含有 `sync.Mutex`、`sync.WaitGroup`、`atomic` 的结构体，**绝不能值传递**。
**Review 要点：** 一个类型的方法，要么全是指针接收者，要么全是值接收者；只要有一个需要修改状态，全部用指针。

---

## b09 字节 vs 字符

**输出：**
```
截断结果: 北京...         ← 取了 6 字节，只有 2 个汉字
len: 51  字符数应为 17
校验失败：名称过长         ← 15 个汉字被判成 45 字节 > 10
```
**原因：** Go 字符串是 UTF-8 字节序列，一个汉字占 3 字节。`len(s)` 和 `s[:n]` 都以**字节**为单位。如果截断位置恰好落在汉字中间，还会产生乱码（`�`）。
**修复：**
- 计算字符数：`utf8.RuneCountInString(s)` 或 `len([]rune(s))`。
- 截断：转成 `[]rune` 再切，`string(r[:n])`。
- 校验"最多 N 个字"时，也要用字符数；如果数据库是 `varchar(N)`，MySQL 按字符数算，也要对应。

**Review 要点：** 所有 `len(str)` 和 `str[a:b]` 出现在校验、展示、截断场景，要问"这里会有中文吗"。

---

## b10 错误包装与判断

**输出：**
```
500 内部错误: serviceGet(7): record not found
500 内部错误: serviceGet2(7): not found in report
```
两处本该是 404，却都变成 500。

**原因：**
1. `serviceGet` 用的是 `%v` 包装，只把错误**转成字符串**，原始错误丢失，`err == ErrNotFound` 恒为 false。
2. `serviceGet2` 用了 `%w` 正确包装，但 `handler` 用**类型断言** `err.(*NotFoundError)` 判断，它只检查最外层，而最外层是 `*fmt.wrapError`。

**修复：**
- 包装用 `%w`。
- 判断用 `errors.Is(err, ErrNotFound)`（比较哨兵错误），`errors.As(err, &nf)`（取出具体类型）。
- 不要用 `==` 或类型断言去判断可能被包装过的错误。

**Review 要点：** 搜索 `err ==`、`err.(*`、`fmt.Errorf(...%v...err)`。Java 里的 `catch (NotFoundException e)` 对应的就是 `errors.As`，别写成断言。

---

## 另外两个高频题（没有做成代码，但 Review 时要留意）

**循环变量捕获：** Go 1.22 之前，`for i := ...  { go func(){ use(i) }() }` 所有协程看到同一个 `i`。**Go 1.22 起每次迭代都是新变量**，已不再是 bug。但老代码、老博客、老课程里还有大量讲解，Review 时先看 `go.mod` 里的 `go` 版本再下结论。

**defer 参数在声明时求值：** `defer fmt.Println(x)` 记录的是此刻的 `x`；`defer func(){ fmt.Println(x) }()` 才是退出时的 `x`。
