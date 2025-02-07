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
		if _, exists := v.aliasValueMap[n]; exists {
			break
		}

		aliasName := n.Value.GetToken().Value
		if node, exists := v.anchorNodeMap[aliasName]; !exists {
			panic(fmt.Sprintf("could not find alias %q", aliasName))
		} else {
			// once the correct alias value is obtained, overwrite with that value.
			v.aliasValueMap[n] = node
		}
	case *ast.AnchorNode:
		anchorName := n.Name.GetToken().Value
		v.anchorNodeMap[anchorName] = n.Value
	}

	return v
}
