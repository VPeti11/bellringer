package main

import (
	"fmt"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

func main() {
	configurate()
	loadTimesFromFile(currentTimeFile)
	loadShortTimes()

	if webEnabled {
		go startWebServer()
		addLog(tr("Web server STARTED", "Web szerver ELINDÍTVA"))
	} else {
		addLog(tr("Web server DISABLED", "Web szerver LETILTVA"))
	}

	go scheduler()

	mainMenu := tview.NewList().
		AddItem("1. "+tr("Schedules", "Időzítések"), "", '1', func() { pages.SwitchToPage("times"); app.SetFocus(pages) }).
		AddItem("2. "+tr("Enable/Disable", "Be/Ki kapcsolás"), "", '2', func() {
			enabled = !enabled
			addLog(fmt.Sprintf(tr("Toggle -> %v", "Funkció BE/KI -> %v"), enabled))
		}).
		AddItem("3. "+tr("Pulse / Alarm mode", "Impulzus/Tűzjelző mód"), "", '3', func() {
			pulseMode = !pulseMode
			if pulseMode {
				go startPulse()
			}
		}).
		AddItem("4. "+tr("Dev console - ADMIN ONLY", "Dev konzol - CSAK KEZELŐNEK"), "", '4', func() { pages.SwitchToPage("dev"); app.SetFocus(pages) }).
		AddItem("6. "+tr("Schedule selection", "Időzítés választás"), "", '6', func() { pages.SwitchToPage("filemenu"); app.SetFocus(pages) }).
		AddItem("7. "+tr("Weekend ringing", "Hétvégén csengessen"), "", '7', func() {
			enableWeekend = !enableWeekend
		}).
		AddItem("8. "+tr("Long ring", "Hosszú csengés"), "", '8', func() { go triggerPulseOnce() }).
		AddItem("9. "+tr("Short ring", "Rövid csengés"), "", '9', func() { go triggerPulseOnceal() }).
		AddItem("10. "+tr("Web server ON/OFF", "Web szerver BE/KI"), "", 'w', func() {
			webEnabled = !webEnabled
			if webEnabled {
				go startWebServer()
			} else {
				stopWebServer()
			}
		}).
		AddItem("11. "+tr("Short ring times", "Rövid csengés időpontok"), "", 'r', func() { pages.SwitchToPage("shorttimes"); app.SetFocus(pages) }).
		AddItem("12. "+tr("Date overrides", "Dátum felülírások-ok"), "", 'o', func() { pages.SwitchToPage("overrides"); app.SetFocus(pages) }).
		AddItem(tr("EMERGENCY STOP", "VÉSZLEÁLLÍTÁS"), "", 'x', func() { go emergencyStop() })

	statusBar := tview.NewTextView().SetDynamicColors(true).SetWrap(true).SetWordWrap(true)
	nextEventWidget := tview.NewTextView().SetDynamicColors(true).SetTextAlign(tview.AlignRight)

	footer := tview.NewFlex().
		AddItem(tview.NewTextView().SetText("Öveges/Bellringer"), 20, 1, false).
		AddItem(nextEventWidget, 0, 1, false)

	layout := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(statusBar, 2, 0, false).
		AddItem(mainMenu, 0, 1, true).
		AddItem(footer, 1, 0, false)

	splashBox := tview.NewTextView().SetDynamicColors(true).SetTextAlign(tview.AlignLeft)
	splashInfo := tview.NewTextView().SetDynamicColors(true).SetTextAlign(tview.AlignCenter)

	splashContent := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(splashBox, 5, 1, false).
		AddItem(splashInfo, 3, 1, false)

	splashLayout := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(nil, 0, 1, false).
		AddItem(tview.NewFlex().SetDirection(tview.FlexColumn).
			AddItem(nil, 0, 1, false).
			AddItem(splashContent, 45, 1, false).
			AddItem(nil, 0, 1, false), 8, 1, false).
		AddItem(nil, 0, 1, false)

	pages.AddPage("times", timesMenu(), true, false)
	pages.AddPage("dev", devConsole(), true, false)
	pages.AddPage("filemenu", fileSelectionMenu(), true, false)
	pages.AddPage("shorttimes", shortTimesMenu(), true, false)
	pages.AddPage("overrides", dateOverrideMenu(), true, false)
	pages.AddPage("splash", splashLayout, true, true)

	startTime := time.Now()
	stopTime := startTime.Add(2 * time.Second)

	updateUI := func(now time.Time) {
		if now.Before(stopTime) && pages.HasPage("splash") {

			boxText := tr(
				`[yellow]╔═══════════════════════════════════════════╗
[yellow]║[white]           BELL RINGER SOFTWARE            [yellow]║
[yellow]║[white]           Created by Vaskó Péter          [yellow]║
[yellow]║[gray]           CC BY-NC-ND 4.0 License         [yellow]║
[yellow]╚═══════════════════════════════════════════╝`,
				`[yellow]╔═══════════════════════════════════════════╗
[yellow]║[white]           BELL RINGER SZOFTVER            [yellow]║
[yellow]║[white] Készítette: Vaskó Péter, jogok fenntartva [yellow]║
[yellow]║[gray]           CC BY-NC-ND 4.0 Licenc          [yellow]║
[yellow]╚═══════════════════════════════════════════╝`)

			infoText := fmt.Sprintf("\n%s: [yellow]%s\n[cyan]%s...",
				tr("Current time", "Pontos idő"),
				now.Format("15:04:05"),
				tr("Loading systems", "Rendszerek betöltése"),
			)

			splashBox.SetText(boxText)
			splashInfo.SetText(infoText)
		}

		eventStr, nextTime, ok := getNextEvent()
		var countdown string
		if ok {
			diff := time.Until(nextTime)
			if diff < 0 {
				diff = 0
			}
			countdown = fmt.Sprintf("%02d:%02d:%02d",
				int(diff.Hours()), int(diff.Minutes())%60, int(diff.Seconds())%60)
		} else {
			countdown = "--:--:--"
			eventStr = tr("no event", "nincs esemény")
		}

		statusBar.SetText(fmt.Sprintf(
			"[yellow]%s:[white] %s  "+
				"[green]%s:[white]%v (%s:%v)  "+
				"[blue]%s:[white]%v  "+
				"[red]%s:[white]%s  "+
				"[yellow]Web:[white]%v  "+
				"[purple]File:[white] %s",

			tr("Time", "Idő"), now.Format("15:04:05"),
			tr("Enabled", "Engedélyezve"), enabled,
			tr("Weekend", "Hétvége"), enableWeekend,
			tr("Pulse", "Impulzus"), pulseMode,
			tr("Status", "Állapot"), statusText,
			webEnabled,
			currentTimeFile,
		))

		nextEventWidget.SetText(fmt.Sprintf(
			"[cyan]%s:[white] %s [magenta]%s:[white] %s",
			tr("Next", "Következő"), eventStr,
			tr("Next ring", "Köv. csengés"), countdown,
		))

		if (now.After(stopTime) || now.Equal(stopTime)) && pages.HasPage("splash") {
			pages.AddPage("main", layout, true, true)
			pages.RemovePage("splash")
			pages.SwitchToPage("main")
			app.SetFocus(mainMenu)
		}
	}

	updateUI(startTime)

	go func() {
		ticker := time.NewTicker(1 * time.Second)
		defer ticker.Stop()

		for t := range ticker.C {
			app.QueueUpdateDraw(func() {
				updateUI(t)
			})
		}
	}()

	if err := app.SetRoot(pages, true).Run(); err != nil {
		panic(err)
	}
}

func shortTimesMenu() tview.Primitive {
	input := tview.NewInputField().SetLabel(tr("Add (HH:MM:SS): ", "Hozzáadás (HH:MM:SS): "))
	list := tview.NewList().SetSelectedFocusOnly(true)

	help := tview.NewTextView().
		SetTextColor(tcell.ColorYellow).
		SetText(tr(" [ENTER] add | [d] delete | [u] refresh | [TAB] switch | [ESC] back", " [ENTER] hozzáadás | [d] törlés | [u] frissítés | [TAB] váltás | [ESC] vissza"))

	refresh := func() {
		loadShortTimes()
		list.Clear()

		shortTimesMu.RLock()
		defer shortTimesMu.RUnlock()

		for t := range shortTimes {
			timeStr := t
			list.AddItem(" - "+timeStr, tr("delete (Enter / d)", "törlés (Enter / d)"), 0, nil)
		}
	}

	add := func(t string) {
		shortTimesMu.Lock()
		shortTimes[t] = true
		shortTimesMu.Unlock()
		saveShortTimes()
		refresh()
	}

	del := func(t string) {
		shortTimesMu.Lock()
		delete(shortTimes, t)
		shortTimesMu.Unlock()
		saveShortTimes()
		refresh()
	}

	input.SetDoneFunc(func(key tcell.Key) {
		if key != tcell.KeyEnter {
			return
		}
		t := strings.TrimSpace(input.GetText())
		if _, err := time.Parse("15:04:05", t); err != nil {
			addLog(tr("Invalid time: ", "Hibás idő: ") + t)
			return
		}
		add(t)
		input.SetText("")
		app.SetFocus(list)
	})

	list.SetSelectedFunc(func(index int, mainText string, secondaryText string, shortcut rune) {
		t := strings.TrimSpace(strings.TrimPrefix(mainText, " - "))
		del(t)
	})

	box := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(list, 0, 1, true).
		AddItem(input, 1, 1, false).
		AddItem(help, 1, 1, false)

	box.SetInputCapture(func(ev *tcell.EventKey) *tcell.EventKey {
		if ev.Key() == tcell.KeyTab {
			if app.GetFocus() == input {
				app.SetFocus(list)
			} else {
				app.SetFocus(input)
			}
			return nil
		}

		if ev.Key() == tcell.KeyEsc {
			pages.SwitchToPage("main")
			return nil
		}

		if app.GetFocus() == list {
			switch ev.Rune() {
			case 'u':
				refresh()
				return nil

			case 'd':
				index := list.GetCurrentItem()
				if index >= 0 {
					mainText, _ := list.GetItemText(index)
					t := strings.TrimSpace(strings.TrimPrefix(mainText, " - "))
					del(t)
				}
				return nil

			case 'a':
				input.SetText("")
				app.SetFocus(input)
				return nil
			}
		}

		return ev
	})

	refresh()
	app.SetFocus(list)

	return box
}

func timesMenu() tview.Primitive {
	currentDay := time.Monday

	napok := map[time.Weekday]string{
		time.Monday:    tr("Monday", "Hétfő"),
		time.Tuesday:   tr("Tuesday", "Kedd"),
		time.Wednesday: tr("Wednesday", "Szerda"),
		time.Thursday:  tr("Thursday", "Csütörtök"),
		time.Friday:    tr("Friday", "Péntek"),
		time.Saturday:  tr("Saturday", "Szombat"),
		time.Sunday:    tr("Sunday", "Vasárnap"),
	}

	input := tview.NewInputField().SetLabel(tr("Time (HH:MM:SS): ", "Idő (HH:MM:SS): "))
	timesInfo := tview.NewTextView().SetDynamicColors(true)
	dayLabel := tview.NewTextView().SetDynamicColors(true)

	dayToInt := func(d time.Weekday) int { return int(d) }
	intToDay := func(i int) time.Weekday { return time.Weekday((i + 7) % 7) }

	updateTimesMenu := func() {
		schedulesMu.RLock()
		list := schedules[currentDay]
		schedulesMu.RUnlock()

		dayLabel.SetText(fmt.Sprintf(tr("[yellow]Day:[white] %s", "[yellow]Nap:[white] %s"), napok[currentDay]))

		if len(list) == 0 {
			timesInfo.SetText(tr("No schedules", "Nincsenek időzítések"))
			return
		}

		timesInfo.SetText(tr("Schedules:\n", "Időzítések:\n") + strings.Join(list, "\n"))
	}

	changeDay := func(delta int) {
		currentDay = intToDay(dayToInt(currentDay) + delta)
		updateTimesMenu()
	}

	updateTimesMenu()

	prevDay := tview.NewButton("<").SetSelectedFunc(func() { changeDay(-1) })
	nextDay := tview.NewButton(">").SetSelectedFunc(func() { changeDay(1) })

	input.SetDoneFunc(func(key tcell.Key) {
		if key != tcell.KeyEnter {
			return
		}

		raw := strings.TrimSpace(input.GetText())

		parsed, err := time.Parse("15:04:05", raw)
		if err != nil || parsed.Format("15:04:05") != raw {
			addLog(tr("Invalid time format (HH:MM:SS only): ", "Hibás időformátum (csak HH:MM:SS): ") + raw)
			return
		}

		timeStr := parsed.Format("15:04:05")

		schedulesMu.RLock()
		for _, t := range schedules[currentDay] {
			if t == timeStr {
				schedulesMu.RUnlock()
				addLog(tr("This time already exists: ", "Ez az idő már létezik: ") + timeStr)
				return
			}
		}
		schedulesMu.RUnlock()

		schedulesMu.Lock()
		schedules[currentDay] = append(schedules[currentDay], timeStr)
		schedulesMu.Unlock()

		addLog(fmt.Sprintf(tr("Time added (%s): %s", "Idő hozzáadva (%s): %s"), napok[currentDay], timeStr))

		saveTimesToFile()
		updateTimesMenu()
		input.SetText("")
	})

	back := tview.NewButton(tr("Back/ESC", "Vissza/ESC")).SetSelectedFunc(func() {
		pages.SwitchToPage("main")
		app.SetFocus(pages)
	})

	input.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEsc {
			pages.SwitchToPage("main")
			app.SetFocus(pages)
			return nil
		}
		return event
	})

	dayControls := tview.NewFlex().
		AddItem(prevDay, 3, 1, false).
		AddItem(dayLabel, 0, 1, false).
		AddItem(nextDay, 3, 1, false)

	root := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(dayControls, 1, 1, false).
		AddItem(timesInfo, 0, 1, false).
		AddItem(input, 1, 1, true).
		AddItem(back, 1, 1, false)

	root.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyLeft:
			changeDay(-1)
			return nil
		case tcell.KeyRight:
			changeDay(1)
			return nil
		}
		return event
	})

	return root
}

