package helpers

import (
	"bufio"
	"os"
	"os/exec"
)

func FlushConsole() {
	command := exec.Command("clear")
	command.Stdout = os.Stdout
	_ = command.Run()
}

func ReadLines(path string) (lines []string, err error) {
	file, err := os.Open(path)
	if err != nil {
		return lines, err
	}
	defer func(file *os.File) {
		_ = file.Close()
	}(file)

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		lines = append(lines, line)
	}
	return lines, nil
}
