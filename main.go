package main

import (
	"bufio"
	"fmt"
	"os"
	"unicode"

	"golang.org/x/term"
)

func main() {
	oldState, err := term.MakeRaw(int(os.Stdin.Fd()))
	if err != nil {
		panic(err)
	}

	defer func() {
		if err := term.Restore(int(os.Stdin.Fd()), oldState); err != nil {
			panic(err)
		}
	}()

	buf := bufio.NewReader(os.Stdin)
	s, err := term.GetState(int(os.Stdin.Fd()))
	if err != nil {
		panic(err)
	}
	fmt.Printf("%#v\r\n", s)

	for {
		rune, _, err := buf.ReadRune()
		if err != nil {
			panic(err)
		}
		if rune == 'q' {
			break
		}
		if unicode.IsControl(rune) {
			fmt.Printf("%d\r\n", rune)
		} else {
			fmt.Printf("%d ('%c')\r\n", rune, rune)
		}
	}
}
