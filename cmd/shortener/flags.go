package main

import "flag"

type Config struct {
	runAddr string
	baseUrl string
}

var flags = new(Config)

func parseFlags() {
	flag.StringVar(&flags.runAddr, "a", "localhost:8080", "HTTP listen address")
	flag.StringVar(&flags.baseUrl, "b", "http://localhost:8080/", "Base URL")

	flag.Parse()
}
