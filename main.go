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
	defaultWindowSize     = webstats.MinWindowSize * 2
	defaultAlarmThreshold = 10
	defaultInputFilepath  = "./input_files/sample_csv.txt"
)

func getFlags() (uint64, uint, uint, string, error) {
	interval := flag.Uint64("interval", defaultInterval, "integer in seconds")
	windowSize := flag.Uint("window-retention", defaultWindowSize, fmt.Sprintf("integer in seconds (min %d)", defaultWindowSize))
	alarmThreshold := flag.Uint("alarm-threshold", defaultAlarmThreshold, "triggers alarm on request/per second over 2 mins")
	inputFilepath := flag.String("input-filepath", defaultInputFilepath, "triggers alarm on request/per second over 2 mins")
	errStrings := []string{}

	flag.Parse()

	if *interval < 1 {
		errStrings = append(errStrings, "interval cannot be < 1")
	}

	if *windowSize < webstats.MinWindowSize {
		errStrings = append(errStrings, fmt.Sprintf("window-retention cannot be < %d", webstats.MinWindowSize))
	}

	if *windowSize > webstats.MaxWindowSize {
		errStrings = append(errStrings, fmt.Sprintf("window-retention cannot be > %d bytes", webstats.MaxWindowSize))
	}

	if *alarmThreshold < 1 {
		errStrings = append(errStrings, "alarmThreshold cannot be < 1")
	}

	if *interval*2 > uint64(*windowSize) {
		errStrings = append(errStrings, "window must be able to hold at least two intervals")
	}

	if _, err := os.Stat(*inputFilepath); os.IsNotExist(err) {
		errStrings = append(errStrings, fmt.Sprintf("input filepath %s does not exist", *inputFilepath))
	}

	var err error
	if len(errStrings) > 0 {
		err = errors.New(strings.Join(errStrings, "\n"))
	}

	return *interval, *windowSize, *alarmThreshold, *inputFilepath, err
}

func main() {
	interval, windowSize, threshold, inputFilepath, err := getFlags()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	reader, err := os.Open(inputFilepath)
	if err != nil {
		fmt.Printf("error opening input file: %v\n", err)
		os.Exit(1)
	}

	inputCh := make(chan parsing.WebServerLogData, 100)
	outputCh := make(chan analytics.ProcessAndOutputData)
	defer close(outputCh)

	go parsing.ParseWebServerLogDataWithChannel(reader, inputCh)
	go analytics.ProcessStats(outputCh)

	firstEntry := <-inputCh
	webStats, err := webstats.InitWebStats(windowSize, threshold, firstEntry.Date)
	if err != nil {
		fmt.Printf("error initializing webstats: %v\n", err)
		os.Exit(1)
	}
	webStats.AddEntry(firstEntry.RequestSection(), firstEntry.Date)
	scheduleProcess := initScheduleInterval(firstEntry.Date-1, interval)

	for data := range inputCh {
		if !scheduleProcess.isScheduled() && scheduleProcess.shouldSchedule(data.Date) {
			scheduleProcess.schedule(data.Date)
		}

		if scheduleProcess.shouldProcess(data.Date) {
			outputCh <- &analytics.SectionData{
				LatestTime: scheduleProcess.timeToProcess,
				Window:     webStats.GetWindowForRange(scheduleProcess.timeToProcess, scheduleProcess.secondsAgo),
			}
			scheduleProcess.markAsProcessed()
		}

		lastAlarm := webStats.HasTotalTrafficAlarm()
		webStats.AddEntry(data.RequestSection(), data.Date)
		curAlarm := webStats.HasTotalTrafficAlarm()
		if lastAlarm != curAlarm {
			outputCh <- analytics.TotalHitsAlarm{Flag: curAlarm, Hits: webStats.TotalHitsForLast2Min(), CurrentTime: webStats.LatestTime()}
		}
	}

	if scheduleProcess.lastTimeProcessed < webStats.LatestTime() {
		// Flush remaining
		numSecondsLeft := webStats.LatestTime() - scheduleProcess.lastTimeProcessed
		if numSecondsLeft > interval {
			numSecondsLeft = interval // if lastTimeProcessed was greater than the interval, only process the last interval
		}

		outputCh <- &analytics.SectionData{
			LatestTime: webStats.LatestTime(),
			Window:     webStats.GetWindowForRange(webStats.LatestTime(), numSecondsLeft),
		}
	}
}
