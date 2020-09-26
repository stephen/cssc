package ast

import (
	"fmt"
	"reflect"
)

// Walk walks the AST starting from the input node.
func Walk(start Node, visit func(n Node)) {
	if start == nil {
		return
	}

	visit(start)
	switch s := start.(type) {
	case *Stylesheet:
		for _, n := range s.Nodes {
			Walk(n, visit)
		}

	case *QualifiedRule:
		Walk(s.Prelude, visit)
		Walk(s.Block, visit)

	case *SelectorList:
		for _, sel := range s.Selectors {
			Walk(sel, visit)
		}

	case *Selector:
		for _, part := range s.Parts {
			Walk(part, visit)
		}

	case *AtRule:
		for _, p := range s.Preludes {
			Walk(p, visit)
		}
		Walk(s.Block, visit)

	case *MediaQueryList:
		for _, mq := range s.Queries {
			Walk(mq, visit)
		}

	case *MediaQuery:
		for _, part := range s.Parts {
			Walk(part, visit)
		}

	case *MediaFeaturePlain:
		Walk(s.Property, visit)
		Walk(s.Value, visit)

	case *QualifiedRuleBlock:
		for _, r := range s.Rules {
			Walk(r, visit)
		}

	case *Declaration:
		for _, v := range s.Values {
			Walk(v, visit)
		}

	case *DeclarationBlock:
		for _, decl := range s.Declarations {
			Walk(decl, visit)
		}

	case *AttributeSelector:
		Walk(s.Value, visit)

	case *MediaFeatureRange:
		Walk(s.LeftValue, visit)
		Walk(s.RightValue, visit)

	case *MathExpression:
		Walk(s.Left, visit)
		Walk(s.Right, visit)

	case *KeyframeSelectorList:
		for _, k := range s.Selectors {
			Walk(k, visit)
		}

	case *Function:
		for _, arg := range s.Arguments {
			Walk(arg, visit)
		}

	case *ClassSelector:
	case *Comma:
	case *Comment:
	case *IDSelector:
	case *String:
	case *TypeSelector:
	case *CombinatorSelector:
	case *PseudoClassSelector:
	case *ANPlusB:
	case *PseudoElementSelector:
	case *HexColor:
	case *Percentage:
	case *Dimension:
	case *Whitespace:
	case *Identifier:

	default:
		panic(fmt.Errorf("unknown node type: %s", reflect.TypeOf(s).String()))
	}
}
