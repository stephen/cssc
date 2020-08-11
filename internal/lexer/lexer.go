// Package lexer implements a css lexer that is meant to be
// used in conjuction with its sibling parser package.
package lexer

import (
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"
)

// Lexer lexes the input source. Callers push the lexer
// along with calls to Next(), which populate the current
// token and literals.
type Lexer struct {
	// ch is the last rune consumed with step(). If there are
	// no more runes, ch is -1.
	ch rune

	// pos is the current byte (not codepoint) offset within source.
	pos int

	// lastPos is the last byte (not codepoint) offset within source.
	lastPos int

	// source is the current source code being lexed.
	source string

	// Current is the last token lexed by Next().
	Current Token

	// CurrentString is the last literal string lexed by Next(). It
	// is not cleared between valid literals.
	CurrentString string

	// CurrentNumeral is the last literal numeral lexed by Next(). It
	// is not cleared between valid literals.
	CurrentNumeral string
}

// NewLexer creates a new lexer for the source.
func NewLexer(source string) *Lexer {
	l := &Lexer{
		source: source,
	}
	l.step()
	return l
}

// Next consumes the most recent r.
func (l *Lexer) Next() {
	// Run in a for-loop so that some types (e.g. whitespace) can use continue to
	// move on to the next token. Other codepaths will end in a return statement
	// at the end of a single iteration.
	for {
		switch l.ch {
		case -1:
			l.Current = EOF
			return

		case ';':
			l.Current = Semicolon
			l.step()

		case ':':
			l.Current = Colon
			l.step()

		case '+':
			if startsNumber(l.ch, l.peek(0), l.peek(1)) {
				l.nextNumericToken()
				return
			}

			l.nextDelimToken()

		case '-':
			if startsNumber(l.ch, l.peek(0), l.peek(1)) {
				l.nextNumericToken()
				return
			}

			if p0, p1 := l.peek(0), l.peek(1); p0 == '-' && p1 == '>' {
				l.Current = CDC
				return
			}

			// XXX: identifier

			l.nextDelimToken()

		case '<':
			if p0, p1, p2 := l.peek(0), l.peek(1), l.peek(2); p0 == '!' && p1 == '-' && p2 == '-' {
				l.Current = CDO
				return
			}

			// Otherwise save it as a delimiter.
			l.nextDelimToken()

		case '@':
			if startsIdentifier(l.peek(0), l.peek(1), l.peek(2)) {
				l.step() // Consume @.
				l.Current = AtKeyword

				start := l.lastPos
				l.nextName()
				l.CurrentString = l.source[start:l.lastPos]
				return
			}

			l.nextDelimToken()

		case '{':
			l.Current = LCurly
			l.step()

		case '}':
			l.Current = RCurly
			l.step()

		case '.':
			if unicode.IsDigit(l.peek(0)) {
				l.nextNumericToken()
				return
			}

			// XXX: support number parsing here too.

			l.nextDelimToken()

		case '/':
			l.step()
			if l.ch != '*' {
				l.errorf("expected * but got %c", l.ch)
			}
			start, end := l.lastPos, -1

		commentToken:
			for {
				switch l.ch {
				case '*':
					l.step()
					if l.ch == '/' {
						end = l.lastPos
						l.step()
						break commentToken
					}
					l.step()
				case -1:
					l.errorf("unexpected EOF")
				default:
					l.step()
				}
			}
			l.Current = Comment
			l.CurrentString = l.source[start:end]

		case '"', '\'':
			mark := l.ch

			l.step()
			start, end := l.lastPos, -1

		stringToken:
			for {
				switch l.ch {
				case mark:
					end = l.lastPos
					l.step()
					break stringToken
				case '\n':
					l.errorf("unexpected newline")
				case '\\':
					l.step()

					switch l.ch {
					case '\n':
						l.step()
					case -1:
						l.errorf("unexpected EOF")
					default:
						if startsEscape(l.ch, l.peek(0)) {
							l.nextEscaped()
						}
					}
				case -1:
					l.errorf("unexpected EOF")
				default:
					l.step()
				}
			}

			l.Current = String
			l.CurrentString = l.source[start:end]

		default:
			if isWhitespace(l.ch) {
				l.step()
				// Don't return out because we only processed whitespace and
				// there's nothing interesting for the caller yet. We don't emit
				// whitespace-token.
				continue
			}

			if unicode.IsDigit(l.ch) {
				l.nextNumericToken()
			}

			// https://www.w3.org/TR/css-syntax-3/#consume-ident-like-token
			if isNameStartCodePoint(l.ch) {
				start := l.lastPos
				l.nextName()
				l.CurrentString = l.source[start:l.lastPos]

				// Here, we need to special case the url function because it supports unquoted string content.
				if strings.ToLower(l.CurrentString) == "url" && l.peek(0) == '(' {
					for i := l.lastPos; i < len(l.source) && isWhitespace(l.peek(0)); i++ {
						l.step()
					}

					if p0 := l.peek(0); p0 == '\'' || p0 == '"' {
						l.Current = FunctionStart
						break
					}

					// XXX: url token

					break
				}

				// Otherwise, it's probably a normal function.
				if l.peek(0) == '(' {
					l.Current = FunctionStart
					break
				}

				l.Current = Ident
			}

		}

		return
	}
}

