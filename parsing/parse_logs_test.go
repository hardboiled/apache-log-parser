package parsing_test

import (
	"fmt"
	"testing"

	"github.com/hardboiled/apache-log-parser/parsing"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestParseLogs(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "ParseLogs Suite")
}

var _ = Describe("WebServerLogData", func() {
	It("returns correcet RequestSection", func() {
		fmtStr := "GET %s HTTP/1.0"
		ld := parsing.WebServerLogData{
			Request: fmt.Sprintf(fmtStr, "/api"),
		}

		Expect(ld.RequestSection()).To(Equal("/api"))
		ld.Request = fmt.Sprintf(fmtStr, "/some-other-section")
		Expect(ld.RequestSection()).To(Equal("/some-other-section"))
		ld.Request = fmt.Sprintf(fmtStr, "/api/bye/")
		Expect(ld.RequestSection()).To(Equal("/api"))
		ld.Request = fmt.Sprintf(fmtStr, "/some-other-section/hello/2")
		Expect(ld.RequestSection()).To(Equal("/some-other-section"))
	})
})
