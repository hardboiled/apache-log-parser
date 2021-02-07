package main

import (
	"fmt"
	"os"

	"github.com/hardboiled/apache-log-parser/parsing"
	"github.com/hardboiled/apache-log-parser/webstats"
)

func main() {
	reader, err := os.Open("./input_files/sample_csv.txt")
	if err != nil {
		fmt.Printf("error opening input file %v\n", err)
		os.Exit(1)
	}

	ch := make(chan parsing.WebServerLogData, 100)

	webStats, err := webstats.InitWebStats(10)
	if err != nil {
		fmt.Printf("error initializing webstats %v\n", err)
		os.Exit(1)
	}

	go parsing.ParseWebServerLogDataWithChannel(reader, ch)
	for data := range ch {
		lastAlarm := webStats.HasTotalTrafficAlarm()
		webStats.AddEntry(data.GetRequestSection(), data.Date)
		curAlarm := webStats.HasTotalTrafficAlarm()
		if lastAlarm != curAlarm {
			hits, lastTime := webStats.GetTotalHitsInWindow()
			fmt.Printf("alarm: %t, hits: %d, lastTime: %d\n", curAlarm, hits, lastTime)
		}
	}
}

// TODO don't alert twice
// make sure that last statistical thing is flushed and...
//   note that this is a behavioral assumption
