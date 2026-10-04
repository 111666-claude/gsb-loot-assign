package lootassign

import "testing"

func TestSingleMemberGetsEverything(t *testing.T) {
	a := New()
	members := []Member{{Name: "solo", Weight: 1}}
	items := []Item{{ID: "i1", Tier: 1}, {ID: "i2", Tier: 1}}
	awards, ok := a.Assign("d1", 0, members, items)
	if !ok || len(awards) != 2 {
		t.Fatalf("一个人应该拿走两件，得到 %d 条 ok=%v", len(awards), ok)
	}
	for _, award := range awards {
		if award.Member != "solo" {
			t.Fatalf("独苗应该拿走全部，得到 %s", award.Member)
		}
	}
}

func TestQuotaOfEvenSplitIsExact(t *testing.T) {
	a := New()
	members := []Member{{Name: "a", Weight: 1}, {Name: "b", Weight: 1}}
	items := []Item{}
	for index := 0; index < 10; index++ {
		items = append(items, Item{ID: string(rune('a' + index)), Tier: 1})
	}
	awards, _ := a.Assign("d1", 0, members, items)
	count := map[string]int{}
	for _, award := range awards {
		count[award.Member]++
	}
	if count["a"] != 5 || count["b"] != 5 {
		t.Fatalf("平分应该每人 5 件，得到 a=%d b=%d", count["a"], count["b"])
	}
}

func TestScannedStartsAtZero(t *testing.T) {
	if New().Scanned() != 0 {
		t.Fatal("Scanned 初始应该是 0")
	}
}

func countByMember(awards []Award) map[string]int {
	count := map[string]int{}
	for _, award := range awards {
		count[award.Member]++
	}
	return count
}

func makeItems(n, tier int) []Item {
	items := []Item{}
	for index := 0; index < n; index++ {
		items = append(items, Item{ID: string(rune('a'+index/26)) + string(rune('a'+index%26)), Tier: tier})
	}
	return items
}

func TestShareRemainderBreaksTiesByName(t *testing.T) {
	a := New()
	members := []Member{{Name: "c", Weight: 1}, {Name: "a", Weight: 1}, {Name: "b", Weight: 1}}
	awards, ok := a.Assign("d1", 0, members, makeItems(11, 1))
	if !ok || len(awards) != 11 {
		t.Fatalf("11 件应该分出 11 条，得到 %d 条 ok=%v", len(awards), ok)
	}
	// 每人先拿 3 件，余 2 件小数部分相同，按名字升序补给 a 和 b。
	count := countByMember(awards)
	if count["a"] != 4 || count["b"] != 4 || count["c"] != 3 {
		t.Fatalf("余数应该按名字升序补，得到 a=%d b=%d c=%d", count["a"], count["b"], count["c"])
	}
}

func TestShareRemainderFollowsFraction(t *testing.T) {
	a := New()
	members := []Member{{Name: "a", Weight: 1}, {Name: "b", Weight: 2}, {Name: "c", Weight: 4}}
	awards, ok := a.Assign("d1", 0, members, makeItems(3, 1))
	if !ok || len(awards) != 3 {
		t.Fatalf("3 件应该分出 3 条，得到 %d 条 ok=%v", len(awards), ok)
	}
	// 份额 3/7、6/7、12/7：c 先拿 1 件，余 2 件按小数部分补给 b(6/7) 和 c(5/7)。
	count := countByMember(awards)
	if count["a"] != 0 || count["b"] != 1 || count["c"] != 2 {
		t.Fatalf("余数应该按小数部分降序补，得到 a=%d b=%d c=%d", count["a"], count["b"], count["c"])
	}
}

func TestItemsSortedByTierThenID(t *testing.T) {
	a := New()
	solo := []Member{{Name: "a", Weight: 1}}
	pool := []Item{
		{ID: "i1", Tier: 1},
		{ID: "i3", Tier: 3},
		{ID: "i2", Tier: 3},
		{ID: "i0", Tier: 2},
	}
	awards, _ := a.Assign("d1", 0, solo, pool)
	want := []string{"i2", "i3", "i0", "i1"}
	if len(awards) != len(want) {
		t.Fatalf("应该分出 %d 条，得到 %d 条", len(want), len(awards))
	}
	for index, id := range want {
		if awards[index].Item != id {
			t.Fatalf("第 %d 件应该是 %s，得到 %s", index, id, awards[index].Item)
		}
	}
}

