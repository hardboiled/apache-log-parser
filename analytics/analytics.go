/*Package analytics takes signals from the primary process
 * and outputs relevant information based on the signals sent
 */
package analytics

import (
	"fmt"
	"io"
	"strings"
	"sync"

	"github.com/hardboiled/apache-log-parser/webstats"
)

// ProcessAndOutputData processes the stats provided
type ProcessAndOutputData interface {
	Do(writer io.Writer)
}

// TotalHitsAlarm hello
type TotalHitsAlarm struct {
	Hits        uint64
	CurrentTime uint64
	Flag        bool
}

// Do prints alarms
func (th TotalHitsAlarm) Do(writer io.Writer) {
	fmtStr := "Recovered from high traffic alert - hits = %d, recovered at %d\n"
	if th.Flag {
		fmtStr = "High traffic generated an alert - hits = %d, triggered at %d\n"
	}

	_, err := writer.Write([]byte(fmt.Sprintf(fmtStr, th.Hits, th.CurrentTime)))
	if err != nil {
		fmt.Printf("Error when writing to output: %v\n", err)
	}
}

// SectionData hello
type SectionData struct {
	LatestTime uint64
	Window     []webstats.WindowEntry
}

// Do output of section data
func (sd *SectionData) Do(writer io.Writer) {
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

	output := []string{}

	output = append(
		output,
		fmt.Sprintf("Stats for time range %d - %d", sd.LatestTime-uint64(len(sd.Window)-1), sd.LatestTime),
	)

	totalHitsForWindow := uint64(0)
	for _, v := range sd.Window {
		totalHitsForWindow = totalHitsForWindow + v.TotalHitsForTimeSlot
	}

	output = append(output, fmt.Sprintf("\ttotal hits for this window %d", totalHitsForWindow))

	for _, v := range topSectionsOrderedDesc {
		output = append(output, fmt.Sprintf("\t %s -> hits: %d", v.name, v.hits))
	}

	result := strings.Join(output, "\n") + "\n"
	_, err := writer.Write([]byte(result))
	if err != nil {
		fmt.Printf("Error when writing to output: %v\n", err)
	}
}

// ProcessStats runs calculations and prints results
func ProcessStats(ch chan ProcessAndOutputData, writer io.Writer, wg *sync.WaitGroup) {
	for val := range ch {
		val.Do(writer)
		wg.Done()
	}
}
