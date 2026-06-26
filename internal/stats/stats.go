package stats

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"
)

const DateLayout = "2006-01-02"

type StatList struct {
	Stats []Stat `json:"stats"`
}

type Stat struct {
	Date        string `json:"date"`
	PresetID    string `json:"presetId"`
	DurationMin int    `json:"duration_min"`
	Count       int    `json:"count"`
}

type statGroup struct{ count, duration int }

func SaveStat(stat Stat, currentStats []Stat) []Stat {
	for i, s := range currentStats {
		if stat.Date == s.Date && stat.PresetID == s.PresetID {
			currentStats[i].DurationMin += stat.DurationMin
			currentStats[i].Count++
			return currentStats
		}
	}

	return append(currentStats, stat)
}

func PrintStats(date string, stats []Stat) {
	cleanedDate := strings.ReplaceAll(date, " ", "")

	if len(stats) == 0 {
		log.Fatal("No stats has been made yet")
	}

	if cleanedDate == "today" {
		count := 0
		duration := 0
		statsToday := map[string]statGroup{}
		dateNow := time.Now().Format(DateLayout)
		for _, stat := range stats {
			if dateNow == stat.Date {
				current := statsToday[stat.PresetID]

				statsToday[stat.PresetID] = statGroup{
					count:    current.count + stat.Count,
					duration: current.duration + stat.DurationMin,
				}
				count += stat.Count
				duration += stat.DurationMin
			}
		}

		if count == 0 && duration == 0 {
			log.Fatal("No entries for that date.")
		}

		fmt.Printf("Today (%s): %d pomodoros, %d min\n", dateNow, count, duration)
		for k, v := range statsToday {
			fmt.Printf("- %s: %d pomodoros, %d min\n", k, v.count, v.duration)
		}
		os.Exit(0)
	}

	parsedTime, err := time.Parse(DateLayout, cleanedDate)
	if err != nil {
		log.Fatalf("pomodoro-cli: invalid date '%s' (expected format: YYYY-MM-DD)\n", cleanedDate)
	}

	count := 0
	duration := 0
	statsFromDate := map[string]statGroup{}
	for _, stat := range stats {
		if parsedTime.Format(DateLayout) == stat.Date {
			current := statsFromDate[stat.PresetID]

			statsFromDate[stat.PresetID] = statGroup{
				count:    current.count + stat.Count,
				duration: current.duration + stat.DurationMin,
			}
			count += stat.Count
			duration += stat.DurationMin
		}
	}

	if count == 0 && duration == 0 {
		log.Fatal("No entries for that date.")
	}

	fmt.Printf("%s: %d pomodoros, %d min\n", parsedTime.Format(DateLayout), count, duration)
	for k, v := range statsFromDate {
		fmt.Printf("- %s: %d pomodoros, %d min\n", k, v.count, v.duration)
	}
	os.Exit(0)
}
