package main

import (
	"log"
	"os"

	"lot-control/internal"
	"lot-control/internal/config"
)

func main() {
	cfgPath := getEnv("CONFIG_PATH", "config.yaml")
	cfg, err := config.Load(cfgPath)
	if err != nil {
		log.Fatalf("không đọc được cấu hình: %v", err)
	}

	server, err := internal.New(cfg)
	if err != nil {
		log.Fatalf("không khởi tạo được server: %v", err)
	}

	if err := server.Run(); err != nil {
		log.Fatalf("server dừng: %v", err)
	}
}

func getEnv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
