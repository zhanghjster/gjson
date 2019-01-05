package gjson

type PathNode struct {
	FullPath string

	name  string
	sub   int
	check bool

	parent *PathNode
	child  map[string]*PathNode
}

func (n *PathNode) Get(path []string) *PathNode {
	if n.child == nil || len(path) == 0 {
		return nil
	}

	p0 := path[0]
	if child, ok := n.child[p0]; ok {
		if len(path) == 1 {
			return child
		}

		return child.Get(path[1:])
	}

	return nil
}

func (n *PathNode) HasName(name string) bool {
	_, ok := n.child[name]
	return ok
}

func (n *PathNode) Has(path []string) bool {
	if n.child == nil || len(path) == 0 {
		return false
	}

	if child, ok := n.child[path[0]]; !ok {
		return false
	} else if len(path) == 1 {
		return child.check
	} else {
		return child.Has(path[1:])
	}
}

func (n *PathNode) Add(path []string, full string) bool {
	if len(path) == 0 {
		n.check = true
		n.FullPath = full
		return true
	}

	var p0 = path[0]
	if n.child == nil {
		n.child = make(map[string]*PathNode)
	}

	if _, ok := n.child[p0]; !ok {
		n.child[p0] = &PathNode{parent: n, name: p0}
	} else if len(path) == 1 { // reach leaf
		return false
	}

	if suc := n.child[p0].Add(path[1:], full); suc {
		n.sub++
		return true
	}

	return false
}