func TestSeedPicksRingStart(t *testing.T) {
	team := []Member{{Name: "a", Weight: 1}, {Name: "b", Weight: 1}, {Name: "c", Weight: 1}}
	cases := []struct {
		seed int64
		want string
	}{
		{0, "a"}, {1, "b"}, {2, "c"}, {3, "a"}, {-1, "c"},
	}
	for _, tc := range cases {
		awards, _ := New().Assign("d1", tc.seed, team, makeItems(6, 1))
		if len(awards) == 0 || awards[0].Member != tc.want {
			t.Fatalf("seed=%d 第一件应该给 %s，得到 %+v", tc.seed, tc.want, awards)
		}
	}
}

func TestRingOrdersByWeightThenName(t *testing.T) {
	team := []Member{{Name: "x", Weight: 1}, {Name: "y", Weight: 3}, {Name: "z", Weight: 2}}
	// 环是 y(3)、z(2)、x(1)，seed=1 从 z 开始。
	awards, _ := New().Assign("d1", 1, team, makeItems(6, 1))
	if len(awards) == 0 || awards[0].Member != "z" {
		t.Fatalf("环应该按权重降序排，seed=1 从 z 开始，得到 %+v", awards)
	}
}

func TestSameDropIDSettlesOnce(t *testing.T) {
	a := New()
	team := []Member{{Name: "a", Weight: 1}, {Name: "b", Weight: 1}}
	first, ok1 := a.Assign("d1", 0, team, makeItems(2, 1))
	second, ok2 := a.Assign("d1", 0, team, makeItems(4, 1))
	if !ok1 || ok2 {
		t.Fatalf("第一次应该结算、第二次应该幂等，得到 ok1=%v ok2=%v", ok1, ok2)
	}
	if len(second) != len(first) {
		t.Fatalf("重发应该返回上一次的 %d 条，得到 %d 条", len(first), len(second))
	}
	for index := range first {
		if second[index] != first[index] {
			t.Fatalf("重发结果应该和上一次相同，第 %d 条 %v != %v", index, second[index], first[index])
		}
	}
	again, ok3 := a.Assign("d2", 0, team, makeItems(2, 1))
	if !ok3 || len(again) != 2 {
		t.Fatalf("新的 dropID 应该重新结算，得到 ok=%v 条数=%d", ok3, len(again))
	}
}

func TestAssignIsConservedAndDeterministic(t *testing.T) {
	members := []Member{{Name: "a", Weight: 5}, {Name: "b", Weight: 3}, {Name: "c", Weight: 1}, {Name: "d", Weight: 1}}
	items := makeItems(37, 1)
	first, _ := New().Assign("d1", 7, members, items)
	second, _ := New().Assign("d1", 7, members, items)
	if len(first) != len(items) {
		t.Fatalf("分配条数应该等于物品数 %d，得到 %d", len(items), len(first))
	}
	for index := range first {
		if first[index] != second[index] {
			t.Fatalf("同一份输入重复执行结果应该相同，第 %d 条不同", index)
		}
	}
	quota := shareQuota(len(items), members)
	count := countByMember(first)
	for name, got := range count {
		if got > quota[name] {
			t.Fatalf("%s 拿了 %d 件，超过份额 %d", name, got, quota[name])
		}
	}
	total := 0
	for _, q := range quota {
		total += q
	}
	if total != len(items) {
		t.Fatalf("份额之和应该等于物品数 %d，得到 %d", len(items), total)
	}
}

func TestScannedDoesNotBlowUp(t *testing.T) {
	a := New()
	members := []Member{}
	for index := 0; index < 300; index++ {
		members = append(members, Member{Name: string(rune('a'+index/26)) + string(rune('a'+index%26)) + string(rune('0'+index%10)), Weight: 1})
	}
	a.Assign("d1", 0, members, makeItems(3000, 1))
	if a.Scanned() > 6000 {
		t.Fatalf("3000 件 300 人扫描量应该在 6000 以内，得到 %d", a.Scanned())
	}
}
