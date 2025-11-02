package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/kerimovkk/pow-server/internal/quotes"
	"github.com/kerimovkk/pow-server/internal/server"
	"gopkg.in/yaml.v3"
)

// Config represents the application configuration
type Config struct {
	Server struct {
		Host              string        `yaml:"host"`
		Port              int           `yaml:"port"`
		MaxConnections    int           `yaml:"max_connections"`
		ReadTimeout       time.Duration `yaml:"read_timeout"`
		WriteTimeout      time.Duration `yaml:"write_timeout"`
		ConnectionTimeout time.Duration `yaml:"connection_timeout"`
	} `yaml:"server"`
	PoW struct {
		BaseDifficulty    int           `yaml:"base_difficulty"`
		MaxDifficulty     int           `yaml:"max_difficulty"`
		ChallengeMaxAge   time.Duration `yaml:"challenge_max_age"`
		DynamicAdjustment bool          `yaml:"dynamic_adjustment"`
	} `yaml:"pow"`
	RateLimit struct {
		MaxRequests     int           `yaml:"max_requests"`
		Window          time.Duration `yaml:"window"`
		CleanupInterval time.Duration `yaml:"cleanup_interval"`
	} `yaml:"rate_limit"`
	Quotes struct {
		FilePath string `yaml:"file_path"`
	} `yaml:"quotes"`
}

func main() {
	log.Println("Starting Word of Wisdom server...")

	// Load configuration
	cfg, err := loadConfig("configs/config.yaml")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Load quotes
	quotesManager := quotes.NewManager()
	if err := quotesManager.LoadFromFile(cfg.Quotes.FilePath); err != nil {
		log.Fatalf("Failed to load quotes: %v", err)
	}
	log.Printf("Loaded %d quotes", quotesManager.Count())

	// Create rate limiter
	rateLimiter := server.NewRateLimiter(
		cfg.RateLimit.MaxRequests,
		cfg.RateLimit.Window,
		cfg.RateLimit.CleanupInterval,
	)
	defer rateLimiter.Stop()

	// Create server config
	serverConfig := &server.Config{
		Host:               cfg.Server.Host,
		Port:               cfg.Server.Port,
		MaxConnections:     cfg.Server.MaxConnections,
		ReadTimeout:        cfg.Server.ReadTimeout,
		WriteTimeout:       cfg.Server.WriteTimeout,
		ConnectionTimeout:  cfg.Server.ConnectionTimeout,
		PoWDifficulty:      cfg.PoW.BaseDifficulty,
		PoWChallengeMaxAge: cfg.PoW.ChallengeMaxAge,
	}

	// Create and start server
	srv := server.NewServer(serverConfig, quotesManager, rateLimiter)
	if err := srv.Start(); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}

	// Wait for interrupt signal for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	<-sigChan
	log.Println("\nShutting down server...")

	if err := srv.Stop(); err != nil {
		log.Printf("Error stopping server: %v", err)
	}

	log.Println("Server stopped")
}

// loadConfig loads configuration from YAML file
func loadConfig(path string) (*Config, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var cfg Config
	decoder := yaml.NewDecoder(file)
	if err := decoder.Decode(&cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}
