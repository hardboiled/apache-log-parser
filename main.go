package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"sync"

	"github.com/hardboiled/apache-log-parser/analytics"
	"github.com/hardboiled/apache-log-parser/manage"
	"github.com/hardboiled/apache-log-parser/parsing"
	"github.com/hardboiled/apache-log-parser/webstats"
)

const (
	defaultInterval       = 10
	defaultWindowSize     = webstats.MinWindowSize * 2
	defaultAlarmThreshold = 10
	defaultInputFilepath  = "./input_files/sample_csv.txt"
)

func getFlags() (manage.Config, error) {
	interval := flag.Uint("interval", defaultInterval, "integer in seconds")
	windowSize := flag.Uint("window-retention", defaultWindowSize, fmt.Sprintf("integer in seconds (min %d)", defaultWindowSize))
	alarmThreshold := flag.Uint("alarm-threshold", defaultAlarmThreshold, "triggers alarm on request/per second over 2 mins")
	inputFilepath := flag.String("input-filepath", defaultInputFilepath, "file path to read in")

	flag.Parse()

	return manage.InitConfig(*interval, *windowSize, *alarmThreshold, *inputFilepath)
}

func setupBuffers(config manage.Config) (io.ReadCloser, chan parsing.WebServerLogData, chan analytics.ProcessAndOutputData, error) {
	reader, err := os.Open(config.InputFilepath)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("error opening input file: %v", err)
	}

	inputCh := make(chan parsing.WebServerLogData, 100)
	outputCh := make(chan analytics.ProcessAndOutputData)

	return reader, inputCh, outputCh, nil
}

func main() {
	config, err := getFlags()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	reader, inputCh, outputCh, err := setupBuffers(config)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// setup go routines and channels
	var wg sync.WaitGroup
	defer reader.Close()
	defer close(outputCh)

	// Note: `ParseWebServerLogDataWithChannel` closes the inputCh when finished
	go parsing.ParseWebServerLogDataWithChannel(reader, inputCh)
	go analytics.ProcessStats(outputCh, os.Stdout, &wg)

	// read in first entry to initialize
	firstEntry := <-inputCh
	webStats, err := webstats.InitWebStats(config.WindowSize, config.AlarmThreshold, firstEntry.Date)
	if err != nil {
		fmt.Printf("error initializing webstats: %v\n", err)
		os.Exit(1)
	}
	webStats.AddEntry(firstEntry.RequestSection(), firstEntry.Date)
	scheduleInterval, err := analytics.InitScheduleInterval(firstEntry.Date, config.Interval)
	if err != nil {
		fmt.Printf("error initializing scheduleInterval: %v\n", err)
		os.Exit(1)
	}

	// main loop
	for data := range inputCh {
		// Check to print interval
		if scheduleInterval.ReadyToProcess(data.Date) {
			wg.Add(1)
			outputCh <- &analytics.SectionData{
				LatestTime: scheduleInterval.TimeToProcess(),
				Window:     webStats.GetWindowForRange(scheduleInterval.TimeToProcess(), scheduleInterval.SecondsAgo()),
			}
			scheduleInterval.MarkAsProcessed()
		}

		// Compare alarm state from previous entry, if different, print alarm status
		lastAlarm := webStats.HasTotalTrafficAlarm()
		webStats.AddEntry(data.RequestSection(), data.Date)
		curAlarm := webStats.HasTotalTrafficAlarm()
		if lastAlarm != curAlarm {
			wg.Add(1)
			outputCh <- analytics.TotalHitsAlarm{Flag: curAlarm, Hits: webStats.TotalHitsForLast2Min(), CurrentTime: webStats.LatestTime()}
		}
	}

	// Flush remaining results
	if scheduleInterval.LastTimeProcessed() < webStats.LatestTime() {
		numSecondsLeft := webStats.LatestTime() - scheduleInterval.LastTimeProcessed()
		if numSecondsLeft > config.Interval {
			numSecondsLeft = config.Interval // if lastTimeProcessed was greater than the interval, only process the last interval
		}

		wg.Add(1)
		outputCh <- &analytics.SectionData{
			LatestTime: webStats.LatestTime(),
			Window:     webStats.GetWindowForRange(webStats.LatestTime(), numSecondsLeft),
		}
	}

	wg.Wait()
}
