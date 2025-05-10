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
	Cx      int // Cursor X position
	Cy      int // Cursor Y position
	Loaded  bool // File loaded
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
		Loaded: false,
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
		// Display Debug Info
		if i == 0 {
			e.WriteString(fmt.Sprintf("Rows: %d, Cols: %d, Cx: %d, Cy: %d", e.Rows, e.Cols, e.Cx, e.Cy))
		}
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

// RelativeMoveCursor moves the cursor relative to its current position.
// If the new position is out of bounds, it will move to the edge. It will not wrap, it will not fail.
// TODO(Ben): add "ok" flag to indiciate if the cursor attempted to move out of bounds.
func (e *EditorConfig) RelativeMoveCursor(x, y int) {
	if x < 0 {
		e.Cx = max(e.Cx+x, 0)
	} else {
		e.Cx = min(e.Cx+x, e.Cols-1)
	}

	if y < 0 {
		e.Cy = max(e.Cy+y, 0)
	} else {
		e.Cy = min(e.Cy+y, e.Rows-1)
	}

	e.PositionCursor()
}

func (e *EditorConfig) AbsoluteMoveCursor(x, y int) {
	e.Cx = x
	e.Cy = y

	e.PositionCursor()
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: edit <file>")
		fmt.Println("Version:", Version)
		fmt.Println("Edit is a simple text editor.")
		fmt.Println("Press Ctrl-Q to exit.")
		os.Exit(1)
	}
	filePath := os.Args[1]
	fmt.Println("File Path:", filePath)

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

	// TODO(ben): this can be more efficient
	file, err := os.Open(filePath)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}
	scanner := bufio.NewScanner(file)

	for {
		// Draw the Screen
		editor.WriteString(HideCursor)
		editor.WriteString("\x1b[H")

		// Draw Rows
		if file == nil {
			for i := range editor.Rows {
				// Display Home Screen
				if i == editor.Rows/3 {
					padding := (editor.Cols - len(Version)) / 2
					if padding > 0 {
						editor.WriteString("~")
						padding -= 1
						editor.WriteString(strings.Repeat(" ", padding))
					}
					editor.WriteString(Version)
				} else {
					editor.WriteString("~")
				}
				editor.WriteString(EraseInline)

				if i < editor.Rows-1 {
					editor.WriteString("\r\n")
				}
			}
		}

		// if there is a file, read it
		if file != nil && !editor.Loaded {
			for i := range editor.Rows {
				editor.WriteString("~")
				editor.WriteString(EraseInline)
				if scanner.Scan() {
					line := scanner.Text()
					editor.WriteString(line)
				}
				if i < editor.Rows-1 {
					editor.WriteString("\r\n")
				}
			}
			editor.Loaded = true
		}


		editor.PositionCursor()
		editor.WriteString(ShowCursor)

		byte, err := buf.ReadByte()
		if err != nil {
			log.Fatalf("error reading byte: %v", err)
		}

		switch byte {
		case 'h':
			editor.RelativeMoveCursor(-1, 0)
		case 'j':
			editor.RelativeMoveCursor(0, 1)
		case 'k':
			editor.RelativeMoveCursor(0, -1)
		case 'l':
			editor.RelativeMoveCursor(1, 0)
		case ctrl('u'):
			editor.RelativeMoveCursor(0, -editor.Rows/2)
		case ctrl('d'):
			editor.RelativeMoveCursor(0, editor.Rows/2)

		case ctrl('q'):
			return // probably okay since there's no post-process, for now...
		}
	}
}

// ctrl returns the control + character for the given byte.
func ctrl(c byte) byte {
	return c & 0x1F
}
