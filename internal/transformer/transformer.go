package transformer

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/stephen/cssc/internal/ast"
	"github.com/stephen/cssc/internal/logging"
	"github.com/stephen/cssc/internal/sources"
	"github.com/stephen/cssc/transforms"
)

// Options is the set of options for transformation.
type Options struct {
	// OriginalSource is used to report error locations.
	// XXX: work even if original source is not passed by having janky errors.
	OriginalSource *sources.Source

	// Reporter is the reporter for errors and warnings.
	Reporter logging.Reporter

	// Options is the set of transform options. Some transforms may need additional context passed in.
	transforms.Options

	// ImportReplacements is the set of import references to inline. ImportReplacements must be non-nil
	// if ImportRules is set to ImportRulesInline.
	ImportReplacements map[*ast.AtRule]*ast.Stylesheet
}

// Transform takes a pass over the input AST and runs various
// transforms.
func Transform(s *ast.Stylesheet, opts Options) *ast.Stylesheet {
	t := &transformer{
		Options: opts,
	}

	if opts.Reporter == nil {
		t.Reporter = logging.DefaultReporter
	}

	if opts.CustomProperties != transforms.CustomPropertiesPassthrough {
		t.variables = make(map[string][]ast.Value)
	}

	if opts.CustomMediaQueries != transforms.CustomMediaQueriesPassthrough {
		t.customMedia = make(map[string]*ast.MediaQuery)
	}

	if opts.ImportReplacements == nil && opts.ImportRules == transforms.ImportRulesInline {
		t.Reporter.AddError(fmt.Errorf("ImportRules is set to ImportRulesInline, but ImportReplacements is not set"))
	}

	s.Nodes = t.transformNodes(s.Nodes)

	return s
}

// transformer takes a pass over the AST and makes
// modifications to the AST, depending on the settings.
type transformer struct {
	Options

	variables   map[string][]ast.Value
	customMedia map[string]*ast.MediaQuery
}

func (t *transformer) addError(loc ast.Node, fmt string, args ...interface{}) {
	t.Reporter.AddError(logging.LocationErrorf(t.OriginalSource, loc.Location(), fmt, args...))
}

func (t *transformer) addWarn(loc ast.Node, fmt string, args ...interface{}) {
	t.Reporter.AddError(logging.LocationWarnf(t.OriginalSource, loc.Location(), fmt, args...))
}

func (t *transformer) transformSelectors(nodes []*ast.Selector) []*ast.Selector {
	newNodes := make([]*ast.Selector, 0, len(nodes))
	for _, n := range nodes {
		newParts := make([]ast.SelectorPart, 0, len(n.Parts))
		for index, p := range n.Parts {
			switch part := p.(type) {
			case *ast.PseudoClassSelector:
				if part.Name != "any-link" || t.AnyLink == transforms.AnyLinkPassthrough {
					newParts = append(newParts, p)
					break
				}

				// Replace one of them with :link.
				newParts = append(
					newParts,
					&ast.PseudoClassSelector{Name: "link"},
				)

				// Make a duplicate with :visited.
				duplicate := *n
				duplicate.Parts[index] = &ast.PseudoClassSelector{Name: "visited"}
				newNodes = append(newNodes, &duplicate)

			default:
				newParts = append(newParts, p)
			}
		}

		n.Parts = newParts
		newNodes = append(newNodes, n)
	}
	return newNodes
}

