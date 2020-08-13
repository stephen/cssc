package main

import (
	"log"

	"github.com/davecgh/go-spew/spew"
	"github.com/stephen/cssc/internal/parser"
)

func main() {
	source := `@import "test.css";
@import url("./testing.css");
	@import url(tester.css);
	/* some notes about the next line
	are here */

	.class {}
	#id {}
	body#id {}
	body::after {}
	a:hover {}
	:not(a, b, c) {}
	.one, .two {}
	`
	// .n {
	// 	width: yes;
	// 	height: 2.3px;
	// 	border: -2em;
	// 	content: "test\u005ctest";
	// }

	// #test {
	// 	uhoh: hello;
	// 	img: url(test.com)
	// 	other: url("test.net")
	// }

	log.Println(spew.Sdump(parser.Parse(source)))
}
