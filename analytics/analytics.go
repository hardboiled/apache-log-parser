/*Package analytics takes signals from the primary process
 * and outputs relevant information based on the signals sent
 */
package analytics

import (
	"fmt"

	"github.com/hardboiled/apache-log-parser/webstats"
)

// ProcessAndOutputData processes the stats provided
type ProcessAndOutputData interface {
	Do()
}

// TotalHitsAlarm hello
type TotalHitsAlarm struct {
	Hits        uint64
	CurrentTime uint64
	Flag        bool
}

// Do prints alarms
func (th TotalHitsAlarm) Do() {
	if th.Flag {
		fmt.Printf("High traffic generated an alert - hits = %d, triggered at %d\n", th.Hits, th.CurrentTime)
		return
	}
	fmt.Printf("Recovered from high traffic alert - hits = %d, recovered at %d\n", th.Hits, th.CurrentTime)
}

// SectionData hello
type SectionData struct {
	LatestTime uint64
	Window     []webstats.WindowEntry
}

// Do output of section data
func (sd *SectionData) Do() {
	totalHitsPerSection := map[string]uint64{}
	numSections := 0

	for _, sectionsInTimeSlot := range sd.Window {
		for k, v := range sectionsInTimeSlot.Sections {
			prev := totalHitsPerSection[k]
			if prev == 0 {
				numSections++
			}

			totalHitsPerSection[k] = prev + v
		}
	}

	numberOfSectionsToPrint := 5
	if numSections < numberOfSectionsToPrint {
		numberOfSectionsToPrint = numSections
	}

	type topSection struct {
		name string
		hits uint64
	}
	topSectionsOrderedDesc := []topSection{}

	for i := 0; i < numberOfSectionsToPrint; i++ {
		curMax := uint64(0)
		curSection := ""
		for k, v := range totalHitsPerSection {
			if v >= curMax {
				curMax = v
				curSection = k
			}
		}
		delete(totalHitsPerSection, curSection)

		topSectionsOrderedDesc = append(topSectionsOrderedDesc, topSection{curSection, curMax})
	}

	fmt.Printf("Stats for time range %d - %d\n", sd.LatestTime-uint64(len(sd.Window)-1), sd.LatestTime)
	for _, v := range topSectionsOrderedDesc {
		fmt.Printf("\t %s -> hits: %d\n", v.name, v.hits)
	}
}

// ProcessStats runs calculations and prints results
func ProcessStats(ch chan ProcessAndOutputData) {
	for val := range ch {
		val.Do()
	}
}
