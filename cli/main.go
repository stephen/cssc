package main

import (
	"log"

	"github.com/davecgh/go-spew/spew"
	"github.com/stephen/cssc/internal/lexer"
)

func main() {
	source := `@import "test.css";
/* some notes about the next line
are here */
.n {
	width: yes;
	height: 2.3px;
	border: -2em;
	content: "test\u005ctest";
}

#test {
	uhoh: hello;
	img: url(test.com)
}
`

	log.Println(spew.Sdump(Parse(source)))

}

func Parse(source string) *Stylesheet {
	l := lexer.NewLexer(source)

	ss := &Stylesheet{}

	for l.Current != lexer.EOF {
		l.Next()
		switch l.Current {
		case lexer.AtKeyword:
			rule := &AtRule{
				Name: l.CurrentString,
			}

		atKeyword:
			for {
				l.Next()
				switch l.Current {
				case lexer.Semicolon:
					break atKeyword
				case lexer.EOF:
					panic("uh oh")
				// case {
				default:
					rule.Prelude += l.CurrentString
				}
			}

			ss.Children = append(ss.Children, rule)

		}
		log.Printf("current token: %s (%s%s)", l.Current, l.CurrentNumeral, l.CurrentString)
		l.CurrentNumeral, l.CurrentString = "", ""
	}
	return ss
}

type Stylesheet struct {
	Children []interface{}
}

type AtRule struct {
	Name string

	// XXX: Prelude ("test.css" or media query)
	Prelude string

	// XXX: Block (contents of {})
}
