package analytics_test

import (
	"testing"

	"github.com/hardboiled/apache-log-parser/analytics"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestScheduleInterval(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "ScheduleInterval Suite")
}

var _ = Describe("InitScheduleInterval", func() {
	It("fails for numbers below buffer lines", func() {
		_, err := analytics.InitScheduleInterval(uint64(analytics.BufferForOverlappingLogTimes-1), uint64(analytics.BufferForOverlappingLogTimes))
		Expect(err).ToNot(BeNil())
	})

	It("fails for interval below buffer lines", func() {
		_, err := analytics.InitScheduleInterval(uint64(analytics.BufferForOverlappingLogTimes), uint64(analytics.BufferForOverlappingLogTimes-1))
		Expect(err).ToNot(BeNil())
	})

	It("succeeds on normal params", func() {
		_, err := analytics.InitScheduleInterval(uint64(analytics.BufferForOverlappingLogTimes), uint64(analytics.BufferForOverlappingLogTimes))
		Expect(err).To(BeNil())
	})
})

var _ = Describe("ScheduleInterval", func() {
	It("schedules intervals regularly", func() {
		interval := uint64(8)
		startTime := uint64(analytics.BufferForOverlappingLogTimes)
		si, _ := analytics.InitScheduleInterval(startTime, interval)

		// add 1 for first time, since InitScheduleInterval has already marked startTime as lastTimeProcessed.
		// Thus, we need to advance the counter to the next time to start the interval
		firstTime := startTime + 1
		firstIteration := uint64(0)
		for !si.ReadyToProcess(firstTime) { // get past first process
			firstTime++
			firstIteration++
		}
		si.MarkAsProcessed()

		Expect(firstIteration).To(Equal(interval))

		secondTime := firstTime
		secondIteration := uint64(0)
		for !si.ReadyToProcess(secondTime) { // get past first process
			secondTime++
			secondIteration++
		}

		Expect(secondIteration).To(Equal(interval))
	})
})
