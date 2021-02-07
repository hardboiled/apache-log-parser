package webstats

import "fmt"

const windowSize = 121 // 2 minutes and 1 second

// WebStats keeps track of apache server log stats
type WebStats struct {
	window                [windowSize]uint64
	SectionStats          map[string]uint64
	isTotalTrafficAlerted bool
	totalTrafficThreshold int
	latestTime            uint64
	endingTotalHits       uint64
}

// InitWebStats safely initializes WebStats
func InitWebStats(totalTrafficThreshold int) (WebStats, error) {
	if totalTrafficThreshold <= 0 {
		return WebStats{}, fmt.Errorf("%d is an invalid threshold", totalTrafficThreshold)
	}

	return WebStats{
		totalTrafficThreshold: totalTrafficThreshold,
		SectionStats:          map[string]uint64{},
	}, nil
}

// AddEntry adds an entry and updates statistics
func (ws *WebStats) AddEntry(_section string, timeInSeconds uint64) {
	idx := int(timeInSeconds % windowSize)
	if ws.latestTime < timeInSeconds {
		ws.endingTotalHits = ws.endingTotalHits - ws.window[idx]
		ws.window[idx] = 1
		ws.latestTime = timeInSeconds
	} else {
		ws.window[idx] = ws.window[idx] + 1
	}

	ws.endingTotalHits++
}

// HasTotalTrafficAlarm returns whether alarm is alerted
func (ws *WebStats) HasTotalTrafficAlarm() bool {
	return int(ws.endingTotalHits) > ws.totalTrafficThreshold*(windowSize-1)
}

// GetTotalHitsInWindow returns total hits for window
func (ws *WebStats) GetTotalHitsInWindow() (uint64, uint64) {
	return ws.endingTotalHits, ws.latestTime
}

func (ws *WebStats) fillDifference(lastIdx, diff int) {
	for i := 0; i < diff; i++ {
		curIdx := (i + lastIdx) % windowSize
		ws.window[i] = ws.window[lastIdx]
		lastIdx = curIdx
	}
}

func (ws *WebStats) getStartingIdx(lastIdx int) int {
	return (lastIdx + windowSize + 1) % windowSize
}
