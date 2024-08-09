package main

import (
	"flag"
	"time"
)

type args struct {
	port               int
	runtimeLogInterval time.Duration
}

func getArgs() args {
	port := flag.Int("port", 8888, "serve port")
	runtimeLogIntervalS := flag.Float64("runtime_log_interval_sec", 10., "runtime log interval")
	flag.Parse()

	return args{
		port:               *port,
		runtimeLogInterval: time.Duration(float64(time.Second) * *runtimeLogIntervalS),
	}
}
