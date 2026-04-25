package main

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/faiface/beep"
	"github.com/faiface/beep/effects"
	"github.com/faiface/beep/mp3"
	"github.com/faiface/beep/speaker"
	"go.bug.st/serial"
)

func playMP3() {
	if bellRinging {
		return
	}
	bellRinging = true

	f, err := os.Open("ring.mp3")
	if err != nil {
		bellRinging = false
		addLog("Hangfájl nem található: ring.mp3")
		return
	}

	streamer, format, err := mp3.Decode(f)
	if err != nil {
		bellRinging = false
		addLog("MP3 dekódolási hiba")
		_ = f.Close()
		return
	}

	if !speakerInitialized {
		speaker.Init(format.SampleRate, format.SampleRate.N(time.Second/10))
		speakerInitialized = true
	}

	ctrl = &beep.Ctrl{Streamer: streamer, Paused: false}
	volume = &effects.Volume{
		Streamer: ctrl,
		Base:     2,
		Volume:   0,
		Silent:   false,
	}

	speaker.Play(volume)
}

func jelKuldes(jel string) error {
	if noserial {
		return nil
	}
	if KivalasztottPort == "" {
		addLog("Hiba: nincs kiválasztott port")
		return fmt.Errorf("nincs kiválasztott port")
	}

	if jel != "HIGH" && jel != "LOW" {
		addLog(fmt.Sprintf("Hiba: érvénytelen jel: %s", jel))
		return fmt.Errorf("érvénytelen jel: %s", jel)
	}

	mode := &serial.Mode{
		BaudRate: sebesseg,
	}

	port, err := serial.Open(KivalasztottPort, mode)
	if err != nil {
		addLog(fmt.Sprintf("Hiba port megnyitásakor: %v", err))
		return err
	}
	defer port.Close()

	_, err = port.Write([]byte(jel + "\n"))
	if err != nil {
		addLog(fmt.Sprintf("Hiba jel küldésekor: %v", err))
		return err
	}
	addLog(fmt.Sprintf("Jel elküldve: %s", jel))

	time.Sleep(100 * time.Millisecond)

	buf := make([]byte, 100)
	n, err := port.Read(buf)
	if err != nil {
		addLog(fmt.Sprintf("Hiba a válasz olvasásakor: %v", err))
		return err
	}

	valasz := strings.TrimSpace(string(buf[:n]))
	addLog(fmt.Sprintf("Pico válasza: %s", valasz))

	return nil
}

func stopRing() {
	if !bellRinging {
		return
	}
	bellRinging = false

	fade := 250 * time.Millisecond
	steps := 25
	stepDur := fade / time.Duration(steps)

	go func() {
		for i := 0; i < steps; i++ {
			speaker.Lock()
			volume.Volume -= 1.0 / float64(steps)
			speaker.Unlock()
			time.Sleep(stepDur)
		}

		speaker.Lock()
		ctrl.Paused = true
		speaker.Unlock()
	}()
}
