package main

import "flag"

type args struct {
	port int64
}

func getArgs() args {
	port := flag.Int64("port", 8888, "serve port")
	flag.Parse()

	return args{
		port: *port,
	}
}
