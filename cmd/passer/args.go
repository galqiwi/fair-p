package main

import (
	"flag"
	"fmt"
	"golang.org/x/time/rate"
	"time"
)

type args struct {
	port               int
	runtimeLogInterval time.Duration
	maxThroughput      rate.Limit
	noIPv4             bool
}

func getArgs() (args, error) {
	port := flag.Int("port", 8888, "serve port")
	runtimeLogIntervalS := flag.Float64("runtime_log_interval_sec", 10., "runtime log interval")
	maxThroughput := flag.Float64("max_throughput", 0, "Max throughput (MB/s)")
	noIPv4 := flag.Bool("no_ipv4", false, "disable ipv4 (optimisation for dns64 systems)")
	flag.Parse()

	if *maxThroughput == float64(0) {
		return args{}, fmt.Errorf("max throughput must be greater than zero")
	}

	return args{
		port:               *port,
		runtimeLogInterval: time.Duration(float64(time.Second) * *runtimeLogIntervalS),
		maxThroughput:      rate.Limit(*maxThroughput * 1024 * 1024),
		noIPv4:             *noIPv4,
	}, nil
}
