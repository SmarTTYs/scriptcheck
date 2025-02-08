package reader

import (
	"fmt"
	"github.com/goccy/go-yaml/ast"
)

type anchorWalker struct {
	anchorNodeMap map[string]ast.Node
	aliasValueMap aliasValueMap
}

func (v *anchorWalker) Visit(node ast.Node) ast.Visitor {
	switch n := node.(type) {
	case *ast.AliasNode:
		aliasName := n.Value.GetToken().Value
		if anchorNode, exists := v.anchorNodeMap[aliasName]; !exists {
			panic(fmt.Sprintf("could not find alias %q", aliasName))
		} else {
			v.aliasValueMap[n] = anchorNode
		}
	case *ast.AnchorNode:
		anchorName := n.Name.GetToken().Value
		v.anchorNodeMap[anchorName] = n.Value
	}

	return v
}
