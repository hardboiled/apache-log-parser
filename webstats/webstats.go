/*Package webstats package defines datastructures to help record realtime
 * analytics for apache logs for later processing
 */
package webstats

import "fmt"

const MinWindowSize = 120 // 2 minutes and 1 second

const alarmWindow = MinWindowSize + 1

// WindowEntry holds section data and total hits for a given time entry
type WindowEntry struct {
	Sections             map[string]uint64
	TotalHitsForTimeSlot uint64
}

// WebStats keeps track of apache server log stats
type WebStats struct {
	window                []WindowEntry
	isTotalTrafficAlerted bool
	totalTrafficThreshold uint
	latestTime            uint64
	totalHits             uint64
}

// WindowSize returns length of window
func (ws *WebStats) WindowSize() int {
	return len(ws.window)
}

// TotalHits returns the total hits within the window
func (ws *WebStats) TotalHits() uint64 {
	return ws.totalHits
}

func (ws *WebStats) setTotalHits(hits uint64) {
	ws.totalHits = hits
}

// HitsAtTime gets the hits at time provided
func (ws *WebStats) HitsAtTime(date uint64) uint64 {
	idx := date % uint64(ws.WindowSize())
	return ws.window[idx].TotalHitsForTimeSlot
}

func (ws *WebStats) setHitsAtTime(date, val uint64) {
	idx := date % uint64(ws.WindowSize())
	ws.window[idx].TotalHitsForTimeSlot = val
}

// LatestTime gets the latest time recorded
func (ws *WebStats) LatestTime() uint64 {
	return ws.latestTime
}

// GetWindowForRange returns the window for begin and end inclusive
func (ws *WebStats) GetWindowForRange(begin, end uint64) []WindowEntry {
	beginIdx := (begin + 1) % alarmWindow
	endIdx := (end + 1) % alarmWindow

	if endIdx < beginIdx {
		return append(ws.window[beginIdx:], ws.window[:endIdx]...)
	}

	return ws.window[beginIdx:endIdx]
}

func (ws *WebStats) setLatestTime(date uint64) {
	ws.latestTime = date
}

// InitWebStats safely initializes WebStats
func InitWebStats(windowSize, totalTrafficThreshold uint) (WebStats, error) {
	if totalTrafficThreshold == 0 {
		return WebStats{}, fmt.Errorf("%d is an invalid threshold", totalTrafficThreshold)
	}

	if windowSize < MinWindowSize {
		return WebStats{}, fmt.Errorf("%d is an invalid window size", windowSize)
	}

	return WebStats{
		totalTrafficThreshold: totalTrafficThreshold,
		window:                make([]WindowEntry, windowSize+1),
	}, nil
}

// AddEntry adds an entry and updates statistics
func (ws *WebStats) AddEntry(sectionName string, timeInSeconds uint64) {
	ws.updateStats(timeInSeconds)
	if ws.window[timeInSeconds%alarmWindow].Sections == nil {
		ws.window[timeInSeconds%alarmWindow].Sections = map[string]uint64{}
	}
	section := ws.window[timeInSeconds%alarmWindow].Sections[sectionName]
	ws.window[timeInSeconds%alarmWindow].Sections[sectionName] = section + 1
}

// HasTotalTrafficAlarm returns whether alarm is alerted
func (ws *WebStats) HasTotalTrafficAlarm() bool {
	return ws.totalHits > uint64(ws.totalTrafficThreshold)*uint64(alarmWindow-1)
}

func (ws *WebStats) updateStats(timeInSeconds uint64) {
	hitsAtCurrentTime := ws.HitsAtTime(timeInSeconds)
	latestTime := ws.LatestTime()
	currentTotalHits := ws.TotalHits()

	if ws.LatestTime() < timeInSeconds {
		if latestTime < timeInSeconds-uint64(ws.WindowSize()) {
			currentTotalHits = 0
		} else {
			currentTotalHits = currentTotalHits - hitsAtCurrentTime
		}
		hitsAtCurrentTime = 0
		latestTime = timeInSeconds
	}

	ws.setHitsAtTime(timeInSeconds, hitsAtCurrentTime+1)
	ws.setTotalHits(currentTotalHits + 1)
	ws.setLatestTime(latestTime)
}
