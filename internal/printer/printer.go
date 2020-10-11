package printer

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"reflect"
	"strings"

	"github.com/stephen/cssc/internal/ast"
	"github.com/stephen/cssc/internal/sources"
)

type printer struct {
	options Options
	s       strings.Builder

	sourceMappings   strings.Builder
	lastWritten      int
	lastMappingState mappingState
}

type mappingState struct {
	generatedColumn int32
	originalLine    int32
	originalColumn  int32
}

// Options is a set of options for printing.
type Options struct {
	OriginalSource *sources.Source
}

// Print prints the input AST node into CSS. It should have deterministic
// output.
func Print(in ast.Node, opts Options) (output string, err error) {
	defer func() {
		if rErr := recover(); rErr != nil {
			if errI, ok := rErr.(error); ok {
				output, err = "", errI
				return
			}

			// Re-panic unknown issues.
			panic(err)
		}
	}()

	p := printer{
		options: opts,
	}

	p.print(in)
	p.printMapping()

	return p.s.String(), nil
}

func (p *printer) printMapping() {
	if p.options.OriginalSource == nil {
		return
	}

	b := strings.Builder{}
	b.WriteString(`{"version": 3,"file":"`)
	b.WriteString(p.options.OriginalSource.Path)
	// XXX: sources content and names.
	b.WriteString(`","sourceRoot":"", "sources": ["source.css"], "sourcesContent":[`)
	out, err := json.Marshal(p.options.OriginalSource.Content)
	if err != nil {
		panic(err)
	}
	b.Write(out)
	b.WriteString(`],"names":[],"mappings":"`)
	b.WriteString(p.sourceMappings.String())
	b.WriteString(`"}`)
	p.s.WriteString("\n/*# sourceMappingURL=data:application/json;base64,")
	// XXX: allocation.
	p.s.WriteString(base64.StdEncoding.EncodeToString([]byte(b.String())))
	p.s.WriteString(" */\n")
}

// addMapping should be called from the printer
// when a new symbol needs to be added to the sourcemap.
func (p *printer) addMapping(loc ast.Span) {
	if p.options.OriginalSource == nil {
		return
	}

	newState := p.lastMappingState

	line, col := p.options.OriginalSource.LineAndCol(loc)
	newState.originalLine, newState.originalColumn = line-1, col-1

	// Note that String() here does not reallocate the string.
	for _, ch := range p.s.String()[p.lastWritten:] {
		if ch == '\n' {
			p.lastMappingState.generatedColumn = 0
			newState.generatedColumn = 0
			p.sourceMappings.WriteRune(';')
			continue
		}

		newState.generatedColumn++
	}

	if p.s.Len() > 0 {
		lastByte := p.sourceMappings.String()[p.sourceMappings.Len()-1]
		if lastByte != ';' {
			p.sourceMappings.WriteRune(',')
		}
	}

	p.sourceMappings.Write(VLQEncode(newState.generatedColumn - p.lastMappingState.generatedColumn))
	// XXX: figure out what to do for multiple sources.
	p.sourceMappings.Write(VLQEncode(0))

	p.sourceMappings.Write(VLQEncode(newState.originalLine - p.lastMappingState.originalLine))
	p.sourceMappings.Write(VLQEncode(newState.originalColumn - p.lastMappingState.originalColumn))
	// XXX: 5th item for "names" mapping

	p.lastMappingState = newState
	p.lastWritten = p.s.Len()
}

