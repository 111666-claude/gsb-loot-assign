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
