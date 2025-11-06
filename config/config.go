package config

type Config struct {
	DatabaseURL string
	NumWorkers  int
	DownloadDir string
}

func NewConfig() *Config {
	return &Config{
		DatabaseURL: "postgres://postgres:postgres@localhost:5432/go_concurrency?sslmode=disable",
		NumWorkers:  10,
		DownloadDir: "downloads",
	}
}
