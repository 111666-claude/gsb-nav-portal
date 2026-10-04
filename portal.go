// Package portal 是传送门连通图：双向连通、登记幂等、成环拒绝、删除后重新合并成通道。
// 图永远是若干条互不相交的链：增量维护邻接（每个锚点至多两条）与连通分量标号，
// 连通判定 O(1)；合并/拆分只重标号较小的一侧（small-to-large，摊还 O(log n)），
// 任何操作都不做全图遍历。
package portal

import "sort"

// Link 是一条传送门。
type Link struct {
	ID string
	A  string
	B  string
}

type component struct {
	size int
}

// Graph 是传送门图。
type Graph struct {
	links map[string]Link
	adj   map[string]map[string]string
	root  map[string]*component
	comps map[*component]struct{}
	scan  int
}

// New 建一张空的传送门图。
func New() *Graph {
	return &Graph{
		links: map[string]Link{},
		adj:   map[string]map[string]string{},
		root:  map[string]*component{},
		comps: map[*component]struct{}{},
	}
}

// Add 登记一条传送门，返回是否登记成功；成环时第二个返回值是环上的节点。
func (g *Graph) Add(id, a, b string) (bool, []string) {
	if id == "" || a == "" || b == "" || a == b {
		return false, nil
	}
	if _, dup := g.links[id]; dup {
		return false, nil
	}
	if len(g.adj[a]) >= 2 || len(g.adj[b]) >= 2 {
		return false, nil
	}
	ca, okA := g.root[a]
	cb, okB := g.root[b]
	if okA && okB && ca == cb {
		from, to := a, b
		if to < from {
			from, to = to, from
		}
		return false, g.pathBetween(from, to)
	}

	// 先做分量合并（此时新边尚未接入邻接，walkPath 不会串到另一侧）。
	switch {
	case !okA && !okB:
		c := &component{size: 2}
		g.root[a], g.root[b] = c, c
		g.comps[c] = struct{}{}
	case okA && !okB:
		ca.size++
		g.root[b] = ca
	case !okA && okB:
		cb.size++
		g.root[a] = cb
	default:
		large, small, from := ca, cb, b
		if cb.size > ca.size {
			large, small, from = cb, ca, a
		}
		for _, node := range g.walkPath(from) {
			g.root[node] = large
		}
		delete(g.comps, small)
		large.size += small.size
	}

	g.links[id] = Link{ID: id, A: a, B: b}
	if g.adj[a] == nil {
		g.adj[a] = map[string]string{}
	}
	if g.adj[b] == nil {
		g.adj[b] = map[string]string{}
	}
	g.adj[a][b] = id
	g.adj[b][a] = id
	return true, nil
}

// Remove 删除一条传送门，返回它原本是否存在。
func (g *Graph) Remove(id string) bool {
	link, ok := g.links[id]
	if !ok {
		return false
	}
	delete(g.links, id)
	delete(g.adj[link.A], link.B)
	delete(g.adj[link.B], link.A)

	c := g.root[link.A]
	for _, node := range []string{link.A, link.B} {
		if len(g.adj[node]) == 0 {
			delete(g.root, node)
			c.size--
		}
	}
	if c.size == 0 {
		delete(g.comps, c)
		return true
	}

	hasA := len(g.adj[link.A]) > 0
	hasB := len(g.adj[link.B]) > 0
	if !hasA || !hasB {
		return true
	}

	// 两侧交替各走一步，先到头的一侧较小；只重标号这一侧。
	sideA, sideB := []string{link.A}, []string{link.B}
	prevA, curA := link.B, link.A
	prevB, curB := link.A, link.B
	for hasA || hasB {
		if hasA {
			next := g.nextOn(curA, prevA)
			if next == "" {
				hasA = false
			} else {
				sideA = append(sideA, next)
				prevA, curA = curA, next
			}
		}
		if hasB {
			next := g.nextOn(curB, prevB)
			if next == "" {
				hasB = false
			} else {
				sideB = append(sideB, next)
				prevB, curB = curB, next
			}
		}
		if !hasA || !hasB {
			break
		}
	}
	newSide := sideA
	if hasA && !hasB {
		newSide = sideB
	}
	c.size -= len(newSide)
	split := &component{size: len(newSide)}
	g.comps[split] = struct{}{}
	for _, node := range newSide {
		g.root[node] = split
	}
	return true
}

// Count 返回已经登记过的传送门条数。
func (g *Graph) Count() int { return len(g.links) }

// Reachable 判断两个锚点是否连通。
func (g *Graph) Reachable(a, b string) bool {
	if a == b {
		g.scan++
		return true
	}
	g.scan += 2
	ca, okA := g.root[a]
	cb, okB := g.root[b]
	return okA && okB && ca == cb
}

// Chains 返回合并后的通道，每条通道从字典序小的端点开始。
func (g *Graph) Chains() [][]string {
	seed := map[*component]string{}
	for node, c := range g.root {
		seed[c] = node
	}
	out := [][]string{}
	for _, start := range seed {
		path := g.walkPath(g.endOf(start))
		first, last := path[0], path[len(path)-1]
		if last < first {
			reversed := make([]string, len(path))
			for i, node := range path {
				reversed[len(path)-1-i] = node
			}
			path = reversed
		}
		out = append(out, path)
	}
	sort.Slice(out, func(i, j int) bool {
		return out[i][0] < out[j][0]
	})
	return out
}

// Scanned 是累计扫描的索引条目数（规模观测）。
func (g *Graph) Scanned() int { return g.scan }

// walkPath 从 start（链的端点）沿唯一方向走到另一端。
func (g *Graph) walkPath(start string) []string {
	path := []string{start}
	prev, cur := "", start
	for {
		next := g.nextOn(cur, prev)
		if next == "" {
			return path
		}
		path = append(path, next)
		prev, cur = cur, next
	}
}

// endOf 返回 start 所在链上沿任意方向走到的端点。
func (g *Graph) endOf(start string) string {
	prev, cur := "", start
	for {
		next := g.nextOn(cur, prev)
		if next == "" {
			return cur
		}
		prev, cur = cur, next
	}
}

func (g *Graph) nextOn(cur, prev string) string {
	for candidate := range g.adj[cur] {
		if candidate != prev {
			return candidate
		}
	}
	return ""
}

// pathBetween 返回同一分量内从 a 沿已有路径到 b 的节点序列（不含重复起点）。
func (g *Graph) pathBetween(a, b string) []string {
	for first := range g.adj[a] {
		path := []string{a, first}
		prev, cur := a, first
		for cur != b {
			next := g.nextOn(cur, prev)
			if next == "" {
				path = nil
				break
			}
			path = append(path, next)
			prev, cur = cur, next
		}
		if path != nil {
			return path
		}
	}
	return nil
}