func (t *transformer) transformNodes(nodes []ast.Node) []ast.Node {
	rv := make([]ast.Node, 0, len(nodes))
	for _, value := range nodes {
		switch node := value.(type) {
		case *ast.QualifiedRule:
			func() {
				selList, ok := node.Prelude.(*ast.SelectorList)
				if !ok {
					return
				}

				isRoot := false
				for _, sel := range selList.Selectors {
					if len(sel.Parts) != 1 {
						isRoot = true
						break
					}
				}

				if !isRoot {
					return
				}

				rootSel, ok := selList.Selectors[0].Parts[0].(*ast.PseudoClassSelector)
				if !ok {
					return
				}

				if rootSel.Name != "root" {
					return
				}

				declBlock, ok := node.Block.(*ast.DeclarationBlock)
				if !ok {
					return
				}

				newDecls := make([]ast.Declarationish, 0, len(declBlock.Declarations))
				for _, decl := range declBlock.Declarations {
					switch d := decl.(type) {
					case *ast.Declaration:

						if strings.HasPrefix(d.Property, "--") && t.variables != nil {
							t.variables[d.Property] = d.Values
							continue
						}
						newDecls = append(newDecls, d)
					default:
						newDecls = append(newDecls, d)
					}

				}

				declBlock.Declarations = newDecls
			}()

			selList, ok := node.Prelude.(*ast.SelectorList)
			if !ok {
				t.addError(node.Prelude, "expected selector list for qualified rule")
			}
			selList.Selectors = t.transformSelectors(selList.Selectors)
			node.Block = t.transformBlock(node.Block)

			if node.Block == nil {
				continue
			}
			rv = append(rv, node)

		case *ast.AtRule:
			switch node.Name {
			case "import":
				if t.ImportReplacements == nil {
					rv = append(rv, node)
				}

				imported, ok := t.ImportReplacements[node]
				if !ok {
					rv = append(rv, node)
					break
				}

				if len(node.Preludes) > 1 {
					t.addWarn(node, "@import transform does not yet support @supports or media queries")
				}

				rv = append(rv, imported.Nodes...)

			case "custom-media":
				func() {
					if t.customMedia == nil {
						return
					}

					if len(node.Preludes) != 2 {
						return
					}

					name, ok := node.Preludes[0].(*ast.Identifier)
					if !ok {
						return
					}

					query, ok := node.Preludes[1].(*ast.MediaQuery)
					if !ok {
						return
					}

					t.customMedia[name.Value] = query
				}()

			case "media":
				mq := node.Preludes[0].(*ast.MediaQueryList)
				mq.Queries = t.transformMediaQueries(mq.Queries)
				rv = append(rv, node)

			default:
				rv = append(rv, node)
			}

		default:
			rv = append(rv, value)
		}
	}
	return rv
}

func (t *transformer) transformMediaQueries(queries []*ast.MediaQuery) []*ast.MediaQuery {
	newQueries := make([]*ast.MediaQuery, 0, len(queries))
	for _, q := range queries {
		q.Parts = t.transformMediaQueryParts(q.Parts)
		newQueries = append(newQueries, q)
	}
	return newQueries
}

// addToValue takes an ast.Value and adds diff to it.
// XXX: change this out to only operate on ast.Dimension, since there are no other valid number types now.
func (t *transformer) addToValue(v ast.Value, diff float64) ast.Value {
	if diff == 0 {
		return v
	}

	switch oldValue := v.(type) {
	case *ast.Dimension:
		f, err := strconv.ParseFloat(oldValue.Value, 10)
		if err != nil {
			t.addError(oldValue, "could not parse dimension value to lower media range: %s", oldValue.Value)
			return oldValue
		}
		return &ast.Dimension{Value: strconv.FormatFloat(f+diff, 'f', -1, 64), Unit: oldValue.Unit}

	default:
		t.addError(oldValue, "tried to modify non-numeric value. expected dimension, percentage, or number, but got: %s", reflect.TypeOf(v).String())
		return v
	}
}

func (t *transformer) transformMediaFeatureRange(part *ast.MediaFeatureRange) []ast.MediaQueryPart {
	asIs := []ast.MediaQueryPart{part}
	if t.MediaFeatureRanges == transforms.MediaFeatureRangesPassthrough {
		return asIs
	}

	var newParts []ast.MediaQueryPart
	if part.LeftValue != nil {
		var direction string
		var diff float64
		switch part.Operator {
		case ">":
			diff = -.001
			fallthrough
		case ">=":
			direction = "max"
		case "<":
			diff = .001
			fallthrough
		case "<=":
			direction = "min"
		}

		newParts = append(newParts, &ast.MediaFeaturePlain{
			// XXX: replace this allocation with a lookup.
			Property: &ast.Identifier{Value: fmt.Sprintf("%s-%s", direction, part.Property.Value)},
			Value:    t.addToValue(part.LeftValue, diff),
		})
	}

	if part.RightValue != nil {
		if part.LeftValue != nil {
			newParts = append(newParts, &ast.Identifier{Value: "and"})
		}

		var direction string
		var diff float64
		switch part.Operator {
		case ">":
			diff = .001
			fallthrough
		case ">=":
			direction = "min"
		case "<":
			diff = -.001
			fallthrough
		case "<=":
			direction = "max"
		}

		newParts = append(newParts, &ast.MediaFeaturePlain{
			// XXX: replace this allocation with a lookup.
			Property: &ast.Identifier{Value: fmt.Sprintf("%s-%s", direction, part.Property.Value)},
			Value:    t.addToValue(part.RightValue, diff),
		})
	}

	return newParts
}

