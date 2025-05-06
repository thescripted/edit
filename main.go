package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"

	"golang.org/x/term"
)

const (
	// Escape Sequences
	ClearScreen = "\x1b[2J"
	EraseInline = "\x1b[K"
	// CursorPosition = "\x1b[%d;%dH" // TODO: this needs dynamic args
	HideCursor = "\x1b[?25l"
	ShowCursor = "\x1b[?25h"
)

const Version = "Edit -- Version 0.0.1"

type EditorConfig struct {
	termios *term.State
	Rows    int
	Cols    int
	writer  *bufio.Writer
	Cx      int
	Cy      int
}

func (e *EditorConfig) WriteBytes(b []byte) error {
	_, err := e.writer.Write(b)
	if err != nil {
		return err
	}
	return e.writer.Flush()
}

func (e *EditorConfig) WriteString(s string) error {
	_, err := e.writer.WriteString(s)
	if err != nil {
		return err
	}
	return e.writer.Flush()
}

func NewEditorConfig() (*EditorConfig, error) {
	w, h, err := term.GetSize(int(os.Stdin.Fd()))
	if err != nil {
		return nil, err
	}

	return &EditorConfig{
		termios: nil,
		Rows:    h,
		Cols:    w,
		writer:  bufio.NewWriter(os.Stdout),
	}, nil
}

func (e *EditorConfig) PositionCursor() {
	e.WriteString(fmt.Sprintf("\x1b[%d;%dH", e.Cy+1, e.Cx+1))
}

func (e *EditorConfig) RefreshScreen() {
	e.WriteString(HideCursor)
	e.WriteString("\x1b[H")

	// Draw Rows
	for i := range e.Rows {
		// Display Home Screen
		if i == e.Rows/3 {
			padding := (e.Cols - len(Version)) / 2
			if padding > 0 {
				e.WriteString("~")
				padding -= 1
				e.WriteString(strings.Repeat(" ", padding))
			}
			e.WriteString(Version)
		} else {
			e.WriteString("~")
		}
		e.WriteString(EraseInline)

		if i < e.Rows-1 {
			e.WriteString("\r\n")
		}
	}

	e.PositionCursor()
	e.WriteString(ShowCursor)
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

	// Restore the terminal state on exit
	defer func() { // TODO: editor.Restore() ??
		os.Stdout.WriteString(ClearScreen)
		os.Stdout.WriteString("\x1b[H")

		if err := term.Restore(int(os.Stdin.Fd()), oldState); err != nil {
			log.Fatalf("error restoring terminal: %v", err)
		}

		// debug
		fmt.Printf("%#v\n", editor)
		fmt.Println("Exited")
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
