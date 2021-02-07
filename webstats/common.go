package webstats

type localStats interface {
	GetWindowSize() int
	GetHitsAtTime(uint64) uint64
	setHitsAtTime(uint64, uint64)
	GetTotalHits() uint64
	setTotalHits(uint64)
	GetLatestTime() uint64
	setLatestTime(uint64)
}

func updateStats(stats localStats, timeInSeconds uint64) {
	hitsAtCurrentTime := stats.GetHitsAtTime(timeInSeconds)
	latestTime := stats.GetLatestTime()
	currentTotalHits := stats.GetTotalHits()

	if stats.GetLatestTime() < timeInSeconds {
		if latestTime < timeInSeconds-uint64(stats.GetWindowSize()) {
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
