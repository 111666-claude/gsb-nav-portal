# nav-portal

传送门连通图：双向连通、登记幂等、成环拒绝、删除后重新合并成通道，只用 Go 标准库。

```
go run ./cmd/nav-portal --sample=bidir
go run ./cmd/nav-portal --sample=dup
go run ./cmd/nav-portal --sample=cycle
go run ./cmd/nav-portal --sample=remove
go run ./cmd/nav-portal --sample=merge
go run ./cmd/nav-portal --sample=scan
go test ./...
go vet ./...
```

## 口径（README 为准）

- **双向连通**：一条传送门让两个锚点互达，`Reachable(x, y)` 与 `Reachable(y, x)` 同真。
- **幂等**：同一个 `id` 只登记一次；重复登记返回 `false`，条数与通道都不变。
- **合法性**：锚点名为空或两端相同一律返回 `false`，图不变。
- **链约束与环检测**：每个锚点最多连两条传送门；两端已经连通（会成环）或某一端已经连满两条
  都返回 `false` 且不改图，成环时第二个返回值给出环上的节点（从新边一端沿已有路径到另一端，
  不含重复的起点）。因此图永远是若干条互不相交的链。
- **删除**：`Remove` 之后两个锚点不再互达，通道合并结果随之更新。
- **通道合并**：`Chains` 把首尾相接的传送门合并成一条通道，每个连通分量只输出一条，
  通道从该分量里字典序最小的一端开始依次列出锚点。
- **代价**：连通判定与登记走增量维护的索引，不许每次全图遍历，`Scanned` 不随边数乘操作数放大；
  10 万条传送门、每秒 10 万次登记与查询。

## 输出契约（不改格式）

```
ba=..
count=.. segments=..
ok=.. cycle=..
reachable=..
segments=.. first=..
scanned=..
```

## 目录

```
portal.go             连通、幂等、环检测与通道合并
portal_test.go        单元测试
cmd/nav-portal/       命令行入口
```
