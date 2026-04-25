package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"go.bug.st/serial/enumerator"
	"golang.org/x/crypto/bcrypt"
	"golang.org/x/term"
)

func configurate() {
	var configPath = "config.br"

	if _, err := os.Stat(configPath); err == nil {
		loadAllConfig(configPath)
		return
	}

	reader := bufio.NewReader(os.Stdin)

	fmt.Println("Select language / Válassz nyelvet:")
	fmt.Println("1. English")
	fmt.Println("2. Magyar")
	fmt.Print("Choice/Választás (1/2): ")
	langInput, _ := reader.ReadString('\n')
	langInput = strings.TrimSpace(langInput)

	engVal := "0"
	if langInput == "1" {
		engVal = "1"
		useEnglish = true
	} else {
		useEnglish = false
	}

	askSerialPort(reader)

	fmt.Print(tr("\nEnable web server? (true/false) [true]: ", "\nEngedélyezze a web szervert? (true/false) [true]: "))
	enabledInput, _ := reader.ReadString('\n')
	enabledInput = strings.TrimSpace(enabledInput)
	if enabledInput == "" {
		enabledInput = "true"
	}
	webEnabled = (enabledInput == "true")

	fmt.Print(tr("Web server port [8074]: ", "Port a web szerverhez [8074]: "))
	portInput, _ := reader.ReadString('\n')
	portInput = strings.TrimSpace(portInput)
	if portInput == "" {
		portInput = "8074"
	}
	webPort = portInput

	fmt.Print(tr("Username [admin]: ", "Felhasználónév [admin]: "))
	userInput, _ := reader.ReadString('\n')
	userInput = strings.TrimSpace(userInput)
	if userInput == "" {
		userInput = "admin"
	}
	webUsername = userInput

	fmt.Print(tr("Password [1234]: ", "Jelszó [1234]: "))
	bytePassword, err := term.ReadPassword(int(os.Stdin.Fd()))
	fmt.Println()

	password := "1234"
	if err == nil && len(bytePassword) > 0 {
		password = strings.TrimSpace(string(bytePassword))
	}

	hash, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	webPasswordHash = string(hash)

	saveAllConfig(engVal, configPath)
	clearScreen()
}

func askSerialPort(reader *bufio.Reader) {
	ports, err := enumerator.GetDetailedPortsList()
	if err != nil || len(ports) == 0 {
		addLog(tr("No serial ports found. Serial disabled.", "Nem található soros port. Soros kapcsolat kikapcsolva."))
		noserial = true
		return
	}

	fmt.Println(tr("\nAvailable serial ports:", "\nElérhető soros portok:"))
	for i, port := range ports {
		fmt.Printf("[%d] %s", i, port.Name)
		if port.IsUSB {
			fmt.Printf(" (USB VID: %s PID: %s)", port.VID, port.PID)
		}
		fmt.Println()
	}

	fmt.Print(tr("Select port number (or type 'no'): ", "Válasszon port számot (vagy írjon 'no'-t): "))
	valasz, _ := reader.ReadString('\n')
	valasz = strings.TrimSpace(strings.ToLower(valasz))

	if valasz == "no" || valasz == "" {
		noserial = true
		KivalasztottPort = ""
		return
	}

	var index int
	_, err = fmt.Sscanf(valasz, "%d", &index)
	if err != nil || index < 0 || index >= len(ports) {
		fmt.Println(tr("Invalid selection, disabling serial.", "Érvénytelen választás, soros port kikapcsolva."))
		noserial = true
		KivalasztottPort = "no"
	} else {
		KivalasztottPort = ports[index].Name
		noserial = false
		fmt.Println(tr("Selected: ", "Kiválasztva: "), KivalasztottPort)
	}
}

func loadAllConfig(configPath string) {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return
	}

	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}
		key, value := parts[0], parts[1]

		switch key {
		case "eng":
			useEnglish = (value == "1")
		case "serial_port":
			KivalasztottPort = value
			noserial = (value == "no")
		case "web_enabled":
			webEnabled = (value == "true")
		case "web_port":
			webPort = value
		case "web_username":
			webUsername = value
		case "web_password_hash":
			webPasswordHash = value
		}
	}
}

func saveAllConfig(engVal string, configPath string) {
	if engVal == "" {
		engVal = "0"
		if useEnglish {
			engVal = "1"
		}
	}

	content := fmt.Sprintf(
		"eng=%s\nserial_port=%s\nweb_enabled=%t\nweb_port=%s\nweb_username=%s\nweb_password_hash=%s\n",
		engVal, KivalasztottPort, webEnabled, webPort, webUsername, webPasswordHash,
	)

	os.WriteFile(configPath, []byte(content), 0644)
}