// startsIdentifier implements https://www.w3.org/TR/css-syntax-3/#would-start-an-identifier.
func startsIdentifier(p0, p1, p2 rune) bool {
	switch p0 {
	case '-':
		return p1 == '-' || startsEscape(p1, p2)
	case '\n':
		return false
	default:
		return isNameCodePoint(p1)
	}
}

// startsEscape implements https://www.w3.org/TR/css-syntax-3/#starts-with-a-valid-escape
func startsEscape(p0, p1 rune) bool {
	if p0 != '\\' {
		return false
	}

	if p1 == '\n' {
		return false
	}

	return true
}

// startsNumber implements https://www.w3.org/TR/css-syntax-3/#starts-with-a-number.
func startsNumber(p0, p1, p2 rune) bool {
	if p0 == '+' || p0 == '-' {
		if unicode.IsDigit(p1) {
			return true
		}

		if p1 == '.' && unicode.IsDigit(p2) {
			return true
		}

		return false
	}

	if p0 == '.' && unicode.IsDigit(p1) {
		return true
	}

	return unicode.IsDigit(p0)
}

// nextNumericToken implements https://www.w3.org/TR/css-syntax-3/#consume-a-numeric-token
// and sets the lexer state.
func (l *Lexer) nextNumericToken() {
	start := l.lastPos
	l.nextNumber()
	l.CurrentNumeral = l.source[start:l.lastPos]

	if startsIdentifier(l.ch, l.peek(0), l.peek(1)) {
		dimenStart := l.lastPos
		l.nextName()
		l.CurrentString = l.source[dimenStart:l.lastPos]
		l.Current = Dimension
	} else if l.ch == '%' {
		l.Current = Percentage
	} else {
		l.Current = Number
	}
}

// nextNumber implements https://www.w3.org/TR/css-syntax-3/#consume-a-number
// and consumes a number. We don't distinguish between number and integer because
// it doesn't matter for us.
func (l *Lexer) nextNumber() {
	if l.ch == '+' || l.ch == '-' {
		l.step()
	}

	for unicode.IsDigit(l.ch) {
		l.step()
	}

	if l.ch == '.' && unicode.IsDigit(l.peek(0)) {
		l.step()
		l.step()
	}

	if p0, p1 := l.peek(0), l.peek(1); (l.ch == 'e' || l.ch == 'E') && (unicode.IsDigit(p0) ||
		((p0 == '+' || p0 == '-') && unicode.IsDigit(p1))) {
		l.step()
		if l.ch == '+' || l.ch == '-' {
			l.step()
		}

		for unicode.IsDigit(l.ch) {
			l.step()
		}
	}
}

// nextDelim consumes a codepoint and saves it as a delimiter token.
func (l *Lexer) nextDelimToken() {
	start := l.lastPos
	l.step()
	l.Current = Delim
	l.CurrentString = l.source[start:l.lastPos]
}

