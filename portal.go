// Package portal 是传送门连通图：双向连通、登记幂等、成环拒绝、删除后重新合并成通道。
// 连通判定与登记走增量维护的分量索引，不做全图遍历。
package portal

import "sort"

// Link 是一条传送门。
type Link struct {
	ID string
	A  string
	B  string
}

// Graph 是传送门图。每个锚点最多连两条传送门，因此每个连通分量都是一条链。
type Graph struct {
	links    map[string]Link
	adj      map[string][]string
	comp     map[string]int
	members  map[int]map[string]struct{}
	nextComp int
	scanned  int
}

// New 建一张空的传送门图。
func New() *Graph {
	return &Graph{
		links:   map[string]Link{},
		adj:     map[string][]string{},
		comp:    map[string]int{},
		members: map[int]map[string]struct{}{},
	}
}

// Add 登记一条传送门，返回是否登记成功；成环时第二个返回值是环上的节点。
func (g *Graph) Add(id, a, b string) (bool, []string) {
	if a == "" || b == "" || a == b {
		return false, nil
	}
	if _, dup := g.links[id]; dup {
		return false, nil
	}
	ca, oka := g.comp[a]
	cb, okb := g.comp[b]
	if oka && okb && ca == cb {
		return false, g.pathBetween(a, b)
	}
	if len(g.adj[a]) >= 2 || len(g.adj[b]) >= 2 {
		return false, nil
	}
	g.links[id] = Link{ID: id, A: a, B: b}
	g.adj[a] = append(g.adj[a], b)
	g.adj[b] = append(g.adj[b], a)
	switch {
	case oka && okb:
		if len(g.members[ca]) < len(g.members[cb]) {
			ca, cb = cb, ca
		}
		for node := range g.members[cb] {
			g.comp[node] = ca
			g.members[ca][node] = struct{}{}
		}
		delete(g.members, cb)
	case oka:
		g.comp[b] = ca
		g.members[ca][b] = struct{}{}
	case okb:
		g.comp[a] = cb
		g.members[cb][a] = struct{}{}
	default:
		g.nextComp++
		g.comp[a] = g.nextComp
		g.comp[b] = g.nextComp
		g.members[g.nextComp] = map[string]struct{}{a: {}, b: {}}
	}
	return true, nil
}

// Remove 删除一条传送门，返回它原本是否存在。
func (g *Graph) Remove(id string) bool {
	link, ok := g.links[id]
	if !ok {
		return false
	}
	delete(g.links, id)
	g.adj[link.A] = removeNeighbor(g.adj[link.A], link.B)
	g.adj[link.B] = removeNeighbor(g.adj[link.B], link.A)
	ca := g.comp[link.A]
	if len(g.adj[link.A]) == 0 {
		g.dropNode(link.A)
	}
	if len(g.adj[link.B]) == 0 {
		g.dropNode(link.B)
	}
	if len(g.adj[link.A]) > 0 && len(g.adj[link.B]) > 0 {
		side := g.collectSide(link.A)
		g.nextComp++
		g.members[g.nextComp] = map[string]struct{}{}
		for node := range side {
			delete(g.members[ca], node)
			g.comp[node] = g.nextComp
			g.members[g.nextComp][node] = struct{}{}
		}
	}
	return true
}

// Count 返回已经登记过的传送门条数。
func (g *Graph) Count() int { return len(g.links) }

// Reachable 判断两个锚点是否连通。
func (g *Graph) Reachable(a, b string) bool {
	if a == b {
		return true
	}
	ca, oka := g.comp[a]
	cb, okb := g.comp[b]
	return oka && okb && ca == cb
}

// Chains 返回合并后的通道，每条通道从字典序小的端点开始。
func (g *Graph) Chains() [][]string {
	out := [][]string{}
	for _, set := range g.members {
		endpoints := []string{}
		for node := range set {
			if len(g.adj[node]) == 1 {
				endpoints = append(endpoints, node)
			}
		}
		start := endpoints[0]
		for _, node := range endpoints[1:] {
			if node < start {
				start = node
			}
		}
		out = append(out, g.walkChain(start))
	}
	sort.Slice(out, func(i, j int) bool { return out[i][0] < out[j][0] })
	return out
}

// Scanned 是累计扫描的邻接条目数（规模观测）。
func (g *Graph) Scanned() int { return g.scanned }

// pathBetween 沿已有链从一端走到另一端，给出环上的节点（不含重复的起点）。
func (g *Graph) pathBetween(a, b string) []string {
	start, target := a, b
	if target < start {
		start, target = target, start
	}
	return g.walkChainFrom(start, target)
}

// walkChain 从链的一端走到另一端，列出整条链上的锚点。
func (g *Graph) walkChain(start string) []string {
	return g.walkChainFrom(start, "")
}

// walkChainFrom 从 start 沿链前进，到 target 为止；target 为空则走到链的尽头。
func (g *Graph) walkChainFrom(start, target string) []string {
	path := []string{start}
	prev := ""
	cur := start
	for cur != target {
		next := ""
		for _, neighbor := range g.adj[cur] {
			g.scanned++
			if neighbor != prev {
				next = neighbor
			}
		}
		if next == "" {
			break
		}
		prev, cur = cur, next
		path = append(path, cur)
	}
	return path
}

// collectSide 收集从 start 出发沿链能到达的所有锚点。
func (g *Graph) collectSide(start string) map[string]struct{} {
	side := map[string]struct{}{start: {}}
	for _, node := range g.walkChain(start) {
		side[node] = struct{}{}
	}
	return side
}

// dropNode 把一个不再有任何传送门的锚点从索引里摘掉。
func (g *Graph) dropNode(node string) {
	ca := g.comp[node]
	delete(g.comp, node)
	delete(g.members[ca], node)
	if len(g.members[ca]) == 0 {
		delete(g.members, ca)
	}
}

func removeNeighbor(neighbors []string, target string) []string {
	for index, neighbor := range neighbors {
		if neighbor == target {
			return append(neighbors[:index], neighbors[index+1:]...)
		}
	}
	return neighbors
}
