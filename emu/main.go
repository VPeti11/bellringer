package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/tarm/serial"
)

var lang = "en"

var translations = map[string]map[string]string{
	"en": {
		"ask_lang":   "Choose language (en/hu): ",
		"ask_port":   "Enter serial port (e.g. COM8 or /dev/ttyUSB0): ",
		"port_error": "Failed to open port:",
		"running":    "Pico USB CDC emulator running on port:",
		"high":       "GPIO1 = HIGH",
		"low":        "GPIO1 = LOW",
		"unknown":    "ERROR UNKNOWN COMMAND:",
	},
	"hu": {
		"ask_lang":   "Válassz nyelvet (en/hu): ",
		"ask_port":   "Add meg a soros portot (pl. COM8 vagy /dev/ttyUSB0): ",
		"port_error": "Nem sikerült megnyitni a portot:",
		"running":    "Pico USB CDC emulátor fut a következő porton:",
		"high":       "GPIO1 = MAGAS",
		"low":        "GPIO1 = ALACSONY",
		"unknown":    "HIBA ISMERETLEN PARANCS:",
	},
}

func t(key string) string {
	return translations[lang][key]
}

func AskLanguage() {
	fmt.Print(translations["en"]["ask_lang"])
	input, _ := bufio.NewReader(os.Stdin).ReadString('\n')
	input = strings.TrimSpace(input)

	if input == "hu" {
		lang = "hu"
	}
}

func main() {

	AskLanguage()

	fmt.Print(t("ask_port"))
	inputReader := bufio.NewReader(os.Stdin)
	port, _ := inputReader.ReadString('\n')
	port = strings.TrimSpace(port)

	baud := 115200

	config := &serial.Config{
		Name:        port,
		Baud:        baud,
		ReadTimeout: time.Millisecond * 50,
	}

	s, err := serial.OpenPort(config)
	if err != nil {
		log.Fatal(t("port_error"), err)
	}
	defer s.Close()

	fmt.Println(t("running"), port)

	var buffer strings.Builder

	for {

		buf := make([]byte, 1)
		n, err := s.Read(buf)
		if err != nil || n == 0 {
			continue
		}

		c := string(buf[0])
		if c == "\n" || c == "\r" {
			parancs := buffer.String()
			buffer.Reset()

			switch parancs {
			case "HIGH":
				s.Write([]byte("OK HIGH\n"))
				fmt.Println(t("high"))
			case "LOW":
				s.Write([]byte("OK LOW\n"))
				fmt.Println(t("low"))
			case "":

			default:
				s.Write([]byte("ERR ISMERETLEN\n"))
				fmt.Println(t("unknown"), parancs)
			}
		} else {

			buffer.WriteString(c)
		}
	}
}
