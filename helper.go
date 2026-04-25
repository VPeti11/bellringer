package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"
)

func canTriggerRing() bool {
	now := time.Now()
	if now.Sub(lastRingTime) < debounceDuration {
		return false
	}
	lastRingTime = now
	return true
}

func loadShortTimes() {
	file, err := os.Open("short.txt")
	if err != nil {
		return
	}
	defer file.Close()

	newMap := map[string]bool{}

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		t := strings.TrimSpace(scanner.Text())
		if t == "" {
			continue
		}
		newMap[t] = true
	}

	shortTimesMu.Lock()
	shortTimes = newMap
	shortTimesMu.Unlock()
}

func saveShortTimes() {
	file, err := os.Create("short.txt")
	if err != nil {
		addLog("short.txt mentési hiba: " + err.Error())
		return
	}
	defer file.Close()

	shortTimesMu.RLock()
	defer shortTimesMu.RUnlock()

	for t := range shortTimes {
		file.WriteString(t + "\n")
	}
}

func getNextEvent() (string, time.Time, bool) {

	if !enabled {
		return tr("never", "soha"), time.Time{}, false
	}

	magyarNapok := map[time.Weekday]string{
		time.Monday:    "Hétfő",
		time.Tuesday:   "Kedd",
		time.Wednesday: "Szerda",
		time.Thursday:  "Csütörtök",
		time.Friday:    "Péntek",
		time.Saturday:  "Szombat",
		time.Sunday:    "Vasárnap",
	}

	angolNapok := map[time.Weekday]string{
		time.Monday:    "Monday",
		time.Tuesday:   "Tuesday",
		time.Wednesday: "Wednesday",
		time.Thursday:  "Thursday",
		time.Friday:    "Friday",
		time.Saturday:  "Saturday",
		time.Sunday:    "Sunday",
	}

	now := time.Now()

	var nextTime time.Time
	found := false
	eventStr := ""

	for i := 0; i < 7; i++ {
		currentDay := now.AddDate(0, 0, i)
		weekday := currentDay.Weekday()
		dateStr := currentDay.Format("060102")

		var times []string

		dateOverridesMu.RLock()
		override, exists := dateOverrides[dateStr]
		dateOverridesMu.RUnlock()

		if exists {
			times = override
		} else {
			schedulesMu.RLock()
			times = schedules[weekday]
			schedulesMu.RUnlock()
		}

		for _, t := range times {
			parsed, err := time.Parse("15:04:05", t)
			if err != nil {
				continue
			}

			eventTime := time.Date(
				currentDay.Year(),
				currentDay.Month(),
				currentDay.Day(),
				parsed.Hour(),
				parsed.Minute(),
				parsed.Second(),
				0,
				now.Location(),
			)

			if eventTime.After(now) {
				if !found || eventTime.Before(nextTime) {
					nextTime = eventTime

					dayName := magyarNapok[weekday]
					if useEnglish {
						dayName = angolNapok[weekday]
					}

					eventStr = fmt.Sprintf("%s %s", dayName, t)
					found = true
				}
			}
		}
	}

	return eventStr, nextTime, found
}

func sleepWithDraw(d time.Duration) {
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()
	done := time.After(d)
	for {
		select {
		case <-done:
			return
		case <-ticker.C:
			app.QueueUpdateDraw(func() {})
		}
	}
}

