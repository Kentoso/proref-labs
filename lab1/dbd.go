package main

import (
	_ "embed"
	"encoding/json"
	"fmt"
)

//go:embed example.dbd
var dbdInput string

func main() {
	tokens, err := lex(dbdInput)
	if err != nil {
		fmt.Println("Error lexing:", err)
		return
	}

	for _, token := range tokens {
		fmt.Println(token)
	}

	parser := NewParser(tokens)

	segms, err := parser.ParseDBD()
	if err != nil {
		fmt.Println("Error parsing DBD:", err)
		return
	}

	segmsJson, err := json.MarshalIndent(segms, "", "  ")
	if err != nil {
		fmt.Println("Error marshalling segms:", err)
		return
	}

	fmt.Println(string(segmsJson))
}
