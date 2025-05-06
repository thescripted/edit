package main

import (
	"bufio"
	"log"
	"os"

	"golang.org/x/term"
)

const ClearScreen = "\x1b[2J"
const CursorPosition = "\x1b[H"

type EditorConfig struct {
	termios *term.State
	Rows	int
	Cols	int
}

func NewEditorConfig() (*EditorConfig, error) {
	w,h, err := term.GetSize(int(os.Stdin.Fd()))
	if err != nil {
		return nil, err
	}

	return &EditorConfig{
		termios: nil,
		Rows:    h,
		Cols:    w,
	}, nil
}

func (e *EditorConfig) RefreshScreen() {
	os.Stdout.WriteString(ClearScreen)
	os.Stdout.WriteString(CursorPosition)

	// Draw `~` for each row
	for i := 0; i < e.Rows; i++ {
		os.Stdout.WriteString("~")
		os.Stdout.WriteString("\r\n")
	}

	os.Stdout.WriteString(CursorPosition)
}

func main() {
	oldState, err := term.MakeRaw(int(os.Stdin.Fd()))
	if err != nil {
		log.Fatalf("error making raw terminal: %v", err)
	}

	editor, err := NewEditorConfig()
	if err != nil {
		log.Fatalf("error creating editor config: %v", err)
	}

	defer func() { // TODO: editor.Restore() ?? 
		// Restore the terminal state on exit
		os.Stdout.WriteString(ClearScreen)
		os.Stdout.WriteString(CursorPosition)

		if err := term.Restore(int(os.Stdin.Fd()), oldState); err != nil {
			log.Fatalf("error restoring terminal: %v", err)
		}
	}()

	buf := bufio.NewReader(os.Stdin)
	if _, err = term.GetState(int(os.Stdin.Fd())); err != nil {
		log.Fatalf("error getting terminal state: %v", err)
	}

	for {
		editor.RefreshScreen()

		byte, err := buf.ReadByte()
		if err != nil {
			log.Fatalf("error reading byte: %v", err)
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
