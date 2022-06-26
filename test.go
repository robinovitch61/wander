package main

import (
	"bytes"
	"fmt"
	"github.com/creack/pty"
	"io"
	"os"
)

func getPtyWithoutCommand() (*os.File, error) {
	// this function just pty.StartWithAttrs with command-specific stuff commented out
	pty, tty, err := pty.Open()
	if err != nil {
		return nil, err
	}
	defer func() { _ = tty.Close() }() // Best effort.

	// if sz != nil {
	// 	if err := Setsize(pty, sz); err != nil {
	// 		_ = pty.Close() // Best effort.
	// 		return nil, err
	// 	}
	// }
	// if c.Stdout == nil {
	// 	c.Stdout = tty
	// }
	// if c.Stderr == nil {
	// 	c.Stderr = tty
	// }
	// if c.Stdin == nil {
	// 	c.Stdin = tty
	// }
	//
	// c.SysProcAttr = attrs
	//
	// if err := c.Start(); err != nil {
	// 	_ = pty.Close() // Best effort.
	// 	return nil, err
	// }
	return pty, err
}

func main() {
	myPty, err := getPtyWithoutCommand()
	if err != nil {
		panic(err)
	}

	_, err = myPty.Write([]byte("test\n"))
	if err != nil {
		panic(err)
	}
	_, err = myPty.Write([]byte{4}) // EOT
	if err != nil {
		panic(err)
	}

	buf := new(bytes.Buffer)
	_, err = io.Copy(buf, myPty)
	if err != nil {
		panic(err)
	}
	fmt.Println(buf.String())
}
