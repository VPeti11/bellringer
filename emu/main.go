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

func main() {

	fmt.Print("Add meg a soros portot (pl. COM8 vagy /dev/ttyUSB0): ")
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
		log.Fatal("Nem sikerült megnyitni a portot:", err)
	}
	defer s.Close()

	fmt.Println("Pico USB CDC emulátor fut a következő porton:", port)

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
				fmt.Println("GPIO1 = MAGAS")
			case "LOW":
				s.Write([]byte("OK LOW\n"))
				fmt.Println("GPIO1 = ALACSONY")
			case "":

			default:
				s.Write([]byte("ERR ISMERETLEN\n"))
				fmt.Println("HIBA ISMERETLEN PARANCS:", parancs)
			}
		} else {

			buffer.WriteString(c)
		}
	}
}
