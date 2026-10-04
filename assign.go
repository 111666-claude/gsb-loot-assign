// Package lootassign 是掉落归属：按权重算件数份额、按种子决定轮转起点、分配结果幂等。
package lootassign

import "sort"

// Member 是一个参与分配的成员。
type Member struct {
	Name   string
	Weight int
}

// Item 是一件待分配物品。
type Item struct {
	ID   string
	Tier int
}

// Award 是一条分配结果。
type Award struct {
	Item   string
	Member string
}

// Assigner 是掉落归属台账。
type Assigner struct {
	ledger  map[string][]Award
	scanned int
}

// New 建一份空的归属台账。
func New() *Assigner { return &Assigner{ledger: map[string][]Award{}} }

// Assign 结算一次掉落归属，返回每件物品的归属；第二个返回值表示这次调用是否真的结算了。
// 同一个 dropID 只结算一次，重复调用返回上一次的结果与 false。
func (a *Assigner) Assign(dropID string, seed int64, members []Member, items []Item) ([]Award, bool) {
	if prev, ok := a.ledger[dropID]; ok {
		return prev, false
	}
	out := a.assign(seed, members, items)
	a.ledger[dropID] = out
	return out, true
}

// assign 真正结算一次掉落归属。
func (a *Assigner) assign(seed int64, members []Member, items []Item) []Award {
	// 物品顺序：品阶降序、同品阶按 ID 升序。
	order := append([]Item{}, items...)
	sort.Slice(order, func(i, j int) bool {
		if order[i].Tier != order[j].Tier {
			return order[i].Tier > order[j].Tier
		}
		return order[i].ID < order[j].ID
	})

	// 件数份额：先取整，再把余数件补给小数部分最大的前几名。
	quota := shareQuota(len(order), members)

	// 轮转环：成员按权重降序、同权重按名字升序排成环。
	ring := append([]Member{}, members...)
	sort.Slice(ring, func(i, j int) bool {
		if ring[i].Weight != ring[j].Weight {
			return ring[i].Weight > ring[j].Weight
		}
		return ring[i].Name < ring[j].Name
	})

	n := len(ring)
	if n == 0 || len(order) == 0 {
		return []Award{}
	}

	// 只把还有份额的成员链成环，拿够即摘链，每件物品只看一个成员，
	// 扫描量随物品数增长，不随成员数乘物品数放大。
	live := make([]bool, n)
	active := []int{}
	for i := range ring {
		if quota[ring[i].Name] > 0 {
			live[i] = true
			active = append(active, i)
		}
	}
	if len(active) == 0 {
		return []Award{}
	}
	next := make([]int, n)
	prev := make([]int, n)
	for k, i := range active {
		next[i] = active[(k+1)%len(active)]
		prev[i] = active[(k-1+len(active))%len(active)]
	}

	// 起点是 seed 对成员数取模（非负余数），零份额成员沿环跳过。
	pos := int(((seed % int64(n)) + int64(n)) % int64(n))
	for !live[pos] {
		pos = (pos + 1) % n
	}

	out := make([]Award, 0, len(order))
	for _, item := range order {
		a.scanned++
		member := ring[pos]
		out = append(out, Award{Item: item.ID, Member: member.Name})
		quota[member.Name]--
		if quota[member.Name] == 0 {
			next[prev[pos]] = next[pos]
			prev[next[pos]] = prev[pos]
		}
		pos = next[pos]
	}
	return out
}

// shareQuota 按权重算每人的件数份额：先拿 N×权重÷总权重 向下取整的件数，
// 余下的名次按「小数部分降序、权重降序、名字升序」给前几名各加一件，
// 直到份额之和正好等于 N。
func shareQuota(n int, members []Member) map[string]int {
	quota := map[string]int{}
	if n == 0 || len(members) == 0 {
		return quota
	}
	total := int64(0)
	for _, member := range members {
		total += int64(member.Weight)
	}
	if total <= 0 {
		return quota
	}
	type rank struct {
		name   string
		weight int
		rem    int64
	}
	ranks := make([]rank, 0, len(members))
	awarded := 0
	for _, member := range members {
		scaled := int64(n) * int64(member.Weight)
		base := scaled / total
		quota[member.Name] += int(base)
		awarded += int(base)
		ranks = append(ranks, rank{name: member.Name, weight: member.Weight, rem: scaled % total})
	}
	sort.Slice(ranks, func(i, j int) bool {
		if ranks[i].rem != ranks[j].rem {
			return ranks[i].rem > ranks[j].rem
		}
		if ranks[i].weight != ranks[j].weight {
			return ranks[i].weight > ranks[j].weight
		}
		return ranks[i].name < ranks[j].name
	})
	for i := 0; i < n-awarded && i < len(ranks); i++ {
		quota[ranks[i].name]++
	}
	return quota
}

// Scanned 是累计扫描的成员数（规模观测）。
func (a *Assigner) Scanned() int { return a.scanned }
