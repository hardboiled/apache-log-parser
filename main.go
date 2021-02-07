package main

import (
	"fmt"
	"os"

	"github.com/hardboiled/apache-log-parser/parsing"
)

func main() {
	reader, err := os.Open("./input_files/sample_csv.txt")
	if err != nil {
		fmt.Printf("error opening input file %v\n", err)
		os.Exit(1)
	}

	ch := make(chan parsing.WebServerLogData, 100)
	// TODO don't alert twice
	go parsing.ParseWebServerLogDataWithChannel(reader, ch)
	if err != nil {
		fmt.Printf("unable to parse input file %v\n", err)
		os.Exit(1)
	}
	for data := range ch {
		fmt.Printf("%v", data)
	}
}
