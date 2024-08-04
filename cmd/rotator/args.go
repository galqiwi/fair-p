package main

import (
	"errors"
	"flag"
)

type args struct {
	logPath string
	maxSize int64
}

func getArgs() (args, error) {
	logPath := flag.String("log_path", "", "log file name")
	maxSize := flag.Int64("max_size", 10, "max log file size in MB")
	flag.Parse()

	if *logPath == "" {
		return args{}, errors.New("log_path should not be empty")
	}

	return args{
		logPath: *logPath,
		maxSize: *maxSize,
	}, nil
}
