package main

import (
	"bufio"
	"errors"
	"fmt"
	"log"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/beevik/ntp"
	"github.com/faiface/beep"
	"github.com/faiface/beep/effects"
	"github.com/faiface/beep/mp3"
	"github.com/faiface/beep/speaker"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"go.bug.st/serial"
	"go.bug.st/serial/enumerator"
)

var (
	app             = tview.NewApplication()
	pages           = tview.NewPages()
	enabled         = true
	weekdayTimes    = []string{}
	pulseMode       = false
	enableWeekend   = false
	statusText      = "LOW"
	logLines        = []string{}
	pulseRunning    = false
	currentTimeFile = "idobeall1.txt"
	currentTime     time.Time
	timeMutex       = &sync.Mutex{}
	updateTimesMenu func() // globális frissítő függvény
)

var port serial.Port
var reader *bufio.Reader
var (
	bellRinging bool
	ctrl        *beep.Ctrl
	volume      *effects.Volume
)

func stopRing() {
	if !bellRinging {
		return
	}
	bellRinging = false

	// Fade-out duration
	fade := 250 * time.Millisecond
	steps := 25
	stepDur := fade / time.Duration(steps)

	go func() {
		for i := 0; i < steps; i++ {
			speaker.Lock()
			volume.Volume -= 1.0 / float64(steps) // fade to silence
			speaker.Unlock()
			time.Sleep(stepDur)
		}

		speaker.Lock()
		ctrl.Paused = true
		speaker.Unlock()
	}()
}

func playMP3() {
	if bellRinging {
		return
	}
	bellRinging = true

	f, err := os.Open("ring.mp3")
	if err != nil {
		log.Fatal(err)
	}

	streamer, format, err := mp3.Decode(f)
	if err != nil {
		log.Fatal(err)
	}

	speaker.Init(format.SampleRate, format.SampleRate.N(time.Second/10))

	ctrl = &beep.Ctrl{Streamer: streamer, Paused: false}
	volume = &effects.Volume{
		Streamer: ctrl,
		Base:     2,
		Volume:   0, // normal volume
		Silent:   false,
	}

	speaker.Play(volume)
}

// ---- LOG ----
func addLog(msg string) {
	line := fmt.Sprintf("[%s] %s", currentTime.Format("15:04:05"), msg)
	logLines = append(logLines, line)
	if len(logLines) > 100 {
		logLines = logLines[len(logLines)-100:]
	}
}

func AutoDetect() error {
	ports, err := enumerator.GetDetailedPortsList()
	if err != nil {
		return err
	}

	for _, p := range ports {
		if p.IsUSB && (strings.Contains(strings.ToLower(p.Product), "pico") ||
			strings.Contains(strings.ToLower(p.SerialNumber), "pico")) {

			mode := &serial.Mode{BaudRate: 115200}
			port, err = serial.Open(p.Name, mode)
			if err != nil {
				return err
			}

			reader = bufio.NewReader(port)
			return nil
		}
	}
	addLog("NOPICO")
	return errors.New("no Raspberry Pi Pico detected")
}

func sendCommand(cmd string) error {
	if port == nil {
		return errors.New("port not initialized, call AutoDetect() first")
	}

	fmt.Fprintf(port, "%s\n", cmd)

	port.SetReadTimeout(2 * time.Second)
	resp, err := reader.ReadString('\n')
	if err != nil {
		return err
	}

	if !strings.HasPrefix(resp, "OK") {
		addLog("NOPICO1")
		return fmt.Errorf("pico error: %s", resp)
	}

	return nil
}

// ---- GPIO MOCK ----
func SetHigh() {
	if !enabled {
		return
	}

	if !canRunNow(enableWeekend) {
		addLog("Hétvégi csengés tiltva")
		return
	}
	statusText = "HIGH"
	addLog("GPIO -> HIGH")
	sendCommand("HIGH")
	go playMP3()

	app.QueueUpdateDraw(func() {})
}

