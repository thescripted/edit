package main

import (
	"bufio"
	"os"

	"golang.org/x/term"
)

const ClearScreen = "\x1b[2J"
const CursorPosition = "\x1b[H"

func main() {
	oldState, err := term.MakeRaw(int(os.Stdin.Fd()))
	if err != nil {
		panic(err)
	}

	defer func() {
		// Restore the terminal state on exit
		os.Stdout.WriteString(ClearScreen)
		os.Stdout.WriteString(CursorPosition)

		if err := term.Restore(int(os.Stdin.Fd()), oldState); err != nil {
			panic(err)
		}
	}()

	buf := bufio.NewReader(os.Stdin)
	if _, err = term.GetState(int(os.Stdin.Fd())); err != nil {
		panic(err)
	}

	for {
		refreshScreen()

		byte, err := buf.ReadByte()
		if err != nil {
			panic(err)
		}
		if byte == ctrl('q') {
			break
		}
	}
}

// ctrl returns the control + character for the given byte.
func ctrl(c byte) byte {
	return c & 0x1F
}

func refreshScreen() {
	os.Stdout.WriteString(ClearScreen)
	os.Stdout.WriteString(CursorPosition)

	// Draw `~` for each row
	for range 24 {
		os.Stdout.WriteString("~")
		os.Stdout.WriteString("\r\n")
	}

	os.Stdout.WriteString(CursorPosition)
}
