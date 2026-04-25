package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
)

const (
	ProgramName = "bellringer"
	AuthorName  = "VPeti11"
	GitRepoURL  = "https://github.com/VPeti11/bellringer.git"
)

var lang = "en"

var translations = map[string]map[string]string{
	"en": {
		"welcome":     "Welcome to %s installer",
		"madeby":      "Made by %s",
		"press_enter": "Press Enter to continue...",
		"installing":  "Installing dependencies...",
		"cloning":     "Cloning repository...",
		"removing":    "Removing precompiled binary (if any)...",
		"building":    "Building bellringer...",
		"linking":     "Linking to /usr/bin...",
		"success":     "\nbellringer installed successfully!",
		"run":         "Run it with: bellringer",
		"ask_lang":    "Choose language (en/hu): ",
		"no_pm":       "No supported package manager found (apt/dnf/pacman)",
		"linux_only":  "This installer only supports Linux",
	},
	"hu": {
		"welcome":     "Üdv a %s telepítőben",
		"madeby":      "Készítette: %s",
		"press_enter": "Nyomj Entert a folytatáshoz...",
		"installing":  "Függőségek telepítése...",
		"cloning":     "Repository klónozása...",
		"removing":    "Előre fordított bináris törlése (ha van)...",
		"building":    "bellringer buildelése...",
		"linking":     "Linkelés /usr/bin-be...",
		"success":     "\nbellringer sikeresen telepítve!",
		"run":         "Futtatás: bellringer",
		"ask_lang":    "Válassz nyelvet (en/hu): ",
		"no_pm":       "Nincs támogatott csomagkezelő (apt/dnf/pacman)",
		"linux_only":  "Ez a telepítő csak Linuxot támogat",
	},
}

func t(key string) string {
	return translations[lang][key]
}

func main() {
	AskLanguage()
	CheckLinux()

	ShowWelcome()

	pm := DetectPackageManager()
	if pm == "" {
		log.Fatal(t("no_pm"))
	}

	fmt.Println(t("installing"))
	if err := InstallDependencies(pm); err != nil {
		log.Fatal(err)
	}

	repoDir, err := BellringerDir()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(t("cloning"))
	if err := CloneRepo(repoDir); err != nil {
		log.Fatal(err)
	}

	fmt.Println(t("removing"))
	if err := RemoveExistingBinary(repoDir); err != nil {
		log.Fatal(err)
	}

	fmt.Println(t("building"))
	if err := BuildBinary(repoDir); err != nil {
		log.Fatal(err)
	}

	fmt.Println(t("linking"))
	if err := LinkBinary(repoDir); err != nil {
		log.Fatal(err)
	}

	fmt.Println(t("success"))
	fmt.Println(t("run"))
}

func AskLanguage() {
	fmt.Print(translations["en"]["ask_lang"])
	input, _ := bufio.NewReader(os.Stdin).ReadString('\n')

	if input == "hu\n" {
		lang = "hu"
	}
}

func CheckLinux() {
	if runtime.GOOS != "linux" {
		log.Fatal(t("linux_only"))
	}
}

func BellringerDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".bellringer"), nil
}

func CloneRepo(dir string) error {
	_ = os.RemoveAll(dir)

	cmd := exec.Command("git", "clone", GitRepoURL, dir)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func RemoveExistingBinary(repoDir string) error {
	bin := filepath.Join(repoDir, ProgramName)
	if _, err := os.Stat(bin); err == nil {
		return os.Remove(bin)
	}
	return nil
}

func BuildBinary(repoDir string) error {
	cmd := exec.Command("go", "build", "-o", ProgramName)
	cmd.Dir = repoDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func LinkBinary(repoDir string) error {
	src := filepath.Join(repoDir, ProgramName)
	dst := filepath.Join("/usr/bin", ProgramName)

	_ = exec.Command("sudo", "rm", "-f", dst).Run()

	cmd := exec.Command("sudo", "ln", "-s", src, dst)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func DetectPackageManager() string {
	switch {
	case CommandExists("apt"):
		return "apt"
	case CommandExists("dnf"):
		return "dnf"
	case CommandExists("pacman"):
		return "pacman"
	default:
		return ""
	}
}

func CommandExists(cmd string) bool {
	_, err := exec.LookPath(cmd)
	return err == nil
}

func InstallDependencies(pm string) error {
	switch pm {
	case "apt":
		exec.Command("sudo", "apt", "update").Run()
		return exec.Command("sudo", "apt", "install", "-y", "git", "go").Run()
	case "dnf":
		return exec.Command("sudo", "dnf", "install", "-y", "git", "go").Run()
	case "pacman":
		return exec.Command("sudo", "pacman", "-Syu", "--noconfirm", "git", "go").Run()
	}
	return nil
}

func ShowWelcome() {
	clear()
	fmt.Printf(t("welcome")+"\n", ProgramName)
	fmt.Printf(t("madeby")+"\n\n", AuthorName)
	fmt.Print(t("press_enter"))
	bufio.NewReader(os.Stdin).ReadBytes('\n')
	clear()
}

func clear() {
	cmd := exec.Command("clear")
	cmd.Stdout = os.Stdout
	cmd.Run()
}
