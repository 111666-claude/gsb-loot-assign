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
func (a *Assigner) Assign(dropID string, seed int64, members []Member, items []Item) ([]Award, bool) {
	if prior, ok := a.ledger[dropID]; ok {
		return prior, false
	}

	order := append([]Item{}, items...)
	sort.SliceStable(order, func(i, j int) bool {
		if order[i].Tier != order[j].Tier {
			return order[i].Tier > order[j].Tier
		}
		return order[i].ID < order[j].ID
	})

	ring := append([]Member{}, members...)
	sort.SliceStable(ring, func(i, j int) bool {
		if ring[i].Weight != ring[j].Weight {
			return ring[i].Weight > ring[j].Weight
		}
		return ring[i].Name < ring[j].Name
	})

	n := int64(len(order))
	totalWeight := int64(0)
	for _, member := range ring {
		totalWeight += int64(member.Weight)
	}

	quota := make([]int64, len(ring))
	remaining := make([]int, len(ring))
	if totalWeight > 0 {
		assigned := int64(0)
		type remainder struct {
			pos    int
			rem    int64
			weight int64
			name   string
		}
		remainders := make([]remainder, len(ring))
		for index, member := range ring {
			exact := n * int64(member.Weight)
			quota[index] = exact / totalWeight
			assigned += quota[index]
			remainders[index] = remainder{pos: index, rem: exact % totalWeight, weight: int64(member.Weight), name: member.Name}
		}
		left := n - assigned
		sort.SliceStable(remainders, func(i, j int) bool {
			if remainders[i].rem != remainders[j].rem {
				return remainders[i].rem > remainders[j].rem
			}
			if remainders[i].weight != remainders[j].weight {
				return remainders[i].weight > remainders[j].weight
			}
			return remainders[i].name < remainders[j].name
		})
		for index := int64(0); index < left; index++ {
			quota[remainders[index].pos]++
		}
		for index := range ring {
			remaining[index] = int(quota[index])
		}
	}

	// 用双向循环链表持有仍有份额的成员，每次取件只 O(1) 扫描，
	// 成员拿满后摘环，扫描总量不超过物品数 + 成员数。
	count := len(ring)
	next := make([]int, count)
	prev := make([]int, count)
	for index := range ring {
		next[index] = (index + 1) % count
		prev[index] = (index - 1 + count) % count
	}
	alive := 0
	cursor := 0
	if count > 0 {
		cursor = int(((seed % int64(count)) + int64(count)) % int64(count))
	}
	for index := 0; index < count; index++ {
		if remaining[index] > 0 {
			alive++
		} else if count > 1 {
			a.scanned++
			next[prev[index]] = next[index]
			prev[next[index]] = prev[index]
			if cursor == index {
				cursor = next[index]
			}
		}
	}

	out := []Award{}
	for _, item := range order {
		if alive == 0 {
			break
		}
		a.scanned++
		out = append(out, Award{Item: item.ID, Member: ring[cursor].Name})
		remaining[cursor]--
		if remaining[cursor] == 0 {
			alive--
			if alive > 0 {
				forward := next[cursor]
				next[prev[cursor]] = forward
				prev[forward] = prev[cursor]
				cursor = forward
			}
		} else {
			cursor = next[cursor]
		}
	}

	a.ledger[dropID] = out
	return out, true
}

// Scanned 是累计扫描的成员数（规模观测）。
func (a *Assigner) Scanned() int { return a.scanned }
