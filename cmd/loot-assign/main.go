// Command loot-assign 跑掉落归属样例。
package main

import (
	"flag"
	"fmt"
	"io"
	"os"

	"example.com/lootassign"
)

func items(count, tier int) []lootassign.Item {
	out := []lootassign.Item{}
	for index := 0; index < count; index++ {
		out = append(out, lootassign.Item{ID: fmt.Sprintf("i-%02d", index), Tier: tier})
	}
	return out
}

func firstMember(awards []lootassign.Award) string {
	if len(awards) == 0 {
		return "-"
	}
	return awards[0].Member
}

// Run 执行一次命令行调用，返回退出码。
func Run(args []string, stdout, stderr io.Writer) int {
	flags := flag.NewFlagSet("loot-assign", flag.ContinueOnError)
	flags.SetOutput(stderr)
	sample := flags.String("sample", "share", "share / seed / tier / idem / scan")
	if err := flags.Parse(args); err != nil {
		return 2
	}
	team := []lootassign.Member{{Name: "a", Weight: 1}, {Name: "b", Weight: 1}, {Name: "c", Weight: 1}}
	if *sample == "share" {
		awards, _ := lootassign.New().Assign("d1", 0, team, items(11, 1))
		fmt.Fprintf(stdout, "awarded=%d\n", len(awards))
		return 0
	}
	if *sample == "seed" {
		first0 := firstMember(mustAssign("d0", 0, team, items(6, 1)))
		first1 := firstMember(mustAssign("d1", 1, team, items(6, 1)))
		fmt.Fprintf(stdout, "first0=%s first1=%s\n", first0, first1)
		return 0
	}
	if *sample == "tier" {
		solo := []lootassign.Member{{Name: "a", Weight: 1}}
		pool := []lootassign.Item{{ID: "i1", Tier: 1}, {ID: "i2", Tier: 3}}
		awards, _ := lootassign.New().Assign("d1", 0, solo, pool)
		first := "-"
		if len(awards) > 0 {
			first = awards[0].Item
		}
		fmt.Fprintf(stdout, "first=%s\n", first)
		return 0
	}
	if *sample == "idem" {
		assigner := lootassign.New()
		assigner.Assign("d1", 0, team, items(2, 1))
		awards, second := assigner.Assign("d1", 0, team, items(4, 1))
		fmt.Fprintf(stdout, "second=%v count=%d\n", second, len(awards))
		return 0
	}
	if *sample == "scan" {
		assigner := lootassign.New()
		members := []lootassign.Member{}
		for index := 0; index < 300; index++ {
			members = append(members, lootassign.Member{Name: fmt.Sprintf("m-%03d", index), Weight: 1})
		}
		assigner.Assign("d1", 0, members, items(3000, 1))
		fmt.Fprintf(stdout, "scanned=%d\n", assigner.Scanned())
		return 0
	}
	fmt.Fprintln(stderr, "需要 --sample share|seed|tier|idem|scan")
	return 2
}

func mustAssign(dropID string, seed int64, members []lootassign.Member, pool []lootassign.Item) []lootassign.Award {
	awards, _ := lootassign.New().Assign(dropID, seed, members, pool)
	return awards
}

func main() {
	os.Exit(Run(os.Args[1:], os.Stdout, os.Stderr))
}
