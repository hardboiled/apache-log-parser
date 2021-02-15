package main

const bufferForOverlappingLogTimes = uint64(5)

type scheduleInterval struct {
	lastTimeProcessed uint64
	interval          uint64
	timeToProcess     uint64
}

func initScheduleInterval(startTime, interval uint64) scheduleInterval {
	return scheduleInterval{
		lastTimeProcessed: startTime,
		interval:          interval,
	}
}

func (sp *scheduleInterval) shouldProcess(nextDate uint64) bool {
	return nextDate-sp.timeToProcess > sp.interval+bufferForOverlappingLogTimes
}

func (sp *scheduleInterval) shouldSchedule(nextDate uint64) bool {
	return nextDate-sp.lastTimeProcessed > sp.interval
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
