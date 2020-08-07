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
`)
	for l.Current != EOF {
		l.Next()
		extra := ""
		switch l.Current {
		case Ident:
			extra = l.CurrentName
		case String, Comment:
			extra = l.CurrentString
		}
		log.Printf("current token: %s (%s)", l.Current, extra)
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

	CurrentName   string // Ident
	CurrentString string // Literal

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

		case '@':
			l.Current = At
			l.step()

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
				case -1:
					l.errorf("unexpected EOF")
				default:
					l.step()
				}
			}

			// TODO: full string-token parsing
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

			if isNameStartCodePoint(l.ch) {
				// TODO: ident-token
				l.Current = Ident
				start := l.lastPos
				for isNameCodePoint(l.ch) {
					l.step()
				}

				l.CurrentName = l.source[start:l.lastPos]
				l.step()
			}
		}

		return
	}
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

// peek returns the next unconsumed rune but does not consume it.
func (l *Lexer) peek() rune {
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
	At         // @

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
)

func (t Token) String() string {
	return tokens[t]
}

var tokens = [...]string{
	Illegal: "Illegal",

	EOF: "EOF",

	Comment: "COMMENT",

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
	At:         "@",

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
