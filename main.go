package main

import (
	"fmt"
	"os"

	"github.com/hardboiled/apache-log-parser/analytics"
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
	outputCh := make(chan analytics.ProcessAndOutputData)
	defer close(outputCh)

	webStats, err := webstats.InitWebStats(10)
	if err != nil {
		fmt.Printf("error initializing webstats: %v\n", err)
		os.Exit(1)
	}

	go parsing.ParseWebServerLogDataWithChannel(reader, inputCh)
	go analytics.ProcessStats(outputCh)

	for data := range inputCh {
		lastAlarm := webStats.HasTotalTrafficAlarm()
		webStats.AddEntry(data.RequestSection(), data.Date)
		curAlarm := webStats.HasTotalTrafficAlarm()
		if lastAlarm != curAlarm {
			outputCh <- analytics.TotalHitsAlarm{Flag: curAlarm, Hits: webStats.TotalHits(), CurrentTime: webStats.LatestTime()}
		}
		if data.Date%10 == 0 {
			outputCh <- &analytics.SectionData{
				LatestTime: webStats.LatestTime(),
				Window:     webStats.GetWindowForRange(data.Date-9, data.Date),
			}
		}
	}

}
