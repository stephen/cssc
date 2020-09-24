package parser

import (
	"github.com/stephen/cssc/internal/ast"
	"github.com/stephen/cssc/internal/lexer"
	"github.com/stephen/cssc/internal/logging"
	"github.com/stephen/cssc/internal/sources"
)

// Parse parses an input stylesheet.
func Parse(source *sources.Source) (ss *ast.Stylesheet, err error) {
	p := newParser(source)
	defer func() {
		if rErr := recover(); rErr != nil {
			if errI, ok := rErr.(*lexer.Error); ok {
				ss, err = nil, errI
				return
			}

			if errI, ok := rErr.(error); ok {
				start, end := p.lexer.Range()
				panic(logging.LocationErrorf(source, start, end, "%v", errI))
			}

			// Re-panic unknown issues.
			panic(err)
		}
	}()

	p.parse()
	return p.ss, nil
}

func newParser(source *sources.Source) *parser {
	return &parser{
		lexer: lexer.NewLexer(source),
		ss:    &ast.Stylesheet{},
	}
}

type parser struct {
	lexer *lexer.Lexer
	ss    *ast.Stylesheet
}

func (p *parser) parse() {
	for p.lexer.Current != lexer.EOF {
		switch p.lexer.Current {
		case lexer.At:
			p.parseAtRule()

		case lexer.Semicolon:
			p.lexer.Next()

		case lexer.CDO, lexer.CDC:
			// From https://www.w3.org/TR/css-syntax-3/#parser-entry-points,
			// we'll always assume we're parsing from the top-level, so we can discard CDO/CDC.
			p.lexer.Next()

		case lexer.Comment:
			p.ss.Nodes = append(p.ss.Nodes, &ast.Comment{
				Span: p.lexer.Location(),
				Text: p.lexer.CurrentString,
			})
			p.lexer.Next()

		default:
			p.ss.Nodes = append(p.ss.Nodes, p.parseQualifiedRule(false))
		}

	}
}

func isImportantString(in string) bool {
	return len(in) == 9 &&
		(in[0] == 'i' || in[0] == 'I') &&
		(in[1] == 'm' || in[1] == 'M') &&
		(in[2] == 'p' || in[2] == 'P') &&
		(in[3] == 'o' || in[3] == 'O') &&
		(in[4] == 'r' || in[4] == 'R') &&
		(in[5] == 't' || in[5] == 'T') &&
		(in[6] == 'a' || in[6] == 'A') &&
		(in[7] == 'n' || in[7] == 'N') &&
		(in[8] == 't' || in[8] == 'T')
}

// parseQualifiedRule parses a rule. If isKeyframes is set, the parser will assume
// all preludes are keyframes percentage selectors. Otherwise, it will assume
// the preludes are selector lists.
func (p *parser) parseQualifiedRule(isKeyframes bool) *ast.QualifiedRule {
	r := &ast.QualifiedRule{
		Span: p.lexer.Location(),
	}

	for {
		switch p.lexer.Current {
		case lexer.EOF:
			p.lexer.Errorf("unexpected EOF")

		case lexer.LCurly:
			block := &ast.DeclarationBlock{
				Span: p.lexer.Location(),
			}

			r.Block = block
			p.lexer.Next()

			for p.lexer.Current != lexer.RCurly {
				decl := &ast.Declaration{
					Span:     p.lexer.Location(),
					Property: p.lexer.CurrentString,
				}
				p.lexer.Expect(lexer.Ident)
				p.lexer.Expect(lexer.Colon)
			values:
				for {
					switch p.lexer.Current {
					case lexer.EOF:
						p.lexer.Errorf("unexpected EOF")

					case lexer.Delim:
						if p.lexer.CurrentString != "!" {
							p.lexer.Errorf("unexpected token: %s", p.lexer.CurrentString)
						}
						p.lexer.Next()

						if !isImportantString(p.lexer.CurrentString) {
							p.lexer.Errorf("expected !important, unexpected token: %s", p.lexer.CurrentString)
						}
						p.lexer.Next()
						decl.Important = true

					case lexer.Comma:
						decl.Values = append(decl.Values, &ast.Comma{Span: p.lexer.Location()})
						p.lexer.Next()

					default:
						val := p.parseValue()
						if val == nil {
							if len(decl.Values) == 0 {
								p.lexer.Errorf("declaration must have a value")
							}
							block.Declarations = append(block.Declarations, decl)

							break values
						}

						decl.Values = append(decl.Values, val)
					}
				}

				if p.lexer.Current == lexer.Semicolon {
					p.lexer.Next()
				}
			}
			p.lexer.Next()
			return r

		default:
			if isKeyframes {
				r.Prelude = p.parseKeyframeSelectorList()
				continue
			}

			r.Prelude = p.parseSelectorList()
		}
	}
}

