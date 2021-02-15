package webstats_test

import (
	"testing"

	"github.com/hardboiled/apache-log-parser/webstats"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestScheduleInterval(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "WebStats Suite")
}

var _ = Describe("WebStats", func() {
	var ws webstats.WebStats
	var startTime uint64

	BeforeEach(func() {
		startTime = 1549574340
		ws, _ = webstats.InitWebStats(120, 10, startTime)
	})

	It("adds entry to right section", func() {
		section1 := "a"
		section2 := "b"
		ws.AddEntry(section1, startTime)
		ws.AddEntry(section1, startTime)
		ws.AddEntry(section1, startTime)

		ws.AddEntry(section2, startTime)
		ws.AddEntry(section2, startTime)

		entry := ws.GetWindowForRange(startTime, 0)[0]
		Expect(ws.HitsAtTime(startTime)).To(Equal(uint64(5)))

		Expect(entry.Sections[section1]).To(Equal(uint64(3)))
		Expect(entry.Sections[section2]).To(Equal(uint64(2)))
	})

	It("triggers 2 min alarm properly", func() {
		localWs, _ := webstats.InitWebStats(120, 2, startTime)
		section1 := "a"
		section2 := "b"

		for i := int64(0); i < 120; i++ {
			localWs.AddEntry(section1, startTime)
			localWs.AddEntry(section2, startTime)
		}

		Expect(localWs.HasTotalTrafficAlarm()).To(Equal(false))
		localWs.AddEntry(section2, startTime)
		Expect(localWs.HasTotalTrafficAlarm()).To(Equal(true))

		// add new section
		localWs.AddEntry(section2, startTime+uint64(120))
		Expect(localWs.HasTotalTrafficAlarm()).To(Equal(false))
	})
})
