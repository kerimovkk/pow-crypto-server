package server

import (
	"fmt"
	"log"
	"net"
	"sync/atomic"
	"time"

	"github.com/kerimovkk/pow-server/internal/pow"
	"github.com/kerimovkk/pow-server/internal/quotes"
)

// Server represents the TCP server
type Server struct {
	listener     net.Listener
	config       *Config
	quotes       *quotes.Manager
	rateLimiter  *RateLimiter
	activeConns  atomic.Int32
	shutdownChan chan struct{}
}

// Config holds server configuration
type Config struct {
	Host               string
	Port               int
	MaxConnections     int
	ReadTimeout        time.Duration
	WriteTimeout       time.Duration
	ConnectionTimeout  time.Duration
	PoWDifficulty      int
	PoWChallengeMaxAge time.Duration
}

// NewServer creates a new TCP server
func NewServer(config *Config, quotesManager *quotes.Manager, rateLimiter *RateLimiter) *Server {
	return &Server{
		config:       config,
		quotes:       quotesManager,
		rateLimiter:  rateLimiter,
		shutdownChan: make(chan struct{}),
	}
}

// Start starts the TCP server
func (s *Server) Start() error {
	addr := fmt.Sprintf("%s:%d", s.config.Host, s.config.Port)

	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("failed to start listener: %w", err)
	}

	s.listener = listener
	log.Printf("Server listening on %s", addr)

	// Accept connections
	go s.acceptLoop()

	return nil
}

// acceptLoop accepts incoming connections
func (s *Server) acceptLoop() {
	for {
		select {
		case <-s.shutdownChan:
			return
		default:
		}

		conn, err := s.listener.Accept()
		if err != nil {
			select {
			case <-s.shutdownChan:
				return
			default:
				log.Printf("Accept error: %v", err)
				continue
			}
		}

		// Check max connections
		current := s.activeConns.Load()
		if current >= int32(s.config.MaxConnections) {
			log.Printf("Max connections reached, rejecting %s", conn.RemoteAddr())
			conn.Close()
			continue
		}

		s.activeConns.Add(1)
		go s.handleConnection(conn)
	}
}

// handleConnection handles a single client connection
func (s *Server) handleConnection(conn net.Conn) {
	defer func() {
		conn.Close()
		s.activeConns.Add(-1)
	}()

	// Set overall connection timeout
	deadline := time.Now().Add(s.config.ConnectionTimeout)
	conn.SetDeadline(deadline)

	clientIP := conn.RemoteAddr().(*net.TCPAddr).IP.String()
	log.Printf("New connection from %s", clientIP)

	// Check rate limit
	if !s.rateLimiter.Allow(clientIP) {
		log.Printf("Rate limit exceeded for %s", clientIP)
		s.sendError(conn, ErrorCodeRateLimitExceeded, "Rate limit exceeded")
		return
	}

	// Generate PoW challenge
	challenge, err := pow.GenerateChallenge(clientIP, s.config.PoWDifficulty)
	if err != nil {
		log.Printf("Failed to generate challenge: %v", err)
		s.sendError(conn, ErrorCodeInternalError, "Internal error")
		return
	}

	// Handle challenge-response protocol
	if err := s.handleChallengeResponse(conn, challenge); err != nil {
		log.Printf("Challenge-response failed for %s: %v", clientIP, err)
		return
	}

	log.Printf("Connection from %s completed successfully", clientIP)
}

// Stop gracefully stops the server
func (s *Server) Stop() error {
	close(s.shutdownChan)

	if s.listener != nil {
		return s.listener.Close()
	}

	return nil
}