func (p *parser) parseKeyframeSelectorList() *ast.KeyframeSelectorList {
	l := &ast.KeyframeSelectorList{
		Span: p.lexer.Location(),
	}

	for {
		if p.lexer.Current == lexer.EOF {
			p.lexer.Errorf("unexpected EOF")
		}

		switch p.lexer.Current {
		case lexer.Percentage:
			l.Selectors = append(l.Selectors, &ast.Percentage{
				Span:  p.lexer.Location(),
				Value: p.lexer.CurrentNumeral,
			})

		case lexer.Ident:
			if p.lexer.CurrentString != "from" && p.lexer.CurrentString != "to" {
				p.lexer.Errorf("unexpected string: %s. keyframe selector can only be from, to, or a percentage", p.lexer.CurrentString)
			}
			l.Selectors = append(l.Selectors, &ast.Identifier{
				Span:  p.lexer.Location(),
				Value: p.lexer.CurrentString,
			})

		default:
			p.lexer.Errorf("unexepected token: %s. keyframe selector can only be from, to, or a percentage", p.lexer.Current.String())
		}
		p.lexer.Next()

		if p.lexer.Current == lexer.Comma {
			p.lexer.Next()
			continue
		}

		break
	}

	return l
}

// parseMathExpression does recursive-descent parsing for sums,
// products, then individual values. See https://www.w3.org/TR/css-values-3/#calc-syntax.
func (p *parser) parseMathExpression() ast.Value {
	return p.parseMathSum()
}

func (p *parser) parseMathSum() ast.Value {
	left := p.parseMathProduct()

	for p.lexer.Current == lexer.Delim && (p.lexer.CurrentString == "+" || p.lexer.CurrentString == "-") {
		op := p.lexer.CurrentString
		p.lexer.Expect(lexer.Delim)
		left = &ast.MathExpression{
			Span:     *left.Location(),
			Left:     left,
			Operator: op,
			Right:    p.parseMathProduct(),
		}
	}

	return left
}

func (p *parser) parseMathProduct() ast.Value {
	left := p.parseValue()

	for p.lexer.Current == lexer.Delim && (p.lexer.CurrentString == "*" || p.lexer.CurrentString == "/") {
		op := p.lexer.CurrentString
		p.lexer.Expect(lexer.Delim)
		left = &ast.MathExpression{
			Span:     *left.Location(),
			Left:     left,
			Operator: op,
			Right:    p.parseValue(),
		}
	}

	return left
}

