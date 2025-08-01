package train

import (
	"fmt"
	"strings"
	"time"
)

// formatSchedule formats the Azure API response
func formatSchedule(schedule ScheduleResponse) string {
	var result string
	count := 0
	for i, travel := range schedule.Result.Travels {
		if count >= 5 {
			break
		}
		result += fmt.Sprintf("🚆 %d:\n", i+1)
		for j, train := range travel.Trains {
			if count >= 5 {
				break
			}
			result += fmt.Sprintf("  🚂 %d:\n", j+1)
			result += fmt.Sprintf("    עליה: %s (רציף %d)\n",
				getStationName(fmt.Sprintf("%d", train.OriginStation)), train.OriginPlatform)
			result += fmt.Sprintf("    זמן יציאת הרכבת: %s\n",
				formatTime(train.DepartureTime))
			result += fmt.Sprintf("    אל: %s (רציף %d)\n",
				getStationName(fmt.Sprintf("%d", train.DestinationStation)), train.DestPlatform)
			result += fmt.Sprintf("    זמן הגעה: %s\n",
				formatTime(train.ArrivalTime))
			count++
		}
		result += "\n"
	}
	return result
}

// formatMobileSchedule formats the Mobile API response
func formatMobileSchedule(schedule MobileScheduleResponse) string {
	var result string
	count := 0
	for i, train := range schedule.Data {
		if count >= 5 {
			break
		}
		result += fmt.Sprintf("🚆 %d:\n", i+1)
		result += fmt.Sprintf("  🚂 1:\n")
		result += fmt.Sprintf("    עליה: %s (רציף %d)\n",
			getStationName(fmt.Sprintf("%d", train.OriginStationId)), train.OriginPlatform)
		result += fmt.Sprintf("    זמן יציאת הרכבת: %s\n",
			formatTime(train.DepartureTime))
		result += fmt.Sprintf("    אל: %s (רציף %d)\n",
			getStationName(fmt.Sprintf("%d", train.DestinationStationId)), train.DestinationPlatform)
		result += fmt.Sprintf("    זמן הגעה: %s\n",
			formatTime(train.ArrivalTime))
		result += "\n"
		count++
	}
	return result
}

// getStationName returns the Hebrew name for a station ID
func getStationName(stationID string) string {
	if station, ok := STATIONS[stationID]; ok {
		return station["Heb"]
	}
	return stationID
}

// formatTime formats a time string to HH:MM:SS format
func formatTime(timeStr string) string {
	t, err := time.Parse("2006-01-02T15:04:05", timeStr)
	if err != nil {
		return timeStr
	}
	return t.Format("15:04:05")
}

// splitMessage splits a message into chunks of maximum length
func splitMessage(message string) []string {
	const maxLength = 4000
	var chunks []string
	currentChunk := ""
	lines := strings.Split(message, "\n")

	for _, line := range lines {
		if len(currentChunk)+len(line)+1 > maxLength {
			chunks = append(chunks, currentChunk)
			currentChunk = line + "\n"
		} else {
			currentChunk += line + "\n"
		}
	}

	if currentChunk != "" {
		chunks = append(chunks, currentChunk)
	}

	return chunks
}