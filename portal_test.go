package portal

import "testing"

func TestFreshGraphIsEmpty(t *testing.T) {
	g := New()
	if g.Count() != 0 {
		t.Fatal("空图的传送门条数应该是 0")
	}
	if len(g.Chains()) != 0 {
		t.Fatal("空图的通道应该是空的")
	}
}

func TestSameAnchorIsReachable(t *testing.T) {
	if !New().Reachable("a", "a") {
		t.Fatal("同一个锚点应该算连通")
	}
}

func TestAddThenReachableForward(t *testing.T) {
	g := New()
	if ok, cycle := g.Add("l1", "a", "b"); !ok || len(cycle) != 0 {
		t.Fatal("一条普通传送门应该登记成功")
	}
	if !g.Reachable("a", "b") {
		t.Fatal("登记之后 a 到 b 应该连通")
	}
}

func TestCountAndChainsOnSingleLink(t *testing.T) {
	g := New()
	g.Add("l1", "a", "b")
	if g.Count() != 1 {
		t.Fatalf("一条传送门应该算 1 条，得到 %d", g.Count())
	}
	if len(g.Chains()) != 1 {
		t.Fatalf("一条传送门应该是一段通道，得到 %d", len(g.Chains()))
	}
}

func TestScannedStartsAtZero(t *testing.T) {
	if New().Scanned() != 0 {
		t.Fatal("Scanned 初始应该是 0")
	}
}
