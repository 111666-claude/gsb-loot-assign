// Package lootassign 是掉落归属：按权重算件数份额、按种子决定轮转起点、分配结果幂等。
// 缺陷：份额只取整不做余数分配、物品只按 ID 排不看品阶、种子不参与轮转起点、
// 每件物品都全扫一遍成员、同一个 dropID 重复结算。
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
func (a *Assigner) Assign(dropID string, seed int64, members []Member, items []Item) ([]Award, bool) {
	order := append([]Item{}, items...)
	sort.Slice(order, func(i, j int) bool { return order[i].ID < order[j].ID })
	weight := 0
	for _, member := range members {
		weight += member.Weight
	}
	quota := map[string]int{}
	for _, member := range members {
		quota[member.Name] = len(order) * member.Weight / weight
	}
	out := []Award{}
	for _, item := range order {
		best := -1
		for index, member := range members {
			a.scanned++
			if quota[member.Name] <= 0 {
				continue
			}
			if best < 0 || quota[member.Name] > quota[members[best].Name] {
				best = index
			}
		}
		if best < 0 {
			continue
		}
		quota[members[best].Name]--
		out = append(out, Award{Item: item.ID, Member: members[best].Name})
	}
	a.ledger[dropID] = out
	return out, true
}

// Scanned 是累计扫描的成员数（规模观测）。
func (a *Assigner) Scanned() int { return a.scanned }
