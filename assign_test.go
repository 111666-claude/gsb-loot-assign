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

func TestQuotaRemainderGoesToLargestFraction(t *testing.T) {
	a := New()
	members := []Member{{Name: "a", Weight: 1}, {Name: "b", Weight: 1}, {Name: "c", Weight: 1}}
	items := []Item{}
	for index := 0; index < 11; index++ {
		items = append(items, Item{ID: string(rune('a' + index)), Tier: 1})
	}
	awards, ok := a.Assign("d1", 0, members, items)
	if !ok || len(awards) != 11 {
		t.Fatalf("11 件应该全部分完，得到 %d 条 ok=%v", len(awards), ok)
	}
	count := map[string]int{}
	for _, award := range awards {
		count[award.Member]++
	}
	// 每人 floor(11/3)=3，余 2 件按小数部分并列时权重相同、按名字升序补给 a 和 b。
	if count["a"] != 4 || count["b"] != 4 || count["c"] != 3 {
		t.Fatalf("余数应该补给 a 和 b，得到 a=%d b=%d c=%d", count["a"], count["b"], count["c"])
	}
}

func TestQuotaRemainderPrefersLargerFraction(t *testing.T) {
	a := New()
	members := []Member{{Name: "x", Weight: 1}, {Name: "y", Weight: 3}}
	items := []Item{}
	for index := 0; index < 3; index++ {
		items = append(items, Item{ID: string(rune('a' + index)), Tier: 1})
	}
	awards, _ := a.Assign("d1", 0, members, items)
	count := map[string]int{}
	for _, award := range awards {
		count[award.Member]++
	}
	// x 得 floor(3/4)=0 余 3/4，y 得 floor(9/4)=2 余 1/4，剩 1 件补给小数部分更大的 x。
	if count["x"] != 1 || count["y"] != 2 {
		t.Fatalf("小数部分更大的 x 应该拿余数，得到 x=%d y=%d", count["x"], count["y"])
	}
}

func TestItemsSortedByTierDescThenIDAsc(t *testing.T) {
	a := New()
	members := []Member{{Name: "solo", Weight: 1}}
	items := []Item{
		{ID: "i2", Tier: 1},
		{ID: "i1", Tier: 3},
		{ID: "i3", Tier: 3},
		{ID: "i0", Tier: 2},
	}
	awards, _ := a.Assign("d1", 0, members, items)
	want := []string{"i1", "i3", "i0", "i2"}
	if len(awards) != len(want) {
		t.Fatalf("应该分出 %d 件，得到 %d 条", len(want), len(awards))
	}
	for index, id := range want {
		if awards[index].Item != id {
			t.Fatalf("第 %d 件应该是 %s，得到 %s", index, id, awards[index].Item)
		}
	}
}

func TestSeedPicksRotationStart(t *testing.T) {
	team := []Member{{Name: "a", Weight: 1}, {Name: "b", Weight: 1}, {Name: "c", Weight: 1}}
	items := []Item{}
	for index := 0; index < 6; index++ {
		items = append(items, Item{ID: string(rune('a' + index)), Tier: 1})
	}
	first := map[int64]string{}
	for seed := int64(0); seed < 3; seed++ {
		awards, _ := New().Assign("d1", seed, team, items)
		if len(awards) == 0 {
			t.Fatalf("seed=%d 没有分出任何物品", seed)
		}
		first[seed] = awards[0].Member
	}
	if first[0] != "a" || first[1] != "b" || first[2] != "c" {
		t.Fatalf("起点应该随 seed 轮转，得到 %v", first)
	}
}

func TestRotationSkipsMembersWhoFilledQuota(t *testing.T) {
	a := New()
	members := []Member{{Name: "big", Weight: 3}, {Name: "small", Weight: 1}}
	items := []Item{}
	for index := 0; index < 4; index++ {
		items = append(items, Item{ID: string(rune('a' + index)), Tier: 1})
	}
	awards, _ := a.Assign("d1", 0, members, items)
	count := map[string]int{}
	for _, award := range awards {
		count[award.Member]++
	}
	if count["big"] != 3 || count["small"] != 1 {
		t.Fatalf("big 应该拿 3 件 small 拿 1 件，得到 big=%d small=%d", count["big"], count["small"])
	}
}

func TestSameDropIDSettlesOnlyOnce(t *testing.T) {
	a := New()
	team := []Member{{Name: "a", Weight: 1}, {Name: "b", Weight: 1}}
	first, ok1 := a.Assign("d1", 0, team, []Item{{ID: "i1", Tier: 1}, {ID: "i2", Tier: 1}})
	second, ok2 := a.Assign("d1", 0, team, []Item{{ID: "i3", Tier: 1}, {ID: "i4", Tier: 1}, {ID: "i5", Tier: 1}, {ID: "i6", Tier: 1}})
	if !ok1 {
		t.Fatal("第一次调用应该真的结算")
	}
	if ok2 {
		t.Fatal("重复 dropID 不应该重新结算")
	}
	if len(second) != len(first) {
		t.Fatalf("重复调用应该返回上次结果，得到 %d 条，上次是 %d 条", len(second), len(first))
	}
	for index := range first {
		if second[index] != first[index] {
			t.Fatalf("重复调用结果应该一致，第 %d 条 %v != %v", index, second[index], first[index])
		}
	}
}

func TestAssignIsDeterministic(t *testing.T) {
	members := []Member{{Name: "a", Weight: 2}, {Name: "b", Weight: 1}, {Name: "c", Weight: 5}}
	items := []Item{}
	for index := 0; index < 40; index++ {
		items = append(items, Item{ID: string(rune('a' + index%26)), Tier: index % 4})
	}
	run := func() []Award {
		awards, _ := New().Assign("d1", 7, members, items)
		return awards
	}
	first, second := run(), run()
	if len(first) != len(items) || len(second) != len(items) {
		t.Fatalf("分配条数应该等于物品数，得到 %d 和 %d", len(first), len(second))
	}
	for index := range first {
		if first[index] != second[index] {
			t.Fatalf("同样输入应该得到同样结果，第 %d 条不同", index)
		}
	}
}

func TestScannedDoesNotScaleWithMembersTimesItems(t *testing.T) {
	a := New()
	members := []Member{}
	for index := 0; index < 300; index++ {
		members = append(members, Member{Name: string(rune('a'+index%26)) + string(rune('A'+index/26)), Weight: 1})
	}
	items := []Item{}
	for index := 0; index < 3000; index++ {
		items = append(items, Item{ID: string(rune('a'+index%26)) + string(rune('A'+index/26%26)) + string(rune('0'+index/676)), Tier: 1})
	}
	awards, _ := a.Assign("d1", 0, members, items)
	if len(awards) != len(items) {
		t.Fatalf("应该分出 %d 件，得到 %d 条", len(items), len(awards))
	}
	if a.Scanned() > 2*(len(items)+len(members)) {
		t.Fatalf("Scanned 不应该随成员数乘物品数放大，得到 %d", a.Scanned())
	}
}
