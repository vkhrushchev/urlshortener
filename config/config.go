package config

import (
	"encoding/json"
	"errors"
	"flag"
	"go.uber.org/zap"
	"io"
	"os"
	"strconv"
)

var log = zap.Must(zap.NewDevelopment()).Sugar()

const (
	runAddrDefault  = "localhost:8080"
	baseURLDefault  = "http://localhost:8080"
	grpcAddrDefault = "localhost:18080"
	saltDefault     = "ACKaRDistERI"
)

// Config - структура с описанием конфигурации
type Config struct {
	RunAddr         string `json:"server_address"`
	BaseURL         string `json:"base_url"`
	FileStoragePath string `json:"file_storage_path"`
	DatabaseDSN     string `json:"database_dsn"`
	TrustedSubnet   string `json:"trusted_subnet"`
	EnableHTTPS     bool   `json:"enable_https"`
	GRPCAddr        string `json:"grpc_address"`
	Salt            string `json:"salt"`
}

// ReadConfig - считывает конфигурацию из переменных окружения, параметров командной строки и конфигурационного файла
func ReadConfig() Config {
	var configFilePath string
	var flagConfig Config
	parseFlags(&flagConfig, &configFilePath)

	if configFilePath == "" {
		configFilePath = os.Getenv("CONFIG")
	}

	log.Debugw("config: ", "flagConfig", flagConfig)

	var config Config
	if configFilePath != "" {
		parseJSONConfig(&config, configFilePath)
	}

	log.Debugw("config: ", "config", config)

	overrideConfigByFlags(&config, &flagConfig)
	log.Debugw("overrideConfigByFlags: ", "config", config)
	overrideConfigByEnv(&config)
	log.Debugw("overrideConfigByEnv: ", "config", config)

	return config
}

func parseFlags(config *Config, configFilePath *string) {
	flag.StringVar(&config.RunAddr, "a", runAddrDefault, "HTTP listen address")
	flag.StringVar(&config.BaseURL, "b", baseURLDefault, "Base URL")
	flag.StringVar(&config.FileStoragePath, "f", "", "Short URL JSON storage")
	flag.StringVar(&config.DatabaseDSN, "d", "", "Database DSN")
	flag.BoolVar(&config.EnableHTTPS, "s", false, "Enable HTTPS")
	flag.StringVar(&config.TrustedSubnet, "t", "", "Trusted Subnet")
	flag.StringVar(configFilePath, "c", "", "Configuration file")
	flag.StringVar(configFilePath, "config", "", "Configuration file")
	flag.StringVar(&config.GRPCAddr, "grpc-addr", grpcAddrDefault, "gRPC listen address")
	flag.StringVar(&config.Salt, "salt", saltDefault, "Salt used for authentication")

	flag.Parse()
}

func parseJSONConfig(config *Config, configFilePath string) {
	f, err := os.Open(configFilePath)
	if err != nil {
		log.Fatalf("config: error opening config file: %v", err)
	}

	defer func(f *os.File) {
		if err := f.Close(); err != nil {
			log.Fatalf("config: error closing config file: %v", err)
		}
	}(f)

	if err := json.NewDecoder(f).Decode(&config); err != nil && !errors.Is(err, io.EOF) {
		log.Fatalf("config: error parsing config file: %v", err)
	}
}

func overrideConfigByFlags(config *Config, flagConfig *Config) {
	if config.RunAddr == "" {
		config.RunAddr = flagConfig.RunAddr
	}

	if config.BaseURL == "" {
		config.BaseURL = flagConfig.BaseURL
	}

	if config.FileStoragePath == "" {
		config.FileStoragePath = flagConfig.FileStoragePath
	}

	if config.DatabaseDSN == "" {
		config.DatabaseDSN = flagConfig.DatabaseDSN
	}

	if config.TrustedSubnet == "" {
		config.TrustedSubnet = flagConfig.TrustedSubnet
	}

	if !config.EnableHTTPS {
		config.EnableHTTPS = flagConfig.EnableHTTPS
	}

	if config.GRPCAddr == "" {
		config.GRPCAddr = flagConfig.GRPCAddr
	}

	if config.Salt == "" {
		config.Salt = flagConfig.Salt
	}
}

func overrideConfigByEnv(config *Config) {
	if serverAddrEnv, ok := os.LookupEnv("SERVER_ADDR"); ok && serverAddrEnv != "" {
		config.RunAddr = serverAddrEnv
	}

	if baseURLEnv, ok := os.LookupEnv("BASE_URL"); ok && baseURLEnv != "" {
		config.BaseURL = baseURLEnv
	}

	if fileStoragePathEnv, ok := os.LookupEnv("FILE_STORAGE_PATH"); ok && fileStoragePathEnv != "" {
		config.FileStoragePath = fileStoragePathEnv
	}

	if databaseDSNEnv, ok := os.LookupEnv("DATABASE_DSN"); ok && databaseDSNEnv != "" {
		config.DatabaseDSN = databaseDSNEnv
	}

	if enableHTTPSEnv, ok := os.LookupEnv("ENABLE_HTTPS"); ok && enableHTTPSEnv != "" {
		var err error
		config.EnableHTTPS, err = strconv.ParseBool(enableHTTPSEnv)
		if err != nil {
			log.Fatalf("config: error parsing ENABLE_HTTPS env variable: %v", err)
		}
	}

	if trustedSubnetEnv, ok := os.LookupEnv("TRUSTED_SUBNET"); ok && trustedSubnetEnv != "" {
		config.TrustedSubnet = trustedSubnetEnv
	}

	if grpcAddrEnv, ok := os.LookupEnv("GRPC_ADDR"); ok && grpcAddrEnv != "" {
		config.GRPCAddr = grpcAddrEnv
	}

	if saltEnv, ok := os.LookupEnv("SHORTENER_SALT"); ok && saltEnv != "" {
		config.Salt = saltEnv
	}
}
