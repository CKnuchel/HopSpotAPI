package main

import "hopSpotAPI/internal/config"

func main() {
	cfg := config.Load()
	println("Server starting on port:", cfg.Port)
}
