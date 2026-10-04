# loot-assign

掉落归属：按权重算件数份额、按种子决定轮转起点、分配结果幂等，只用 Go 标准库。

```
go run ./cmd/loot-assign --sample=share
go run ./cmd/loot-assign --sample=seed
go run ./cmd/loot-assign --sample=tier
go run ./cmd/loot-assign --sample=idem
go run ./cmd/loot-assign --sample=scan
go test ./...
go vet ./...
```

## 口径（README 为准）

- **件数份额**：权重总和为 W、物品数为 N 时，每个成员先拿 `N × 权重 ÷ W` 向下取整的件数；
  余下的名次按「小数部分降序、权重降序、名字升序」依次给前几名各加一件，直到份额之和正好等于 N。
- **物品顺序**：物品按品阶降序、同一品阶按 ID 升序排列后依次分配。
- **轮转起点**：成员按权重降序、同权重按名字升序排成一个环，起点是 `seed` 对成员数取模（非负余数）；
  从起点沿环依次分配，每人只拿自己份额内的件数，拿够就跳过。
- **幂等**：同一个 `dropID` 只结算一次，重复调用返回上一次的结果，第二个返回值为 `false`。
- **守恒**：分配出的条数恰好等于物品数；每个成员拿到的条数不超过自己的份额；同一份输入重复执行结果相同。
- **规模**：1 万个物品、100 个成员单次分配 1 秒内完成，`Scanned` 不随成员数乘物品数放大。

## 输出契约（不改格式）

```
awarded=..
first0=.. first1=..
first=..
second=.. count=..
scanned=..
```

## 目录

```
assign.go             份额、轮转与幂等台账
assign_test.go        单元测试
cmd/loot-assign/      命令行入口
```