func devConsole() tview.Primitive {
	console := tview.NewTextView().
		SetDynamicColors(true).
		SetScrollable(true).
		SetChangedFunc(func() {
			app.Draw()
		})

	updateLog := func() {
		console.Clear()
		fmt.Fprintf(console, "%s", tr("[yellow]DEV MODE[-]\n", "[yellow]DEV MÓD[-]\n"))
		fmt.Fprintf(console, "%s", tr("[green]H=HIGH  L=LOW  C=CLEAR  U=UPDATE  B=BACK[-]\n\n", "[green]H=MAGAS L=ALACSONY C=TÖRLÉS U=FRISSÍTÉS B=VISSZA[-]\n\n"))
		fmt.Fprintf(console, "%s", strings.Join(logLines, "\n"))
		console.ScrollToEnd()
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
		case 'c', 'C':
			logLines = nil
		case 'u', 'U':
			updateLog()
			return nil
		case 'b', 'B':
			pages.SwitchToPage("main")
			return nil
		}
		updateLog()
		return event
	})

	updateLog()
	return flex
}

func showNewFilePrompt(updateList func()) {
	form := tview.NewForm()
	inputField := tview.NewInputField().
		SetLabel(tr("File name (letters/numbers only): ", "Fájl neve (csak betű/szám): ")).
		SetFieldWidth(30)

	form.AddFormItem(inputField)

	form.AddButton(tr("Create", "Létrehozás"), func() {
		rawName := inputField.GetText()

		reg := regexp.MustCompile(`[^a-zA-Z0-9]+`)
		cleanName := reg.ReplaceAllString(rawName, "")

		if cleanName != "" {
			fileName := cleanName + ".txt"
			f, err := os.Create(fileName)
			if err == nil {
				f.Close()
				addLog(tr("New file created: ", "Új fájl létrehozva: ") + fileName)
				updateList()
				app.SetRoot(pages, true)
			} else {
				addLog(tr("Error creating file: ", "Hiba a létrehozáskor: ") + err.Error())
			}
		} else {
			addLog(tr("Invalid filename!", "Érvénytelen fájlnév!"))
		}
	})

	form.AddButton(tr("Cancel", "Mégse"), func() {
		app.SetRoot(pages, true)
	})

	form.SetBorder(true).
		SetTitle(tr(" New File ", " Új fájl ")).
		SetTitleAlign(tview.AlignCenter)

	app.SetFocus(form)
	app.SetRoot(form, true)
}

