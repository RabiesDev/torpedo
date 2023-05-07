package utils

import (
	"os"
	"os/exec"
)

func FlushConsole() {
	cmd := exec.Command("clear")
	cmd.Stdout = os.Stdout
	cmd.Run()
}

func DecodeSecret(secret []int) []byte {
	decodeResult := make([]byte, 24)
	runningValue := 0

	for i := 0; i < 24; i++ {
		charCode1 := secret[17+i*2]
		char1 := charCode1
		if charCode1 <= 96 {
			char1 += 32
		}
		char1 = (char1 - 98 - i*34) % 26
		if char1 < 0 {
			char1 += 26
		}

		charCode2 := secret[18+i*2]
		char2 := charCode2
		if charCode2 <= 96 {
			char2 += 32
		}
		char2 = (char2 - 115 - i*34) % 26
		if char2 < 0 {
			char2 += 26
		}

		interim := (char1 << 4) | char2
		offset := 65
		if interim >= 97 {
			offset = 97
		}
		interim -= offset
		if i == 0 {
			runningValue = 2 + interim
		}

		decodeResult[i] = byte((interim+runningValue)%26 + offset)
		runningValue += 3 + interim
	}

	return decodeResult
}
