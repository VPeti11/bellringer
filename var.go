package main

import (
	"bufio"
	"net/http"
	"sync"
	"time"

	"github.com/faiface/beep"
	"github.com/faiface/beep/effects"
	"github.com/gorilla/websocket"
	"github.com/rivo/tview"
	"go.bug.st/serial"
)

var (
	app             = tview.NewApplication()
	pages           = tview.NewPages()
	logLines        = []string{}
	updateTimesMenu func()
	enabled         = true
	pulseMode       = false
	pulseRunning    = false
	enableWeekend   = false
	statusText      = "LOW"

	currentTimeFile = "idobeall1.txt"
	schedulesMu     sync.RWMutex
	schedules       = map[time.Weekday][]string{
		time.Monday:    {},
		time.Tuesday:   {},
		time.Wednesday: {},
		time.Thursday:  {},
		time.Friday:    {},
		time.Saturday:  {},
		time.Sunday:    {},
	}

	KivalasztottPort string
	port             serial.Port
	reader           *bufio.Reader
	noserial         bool

	bellRinging        bool
	speakerInitialized bool
	ctrl               *beep.Ctrl
	volume             *effects.Volume
	lastRingTime       time.Time

	webServer       *http.Server
	webOn           bool
	webEnabled      bool
	webPort         string
	webUsername     string
	webPasswordHash string

	shortTimesMu sync.RWMutex
	shortTimes   = map[string]bool{}

	wsMu      sync.Mutex
	wsClients = make(map[*websocket.Conn]bool)

	dateOverridesMu sync.RWMutex
	dateOverrides   = map[string][]string{}

	useEnglish bool
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

const (
	debounceDuration = 15 * time.Second
	sebesseg         = 115200
)
