package main

const bufferForOverlappingLogTimes = uint64(5)

type scheduleInterval struct {
	lastTimeProcessed uint64
	secondsAgo        uint64
	timeToProcess     uint64
}

func initScheduleInterval(startTime, interval uint64) scheduleInterval {
	return scheduleInterval{
		lastTimeProcessed: startTime - bufferForOverlappingLogTimes,
		secondsAgo:        interval - 1,
	}
}

func (sp *scheduleInterval) shouldProcess(nextDate uint64) bool {
	return sp.isScheduled() && nextDate > sp.timeToProcess+bufferForOverlappingLogTimes
}

func (sp *scheduleInterval) shouldSchedule(nextDate uint64) bool {
	return nextDate > sp.secondsAgo+sp.lastTimeProcessed
}

func (sp *scheduleInterval) schedule(scheduleDate uint64) {
	sp.timeToProcess = scheduleDate
}

func (sp *scheduleInterval) isScheduled() bool {
	return sp.timeToProcess > 0
}

func (sp *scheduleInterval) markAsProcessed() {
	sp.lastTimeProcessed = sp.timeToProcess
	sp.timeToProcess = 0
}
