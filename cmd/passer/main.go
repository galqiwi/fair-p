package main

import (
	"fmt"
	"os"
)

func main() {
	err := Main()
	if err != nil {
		_, _ = fmt.Fprint(os.Stderr, err.Error())
		os.Exit(1)
	}
}

func Main() error {
	args, err := getArgs()
	if err != nil {
		return err
	}
	r, err := NewRunner(args)
	if err != nil {
		return err
	}
	return r.Run()
}
