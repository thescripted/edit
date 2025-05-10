package main

import (
	"bufio"
	"fmt"
	"io"
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
	termios      *term.State
	Rows         int
	Cols         int
	writer       *bufio.Writer
	AbsCx        int
	AbsCy        int
	RelCx        int
	RelCy        int
	Ready        bool
	Content      []string
	ContentWidth int
	ViewStart    int
	ViewEnd      int
}

func (e *EditorConfig) WriteBytes(b []byte) error {
	_, err := e.writer.Write(b)
	if err != nil {
		return err
	}
	return e.writer.Flush()
}

func (e *EditorConfig) WriteString(s string) error {
	stringIntoFrame := s[:min(len(s), e.Cols-2)]
	_, err := e.writer.WriteString(stringIntoFrame)
	if err != nil {
		return err
	}
	return e.writer.Flush()
}

func NewEditorConfig(r io.Reader) (*EditorConfig, error) {
	w, h, err := term.GetSize(int(os.Stdin.Fd()))
	if err != nil {
		return nil, err
	}
	scanner := bufio.NewScanner(r)
	// Write the file content to the editorConfig
	content := []string{}
	for scanner.Scan() {
		line := scanner.Text()
		content = append(content, line)
	}

	return &EditorConfig{
		termios:      nil,
		Rows:         h,
		Cols:         w,
		AbsCx:        2,
		RelCx:        2,
		writer:       bufio.NewWriter(os.Stdout),
		Ready:        false,
		Content:      content,
		ViewEnd:      h,
		ContentWidth: len(content),
	}, nil
}

func (e *EditorConfig) UpdateViewPort() {
	diff := 0
	if e.AbsCy >= e.ViewEnd {
		diff = e.AbsCy - e.ViewEnd
	}
	if e.AbsCy < e.ViewStart {
		diff = e.AbsCy - e.ViewStart
	}
	e.ViewStart += diff
	e.ViewEnd += diff
}

func (e *EditorConfig) PositionCursor() {
	e.WriteString(fmt.Sprintf("\x1b[%d;%dH", e.RelCy+1, e.RelCx+1))
}

func (e *EditorConfig) RefreshScreen() {
	e.WriteString(HideCursor)
	e.WriteString("\x1b[H")

	// Draw Rows
	for i := e.ViewStart; i < e.Rows; i++ {
		// Display Home Screen
		if i == e.Rows/3 && len(e.Content) == 0 {
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
		e.WriteString(e.Content[i])

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
		e.AbsCx = max(e.AbsCx+x, 2)
		e.RelCx = max(e.RelCx+x, 2)
	} else {
		e.AbsCx = min(e.AbsCx+x, e.Cols-1)
		e.RelCx = min(e.RelCx+x, e.Cols-1)
	}

	if y < 0 {
		e.AbsCy = max(e.AbsCy+y, 0)
		e.RelCy = max(e.RelCy+y, 0)
	} else {
		e.AbsCy = min(e.AbsCy+y, e.ContentWidth-1)
		e.RelCy = min(e.RelCy+y, e.Rows-1) // This is basically our Screen size.
	}

	e.UpdateViewPort()
	e.PositionCursor()
}

func (e *EditorConfig) AbsoluteMoveCursor(x, y int) {
	e.AbsCx = x
	e.AbsCy = y

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

	// TODO(ben): this can be more efficient
	file, err := os.Open(filePath)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}
	editor, err := NewEditorConfig(file)
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
	}()

	buf := bufio.NewReader(os.Stdin)
	if _, err = term.GetState(int(os.Stdin.Fd())); err != nil {
		log.Fatalf("error getting terminal state: %v", err)
	}

	for {
		// Draw the Screen
		editor.WriteString(HideCursor)
		editor.WriteString("\x1b[H")

		for i := editor.ViewStart; i < editor.ViewEnd; i++ {
			// Display Home Screen
			if i == editor.Rows/3 && len(editor.Content) == 0 {
				padding := (editor.Cols - len(Version)) / 2
				if padding > 0 {
					editor.WriteString("~")
					padding -= 1
					editor.WriteString(strings.Repeat(" ", padding))
				}
				editor.WriteString(Version)
			} else {
				editor.WriteString("~ ") // add a space
			}
			editor.WriteString(EraseInline)
			editor.WriteString(editor.Content[i])

			if i < editor.ViewEnd-1 {
				editor.WriteString("\r\n")
			}
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
