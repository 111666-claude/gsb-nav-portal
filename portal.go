// Package portal 是传送门连通图：双向连通、登记幂等、成环拒绝、删除后重新合并成通道。
// 缺陷：只登记单向、重复 id 照收、成环不拒、删除不清邻接、通道不合并相邻段、
// 连通判定每次全图遍历。
package portal

// Link 是一条传送门。
type Link struct {
	ID string
	A  string
	B  string
}

// Graph 是传送门图。
type Graph struct {
	links   map[string]Link
	order   []string
	adj     map[string][]string
	scanned int
}

// New 建一张空的传送门图。
func New() *Graph {
	return &Graph{links: map[string]Link{}, adj: map[string][]string{}}
}

// Add 登记一条传送门，返回是否登记成功；成环时第二个返回值是环上的节点。
func (g *Graph) Add(id, a, b string) (bool, []string) {
	g.adj[a] = append(g.adj[a], b)
	g.links[id] = Link{ID: id, A: a, B: b}
	g.order = append(g.order, id)
	return true, nil
}

// Remove 删除一条传送门，返回它原本是否存在。
func (g *Graph) Remove(id string) bool {
	if _, ok := g.links[id]; !ok {
		return false
	}
	delete(g.links, id)
	return true
}

// Count 返回已经登记过的传送门条数。
func (g *Graph) Count() int { return len(g.order) }

// Reachable 判断两个锚点是否连通。
func (g *Graph) Reachable(a, b string) bool {
	if a == b {
		return true
	}
	seen := map[string]bool{a: true}
	stack := []string{a}
	for len(stack) > 0 {
		node := stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		for _, next := range g.adj[node] {
			g.scanned++
			if next == b {
				return true
			}
			if !seen[next] {
				seen[next] = true
				stack = append(stack, next)
			}
		}
	}
	return false
}

// Chains 返回合并后的通道，每条通道从字典序小的端点开始。
func (g *Graph) Chains() [][]string {
	out := [][]string{}
	for _, id := range g.order {
		link := g.links[id]
		out = append(out, []string{link.A, link.B})
	}
	return out
}

// Scanned 是累计扫描的邻接条目数（规模观测）。
func (g *Graph) Scanned() int { return g.scanned }