// parseValue parses a possible ast value at the current position. Callers
// can set allowMathOperators if the enclosing context allows math expressions.
// See: https://www.w3.org/TR/css-values-4/#math-function.
func (p *parser) parseValue() ast.Value {
	switch p.lexer.Current {
	case lexer.Dimension:
		defer p.lexer.Next()
		return &ast.Dimension{
			Span: p.lexer.Location(),

			Unit:  p.lexer.CurrentString,
			Value: p.lexer.CurrentNumeral,
		}

	case lexer.Percentage:
		defer p.lexer.Next()
		return &ast.Dimension{
			Span:  p.lexer.Location(),
			Unit:  "%",
			Value: p.lexer.CurrentNumeral,
		}

	case lexer.Number:
		defer p.lexer.Next()
		return &ast.Dimension{
			Span:  p.lexer.Location(),
			Value: p.lexer.CurrentNumeral,
		}

	case lexer.Ident:
		defer p.lexer.Next()
		return &ast.Identifier{
			Span:  p.lexer.Location(),
			Value: p.lexer.CurrentString,
		}

	case lexer.Hash:
		defer p.lexer.Next()
		return &ast.HexColor{
			Span: p.lexer.Location(),
			RGBA: p.lexer.CurrentString,
		}

	case lexer.String:
		defer p.lexer.Next()
		return &ast.String{
			Span:  p.lexer.Location(),
			Value: p.lexer.CurrentString,
		}

	case lexer.FunctionStart:
		fn := &ast.Function{
			Span: p.lexer.Location(),
			Name: p.lexer.CurrentString,
		}
		p.lexer.Next()

	arguments:
		for {
			switch p.lexer.Current {
			case lexer.RParen:
				p.lexer.Next()
				break arguments
			case lexer.Comma:
				fn.Arguments = append(fn.Arguments, &ast.Comma{
					Span: p.lexer.Location(),
				})
				p.lexer.Next()
			default:
				if fn.IsMath() {
					fn.Arguments = append(fn.Arguments, p.parseMathExpression())
					continue
				}
				fn.Arguments = append(fn.Arguments, p.parseValue())
			}
		}

		return fn
	default:
		return nil
	}
}

func (p *parser) parseAtRule() {
	switch p.lexer.CurrentString {
	case "import":
		p.parseImportAtRule()

	case "media":
		p.parseMediaAtRule()

	case "keyframes", "-webkit-keyframes":
		p.parseKeyframes()

	case "custom-media":
		p.parseCustomMediaAtRule()

	default:
		p.lexer.Errorf("unsupported at rule: %s", p.lexer.CurrentString)
	}
}

// parseImportAtRule parses an import at rule. It roughly implements
// https://www.w3.org/TR/css-cascade-4/#at-import.
func (p *parser) parseImportAtRule() {
	prelude := &ast.String{}

	imp := &ast.AtRule{
		Span:     p.lexer.Location(),
		Name:     p.lexer.CurrentString,
		Preludes: []ast.AtPrelude{prelude},
	}
	p.lexer.Next()

	switch p.lexer.Current {
	case lexer.URL:
		prelude.Span = p.lexer.Location()
		prelude.Value = p.lexer.CurrentString
		p.ss.Imports = append(p.ss.Imports, ast.ImportSpecifier{
			Value:  prelude.Value,
			AtRule: imp,
		})
		p.lexer.Next()

	case lexer.FunctionStart:
		if p.lexer.CurrentString != "url" {
			p.lexer.Errorf("@import target must be a url or string")
		}
		p.lexer.Next()

		prelude.Span = p.lexer.Location()
		prelude.Value = p.lexer.CurrentString
		p.ss.Imports = append(p.ss.Imports, ast.ImportSpecifier{
			Value:  prelude.Value,
			AtRule: imp,
		})
		p.lexer.Expect(lexer.String)
		p.lexer.Expect(lexer.RParen)

	case lexer.String:
		prelude.Span = p.lexer.Location()
		prelude.Value = p.lexer.CurrentString
		p.ss.Imports = append(p.ss.Imports, ast.ImportSpecifier{
			Value:  prelude.Value,
			AtRule: imp,
		})
		p.lexer.Expect(lexer.String)

	default:
		p.lexer.Errorf("unexpected import specifier")
	}

	// XXX: also support @supports.
	mq := p.parseMediaQueryList()
	if mq != nil {
		imp.Preludes = append(imp.Preludes, mq)
	}

	p.ss.Nodes = append(p.ss.Nodes, imp)
}

