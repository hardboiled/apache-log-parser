package analytics

const bufferForOverlappingLogTimes = uint64(5)

// ScheduleInterval helps determine when to output interval statistics, specifically `SectionData`
type ScheduleInterval struct {
	lastTimeProcessed uint64
	secondsAgo        uint64
	timeToProcess     uint64
}

// ReadyToProcess returns true if the interval should be outputted
func (sp *ScheduleInterval) ReadyToProcess(nextDate uint64) bool {
	// shouldSchedule works in concert with shouldProcess. An interval is scheduled to be processed
	// based on the nextDate, and then after we've read in a few more lines based on the
	// `bufferForOverlappingLogTimes`, we assume that the interval has been completely read in and
	// then it's available to be processed.
	if !sp.isScheduled() && sp.shouldSchedule(nextDate) {
		sp.schedule(nextDate)
	}

	return sp.shouldProcess(nextDate)
}

// TimeToProcess getter for `timeToProcess`
func (sp *ScheduleInterval) TimeToProcess() uint64 {
	return sp.timeToProcess
}

// SecondsAgo getter for `secondsAgo`
func (sp *ScheduleInterval) SecondsAgo() uint64 {
	return sp.secondsAgo
}

// LastTimeProcessed getter for `lastTimeProcessed`
func (sp *ScheduleInterval) LastTimeProcessed() uint64 {
	return sp.lastTimeProcessed
}

// MarkAsProcessed should be called after interval has been processed
func (sp *ScheduleInterval) MarkAsProcessed() {
	sp.lastTimeProcessed = sp.timeToProcess
	sp.timeToProcess = 0
}

// InitScheduleInterval initializes the ScheduleInterval structure properly
func InitScheduleInterval(startTime, interval uint64) ScheduleInterval {
	return ScheduleInterval{
		// we return a `lastTimeProcessed` as startTime - bufferForOverlappingLogTimes
		// because apache log lines are not guaranteed to be printed exactly in order by time
		// so we backtrack a few seconds to avoid missing older log lines that are after the
		// first line in the input
		lastTimeProcessed: startTime - bufferForOverlappingLogTimes,
		secondsAgo:        interval - 1,
	}
}

func (sp *ScheduleInterval) shouldProcess(nextDate uint64) bool {
	return sp.isScheduled() && nextDate > sp.timeToProcess+bufferForOverlappingLogTimes
}

func (sp *ScheduleInterval) shouldSchedule(nextDate uint64) bool {
	return nextDate > sp.secondsAgo+sp.lastTimeProcessed
}

func (sp *ScheduleInterval) schedule(scheduleDate uint64) {
	sp.timeToProcess = scheduleDate
}

func (sp *ScheduleInterval) isScheduled() bool {
	return sp.timeToProcess > 0
}