// print prints the current ast node to the printer output.
func (p *printer) print(in ast.Node) {
	switch node := in.(type) {
	case *ast.Stylesheet:
		for _, n := range node.Nodes {
			p.print(n)
		}

	case *ast.AtRule:
		p.addMapping(node.Span)
		p.s.WriteRune('@')
		p.s.WriteString(node.Name)
		if len(node.Preludes) > 0 {
			p.s.WriteRune(' ')
			for i, prelude := range node.Preludes {
				p.print(prelude)

				if i+1 < len(node.Preludes) {
					p.s.WriteRune(' ')
				}
			}
		}

		if node.Block != nil {
			p.s.WriteRune('{')
			p.print(node.Block)
			p.s.WriteRune('}')
		} else {
			p.s.WriteRune(';')
		}

	case *ast.SelectorList:
		for i, s := range node.Selectors {
			p.print(s)

			if i+1 < len(node.Selectors) {
				p.s.WriteRune(',')
			}
		}

	case *ast.KeyframeSelectorList:
		for _, s := range node.Selectors {
			p.print(s)
		}

	case *ast.QualifiedRule:
		p.addMapping(node.Location())
		p.print(node.Prelude)

		p.s.WriteRune('{')
		p.print(node.Block)
		p.s.WriteRune('}')

	case *ast.QualifiedRuleBlock:
		for _, r := range node.Rules {
			p.print(r)
		}

	case *ast.DeclarationBlock:
		for i, d := range node.Declarations {
			p.print(d)

			if i+1 < len(node.Declarations) {
				p.s.WriteRune(';')
			}
		}

	case *ast.Declaration:
		p.s.WriteString(node.Property)
		p.s.WriteRune(':')
		for i, val := range node.Values {
			p.print(val)

			// Print space if we're not the last value and the previous or current
			// value was not a comma.
			if i+1 < len(node.Values) {
				if _, nextIsComma := node.Values[i+1].(*ast.Comma); !nextIsComma {
					if _, isComma := val.(*ast.Comma); !isComma {
						p.s.WriteRune(' ')
					}
				}
			}
		}

		if node.Important {
			p.s.WriteString("!important")
		}

	case *ast.Comma:
		p.s.WriteRune(',')

	case *ast.Dimension:
		p.s.WriteString(node.Value)
		p.s.WriteString(node.Unit)

	case *ast.Percentage:
		p.s.WriteString(node.Value)
		p.s.WriteRune('%')

	case *ast.String:
		p.s.WriteRune('"')
		p.s.WriteString(node.Value)
		p.s.WriteRune('"')

	case *ast.Identifier:
		p.s.WriteString(node.Value)

	case *ast.Function:
		p.s.WriteString(node.Name)
		p.s.WriteRune('(')
		for _, arg := range node.Arguments {
			p.print(arg)
		}
		p.s.WriteRune(')')

	case *ast.MathExpression:
		p.print(node.Left)
		p.s.WriteString(node.Operator)
		p.print(node.Right)

	case *ast.Whitespace:
		p.s.WriteRune(' ')

	case *ast.Selector:
		for i, part := range node.Parts {
			if _, isWhitespace := part.(*ast.Whitespace); i+1 >= len(node.Parts) && isWhitespace {
				continue
			}

			p.print(part)
		}

	case *ast.AttributeSelector:
		p.s.WriteRune('[')
		p.s.WriteString(node.Property)
		if node.Value != nil {
			p.s.WriteRune('=')
			p.print(node.Value)
		}
		p.s.WriteRune(']')

	case *ast.TypeSelector:
		p.s.WriteString(node.Name)

	case *ast.ClassSelector:
		p.s.WriteRune('.')
		p.s.WriteString(node.Name)

	case *ast.IDSelector:
		p.s.WriteRune('#')
		p.s.WriteString(node.Name)

	case *ast.CombinatorSelector:
		p.s.WriteString(node.Operator)

	case *ast.PseudoElementSelector:
		p.s.WriteRune(':')
		p.print(node.Inner)

	case *ast.HexColor:
		p.s.WriteRune('#')
		p.s.WriteString(node.RGBA)

	case *ast.PseudoClassSelector:
		p.s.WriteRune(':')
		p.s.WriteString(node.Name)
		if node.Arguments != nil {
			p.s.WriteRune('(')
			p.print(node.Arguments)
			p.s.WriteRune(')')
		}

	case *ast.ANPlusB:
		if node.A != "" {
			p.s.WriteString(node.A)
		}
		p.s.WriteRune('n')
		if node.B != "" {
			p.s.WriteRune('+')
			p.s.WriteString(node.B)
		}

	case *ast.MediaQueryList:
		for i, q := range node.Queries {
			p.print(q)

			if i+1 < len(node.Queries) {
				p.s.WriteRune(',')
			}
		}

	case *ast.MediaQuery:
		for i, part := range node.Parts {
			p.print(part)

			if i+1 < len(node.Parts) {
				p.s.WriteRune(' ')
			}
		}

	case *ast.MediaFeaturePlain:
		p.s.WriteRune('(')
		p.print(node.Property)
		if node.Value != nil {
			p.s.WriteRune(':')
			p.print(node.Value)
		}
		p.s.WriteRune(')')

	case *ast.MediaFeatureRange:
		p.s.WriteRune('(')
		if node.LeftValue != nil {
			p.print(node.LeftValue)
			p.s.WriteString(node.Operator)
		}
		p.print(node.Property)
		if node.RightValue != nil {
			p.s.WriteString(node.Operator)
			p.print(node.RightValue)
		}
		p.s.WriteRune(')')

	default:
		panic(fmt.Sprintf("unknown ast node: %s", reflect.TypeOf(in).String()))
	}

}
