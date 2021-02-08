package webstats

import "fmt"

const totalHitsWindowSize = 121 // 2 minutes and 1 second

// WebStats keeps track of apache server log stats
type WebStats struct {
	window                [totalHitsWindowSize]uint64
	Sections              map[string]SectionStats
	isTotalTrafficAlerted bool
	totalTrafficThreshold int
	latestTime            uint64
	totalHits             uint64
}

// GetWindowSize returns length of window
func (ws *WebStats) WindowSize() int {
	return totalHitsWindowSize
}

// GetTotalHits returns the total hits within the window
func (ws *WebStats) TotalHits() uint64 {
	return ws.totalHits
}

func (ws *WebStats) setTotalHits(hits uint64) {
	ws.totalHits = hits
}

// GetHitsAtTime gets the hits at time provided
func (ws *WebStats) HitsAtTime(date uint64) uint64 {
	idx := date % uint64(ws.WindowSize())
	return ws.window[idx]
}

func (ws *WebStats) setHitsAtTime(date, val uint64) {
	idx := date % uint64(ws.WindowSize())
	ws.window[idx] = val
}

// GetLatestTime gets the latest time recorded
func (ws *WebStats) LatestTime() uint64 {
	return ws.latestTime
}

func (ws *WebStats) setLatestTime(date uint64) {
	ws.latestTime = date
}

// InitWebStats safely initializes WebStats
func InitWebStats(totalTrafficThreshold int) (WebStats, error) {
	if totalTrafficThreshold <= 0 {
		return WebStats{}, fmt.Errorf("%d is an invalid threshold", totalTrafficThreshold)
	}

	return WebStats{
		totalTrafficThreshold: totalTrafficThreshold,
		Sections:              map[string]SectionStats{},
	}, nil
}

// AddEntry adds an entry and updates statistics
func (ws *WebStats) AddEntry(sectionName string, timeInSeconds uint64) {
	section, ok := ws.Sections[sectionName]
	if !ok {
		section = SectionStats{
			latestTime: timeInSeconds,
		}
	}

	updateStats(&section, timeInSeconds)
	ws.Sections[sectionName] = section

	updateStats(ws, timeInSeconds)
}

// HasTotalTrafficAlarm returns whether alarm is alerted
func (ws *WebStats) HasTotalTrafficAlarm() bool {
	return int(ws.totalHits) > ws.totalTrafficThreshold*(totalHitsWindowSize-1)
}