// parseKeyframes parses a keyframes at rule. It roughly implements
// https://www.w3.org/TR/css-animations-1/#keyframes
func (p *parser) parseKeyframes() {
	r := &ast.AtRule{
		Span: p.lexer.Location(),
		Name: p.lexer.CurrentString,
	}
	p.lexer.Next()

	switch p.lexer.Current {
	case lexer.String:
		r.Preludes = append(r.Preludes, &ast.String{
			Span:  p.lexer.Location(),
			Value: p.lexer.CurrentString,
		})

	case lexer.Ident:
		r.Preludes = append(r.Preludes, &ast.Identifier{
			Span:  p.lexer.Location(),
			Value: p.lexer.CurrentString,
		})

	default:
		p.lexer.Errorf("unexpected token %s, expected string or identifier for keyframes", p.lexer.Current.String())
	}
	p.lexer.Next()

	block := &ast.QualifiedRuleBlock{
		Span: p.lexer.Location(),
	}
	r.Block = block
	p.lexer.Expect(lexer.LCurly)
	for {
		switch p.lexer.Current {
		case lexer.EOF:
			p.lexer.Errorf("unexpected EOF")

		case lexer.RCurly:
			p.ss.Nodes = append(p.ss.Nodes, r)
			p.lexer.Next()
			return

		default:
			block.Rules = append(block.Rules, p.parseQualifiedRule(true))
		}
	}
}

// parseMediaAtRule parses a media at rule. It roughly implements
// https://www.w3.org/TR/mediaqueries-4/#media.
func (p *parser) parseMediaAtRule() {
	r := &ast.AtRule{
		Span: p.lexer.Location(),
		Name: p.lexer.CurrentString,
	}
	p.lexer.Next()

	r.Preludes = []ast.AtPrelude{p.parseMediaQueryList()}

	block := &ast.QualifiedRuleBlock{
		Span: p.lexer.Location(),
	}
	r.Block = block
	p.lexer.Expect(lexer.LCurly)
	for {
		switch p.lexer.Current {
		case lexer.EOF:
			p.lexer.Errorf("unexpected EOF")

		case lexer.RCurly:
			p.ss.Nodes = append(p.ss.Nodes, r)
			p.lexer.Next()
			return

		default:
			block.Rules = append(block.Rules, p.parseQualifiedRule(false))
		}
	}
}

func (p *parser) parseMediaQueryList() *ast.MediaQueryList {
	l := &ast.MediaQueryList{
		Span: p.lexer.Location(),
	}

	for {
		if p.lexer.Current == lexer.EOF {
			p.lexer.Errorf("unexpected EOF")
		}

		q := p.parseMediaQuery()
		if q != nil {
			l.Queries = append(l.Queries, q)
		}

		if p.lexer.Current == lexer.Comma {
			p.lexer.Next()
			continue
		}

		break
	}

	if len(l.Queries) == 0 {
		return nil
	}

	return l
}

func (p *parser) parseMediaQuery() *ast.MediaQuery {
	q := &ast.MediaQuery{
		Span: p.lexer.Location(),
	}

	for {
		switch p.lexer.Current {
		case lexer.EOF:
			p.lexer.Errorf("unexpected EOF")
			return q

		case lexer.LParen:
			q.Parts = append(q.Parts, p.parseMediaFeature())

		case lexer.Ident:
			q.Parts = append(q.Parts, p.parseValue().(*ast.Identifier))

		default:
			if len(q.Parts) > 0 {
				return q
			}

			return nil
		}
	}
}

