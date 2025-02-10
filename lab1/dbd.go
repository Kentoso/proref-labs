package main

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"os"
)

//go:embed example.dbd
var dbdInput string

func main() {
	tokens, err := lex(dbdInput)
	if err != nil {
		fmt.Println("Error lexing:", err)
		return
	}

	//for _, token := range tokens {
	//	fmt.Println(token)
	//}

	parser := NewParser(tokens)

	segms, err := parser.ParseDBD()
	if err != nil {
		fmt.Println("Error parsing DBD:", err)
		return
	}
	fmt.Printf("Parsed %d segments\n", len(segms))

	for _, segm := range segms {
		attrMessage := ""
		for _, attr := range segm.Attributes {
			attrMessage += fmt.Sprintf("%s=%s ", attr.Key, attr.Value)
		}

		fmt.Printf("Got SEGM %s\n", attrMessage)
	}

	segmsJson, err := json.MarshalIndent(segms, "", "  ")
	if err != nil {
		fmt.Println("Error marshalling segms:", err)
		return
	}

	f, err := os.Create("example.json")
	if err != nil {
		fmt.Println("Error creating file:", err)
		return
	}

	defer f.Close()

	_, err = f.Write(segmsJson)
	if err != nil {
		fmt.Println("Error writing to file:", err)
		return
	}
}
