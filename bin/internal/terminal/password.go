package terminal

import (
	"fmt"
	"os"

	"golang.org/x/term"
)

// ReadPassword reads a password from stdin, printing '*' for each character.
func ReadPassword(prompt string) (string, error) {
	fmt.Print(prompt)

	fd := int(os.Stdin.Fd())

	oldState, err := term.MakeRaw(fd)
	if err != nil {
		// Fallback: read without echo (no * but still invisible).
		fmt.Println()
		bytePwd, err := term.ReadPassword(fd)
		if err != nil {
			return "", err
		}
		return string(bytePwd), nil
	}
	defer term.Restore(fd, oldState)

	var pwd []byte
	buf := make([]byte, 1)

	for {
		n, err := os.Stdin.Read(buf)
		if err != nil || n != 1 {
			break
		}

		switch buf[0] {
		case 13, 10: // CR / LF → done
			fmt.Print("\r\n")
			return string(pwd), nil
		case 127, 8: // DEL / Backspace
			if len(pwd) > 0 {
				pwd = pwd[:len(pwd)-1]
				fmt.Print("\b \b")
			}
		case 3: // Ctrl+C
			fmt.Print("\r\n")
			os.Exit(1)
		default:
			if buf[0] >= 32 && buf[0] <= 126 { // printable
				pwd = append(pwd, buf[0])
				fmt.Print("*")
			}
		}
	}

	return string(pwd), nil
}
