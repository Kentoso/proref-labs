package main

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"strings"
)

func removeLineNumber(line string) (string, error) {
	if len(line) == 80 {
		return line[:72], nil
	}

	return "", fmt.Errorf("invalid line length: %d", len(line))
}

func getContinuationLine(line string) string {
	return line[14:]
}

func preprocessLines(input string) ([]string, error) {
	lines := strings.Split(input, "\n")
	var logicalLines []string
	var cleanedLines []string

	for i := 0; i < len(lines); i++ {
		line, err := removeLineNumber(lines[i])
		if err != nil {
			return nil, err
		}

		cleanedLines = append(cleanedLines, line)
	}

	for i := 0; i < len(cleanedLines); i++ {
		line := cleanedLines[i]
		if line[len(line)-1] == '*' {
			if i+1 < len(cleanedLines) {
				nextLine := cleanedLines[i+1]
				line = strings.TrimRight(line[:len(line)-1], " ") + getContinuationLine(nextLine)
				i++
			} else {
				return nil, fmt.Errorf("%d: missing continuation line", i)
			}
		}

		logicalLines = append(logicalLines, line)
	}

	return logicalLines, nil
}

//go:embed example.dbd
var dbdInput string

func main() {
	logicalLines, err := preprocessLines(dbdInput)
	if err != nil {
		fmt.Println("Error preprocessing lines:", err)
		return
	}

	var tokens []Token

	for _, line := range logicalLines {
		tokens = append(tokens, lex(line)...)
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