func loadTimesFromFile(filename string) {

	if filename == "" {
		filename = "idobeall.txt"
	}

	if !strings.HasSuffix(filename, ".txt") {
		addLog("Csak .txt fájlokat lehet betölteni: " + filename)
		return
	}

	currentTimeFile = filename
	logLines = nil

	dateOverridesMu.Lock()
	dateOverrides = map[string][]string{}
	dateOverridesMu.Unlock()

	file, err := os.Open(filename)
	if err != nil {
		addLog(fmt.Sprintf("%s nem található, új fájl létrehozva", filename))

		newFile, err := os.Create(filename)
		if err != nil {
			addLog("Nem sikerült létrehozni a fájlt: " + err.Error())
			return
		}
		newFile.Close()
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)

	newSchedules := map[time.Weekday][]string{
		time.Monday: {}, time.Tuesday: {}, time.Wednesday: {},
		time.Thursday: {}, time.Friday: {}, time.Saturday: {}, time.Sunday: {},
	}

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		parts := strings.Split(line, "=")
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		timesStr := strings.TrimSpace(parts[1])

		if len(key) == 6 {
			if _, err := time.Parse("060102", key); err == nil {

				if timesStr == "" {
					dateOverridesMu.Lock()
					dateOverrides[key] = []string{}
					dateOverridesMu.Unlock()
					continue
				}

				times := strings.Split(timesStr, ",")
				for i := range times {
					times[i] = strings.TrimSpace(times[i])
				}

				dateOverridesMu.Lock()
				dateOverrides[key] = times
				dateOverridesMu.Unlock()

				continue
			}
		}

		var day time.Weekday

		switch key {
		case "Monday":
			day = time.Monday
		case "Tuesday":
			day = time.Tuesday
		case "Wednesday":
			day = time.Wednesday
		case "Thursday":
			day = time.Thursday
		case "Friday":
			day = time.Friday
		case "Saturday":
			day = time.Saturday
		case "Sunday":
			day = time.Sunday
		default:
			addLog("Ismeretlen kulcs a fájlban: " + key)
			continue
		}

		if timesStr == "" {
			newSchedules[day] = []string{}
			continue
		}

		times := strings.Split(timesStr, ",")
		for i := range times {
			times[i] = strings.TrimSpace(times[i])
		}

		newSchedules[day] = times
	}

	if err := scanner.Err(); err != nil {
		addLog("Hiba a fájl olvasása közben: " + err.Error())
		return
	}

	schedulesMu.Lock()
	schedules = newSchedules
	schedulesMu.Unlock()

	total := 0
	for _, t := range newSchedules {
		total += len(t)
	}

	addLog(fmt.Sprintf("%d időzítés betöltve a %s fájlból", total, filename))

	if updateTimesMenu != nil {
		updateTimesMenu()
	}
}

func saveTimesToFile() {
	if currentTimeFile == "" {
		addLog("Nincs kiválasztva fájl a mentéshez")
		return
	}

	file, err := os.Create(currentTimeFile)
	if err != nil {
		addLog("Nem sikerült menteni az időzítéseket: " + err.Error())
		return
	}
	defer file.Close()

	schedulesMu.RLock()
	for day, times := range schedules {
		dayStr := day.String()
		line := dayStr + "=" + strings.Join(times, ",") + "\n"
		_, _ = file.WriteString(line)
	}
	schedulesMu.RUnlock()

	dateOverridesMu.RLock()
	for date, times := range dateOverrides {

		line := date + "="

		if len(times) > 0 {
			line += strings.Join(times, ",")
		}

		line += "\n"

		_, _ = file.WriteString(line)
	}
	dateOverridesMu.RUnlock()

	addLog("Időzítések + override-ok mentve a " + currentTimeFile + " fájlba")
}

func listAllFiles() []string {
	var files []string
	entries, err := os.ReadDir(".")
	if err != nil {
		return files
	}
	for _, entry := range entries {
		name := entry.Name()
		if !entry.IsDir() && strings.HasSuffix(name, ".txt") {
			if name == "serial.txt" || name == "webconfig.txt" || name == "short.txt" || name == "log.txt" {
				continue
			}
			files = append(files, name)
		}
	}
	return files
}

func canRunNow() bool {
	weekday := time.Now().Weekday()
	isWeekend := weekday == time.Saturday || weekday == time.Sunday
	return !isWeekend || enableWeekend
}

func clearScreen() {
	if runtime.GOOS == "windows" {
		cmd := exec.Command("cmd", "/c", "cls")
		cmd.Stdout = os.Stdout
		cmd.Run()
	} else {
		fmt.Print("\033[H\033[2J")
	}
}

func tr(en, hu string) string {
	if useEnglish {
		return en
	}
	return hu
}

func centerText(text string, width int) string {
	if len(text) >= width {
		return text
	}
	padding := (width - len(text)) / 2
	return strings.Repeat(" ", padding) + text
}
