package main

import (
	"fmt"
	"os"
)

func main() {
	err := Main()
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
}

func Main() error {
	r, err := NewRunner(getArgs())
	if err != nil {
		return err
	}
	return r.Run()
}
