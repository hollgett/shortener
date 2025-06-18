package config

import (
	"flag"
	"os"
	"sync"

	"github.com/hollgett/shortener.git/pkg/autoloaderenv"
)

type ShortenerConfig struct {
	once        sync.Once
	debug       bool
	Addr        string `env:"SERVER_ADDRESS"`
	BaseURL     string `env:"BASE_URL"`
	FilePath    string `env:"FILE_STORAGE_PATH"`
	DatabaseDSN string `env:"DATABASE_DSN"`
	SecretKey   string `env:"SECRET_KEY"`
}

// NewConfig return struct config with filled args.
//
// priority load cfg env -> flag cmd -> default value
func NewConfig() *ShortenerConfig {
	cfg := ShortenerConfig{}
	cfg.fillConfig()
	return &cfg
}

func (s *ShortenerConfig) fillConfig() {
	s.once.Do(func() {
		flag.StringVar(&s.Addr, "a", "localhost:8080", "set address for server, similar host:port")
		flag.StringVar(&s.BaseURL, "b", "http://localhost:8080", "set static address for server")
		flag.StringVar(&s.FilePath, "f", "tmp/short-url-db.json", "set filestorage mode")
		flag.StringVar(&s.DatabaseDSN, "d", "", "set database PostgreSQL mode")
		flag.BoolVar(&s.debug, "t", false, "setup debug mode")

		flag.Parse()

		if s.debug {
			autoloaderenv.LoadEnv()
		}

		addr, ok := os.LookupEnv("SERVER_ADDRESS")
		if ok {
			s.Addr = addr
		}
		baseURL, ok := os.LookupEnv("BASE_URL")
		if ok {
			s.BaseURL = baseURL
		}
		filePath, ok := os.LookupEnv("FILE_STORAGE_PATH")
		if ok {
			s.FilePath = filePath
		}
		databaseDSN, ok := os.LookupEnv("DATABASE_DSN")
		if ok {
			s.DatabaseDSN = databaseDSN
		}
		secretKey, ok := os.LookupEnv("SECRET_KEY")
		if ok {
			s.SecretKey = secretKey
		} else {
			s.SecretKey = "secret_key"
		}
	})

}
