package main

import (
	"flag"
	"os"
	"strconv"
)

// Config - структура с описанием конфигурации
type Config struct {
	runAddr         string
	baseURL         string
	fileStoragePath string
	databaseDSN     string
	enableHTTPS     bool
}

var flags = new(Config)

func parseFlags() {
	flag.StringVar(&flags.runAddr, "a", "localhost:8080", "HTTP listen address")
	flag.StringVar(&flags.baseURL, "b", "http://localhost:8080/", "Base URL")
	flag.StringVar(&flags.fileStoragePath, "f", "", "Short URL JSON storage")
	flag.StringVar(&flags.databaseDSN, "d", "", "Database DSN")
	flag.BoolVar(&flags.enableHTTPS, "s", false, "Enable HTTPS")

	flag.Parse()

	if serverAddrEnv := os.Getenv("SERVER_ADDR"); serverAddrEnv != "" {
		flags.runAddr = serverAddrEnv
	}

	if baseURLEnv := os.Getenv("BASE_URL"); baseURLEnv != "" {
		flags.baseURL = baseURLEnv
	}

	if fileStoragePathEnv := os.Getenv("FILE_STORAGE_PATH"); fileStoragePathEnv != "" {
		flags.fileStoragePath = fileStoragePathEnv
	}

	if databaseDSNEnv := os.Getenv("DATABASE_DSN"); databaseDSNEnv != "" {
		flags.databaseDSN = databaseDSNEnv
	}

	if enableHTTPSEnv := os.Getenv("ENABLE_HTTPS"); enableHTTPSEnv != "" {
		var err error
		flags.enableHTTPS, err = strconv.ParseBool(enableHTTPSEnv)
		if err != nil {
			log.Fatalf("Error parsing ENABLE_HTTPS: %v", err)
		}
	}
}
