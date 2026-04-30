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
					if enabled && !eStop {
						go triggerPulseOnceal()
					}

				} else {
					if enabled && !eStop {
						go triggerPulseOnce()
					}
				}
			}
		}
	}
}

func toggleScheduler() {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		nowTime := time.Now()
		nowStr := nowTime.Format("060102 150405")

		toggleMu.Lock()

		for i := len(toggleQueue) - 1; i >= 0; i-- {

			if toggleQueue[i].Time.Before(nowTime) && toggleQueue[i].Time.Format("060102 150405") != nowStr {
				addLog("Expired task removed: " + toggleQueue[i].Time.Format("06-01-02 15:04:05"))
				toggleQueue = append(toggleQueue[:i], toggleQueue[i+1:]...)
			}
		}

		var executedIndex = -1
		var taskToExecute *ScheduledToggle

		for i, t := range toggleQueue {
			if t.Time.Format("060102 150405") == nowStr {
				taskToExecute = &t
				executedIndex = i
				break
			}
		}

		if taskToExecute != nil {
			enabled = taskToExecute.State

			addLog(fmt.Sprintf("Scheduled toggle executed -> %v (%s)",
				taskToExecute.State,
				taskToExecute.Time.Format("06-01-02 15:04:05"),
			))

			toggleQueue = append(toggleQueue[:executedIndex], toggleQueue[executedIndex+1:]...)
		}

		toggleMu.Unlock()
	}
}

func emergencyStop() {
	if !eStop {
		addLog("VÉSZLEÁLLÍTÁS AKTIVÁLVA")
		pulseMode = false
		pulseRunning = false
		bellRinging = false

		go SetLow()

		if webOn {
			stopWebServer()
		}

		enabled = false
		eStop = true
	} else {
		addLog("VÉSZLEÁLLÍTÁS FELDOLDVA")

		enabled = true

		if !webOn {
			startWebServer()
		}
		eStop = false
	}

	app.QueueUpdateDraw(func() {})
}

func SetHigh() {
	if bellRinging {
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
