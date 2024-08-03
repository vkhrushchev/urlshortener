package main

import (
	"flag"
	"os"
)

type Config struct {
	runAddr            string
	baseURL            string
	fileStoragePathEnv string
	databaseDSN        string
}

var flags = new(Config)

func parseFlags() {
	flag.StringVar(&flags.runAddr, "a", "localhost:8080", "HTTP listen address")
	flag.StringVar(&flags.baseURL, "b", "http://localhost:8080/", "Base URL")
	flag.StringVar(&flags.fileStoragePathEnv, "f", "./short_url_json_storage", "Short URL JSON storage")
	flag.StringVar(&flags.databaseDSN, "d", "postgres://yp-sandbox:yp-sandbox@localhost:5432/yp-sandbox", "Database DSN")

	flag.Parse()

	if serverAddrEnv := os.Getenv("SERVER_ADDR"); serverAddrEnv != "" {
		flags.runAddr = serverAddrEnv
	}

	if baseURLEnv := os.Getenv("BASE_URL"); baseURLEnv != "" {
		flags.baseURL = baseURLEnv
	}

	if fileStoragePathEnv := os.Getenv("FILE_STORAGE_PATH"); fileStoragePathEnv != "" {
		flags.fileStoragePathEnv = fileStoragePathEnv
	}

	if databaseDSNEnv := os.Getenv("DATABASE_DSN"); databaseDSNEnv != "" {
		flags.databaseDSN = databaseDSNEnv
	}
}