func SetLow() {
	statusText = "LOW"
	addLog("GPIO -> LOW")
	sendCommand("LOW")
	stopRing()
	app.QueueUpdateDraw(func() {})
}

// ---- MAIN ----
func main() {
	if err := AutoDetect(); err != nil {
		addLog("Pico nem található, offline mód")
	}
	loadTimesFromFile(currentTimeFile)

	// --- NTP idő lekérése ---
	ntpTime, err := ntp.Time("pool.ntp.org")
	if err != nil {
		fmt.Println("NTP lekérés sikertelen, gépi időt használok")
		currentTime = time.Now()
	} else {
		currentTime = ntpTime
	}

	go clockTicker()
	go scheduler()

	// --- Main menu ---
	mainMenu := tview.NewList().
		AddItem("1. Időzítések", "", '1', func() {
			pages.SwitchToPage("times")
			app.SetFocus(pages)
		}).
		AddItem("2. Be/Ki kapcsolás", "", '2', func() {
			enabled = !enabled
			addLog(fmt.Sprintf("Funkció BE/KI -> %v", enabled))
		}).
		AddItem("3. Impulzus/Tűzjelző mód", "", '3', func() {
			pulseMode = !pulseMode
			addLog(fmt.Sprintf("Impulzus mód -> %v", pulseMode))
			if pulseMode {
				startPulse()
			}
		}).
		AddItem("4. Dev konzol - CSAK KEZELŐNEK", "", '4', func() {
			pages.SwitchToPage("dev")
			app.SetFocus(pages)
		}).
		AddItem("5. Idő beállítása - CSAK KEZELŐNEK", "", '5', func() {
			pages.SwitchToPage("settime")
			app.SetFocus(pages)
		}).
		AddItem("6. Időzítés választás", "", '6', func() {
			pages.SwitchToPage("filemenu")
			app.SetFocus(pages)
		}).
		AddItem("7. Hétvégén csengessen", "", '6', func() {
			enableWeekend = !enableWeekend
			addLog(fmt.Sprintf("Hétvége -> %v", enableWeekend))
		})

	statusBar := tview.NewTextView().SetDynamicColors(true)

	layout := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(statusBar, 1, 1, false).
		AddItem(mainMenu, 0, 1, true)

	// --- Oldal hozzáadása ---
	pages.AddPage("main", layout, true, true)
	pages.AddPage("times", timesMenu(), true, false)
	pages.AddPage("dev", devConsole(), true, false)
	pages.AddPage("settime", setTimeMenu(), true, false)
	pages.AddPage("filemenu", fileSelectionMenu(), true, false)

	// Status bar frissítés
	go func() {
		for {
			timeMutex.Lock()
			ct := currentTime
			timeMutex.Unlock()
			app.QueueUpdateDraw(func() {
				statusBar.SetText(fmt.Sprintf(
					"[yellow]Idő:[white] %s  [white],[green]Engedélyezve:[white]%v [white](Hétvége:%v)  [white],[blue]Impulzus:[white]%v  [white],[red]Állapot:[white]%s [white],[red]Karbantartó: [white]Vaskó Péter[white], [red]Bellringer@Oveges",
					ct.Format("15:04:05"),
					enabled,
					enableWeekend,
					pulseMode,
					statusText,
				))
			})

			time.Sleep(time.Second)
		}
	}()

	if err := app.SetRoot(pages, true).Run(); err != nil {
		panic(err)
	}
}

// ---- CLOCK ----
func clockTicker() {
	for {
		time.Sleep(time.Second)
		timeMutex.Lock()
		currentTime = currentTime.Add(time.Second)
		timeMutex.Unlock()
		app.QueueUpdateDraw(func() {})
	}
}

