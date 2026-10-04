// Command nav-portal 跑传送门连通图样例。
package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	"example.com/portal"
)

// Run 执行一次命令行调用，返回退出码。
func Run(args []string, stdout, stderr io.Writer) int {
	flags := flag.NewFlagSet("nav-portal", flag.ContinueOnError)
	flags.SetOutput(stderr)
	sample := flags.String("sample", "bidir", "bidir / dup / cycle / remove / merge / scan")
	if err := flags.Parse(args); err != nil {
		return 2
	}
	if *sample == "bidir" {
		g := portal.New()
		g.Add("l1", "a", "b")
		fmt.Fprintf(stdout, "ba=%v\n", g.Reachable("b", "a"))
		return 0
	}
	if *sample == "dup" {
		g := portal.New()
		g.Add("l1", "a", "b")
		g.Add("l1", "a", "b")
		fmt.Fprintf(stdout, "count=%d segments=%d\n", g.Count(), len(g.Chains()))
		return 0
	}
	if *sample == "cycle" {
		g := portal.New()
		g.Add("l1", "a", "b")
		g.Add("l2", "b", "c")
		ok, cycle := g.Add("l3", "c", "a")
		fmt.Fprintf(stdout, "ok=%v cycle=%s\n", ok, strings.Join(cycle, ","))
		return 0
	}
	if *sample == "remove" {
		g := portal.New()
		g.Add("l1", "a", "b")
		g.Remove("l1")
		fmt.Fprintf(stdout, "reachable=%v\n", g.Reachable("a", "b"))
		return 0
	}
	if *sample == "merge" {
		g := portal.New()
		g.Add("l1", "a", "b")
		g.Add("l2", "b", "c")
		chains := g.Chains()
		first := ""
		if len(chains) > 0 {
			first = strings.Join(chains[0], ",")
		}
		fmt.Fprintf(stdout, "segments=%d first=%s\n", len(chains), first)
		return 0
	}
	if *sample == "scan" {
		g := portal.New()
		for index := 0; index < 3000; index++ {
			g.Add(fmt.Sprintf("l-%d", index), fmt.Sprintf("n%d", index), fmt.Sprintf("n%d", index+1))
		}
		for index := 0; index < 3000; index++ {
			g.Reachable("n0", "n3000")
		}
		fmt.Fprintf(stdout, "scanned=%d\n", g.Scanned())
		return 0
	}
	fmt.Fprintln(stderr, "需要 --sample bidir|dup|cycle|remove|merge|scan")
	return 2
}

func main() {
	os.Exit(Run(os.Args[1:], os.Stdout, os.Stderr))
}
