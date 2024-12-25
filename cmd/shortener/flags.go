package main

import (
	"encoding/json"
	"errors"
	"flag"
	"io"
	"os"
	"strconv"
)

const (
	runAddrDefault = "localhost:8080"
	baseURLDefault = "http://localhost:8080"
)

var (
	runAddr         string
	baseURL         string
	fileStoragePath string
	databaseDSN     string
	enableHTTPS     bool
	configFile      string
)

// Config - структура с описанием конфигурации
type Config struct {
	RunAddr         string `json:"server_address"`
	BaseURL         string `json:"base_url"`
	FileStoragePath string `json:"file_storage_path"`
	DatabaseDSN     string `json:"database_dsn"`
	EnableHTTPS     bool   `json:"enable_https"`
}

func readConfig() Config {
	parseFlags()
	if configFile == "" {
		configFile = os.Getenv("CONFIG")
	}

	var jsonConfig Config
	if configFile != "" {
		parseJSONConfig(&jsonConfig)
	}

	log.Debugw("config: ", "config", jsonConfig)

	overrideConfigByFlags(&jsonConfig)
	log.Debugw("overrideConfigByFlags: ", "config", jsonConfig)
	overrideConfigByEnv(&jsonConfig)
	log.Debugw("overrideConfigByEnv: ", "config", jsonConfig)

	return jsonConfig
}

func parseFlags() {
	flag.StringVar(&runAddr, "a", runAddrDefault, "HTTP listen address")
	flag.StringVar(&baseURL, "b", baseURLDefault, "Base URL")
	flag.StringVar(&fileStoragePath, "f", "", "Short URL JSON storage")
	flag.StringVar(&databaseDSN, "d", "", "Database DSN")
	flag.BoolVar(&enableHTTPS, "s", false, "Enable HTTPS")
	flag.StringVar(&configFile, "c", "", "Configuration file")
	flag.StringVar(&configFile, "config", "", "Configuration file")

	flag.Parse()
}

func parseJSONConfig(config *Config) {
	f, err := os.Open(configFile)
	if err != nil {
		log.Fatalf("Error opening config file: %v", err)
	}

	defer func(f *os.File) {
		if err := f.Close(); err != nil {
			log.Fatalf("Error closing config file: %v", err)
		}
	}(f)

	if err := json.NewDecoder(f).Decode(&config); err != nil && !errors.Is(err, io.EOF) {
		log.Fatalf("Error parsing config file: %v", err)
	}
}

func overrideConfigByFlags(config *Config) {
	if runAddr != "" && runAddr != runAddrDefault {
		config.RunAddr = runAddr
	} else if config.RunAddr == "" {
		config.RunAddr = runAddrDefault
	}

	if baseURL != "" && baseURL != baseURLDefault {
		config.BaseURL = baseURL
	} else if config.BaseURL == "" {
		config.BaseURL = baseURLDefault
	}

	if fileStoragePath != "" {
		config.FileStoragePath = fileStoragePath
	}

	if databaseDSN != "" {
		config.DatabaseDSN = databaseDSN
	}
}

func overrideConfigByEnv(config *Config) {
	if serverAddrEnv, ok := os.LookupEnv("SERVER_ADDR"); ok {
		config.RunAddr = serverAddrEnv
	}

	if baseURLEnv, ok := os.LookupEnv("BASE_URL"); ok {
		config.BaseURL = baseURLEnv
	}

	if fileStoragePathEnv, ok := os.LookupEnv("FILE_STORAGE_PATH"); ok {
		config.FileStoragePath = fileStoragePathEnv
	}

	if databaseDSNEnv, ok := os.LookupEnv("DATABASE_DSN"); ok {
		config.DatabaseDSN = databaseDSNEnv
	}

	if enableHTTPSEnv, ok := os.LookupEnv("ENABLE_HTTPS"); ok {
		var err error
		config.EnableHTTPS, err = strconv.ParseBool(enableHTTPSEnv)
		if err != nil {
			log.Fatalf("Error parsing ENABLE_HTTPS env variable: %v", err)
		}
	}
}
