package manage

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/hardboiled/apache-log-parser/webstats"
)

type Config struct {
	Interval       uint64 //
	WindowSize     uint
	AlarmThreshold uint
	InputFilepath  string
	OutputFilepath string
}

func InitConfig(interval, windowSize, alarmThreshold uint, inputFilepath, outputFilepath string) (Config, error) {
	errStrings := []string{}
	if interval < 1 {
		errStrings = append(errStrings, "interval cannot be < 1")
	}

	if windowSize < webstats.MinWindowSize {
		errStrings = append(errStrings, fmt.Sprintf("window-retention cannot be < %d", webstats.MinWindowSize))
	}

	if windowSize > webstats.MaxWindowSize {
		errStrings = append(errStrings, fmt.Sprintf("window-retention cannot be > %d bytes", webstats.MaxWindowSize))
	}

	if alarmThreshold < 1 {
		errStrings = append(errStrings, "alarmThreshold cannot be < 1")
	}

	if interval*2 > windowSize {
		errStrings = append(errStrings, "window must be able to hold at least two intervals")
	}

	if _, err := os.Stat(inputFilepath); os.IsNotExist(err) {
		errStrings = append(errStrings, fmt.Sprintf("input filepath %s does not exist", inputFilepath))
	}

	if outputFilepath != "" {
		if _, err := os.Stat(outputFilepath); os.IsNotExist(err) {
			errStrings = append(errStrings, fmt.Sprintf("input filepath %s does not exist", inputFilepath))
		}
	}

	var err error
	if len(errStrings) > 0 {
		err = errors.New(strings.Join(errStrings, "\n"))
	}

	return Config{
		Interval:       uint64(interval),
		WindowSize:     windowSize,
		AlarmThreshold: alarmThreshold,
		InputFilepath:  inputFilepath,
		OutputFilepath: outputFilepath,
	}, err
}