func (p *parser) parseMediaFeature() ast.MediaFeature {
	startLoc := p.lexer.Location()
	p.lexer.Expect(lexer.LParen)

	firstValue := p.parseValue()

	switch p.lexer.Current {
	case lexer.RParen:
		p.lexer.Next()
		ident, ok := firstValue.(*ast.Identifier)
		if !ok {
			// XXX: this location is wrong. also, can't figure out type since we lost the lexer value.
			p.lexer.Errorf("expected identifier in media feature with no value")
		}

		return &ast.MediaFeaturePlain{
			Span:     startLoc,
			Property: ident,
		}

	case lexer.Colon:
		p.lexer.Next()
		ident, ok := firstValue.(*ast.Identifier)
		if !ok {
			// XXX: this location is wrong. also, can't figure out type since we lost the lexer value.
			p.lexer.Errorf("expected identifier in non-range media feature")
		}

		secondValue := p.parseValue()

		p.lexer.Expect(lexer.RParen)
		return &ast.MediaFeaturePlain{
			Span:     startLoc,
			Property: ident,
			Value:    secondValue,
		}

	case lexer.Delim:
		r := &ast.MediaFeatureRange{
			Span:      startLoc,
			LeftValue: firstValue,
		}
		r.Operator = p.parseMediaRangeOperator()

		secondValue := p.parseValue()

		maybeIdent, ok := secondValue.(*ast.Identifier)
		if !ok {
			// If the first value was an identifier, then we'll call that the property.
			maybeIdent, ok := firstValue.(*ast.Identifier)
			if !ok {
				p.lexer.Errorf("expected identifier")
			}

			r.LeftValue = nil
			r.Property = maybeIdent
			r.RightValue = secondValue

			p.lexer.Expect(lexer.RParen)
			return r
		}
		r.Property = maybeIdent

		if p.lexer.Current == lexer.Delim {
			op := p.parseMediaRangeOperator()
			if op != r.Operator {
				p.lexer.Errorf("operators in a range must be the same")
			}
			r.RightValue = p.parseValue()
		}

		p.lexer.Expect(lexer.RParen)
		return r
	}

	p.lexer.Errorf("unexpected token: %s", p.lexer.Current.String())
	return nil
}

var (
	mediaOperatorLT  = "<"
	mediaOperatorLTE = "<="
	mediaOperatorGT  = ">"
	mediaOperatorGTE = ">="
)

func (p *parser) parseMediaRangeOperator() string {
	operator := p.lexer.CurrentString
	p.lexer.Next()

	if p.lexer.Current == lexer.Delim {
		if p.lexer.CurrentString != "=" || (operator != "<" && operator != ">") {
			p.lexer.Errorf("unexpected token: %s", p.lexer.Current.String())
		}

		p.lexer.Next()

		switch operator {
		case "<":
			return mediaOperatorLTE
		case ">":
			return mediaOperatorGTE
		default:
			p.lexer.Errorf("unknown operator: %s", operator)
		}
	}

	switch operator {
	case "<":
		return mediaOperatorLT
	case ">":
		return mediaOperatorGT
	default:
		p.lexer.Errorf("unknown operator: %s", operator)
		return ""
	}
}

// parseCustomMediaAtRule parses a @custom-media rule.
// See: https://www.w3.org/TR/mediaqueries-5/#custom-mq.
func (p *parser) parseCustomMediaAtRule() {
	r := &ast.AtRule{
		Span: p.lexer.Location(),
		Name: p.lexer.CurrentString,
	}
	p.lexer.Next()

	maybeName := p.parseValue()
	name, ok := maybeName.(*ast.Identifier)
	if !ok {
		// XXX: show received type
		p.lexer.Errorf("expected identifier")
	}

	r.Preludes = append(r.Preludes, name)
	queries := p.parseMediaQueryList()
	if len(queries.Queries) != 1 {
		p.lexer.Errorf("@custom-media rule requires a single media query argument")
	}
	r.Preludes = append(r.Preludes)
	r.Preludes = append(r.Preludes, queries.Queries[0])

	p.ss.Nodes = append(p.ss.Nodes, r)
}
