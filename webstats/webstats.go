/*Package webstats package defines datastructures to help record realtime
 * analytics for apache logs for later processing
 */
package webstats

import "fmt"

const twoMinutes = 120 // 2 minutes

// MinWindowSize is the smallest size that the WebStats window
//  can be initialized to.
const MinWindowSize = twoMinutes // 2 minutes

// MaxWindowSize is the max allowed retention of webStats
const MaxWindowSize = (1 << 20) // 1MB

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
	totalHitsForLast2Min  uint64
}

// WindowSize returns length of window
func (ws *WebStats) WindowSize() int {
	return len(ws.window)
}

// TotalHitsForLast2Min returns the total hits within the window
func (ws *WebStats) TotalHitsForLast2Min() uint64 {
	return ws.totalHitsForLast2Min
}

func (ws *WebStats) setTotalHitsForLast2Min(hits uint64) {
	ws.totalHitsForLast2Min = hits
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
func (ws *WebStats) GetWindowForRange(curTime uint64, timeRangeForLastSeconds uint64) []WindowEntry {
	beginRange := curTime - timeRangeForLastSeconds
	windowSize := uint64(len(ws.window))
	endIdx := (curTime + 1) % windowSize // endIdx is exclusive when mapping slices

	// for the timeRangeForLastSeconds, if it's 10, we'd want the last 10 seconds worth of time.
	//   If curTime is 11, we'd want seconds 2 - 11 (which would be all hits from the last 10 seconds)
	//   Thus, we have to add one to the beginRange to avoid including hits at time 1 here.
	beginIdx := (beginRange + 1) % windowSize

	if endIdx < beginIdx {
		return append(ws.window[beginIdx:], ws.window[:endIdx]...)
	}

	return ws.window[beginIdx:endIdx]
}

func (ws *WebStats) setLatestTime(date uint64) {
	ws.latestTime = date
}

// InitWebStats safely initializes WebStats
func InitWebStats(windowSize, totalTrafficThreshold uint, startTime uint64) (WebStats, error) {
	if totalTrafficThreshold == 0 {
		return WebStats{}, fmt.Errorf("%d is an invalid threshold", totalTrafficThreshold)
	}

	if windowSize < MinWindowSize {
		return WebStats{}, fmt.Errorf("%d is an invalid window size", windowSize)
	}

	return WebStats{
		totalTrafficThreshold: totalTrafficThreshold,
		window:                make([]WindowEntry, int(windowSize)),
		latestTime:            startTime,
	}, nil
}

// AddEntry adds an entry and updates statistics
func (ws *WebStats) AddEntry(sectionName string, timeInSeconds uint64) {
	ws.updateStats(timeInSeconds)
	curIdx := timeInSeconds % uint64(len(ws.window))
	if ws.window[curIdx].Sections == nil {
		ws.window[curIdx].Sections = map[string]uint64{}
	}
	section := ws.window[curIdx].Sections[sectionName]
	ws.window[curIdx].Sections[sectionName] = section + 1
}

// HasTotalTrafficAlarm returns whether alarm is alerted
func (ws *WebStats) HasTotalTrafficAlarm() bool {
	// To calcuate whether we've exceeded the average threshold allowed, we have to use multiplication
	//   of the threshold by the window size, since dividing integers can result in loss of data.
	return ws.totalHitsForLast2Min > uint64(ws.totalTrafficThreshold*twoMinutes)
}

func (ws *WebStats) updateStats(timeInSeconds uint64) {
	currentTotalHitsForLast2Min := ws.TotalHitsForLast2Min()
	hitsForCurrentTime := ws.HitsAtTime(timeInSeconds)

	if ws.LatestTime() < timeInSeconds {
		if ws.LatestTime() <= timeInSeconds-uint64(twoMinutes) {
			// if no hits have come in for the last two minutes, reset counter
			currentTotalHitsForLast2Min = 0
		} else {
			// subtract hits from 2 mins ago, since now we have a new latest time
			currentTotalHitsForLast2Min = currentTotalHitsForLast2Min - ws.HitsAtTime(timeInSeconds-twoMinutes)
		}

		// have to zero out entries in window for any gaps between latest time recorded and current time.
		//   Otherwise, stale calculations for the previous window could be left behind and cause future
		//   calculations to be wrong.
		for i := ws.LatestTime() + 1; i < timeInSeconds; i++ {
			ws.setHitsAtTime(i, 0)
		}

		hitsForCurrentTime = 0
		ws.setLatestTime(timeInSeconds)
	}

	ws.setHitsAtTime(timeInSeconds, hitsForCurrentTime+1)
	ws.setTotalHitsForLast2Min(currentTotalHitsForLast2Min + 1)
}
