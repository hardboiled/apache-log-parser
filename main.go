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
		fmt.Printf("error opening input file: %v\n", err)
		os.Exit(1)
	}

	inputCh := make(chan parsing.WebServerLogData, 100)

	webStats, err := webstats.InitWebStats(10)
	if err != nil {
		fmt.Printf("error initializing webstats: %v\n", err)
		os.Exit(1)
	}

	go parsing.ParseWebServerLogDataWithChannel(reader, inputCh)

	for data := range inputCh {
		lastAlarm := webStats.HasTotalTrafficAlarm()
		webStats.AddEntry(data.GetRequestSection(), data.Date)
		curAlarm := webStats.HasTotalTrafficAlarm()
		if lastAlarm != curAlarm {
			fmt.Printf("alarm: %t, hits: %d, lastTime: %d\n", curAlarm, webStats.GetTotalHits(), webStats.GetLatestTime())
		}
		if data.Date%10 == 0 {
			fmt.Println("Sections:")
			for k, v := range webStats.Sections {
				fmt.Printf("%s -> hits: %d, latestTime: %d\n", k, v.GetTotalHits(), v.GetLatestTime())
			}
		}
	}
}

// TODO don't alert twice
// make sure that last statistical thing is flushed and...
//   note that this is a behavioral assumption
