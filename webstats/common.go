package webstats

type localStats interface {
	WindowSize() int
	HitsAtTime(uint64) uint64
	setHitsAtTime(uint64, uint64)
	TotalHits() uint64
	setTotalHits(uint64)
	LatestTime() uint64
	setLatestTime(uint64)
}

func updateStats(stats localStats, timeInSeconds uint64) {
	hitsAtCurrentTime := stats.HitsAtTime(timeInSeconds)
	latestTime := stats.LatestTime()
	currentTotalHits := stats.TotalHits()

	if stats.LatestTime() < timeInSeconds {
		if latestTime < timeInSeconds-uint64(stats.WindowSize()) {
			currentTotalHits = 0
		} else {
			currentTotalHits = currentTotalHits - hitsAtCurrentTime
		}
		hitsAtCurrentTime = 0
		latestTime = timeInSeconds
	}

	stats.setHitsAtTime(timeInSeconds, hitsAtCurrentTime+1)
	stats.setTotalHits(currentTotalHits + 1)
	stats.setLatestTime(latestTime)
}