func (t *transformer) transformMediaQueryParts(parts []ast.MediaQueryPart) []ast.MediaQueryPart {
	newParts := make([]ast.MediaQueryPart, 0, len(parts))
	for _, p := range parts {
		switch part := p.(type) {
		case *ast.MediaFeaturePlain:
			if part.Value != nil || !strings.HasPrefix(part.Property.Value, "--") {
				newParts = append(newParts, p)
				break
			}

			if t.customMedia == nil {
				newParts = append(newParts, p)
				break
			}

			replacement, ok := t.customMedia[part.Property.Value]
			if !ok {
				newParts = append(newParts, p)
				break
			}

			newParts = append(newParts, replacement.Parts...)

		case *ast.MediaFeatureRange:
			newParts = append(newParts, t.transformMediaFeatureRange(part)...)

		default:
			newParts = append(newParts, p)
		}
	}
	return newParts
}

func (t *transformer) transformBlock(block ast.Block) ast.Block {
	switch node := block.(type) {
	case *ast.QualifiedRuleBlock:
		// 	for _, d := range node.Rules {
		// 		// t.transform(d)
		// 	}
		if len(node.Rules) == 0 {
			return nil
		}

	case *ast.DeclarationBlock:
		node.Declarations = t.transformDeclarations(node.Declarations)
		if len(node.Declarations) == 0 {
			return nil
		}

	default:
		panic("unknown block")
	}

	return block
}

func (t *transformer) transformDeclarations(decls []ast.Declarationish) []ast.Declarationish {
	newDecls := make([]ast.Declarationish, 0, len(decls))
	for _, decl := range decls {
		switch d := decl.(type) {
		case *ast.Declaration:
			d.Values = t.transformValues(d.Values)
			newDecls = append(newDecls, d)
		default:
			newDecls = append(newDecls, d)
		}
	}

	return newDecls
}

func (t *transformer) transformValues(values []ast.Value) []ast.Value {
	rv := make([]ast.Value, 0, len(values))
	for _, value := range values {
		switch v := value.(type) {
		case *ast.Function:
			newValues := []ast.Value{v}
			func() {
				if v.Name != "var" {
					return
				}

				if t.variables == nil {
					return
				}

				if len(v.Arguments) == 0 {
					t.addError(v, "expected at least one argument to var()")
					return
				}

				varName, ok := v.Arguments[0].(*ast.Identifier)
				if !ok {
					t.addError(v, "expected identifier as argument to var()")
					return
				}

				vals, ok := t.variables[varName.Value]
				if !ok {
					// The first argument is the value, the second is a comma.
					if len(v.Arguments) > 2 {
						newValues = v.Arguments[2:]
						return
					}

					t.addWarn(v, "use of undefined variable without fallback: %s", varName.Value)
					return
				}

				newValues = vals
			}()

			func() {
				if v.Name != "calc" {
					return
				}

				if t.CalcReduction == transforms.CalcReductionPassthrough {
					return
				}

				if len(v.Arguments) != 1 {
					t.addWarn(v, "expected single argument for calc()")
					return
				}

				args := t.transformValues([]ast.Value{v.Arguments[0]})
				if len(args) != 1 {
					t.addWarn(v, "expected single argument for calc()")
					return
				}

				arg, ok := args[0].(*ast.MathExpression)
				if !ok {
					return
				}

				l, r := t.transformValues([]ast.Value{arg.Left}), t.transformValues([]ast.Value{arg.Right})
				if len(l) != 1 {
					t.addWarn(arg.Left, "expected left-hand side of math expression to be a single value")
					return
				}
				if len(r) != 1 {
					t.addWarn(arg.Right, "expected right-hand side of math expression to be a single value")
					return
				}

				newValue := t.evaluateMathExpression(l[0], r[0], arg.Operator)
				if newValue == nil {
					return
				}

				newValues = []ast.Value{newValue}
			}()

			rv = append(rv, newValues...)

		default:
			rv = append(rv, v)
		}
	}

	return rv
}

func (t *transformer) doMath(left, right, op string) (float64, error) {
	leftValue, err := strconv.ParseFloat(left, 10)
	if err != nil {
		return 0, fmt.Errorf("could not parse dimension value: %s", left)
	}

	rightValue, err := strconv.ParseFloat(right, 10)
	if err != nil {
		return 0, fmt.Errorf("could not parse dimension value: %s", right)
	}
	switch op {
	case "+":
		return leftValue + rightValue, nil
	case "-":
		return leftValue - rightValue, nil
	case "*":
		return leftValue * rightValue, nil
	case "/":
		if rightValue == 0 {
			return 0, fmt.Errorf("cannot divide by zero")
		}
		return leftValue / rightValue, nil
	default:
		return 0, fmt.Errorf("unknown op: %s", op)
	}
}

