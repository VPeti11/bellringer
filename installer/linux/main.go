package installer

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"time"
)

// --- Constants and Config ---

const (
	ProgramName  = "bellringer"
	AuthorName   = "VPeti11"
	GitRepoURL   = "https://vpeti11/bellringer.git"
	LocalRepoDir = "yourrepo"
)

var Dependencies = map[string][]string{
	"apt":    {"go", "git"},
	"dnf":    {"go", "git"},
	"pacman": {"go", "git"},
}

// --- Main ---  Update this too to reflect your codebase

func main() {
	if err := CheckLinuxPlatform(); err != nil {
		log.Fatalf("Unsupported platform: %v", err)
	}

	ShowWelcomeMessage()

	pkgManager := DetectPackageManager()
	if pkgManager == "" {
		log.Fatalf("No supported package manager found.")
	}

	fmt.Printf("Using package manager: %s\n", pkgManager)
	if err := InstallDependencies(pkgManager); err != nil {
		log.Fatalf("Failed to install dependencies: %v", err)
	}

	fmt.Println("Cloning git repository...")
	if err := CloneGitRepo(GitRepoURL); err != nil {
		log.Fatalf("Failed to clone repo: %v", err)
	}

	if err := ChangeDirectory(LocalRepoDir); err != nil {
		log.Fatalf("Failed to change directory: %v", err)
	}

	fmt.Println("Building Go binary...")
	if err := InstallGoBinary("main.go", ProgramName); err != nil {
		log.Fatalf("Failed to build Go binary: %v", err)
	}

	PromptContinue("All installation steps completed successfully! Press Enter to exit.")
}

func ClearScreen() {
	cmd := exec.Command("clear")
	cmd.Stdout = os.Stdout
	_ = cmd.Run()
}

func CommandExists(cmd string) bool {
	_, err := exec.LookPath(cmd)
	return err == nil
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

func InstallDependencies(manager string) error {
	pkgs, ok := Dependencies[manager]
	if !ok {
		return fmt.Errorf("unsupported package manager: %s", manager)
	}

	switch manager {
	case "apt":
		_ = exec.Command("sudo", "apt", "update").Run()
		args := append([]string{"apt", "install", "-y"}, pkgs...)
		return exec.Command("sudo", args...).Run()
	case "dnf":
		args := append([]string{"dnf", "install", "-y"}, pkgs...)
		return exec.Command("sudo", args...).Run()
	case "pacman":
		args := append([]string{"pacman", "-Syu", "--noconfirm"}, pkgs...)
		return exec.Command("sudo", args...).Run()
	}
	return nil
}

func InstallGoBinary(sourceFile string, outName string) error {
	outPath := filepath.Join("/usr/bin", outName)
	cmd := exec.Command("go", "build", "-o", outPath, sourceFile)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return err
	}
	return exec.Command("chmod", "+x", outPath).Run()
}

func CloneGitRepo(url string) error {
	cmd := exec.Command("git", "clone", url)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func ChangeDirectory(dir string) error {
	return os.Chdir(dir)
}

func PromptContinue(message string) {
	ClearScreen()
	fmt.Println(message)
	fmt.Println("Press Enter to continue...")
	bufio.NewReader(os.Stdin).ReadBytes('\n')
}

func CheckLinuxPlatform() error {
	if runtime.GOOS != "linux" {
		return fmt.Errorf("this installer only supports Linux")
	}
	return nil
}

func SleepMessage(message string, duration time.Duration) {
	fmt.Println(message)
	time.Sleep(duration)
}

func ShowWelcomeMessage() {
	ClearScreen()
	fmt.Printf("Welcome to the %s installer\n", ProgramName)
	fmt.Printf("Made by %s\n\n", AuthorName)
	fmt.Println("Press Enter to continue...")
	bufio.NewReader(os.Stdin).ReadBytes('\n')
}
