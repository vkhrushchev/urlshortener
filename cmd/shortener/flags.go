package main

import (
	"flag"
	"os"
)

type Config struct {
	runAddr string
	baseURL string
}

var flags = new(Config)

func parseFlags() {
	flag.StringVar(&flags.runAddr, "a", "localhost:8080", "HTTP listen address")
	flag.StringVar(&flags.baseURL, "b", "http://localhost:8080/", "Base URL")

	flag.Parse()

	if serverAddrEnv := os.Getenv("SERVER_ADDR"); serverAddrEnv != "" {
		flags.runAddr = serverAddrEnv
	}

	if baseURLEnv := os.Getenv("BASE_URL"); baseURLEnv != "" {
		flags.baseURL = baseURLEnv
	}
}
