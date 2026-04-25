package main

import (
	"fmt"
	"os"
	"time"
)

func addLog(msg string) {
	line := fmt.Sprintf("[%s] %s", time.Now().Format("2006-01-02 15:04:05"), msg)

	logLines = append(logLines, line)
	if len(logLines) > 500 {
		logLines = logLines[len(logLines)-500:]
	}

	f, err := os.OpenFile("log.txt", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err == nil {
		defer f.Close()
		fmt.Fprintln(f, line)
	}
}

func startPulse() {
	if pulseRunning {
		return
	}
	pulseRunning = true

	go func() {
		for pulseMode {
			SetHigh()
			sleepWithDraw(1 * time.Second)
			if !pulseMode {
				break
			}
			SetLow()
			sleepWithDraw(1 * time.Second)
		}
		SetLow()
		pulseRunning = false
	}()
}

func triggerPulseOnce() {
	if !canTriggerRing() {
		addLog("Debounce aktív (triggerPulseOnce)")
		return
	}

	SetLow()
	time.Sleep(500 * time.Millisecond)

	SetHigh()
	sleepWithDraw(8 * time.Second)

	SetLow()

}

func triggerPulseOnceal() {
	if !canTriggerRing() {
		addLog("Debounce aktív (triggerPulseOnce)")
		return
	}

	SetLow()
	time.Sleep(500 * time.Millisecond)

	SetHigh()
	sleepWithDraw(2 * time.Second)

	SetLow()
	time.Sleep(700 * time.Millisecond)

	SetHigh()
	sleepWithDraw(2 * time.Second)

	SetLow()

}

func scheduler() {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for range ticker.C {
		broadcastStatus()

		if !enabled {
			continue
		}

		nowTime := time.Now()
		now := nowTime.Format("15:04:05")

		todayStr := nowTime.Format("060102")
		weekday := nowTime.Weekday()

		var times []string

		dateOverridesMu.RLock()
		override, exists := dateOverrides[todayStr]
		dateOverridesMu.RUnlock()

		if exists {
			times = override
		} else {
			schedulesMu.RLock()
			times = schedules[weekday]
			schedulesMu.RUnlock()
		}

		for _, t := range times {
			if t == now {
				addLog("IDŐZÍTÉS AKTIVÁLVA: " + t)

				shortTimesMu.RLock()
				short := shortTimes[t]
				shortTimesMu.RUnlock()

				if short {
					go triggerPulseOnceal()
				} else {
					go triggerPulseOnce()
				}
			}
		}
	}
}

func emergencyStop() {
	addLog("VÉSZLEÁLLÍTÁS AKTIVÁLVA")
	pulseMode = false
	pulseRunning = false
	bellRinging = false
	go SetLow()

	if webOn {
		stopWebServer()
	}

	enabled = false
	app.QueueUpdateDraw(func() {})
}

func SetHigh() {
	if bellRinging {
		return
	}
	if !enabled {
		return
	}
	if !pulseMode {
		if !canRunNow() {
			addLog("Hétvégi csengés tiltva")
			return
		}
	}
	statusText = "HIGH"
	addLog("GPIO -> HIGH")
	jelKuldes("HIGH")
	go playMP3()

	app.QueueUpdateDraw(func() {})
}

func SetLow() {
	statusText = "LOW"
	addLog("GPIO -> LOW")
	jelKuldes("LOW")
	if bellRinging {
		stopRing()
	}
	app.QueueUpdateDraw(func() {})
}
