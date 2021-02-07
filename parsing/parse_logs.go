package parsing

import (
	"io"
	"strings"

	"github.com/gocarina/gocsv"
)

// WebServerLogData represents one line from a csv (only export used lines)
type WebServerLogData struct {
	RemoteHost string `csv:"remotehost"`
	Rfc931     string `csv:"rfc931"`
	AuthUser   string `csv:"authuser"`
	Date       uint64 `csv:"date"`
	Request    string `csv:"request"`
	Status     uint64 `csv:"status"`
	Bytes      uint64 `csv:"bytes"`
}

// ParseWebServerLogDataWithChannel will send back each parsed line through the provided channel
func ParseWebServerLogDataWithChannel(stream io.ReadCloser, c chan WebServerLogData) error {
	return gocsv.UnmarshalToChan(stream, c)
}

// GetRequestSection takes the request and finds the section associated with it
func (ld *WebServerLogData) GetRequestSection() string {
	endIdx := strings.IndexAny(ld.Request[1:], " /") + 1
	return ld.Request[:endIdx]
}
