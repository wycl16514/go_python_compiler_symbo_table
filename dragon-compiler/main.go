package main

import (
	"fmt"
	"io"
	"lexer"
	"simple_parser"
)

func main() {
	source := "{int x; char y; {bool y; x; y;} x; y;}"
	my_lexer := lexer.NewLexer(source)
	parser := simple_parser.NewSimpleParser(my_lexer)
	err := parser.Parse()
	if err == io.EOF || err == nil {
		fmt.Println("\nparsing success")
	} else {
		fmt.Println("source is ilegal : ", err)
	}
}