// nextName consumes and returns a name, stepping the lexer forward.
func (l *Lexer) nextName() {
	for isNameCodePoint(l.ch) {
		l.step()
	}
}

// nextEscaped consumes and returns an escaped codepoint, stepping the lexer forward.
// It implements https://www.w3.org/TR/css-syntax-3/#consume-escaped-code-point.
//
// Note that we do not need to interpret the codepoint for our purposes - we can record
// the byte offsets as-is for transformation purposes.
func (l *Lexer) nextEscaped() {
	l.step()
	for i := 0; i < 5 && isHexDigit(l.ch); i++ {
		l.step()
		if isWhitespace(l.ch) {
			l.step()
		}
	}
}

func isHexDigit(r rune) bool {
	return unicode.IsDigit(r) || (r >= 'A' && r <= 'F') || (r >= 'a' && r <= 'f')
}

// isWhitespace implements https://www.w3.org/TR/css-syntax-3/#whitespace.
func isWhitespace(r rune) bool {
	return r == '\n' || r == '\u0009' || r == ' '
}

// isNameStartCodePoint implements https://www.w3.org/TR/css-syntax-3/#name-start-code-point.
func isNameStartCodePoint(r rune) bool {
	return unicode.IsLetter(r) || int32(r) >= 0x80 || r == '_'
}

// isNameCodePoint implements https://www.w3.org/TR/css-syntax-3/#name-code-point.
func isNameCodePoint(r rune) bool {
	return isNameStartCodePoint(r) || unicode.IsNumber(r)
}

func (l *Lexer) errorf(f string, args ...interface{}) {
	panic(fmt.Sprintf(f, args...))
}

// step consumes the next unicode rune and stores it.
func (l *Lexer) step() {
	cp, size := utf8.DecodeRuneInString(l.source[l.pos:])

	if size == 0 {
		l.ch = -1
		return
	}

	l.ch = cp
	l.lastPos = l.pos
	l.pos += size
}

// peek returns the next ith unconsumed rune but does not consume it.
// i is 0-indexed (0 is one ahead, 1 is two ahead, etc.)
func (l *Lexer) peek(i int) rune {
	cp, size := utf8.DecodeRuneInString(l.source[l.pos:])
	if size == 0 {
		return -1
	}
	return cp
}

// Token is the set of lexical tokens in CSS.
type Token int

// https://www.w3.org/TR/css-syntax-3/#consume-token
const (
	Illegal Token = iota

	EOF

	Comment

	NumberSign    // #
	Apostrophe    // '
	Plus          // +
	Comma         // ,
	Hyphen        // -
	Period        // .
	Colon         // :
	Semicolon     // ;
	AtKeyword     // @
	FunctionStart // something(

	Backslash // \

	LessThan    // <
	GreaterThan // >

	LParen // (
	RParen // )

	CDO // <!--
	CDC // -->

	LBracket // [
	RBracket // ]

	LCurly // {
	RCurly // }

	Number     // Number literal
	Percentage // Percentage literal
	Dimension  // Dimension literal
	String     // String literal
	Ident      // Identifier
	Delim      // Delimiter (used for preserving tokens for subprocessors)
)

func (t Token) String() string {
	return tokens[t]
}

var tokens = [...]string{
	Illegal: "Illegal",

	EOF: "EOF",

	Comment: "COMMENT",
	Delim:   "DELIMITER",

	Number:     "NUMBER",
	Percentage: "PERCENTAGE",
	Dimension:  "DIMENSION",
	String:     "STRING",
	Ident:      "IDENT",

	NumberSign:    "#",
	Apostrophe:    "'",
	Plus:          "+",
	Comma:         ",",
	Hyphen:        "-",
	Period:        ".",
	Colon:         ":",
	Semicolon:     ";",
	AtKeyword:     "@",
	FunctionStart: "FUNCTION",

	Backslash: `\`,

	LessThan:    "<",
	GreaterThan: ">",

	CDO: "<!--",
	CDC: "-->",

	LParen: "(",
	RParen: ")",

	LBracket: "[",
	RBracket: "]",

	LCurly: "{",
	RCurly: "}",
}
