package webstats

const sectionHitsWindowSize = 10 // 10 seconds

// SectionStats keeps track of individual section stats
type SectionStats struct {
	window     [sectionHitsWindowSize]uint64
	latestTime uint64
	totalHits  uint64
}

// GetWindowSize returns length of window
func (s *SectionStats) WindowSize() int {
	return sectionHitsWindowSize
}

// GetTotalHits returns the total hits within the window
func (s *SectionStats) TotalHits() uint64 {
	return s.totalHits
}

func (s *SectionStats) setTotalHits(hits uint64) {
	s.totalHits = hits
}

func (s *SectionStats) setLatestTime(date uint64) {
	s.latestTime = date
}

// GetHitsAtTime gets the hits at time provided
func (s *SectionStats) HitsAtTime(date uint64) uint64 {
	idx := date % uint64(s.WindowSize())
	return s.window[idx]
}

func (s *SectionStats) setHitsAtTime(date, val uint64) {
	idx := date % uint64(s.WindowSize())
	s.window[idx] = val
}

// GetLatestTime gets the latest time recorded
func (s *SectionStats) LatestTime() uint64 {
	return s.latestTime
}