func fileSelectionMenu() tview.Primitive {
	list := tview.NewList()

	var updateList func()
	updateList = func() {
		list.Clear()
		files := listAllFiles()
		for _, f := range files {
			fname := f
			list.AddItem(fname, tr("ENTER: Load file", "ENTER: Fájl betöltése"), 0, func() {
				loadTimesFromFile(fname)
				pages.SwitchToPage("times")
				app.SetFocus(pages)
			})
		}
		list.AddItem(tr("Create new file", "Új fájl létrehozása"), "", 0, func() {
			showNewFilePrompt(updateList)
		})
		list.AddItem(tr("Back/ESC", "Vissza/ESC"), "", 0, func() {
			pages.SwitchToPage("main")
			app.SetFocus(pages)
		})
	}

	updateList()
	return list
}

func dateOverrideMenu() tview.Primitive {
	inputDate := tview.NewInputField().SetLabel(tr("Date (YYMMDD): ", "Dátum (YYMMDD): "))
	inputTimes := tview.NewInputField().SetLabel(tr("Times (HH:MM:SS, ...): ", "Idők (HH:MM:SS, ...): "))
	list := tview.NewList().SetSelectedFocusOnly(true)

	var selected string

	refresh := func() {
		list.Clear()
		dateOverridesMu.RLock()
		defer dateOverridesMu.RUnlock()

		for d, times := range dateOverrides {
			display := d + " -> "
			if len(times) == 0 {
				display += tr("NONE (disabled)", "NINCS (tiltva)")
			} else {
				display += strings.Join(times, ", ")
			}

			ds := d
			list.AddItem(display, tr("Press Enter to delete", "Nyomj Enter-t a törléshez"), 0, func() {
				selected = ds
			})
		}
	}

	add := func(date string, times []string) {
		dateOverridesMu.Lock()
		dateOverrides[date] = append(dateOverrides[date], times...)
		dateOverridesMu.Unlock()

		addLog(tr("Override added: ", "Override hozzáadva: ") + date)
		refresh()
	}

	del := func(date string) {
		dateOverridesMu.Lock()
		if _, ok := dateOverrides[date]; ok {
			delete(dateOverrides, date)
			addLog(tr("Override deleted: ", "Override törölve: ") + date)
		}
		dateOverridesMu.Unlock()
		refresh()
	}

	inputTimes.SetDoneFunc(func(key tcell.Key) {
		if key != tcell.KeyEnter {
			return
		}

		dateStr := strings.TrimSpace(inputDate.GetText())
		rawTimes := strings.TrimSpace(inputTimes.GetText())

		parsedDate, err := time.Parse("060102", dateStr)
		if err != nil {
			addLog(tr("Invalid date format: ", "Hibás dátum formátum: ") + dateStr)
			return
		}

		now := time.Now()
		today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
		if parsedDate.Before(today) {
			addLog(tr("Error: Cannot set past date!", "Hiba: Múltbéli dátum nem adható meg!"))
			return
		}

		var times []string
		if rawTimes != "" {
			parts := strings.Split(rawTimes, ",")
			for _, t := range parts {
				t = strings.TrimSpace(t)
				if _, err := time.Parse("15:04:05", t); err != nil {
					addLog(tr("Invalid time format: ", "Hibás idő formátum: ") + t)
					return
				}
				times = append(times, t)
			}
		}

		add(dateStr, times)

		inputDate.SetText("")
		inputTimes.SetText("")
		app.SetFocus(inputDate)
	})

	list.SetSelectedFunc(func(index int, mainText string, secondaryText string, shortcut rune) {
		if selected != "" {
			del(selected)
			selected = ""
		}
	})

	layout := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(list, 0, 2, false).
		AddItem(inputDate, 1, 1, true).
		AddItem(inputTimes, 1, 1, false)

	layout.SetInputCapture(func(ev *tcell.EventKey) *tcell.EventKey {
		switch ev.Key() {
		case tcell.KeyTab:
			f := app.GetFocus()
			switch f {
			case inputDate:
				app.SetFocus(inputTimes)
			case inputTimes:
				app.SetFocus(list)
			default:
				app.SetFocus(inputDate)
			}
			return nil

		case tcell.KeyEsc:
			saveTimesToFile()
			pages.SwitchToPage("main")
			app.SetFocus(pages)
			return nil
		}

		if app.GetFocus() == list {
			switch ev.Rune() {
			case 'd':
				if selected != "" {
					del(selected)
					selected = ""
				}
				return nil
			case 'a':
				app.SetFocus(inputDate)
				return nil
			}
		}

		return ev
	})

	refresh()
	app.SetFocus(inputDate)

	return layout
}