func timesMenu() tview.Primitive {
	input := tview.NewInputField().SetLabel("Idő HH:MM:SS): ")
	timesInfo := tview.NewTextView().SetDynamicColors(true)

	// A globális updateTimesMenu változóhoz rendeljük a frissítő függvényt
	updateTimesMenu = func() {
		if len(weekdayTimes) == 0 {
			timesInfo.SetText("Nincsenek időzítések")
		} else {
			timesInfo.SetText("Időzítések:\n" + strings.Join(weekdayTimes, ", "))
		}
	}

	updateTimesMenu() // azonnali frissítés

	// Új idő hozzáadása
	input.SetDoneFunc(func(key tcell.Key) {
		if key == tcell.KeyEnter {
			txt := input.GetText()
			_, err := time.Parse("15:04:05", txt) // csak ellenőrzés, t nem kell
			if err != nil {
				addLog("Hibás időformátum: " + txt)
				return
			}
			// Idő hozzáadása
			weekdayTimes = append(weekdayTimes, txt)
			addLog("Idő hozzáadva: " + txt)
			saveTimesToFile() // mentés fájlba
			updateTimesMenu() // frissítés az UI-ban
			input.SetText("") // mező ürítése
		}
	})

	// Vissza gomb
	back := tview.NewButton("Vissza/ESC").SetSelectedFunc(func() {
		pages.SwitchToPage("main")
		app.SetFocus(pages)
	})

	// ESC gomb kezelése
	input.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEsc {
			pages.SwitchToPage("main")
			app.SetFocus(pages)
			return nil
		}
		return event
	})

	return tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(timesInfo, 0, 1, false).
		AddItem(input, 1, 1, true).
		AddItem(back, 1, 1, false)
}

// ---- DEV CONSOLE ----
func devConsole() tview.Primitive {
	console := tview.NewTextView().SetDynamicColors(true)

	updateLog := func() {
		console.SetText(
			"DEV MÓD\n" +
				"H=HIGH  L=LOW  T=TRIGGER  C=CLEAR  B=BACK\n\n" +
				strings.Join(logLines, "\n"),
		)
	}

	flex := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(console, 0, 1, true)

	flex.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Rune() {
		case 'h', 'H':
			go SetHigh()
		case 'l', 'L':
			go SetLow()
		case 't', 'T':
			addLog("MANUÁLIS IDŐ TRIGGER")
			go triggerPulseOnce()
		case 'c', 'C':
			logLines = nil
		case 'b', 'B':
			pages.SwitchToPage("main")
			app.SetFocus(pages)
			return nil
		}
		updateLog()
		return nil
	})

	updateLog()
	return flex
}

// ---- SET TIME MENU ----
func setTimeMenu() tview.Primitive {
	form := tview.NewForm().
		AddInputField("Idő (HH:MM:SS)", "", 8, nil, nil).
		AddButton("Vissza/ESC", func() {
			pages.SwitchToPage("main")
			app.SetFocus(pages)
		})

	form.SetBorder(true).SetTitle("Idő beállítása").SetTitleAlign(tview.AlignCenter)

	input := form.GetFormItemByLabel("Idő (HH:MM:SS)").(*tview.InputField)
	input.SetDoneFunc(func(key tcell.Key) {
		if key == tcell.KeyEnter {
			txt := input.GetText()
			t, err := time.Parse("15:04:05", txt)
			if err != nil {
				addLog("Hibás időformátum")
				return
			}
			timeMutex.Lock()
			currentTime = time.Date(
				currentTime.Year(),
				currentTime.Month(),
				currentTime.Day(),
				t.Hour(),
				t.Minute(),
				t.Second(),
				0,
				currentTime.Location(),
			)
			timeMutex.Unlock()
			addLog("Idő kézi beállítva: " + txt)
			input.SetText("")
		}
	})

	input.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEsc {
			pages.SwitchToPage("main")
			app.SetFocus(pages)
			return nil
		}
		return event
	})

	return form
}

// ---- PULSE ----
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
	// biztosítjuk, hogy mindig LOW-ról induljon
	SetLow()
	time.Sleep(500 * time.Millisecond) // rövid delay a biztonságos váltáshoz

	// 1 másodperces HIGH
	SetHigh()
	sleepWithDraw(3 * time.Second)

	// vissza LOW-ra
	SetLow()

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

