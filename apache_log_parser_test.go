package main_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestApacheLogParser(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "ApacheLogParser Suite")
}
