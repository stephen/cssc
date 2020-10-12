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
				panic(logging.LocationErrorf(source, p.lexer.TokenSpan(), "%v", errI))
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
		Span: p.lexer.TokenSpan(),
	}

	for {
		switch p.lexer.Current {
		case lexer.EOF:
			p.lexer.Errorf("unexpected EOF")

		case lexer.LCurly:
			r.Block = p.parseDeclarationBlock()
			r.End = r.Block.Location().End
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

// parseDeclarationBlock parses a {} block with declarations, e.g.
// { width: 1px; }.
func (p *parser) parseDeclarationBlock() *ast.DeclarationBlock {
	block := &ast.DeclarationBlock{
		Span: p.lexer.TokenSpan(),
	}
	p.lexer.Next()

	for p.lexer.Current != lexer.RCurly {
		decl := &ast.Declaration{
			Span:     p.lexer.TokenSpan(),
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
				decl.End = p.lexer.TokenEnd()
				p.lexer.Next()
				decl.Important = true

			case lexer.Comma:
				decl.Values = append(decl.Values, &ast.Comma{Span: p.lexer.TokenSpan()})
				p.lexer.Next()

			default:
				val := p.parseValue()
				if val == nil {
					if len(decl.Values) == 0 {
						p.lexer.Errorf("declaration must have a value")
					}
					if lastValueEnd := decl.Values[len(decl.Values)-1].Location().End; lastValueEnd > decl.End {
						decl.End = lastValueEnd
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
	block.End = p.lexer.TokenEnd()
	p.lexer.Next()
	return block
}

func (p *parser) parseKeyframeSelectorList() *ast.KeyframeSelectorList {
	l := &ast.KeyframeSelectorList{
		Span: p.lexer.TokenSpan(),
	}

	for {
		if p.lexer.Current == lexer.EOF {
			p.lexer.Errorf("unexpected EOF")
		}

		switch p.lexer.Current {
		case lexer.Percentage:
			l.Selectors = append(l.Selectors, &ast.Percentage{
				Span:  p.lexer.TokenSpan(),
				Value: p.lexer.CurrentNumeral,
			})

		case lexer.Ident:
			if p.lexer.CurrentString != "from" && p.lexer.CurrentString != "to" {
				p.lexer.Errorf("unexpected string: %s. keyframe selector can only be from, to, or a percentage", p.lexer.CurrentString)
			}
			l.Selectors = append(l.Selectors, &ast.Identifier{
				Span:  p.lexer.TokenSpan(),
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

	if len(l.Selectors) == 0 {
		p.lexer.Errorf("keyframes rule must have at least one selector (from, to, or a percentage)")
	}

	l.End = l.Selectors[len(l.Selectors)-1].Location().End
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

		right := p.parseMathProduct()

		span := left.Location()
		span.End = right.Location().End

		left = &ast.MathExpression{
			Span:     span,
			Left:     left,
			Operator: op,
			Right:    right,
		}
	}

	return left
}

func (p *parser) parseMathProduct() ast.Value {
	left := p.parseValue()

	for p.lexer.Current == lexer.Delim && (p.lexer.CurrentString == "*" || p.lexer.CurrentString == "/") {
		op := p.lexer.CurrentString
		p.lexer.Expect(lexer.Delim)

		right := p.parseValue()

		span := left.Location()
		span.End = right.Location().End

		left = &ast.MathExpression{
			Span:     span,
			Left:     left,
			Operator: op,
			Right:    right,
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
			Span: p.lexer.TokenSpan(),

			Unit:  p.lexer.CurrentString,
			Value: p.lexer.CurrentNumeral,
		}

	case lexer.Percentage:
		defer p.lexer.Next()
		return &ast.Dimension{
			Span:  p.lexer.TokenSpan(),
			Unit:  "%",
			Value: p.lexer.CurrentNumeral,
		}

	case lexer.Number:
		defer p.lexer.Next()
		return &ast.Dimension{
			Span:  p.lexer.TokenSpan(),
			Value: p.lexer.CurrentNumeral,
		}

	case lexer.Ident:
		defer p.lexer.Next()
		return &ast.Identifier{
			Span:  p.lexer.TokenSpan(),
			Value: p.lexer.CurrentString,
		}

	case lexer.Hash:
		defer p.lexer.Next()
		return &ast.HexColor{
			Span: p.lexer.TokenSpan(),
			RGBA: p.lexer.CurrentString,
		}

	case lexer.String:
		defer p.lexer.Next()
		return &ast.String{
			Span:  p.lexer.TokenSpan(),
			Value: p.lexer.CurrentString,
		}

	case lexer.FunctionStart:
		fn := &ast.Function{
			Span: p.lexer.TokenSpan(),
			Name: p.lexer.CurrentString,
		}
		p.lexer.Next()

	arguments:
		for {
			switch p.lexer.Current {
			case lexer.RParen:
				fn.End = p.lexer.TokenEnd()
				p.lexer.Next()
				break arguments
			case lexer.Comma:
				fn.Arguments = append(fn.Arguments, &ast.Comma{
					Span: p.lexer.TokenSpan(),
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

	case "font-face":
		p.parseFontFace()

	default:
		p.lexer.Errorf("unsupported at rule: %s", p.lexer.CurrentString)
	}
}

// parseImportAtRule parses an import at rule. It roughly implements
// https://www.w3.org/TR/css-cascade-4/#at-import.
func (p *parser) parseImportAtRule() {
	prelude := &ast.String{}

	imp := &ast.AtRule{
		Span:     p.lexer.TokenSpan(),
		Name:     p.lexer.CurrentString,
		Preludes: []ast.AtPrelude{prelude},
	}
	p.lexer.Next()

	switch p.lexer.Current {
	case lexer.URL:
		prelude.Span = p.lexer.TokenSpan()
		prelude.Value = p.lexer.CurrentString
		p.ss.Imports = append(p.ss.Imports, ast.ImportSpecifier{
			Value:  prelude.Value,
			AtRule: imp,
		})
		p.lexer.Next()

	case lexer.FunctionStart:
		prelude.Span = p.lexer.TokenSpan()
		if p.lexer.CurrentString != "url" {
			p.lexer.Errorf("@import target must be a url or string")
		}
		p.lexer.Next()

		prelude.Value = p.lexer.CurrentString
		p.ss.Imports = append(p.ss.Imports, ast.ImportSpecifier{
			Value:  prelude.Value,
			AtRule: imp,
		})
		p.lexer.Expect(lexer.String)
		prelude.Span.End = p.lexer.TokenEnd()
		p.lexer.Expect(lexer.RParen)

	case lexer.String:
		prelude.Span = p.lexer.TokenSpan()
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

	imp.End = imp.Preludes[len(imp.Preludes)-1].Location().End
	p.ss.Nodes = append(p.ss.Nodes, imp)
}

// parseKeyframes parses a keyframes at rule. It roughly implements
// https://www.w3.org/TR/css-animations-1/#keyframes
func (p *parser) parseKeyframes() {
	r := &ast.AtRule{
		Span: p.lexer.TokenSpan(),
		Name: p.lexer.CurrentString,
	}
	p.lexer.Next()

	switch p.lexer.Current {
	case lexer.String:
		r.Preludes = append(r.Preludes, &ast.String{
			Span:  p.lexer.TokenSpan(),
			Value: p.lexer.CurrentString,
		})

	case lexer.Ident:
		r.Preludes = append(r.Preludes, &ast.Identifier{
			Span:  p.lexer.TokenSpan(),
			Value: p.lexer.CurrentString,
		})

	default:
		p.lexer.Errorf("unexpected token %s, expected string or identifier for keyframes", p.lexer.Current.String())
	}
	p.lexer.Next()

	block := &ast.QualifiedRuleBlock{
		Span: p.lexer.TokenSpan(),
	}
	r.Block = block
	p.lexer.Expect(lexer.LCurly)
	for {
		switch p.lexer.Current {
		case lexer.EOF:
			p.lexer.Errorf("unexpected EOF")

		case lexer.RCurly:
			p.ss.Nodes = append(p.ss.Nodes, r)
			block.End = p.lexer.TokenEnd()
			r.End = block.End
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
		Span: p.lexer.TokenSpan(),
		Name: p.lexer.CurrentString,
	}
	p.lexer.Next()

	r.Preludes = []ast.AtPrelude{p.parseMediaQueryList()}

	block := &ast.QualifiedRuleBlock{
		Span: p.lexer.TokenSpan(),
	}
	r.Block = block
	p.lexer.Expect(lexer.LCurly)
	for {
		switch p.lexer.Current {
		case lexer.EOF:
			p.lexer.Errorf("unexpected EOF")

		case lexer.RCurly:
			p.ss.Nodes = append(p.ss.Nodes, r)
			block.End = p.lexer.TokenEnd()
			r.End = block.End
			p.lexer.Next()
			return

		default:
			block.Rules = append(block.Rules, p.parseQualifiedRule(false))
		}
	}
}

func (p *parser) parseMediaQueryList() *ast.MediaQueryList {
	l := &ast.MediaQueryList{
		Span: p.lexer.TokenSpan(),
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

	l.End = l.Queries[len(l.Queries)-1].End
	return l
}

func (p *parser) parseMediaQuery() *ast.MediaQuery {
	q := &ast.MediaQuery{
		Span: p.lexer.TokenSpan(),
	}

	for {
		switch p.lexer.Current {
		case lexer.EOF:
			p.lexer.Errorf("unexpected EOF")

		case lexer.LParen:
			q.Parts = append(q.Parts, p.parseMediaFeature())

		case lexer.Ident:
			q.Parts = append(q.Parts, p.parseValue().(*ast.Identifier))

		default:
			if len(q.Parts) > 0 {
				q.End = q.Parts[len(q.Parts)-1].Location().End
				return q
			}

			return nil
		}
	}
}

func (p *parser) parseMediaFeature() ast.MediaFeature {
	startLoc := p.lexer.TokenSpan()
	p.lexer.Expect(lexer.LParen)

	firstValue := p.parseValue()

	switch p.lexer.Current {
	case lexer.RParen:
		startLoc.End = p.lexer.TokenEnd()
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

		plain := &ast.MediaFeaturePlain{
			Span:     startLoc,
			Property: ident,
			Value:    secondValue,
		}
		// XXX: this one excludes the right paren even though the left paren is included.
		plain.End = p.lexer.TokenEnd()
		p.lexer.Expect(lexer.RParen)
		return plain

	case lexer.Delim:
		r := &ast.MediaFeatureRange{
			Span:      startLoc,
			LeftValue: firstValue,
		}
		r.Operator = p.parseMediaRangeOperator()

		secondValue := p.parseValue()

		maybeIdent, ok := secondValue.(*ast.Identifier)
		if !ok {
			// Since the second value isn't an identifier, we expect something like
			// width < 600px, so the first value must have been an identifier.
			maybeIdent, ok := firstValue.(*ast.Identifier)
			if !ok {
				p.lexer.Errorf("expected identifier")
			}

			r.LeftValue = nil
			r.Property = maybeIdent
			r.RightValue = secondValue

			r.End = p.lexer.TokenEnd()
			p.lexer.Expect(lexer.RParen)
			return r
		}

		// Otherwise, the second value is an identifier and we are looking at something like
		// 600px < width < 800px.
		r.Property = maybeIdent

		if p.lexer.Current == lexer.Delim {
			op := p.parseMediaRangeOperator()
			if op != r.Operator {
				p.lexer.Errorf("operators in a range must be the same")
			}
			r.RightValue = p.parseValue()
		}

		r.End = p.lexer.TokenEnd()
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
		Span: p.lexer.TokenSpan(),
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
	r.Preludes = append(r.Preludes, queries.Queries[0])
	r.End = queries.Queries[0].End

	p.ss.Nodes = append(p.ss.Nodes, r)
}

// parseFontFace parses an @font-face rule.
// See: https://www.w3.org/TR/css-fonts-4/#font-face-rule
func (p *parser) parseFontFace() {
	r := &ast.AtRule{
		Span: p.lexer.TokenSpan(),
		Name: p.lexer.CurrentString,
	}
	p.lexer.Next()

	r.Block = p.parseDeclarationBlock()
	r.End = r.Block.Location().End

	p.ss.Nodes = append(p.ss.Nodes, r)
}
