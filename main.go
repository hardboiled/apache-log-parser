package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/hardboiled/apache-log-parser/analytics"
	"github.com/hardboiled/apache-log-parser/parsing"
	"github.com/hardboiled/apache-log-parser/webstats"
)

const (
	defaultInterval       = 10
	defaultWindowSize     = 120
	defaultAlarmThreshold = 10
)

func getFlags() (uint, uint, uint, error) {
	interval := flag.Uint("interval", defaultInterval, "integer in seconds")
	windowSize := flag.Uint("window-retention", defaultWindowSize, fmt.Sprintf("integer in seconds (min %d)", defaultWindowSize))
	alarmThreshold := flag.Uint("alarm-threshold", 1<<10, "triggers alarm on request/per second over 2 mins")
	errStrings := []string{}

	flag.Parse()

	if *interval < 1 {
		errStrings = append(errStrings, "interval cannot be < 1")
	}

	if *windowSize < webstats.MinWindowSize {
		errStrings = append(errStrings, fmt.Sprintf("window-retention cannot be < %d", defaultWindowSize))
	}

	if *alarmThreshold < 1 {
		errStrings = append(errStrings, "alarmThreshold cannot be < 1")
	}

	var err error
	if len(errStrings) > 0 {
		err = errors.New(strings.Join(errStrings, "\n"))
	}

	return *interval, *windowSize, *alarmThreshold, err
}

func main() {
	interval, windowSize, threshold, err := getFlags()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	reader, err := os.Open("./input_files/sample_csv.txt")
	if err != nil {
		fmt.Printf("error opening input file: %v\n", err)
		os.Exit(1)
	}

	inputCh := make(chan parsing.WebServerLogData, 100)
	outputCh := make(chan analytics.ProcessAndOutputData)
	defer close(outputCh)

	webStats, err := webstats.InitWebStats(windowSize, threshold)
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
		if data.Date%uint64(interval) == 0 {
			outputCh <- &analytics.SectionData{
				LatestTime: webStats.LatestTime(),
				Window:     webStats.GetWindowForRange(data.Date-9, data.Date),
			}
		}
	}

}