// evaluateMathExpression attempts to add l + r. If sub is true, subtraction will be applied
// instead. For sum to succeed, both l and r must be of the same type.
// See notes from https://www.w3.org/TR/css-values-3/#calc-type-checking.
func (t *transformer) evaluateMathExpression(l, r ast.Value, op string) ast.Value {
	if expr, ok := l.(*ast.MathExpression); ok {
		if evaluated := t.evaluateMathExpression(expr.Left, expr.Right, expr.Operator); evaluated != nil {
			l = evaluated
		}
	}

	if expr, ok := r.(*ast.MathExpression); ok {
		if evaluated := t.evaluateMathExpression(expr.Left, expr.Right, expr.Operator); evaluated != nil {
			r = evaluated
		}
	}

	switch op {
	case "+", "-":
		switch left := l.(type) {
		case *ast.Dimension:
			right, ok := r.(*ast.Dimension)
			if !ok {
				return nil
			}

			if left.Unit != right.Unit {
				if left.Unit == "" && right.Unit != "" {
					// Invalid, because we cannot mix number types and lengths, e.g. (2 + 5rem).
					t.addError(left, "cannot add number type and %s type together", right.Unit)
				}

				if left.Unit != "" && right.Unit == "" {
					// Invalid, because we cannot mix number types and lengths, e.g. (5rem + 2).
					t.addError(left, "cannot add number type and %s type together", left.Unit)
				}

				// Valid css, but we cannot reduce (e.g. 2px + 3rem).
				return nil
			}

			newValue, err := t.doMath(left.Value, right.Value, op)
			if err != nil {
				t.addError(l, err.Error())
				return nil
			}

			return &ast.Dimension{
				Value: strconv.FormatFloat(newValue, 'f', -1, 64),
				Unit:  left.Unit,
			}

		// case *ast.MathExpression:
		// if left.op is not + or -, return nil

		// We know left.Right is an ast.Dimension because of the structure of the parse tree
		// We know Right is an ast.Dimension
		// If left.Right.Unit != right.Unit and left.Left.Unit != right.Unit, return nil
		// left.Right.Unit case: doMath on left.Right.Value and right.Value and return left with Right = [the result of doMath]
		// left.Left.Unit case: doMath on left.Left.Value and right.Value and return left with Left = [the result of doMath]

		default:
			t.addError(l, "cannot perform %s on this type", op)
			return nil
		}

	case "*":
		leftAsDimension, leftIsDimension := l.(*ast.Dimension)
		rightAsDimension, rightIsDimension := r.(*ast.Dimension)
		if !leftIsDimension || !rightIsDimension {
			return nil
		}

		if leftAsDimension.Unit != "" && rightAsDimension.Unit != "" {
			t.addError(l, "one side of multiplication must be a number (non-percentage/dimension)")
			return nil
		}

		maybeWithUnit, number := leftAsDimension, rightAsDimension
		if leftAsDimension.Unit == "" {
			maybeWithUnit, number = rightAsDimension, leftAsDimension
		}

		// XXX: handle cases like calc(var(--x) * 2 * 2)

		newValue, err := t.doMath(maybeWithUnit.Value, number.Value, op)
		if err != nil {
			t.addError(l, err.Error())
			return nil
		}

		return &ast.Dimension{
			Value: strconv.FormatFloat(newValue, 'f', -1, 64),
			Unit:  maybeWithUnit.Unit,
		}

	case "/":
		rightAsDimension, rightIsDimension := r.(*ast.Dimension)
		if !rightIsDimension || rightAsDimension.Unit != "" {
			t.addError(l, "right side of division must be a number (non-percentage/dimension)")
			return nil
		}

		switch left := l.(type) {
		case *ast.Dimension:
			newValue, err := t.doMath(left.Value, rightAsDimension.Value, op)
			if err != nil {
				t.addError(l, err.Error())
				return nil
			}

			return &ast.Dimension{
				Value: strconv.FormatFloat(newValue, 'f', -1, 64),
				Unit:  left.Unit,
			}

			// case *ast.MathExpression

		default:
			t.addError(l, "cannot perform %s on this type", op)
			return nil
		}

	default:
		t.addError(l, "unknown op: %s", op)
		return nil
	}
}
