package main

import (
	"log"
	"unicode"
	"unicode/utf8"

	"github.com/samsarahq/go/oops"
)

func main() {
	l := NewLexer(`@import "test.css";
/* some notes about the next line
are here */
.n {
	width: yes;
}
`)

	for l.Current != EOF {
		l.Next()
		log.Printf("current token: %s (%s)", l.Current, l.CurrentLiteral)
	}
}

// Lexer lexes the input source. Callers push the lexer
// along with calls to Next(), which populate the current
// token.
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

	// CurrentLiteral is the last literal lexed by Next(). It
	// is not cleared between valid literals.
	CurrentLiteral string // Literal

	Errors []error
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
			log.Println("colon")
			l.Current = Colon
			l.step()

		case '@':
			l.step()
			if startsIdentifier(l.peek(0), l.peek(1), l.peek(2)) {
				l.Current = AtKeyword
				l.CurrentLiteral = l.nextName()
				return
			}

			l.Current = Delim
			l.CurrentLiteral = string(l.ch)

		case '{':
			l.Current = LCurly
			l.step()

		case '}':
			l.Current = RCurly
			l.step()

		case '.':
			// XXX: support number parsing here too.
			start := l.lastPos
			l.step()
			l.Current = Delim
			l.CurrentLiteral = l.source[start:l.lastPos]

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
			l.CurrentLiteral = l.source[start:end]

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
				case -1:
					l.errorf("unexpected EOF")
				default:
					l.step()
				}
			}

			// TODO: full string-token parsing
			l.Current = String
			l.CurrentLiteral = l.source[start:end]

		default:
			if isWhitespace(l.ch) {
				l.step()
				// Don't return out because we only processed whitespace and
				// there's nothing interesting for the caller yet. We don't emit
				// whitespace-token.
				continue
			}

			// consume a name
			if isNameStartCodePoint(l.ch) {
				l.Current = Ident
				l.CurrentLiteral = l.nextName()
			}

		}

		return
	}
}

// startsIdentifier implements https://www.w3.org/TR/css-syntax-3/#would-start-an-identifier.
func startsIdentifier(p0, p1, p2 rune) bool {
	switch p0 {
	case '-':
		return p1 == '-' || isEscape(p1, p2)
	case '\n':
		return false
	default:
		return isNameCodePoint(p1)
	}
}

// isEscape implements https://www.w3.org/TR/css-syntax-3/#starts-with-a-valid-escape
func isEscape(p0, p1 rune) bool {
	if p0 != '\\' {
		return false
	}

	if p1 == '\n' {
		return false
	}

	return true
}

// nextName consumes and returns the a name, stepping the lexer forward.
func (l *Lexer) nextName() string {
	start := l.lastPos
	for isNameCodePoint(l.ch) {
		l.step()
	}

	return l.source[start:l.lastPos]
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

func (l *Lexer) errorf(fmt string, args ...interface{}) {
	panic(oops.Errorf(fmt, args...))
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

	NumberSign // #
	Apostrophe // '
	Plus       // +
	Comma      // ,
	Hyphen     // -
	Period     // .
	Colon      // :
	Semicolon  // ;
	AtKeyword  // @

	Backslash // \

	LessThan    // <
	GreaterThan // >

	LParen // (
	RParen // )

	LBracket // [
	RBracket // ]

	LCurly // {
	RCurly // }

	Digit  // Number literal
	String // String literal
	Ident  // Identifier
	Delim  // Delimiter (used for preserving tokens for subprocessors)
)

func (t Token) String() string {
	return tokens[t]
}

var tokens = [...]string{
	Illegal: "Illegal",

	EOF: "EOF",

	Comment: "COMMENT",
	Delim:   "DELIMITER",

	Digit:  "DIGIT",
	String: "STRING",
	Ident:  "IDENT",

	NumberSign: "#",
	Apostrophe: "'",
	Plus:       "+",
	Comma:      ",",
	Hyphen:     "-",
	Period:     ".",
	Colon:      ":",
	Semicolon:  ";",
	AtKeyword:  "@",

	Backslash: `\`,

	LessThan:    "<",
	GreaterThan: ">",

	LParen: "(",
	RParen: ")",

	LBracket: "[",
	RBracket: "]",

	LCurly: "{",
	RCurly: "}",
}