// ---- SCHEDULER ----
func scheduler() {
	for {
		if !enabled {
			time.Sleep(time.Second)
			continue
		}

		timeMutex.Lock()
		now := currentTime.Format("15:04:05")
		timeMutex.Unlock()

		for _, t := range weekdayTimes {
			if t == now {
				addLog("IDŐZÍTÉS AKTIVÁLVA: " + t)
				go triggerPulseOnce() // egyszeri HIGH/LOW a triggerhez
			}
		}

		time.Sleep(time.Second)
	}
}

// ---- FILE LOAD/SAVE ----
func loadTimesFromFile(filename string) {
	weekdayTimes = nil

	if filename == "" {
		filename = "idobeall.txt"
	}

	if !strings.HasSuffix(filename, ".txt") {
		addLog("Csak .txt fájlokat lehet betölteni: " + filename)
		return
	}

	currentTimeFile = filename
	logLines = nil

	file, err := os.Open(filename)
	if err != nil {
		addLog(fmt.Sprintf("%s nem található, új fájl jön létre", filename))
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
	for scanner.Scan() {
		line := strings.TrimSpace(strings.ReplaceAll(scanner.Text(), "\r", ""))
		line = strings.TrimPrefix(line, "\ufeff") // BOM eltávolítása
		if line != "" {
			weekdayTimes = append(weekdayTimes, line)
		}
	}

	if err := scanner.Err(); err != nil {
		addLog("Hiba a " + filename + " olvasása közben: " + err.Error())
	} else {
		addLog(fmt.Sprintf("%d időzítés betöltve a %s fájlból", len(weekdayTimes), filename))
	}

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

	for _, t := range weekdayTimes {
		_, _ = file.WriteString(t + "\n")
	}
	addLog("Időzítések mentve a " + currentTimeFile + " fájlba")
}

// ---- FILE SELECTION MENU ----
func fileSelectionMenu() tview.Primitive {
	list := tview.NewList()

	var updateList func()
	updateList = func() {
		list.Clear()
		files := listAllFiles()
		for _, f := range files {
			fname := f
			list.AddItem(fname, "Load this file", 0, func() {
				loadTimesFromFile(fname)
				pages.SwitchToPage("times")
				app.SetFocus(pages)
			})
		}
		list.AddItem("Új fájl létrehozása", "Create a new schedule file", 0, func() {
			showNewFilePrompt(updateList)
		})
		list.AddItem("Vissza/ESC", "Return to main menu", 0, func() {
			pages.SwitchToPage("main")
			app.SetFocus(pages)
		})
	}

	updateList()
	return list
}

func listAllFiles() []string {
	var files []string
	entries, err := os.ReadDir(".")
	if err != nil {
		return files
	}
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".txt") {
			files = append(files, entry.Name())
		}
	}
	return files
}

func showNewFilePrompt(updateList func()) {
	form := tview.NewForm()
	inputField := tview.NewInputField().SetLabel("Fájl neve").SetFieldWidth(20)
	form.AddFormItem(inputField)

	form.AddButton("Létrehozás", func() {
		name := inputField.GetText()
		if name != "" {
			if !strings.HasSuffix(name, ".txt") {
				name += ".txt"
			}
			f, err := os.Create(name)
			if err == nil {
				f.Close()
				addLog("Új fájl létrehozva: " + name)
				updateList()
			} else {
				addLog("Nem sikerült létrehozni a fájlt: " + err.Error())
			}
		}
		app.SetRoot(pages, true)
	})

	form.AddButton("Mégse", func() {
		app.SetRoot(pages, true)
	})

	form.SetBorder(true).SetTitle("Új fájl létrehozása").SetTitleAlign(tview.AlignCenter)
	app.SetRoot(form, true)
}

func canRunNow(enableWeekend bool) bool {
	timeMutex.Lock()
	ct := currentTime
	timeMutex.Unlock()

	weekday := ct.Weekday()
	isWeekend := weekday == time.Saturday || weekday == time.Sunday

	if isWeekend && !enableWeekend {
		return false
	}
	return true
}
