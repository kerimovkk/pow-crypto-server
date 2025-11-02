package client

import (
	"fmt"
	"log"
	"net"
	"time"

	"github.com/kerimovkk/pow-server/internal/pow"
	"github.com/kerimovkk/pow-server/internal/protocol"
)

// Client represents a TCP client
type Client struct {
	serverAddr string
	timeout    time.Duration
}

// NewClient creates a new client
func NewClient(serverAddr string, timeout time.Duration) *Client {
	return &Client{
		serverAddr: serverAddr,
		timeout:    timeout,
	}
}

// GetQuote connects to the server and retrieves a quote
func (c *Client) GetQuote() (string, error) {
	// Connect to server
	conn, err := net.DialTimeout("tcp", c.serverAddr, c.timeout)
	if err != nil {
		return "", fmt.Errorf("failed to connect to server: %w", err)
	}
	defer conn.Close()

	// Set overall deadline
	conn.SetDeadline(time.Now().Add(c.timeout))

	// Send challenge request
	log.Println("Sending challenge request...")
	msg := &protocol.Message{
		Type:    protocol.MessageTypeChallengeRequest,
		Payload: []byte{}, // Empty payload
	}

	if err := protocol.WriteMessage(conn, msg); err != nil {
		return "", fmt.Errorf("failed to send challenge request: %w", err)
	}

	// Receive challenge response
	log.Println("Waiting for challenge...")
	msg, err = protocol.ReadMessage(conn)
	if err != nil {
		return "", fmt.Errorf("failed to read challenge response: %w", err)
	}

	if msg.Type == protocol.MessageTypeError {
		errMsg, _ := protocol.DecodeError(msg.Payload)
		return "", fmt.Errorf("server error: %s", errMsg.Message)
	}

	if msg.Type != protocol.MessageTypeChallengeResponse {
		return "", fmt.Errorf("unexpected message type: %d", msg.Type)
	}

	challengeResp, err := protocol.DecodeChallengeResponse(msg.Payload)
	if err != nil {
		return "", fmt.Errorf("failed to decode challenge: %w", err)
	}

	log.Printf("Received challenge with difficulty: %d (IP: %s)", challengeResp.Difficulty, challengeResp.ClientIP)

	// Solve the PoW challenge
	log.Println("Solving PoW challenge...")
	challenge := &pow.Challenge{
		Difficulty: challengeResp.Difficulty,
		Timestamp:  challengeResp.Timestamp,
		Data:       challengeResp.RandomData,
		ClientIP:   challengeResp.ClientIP,
	}

	startTime := time.Now()
	nonce, err := pow.Solve(challenge)
	if err != nil {
		return "", fmt.Errorf("failed to solve challenge: %w", err)
	}
	solveTime := time.Since(startTime)
	log.Printf("Challenge solved in %v (nonce: %d)", solveTime, nonce)

	// Send solution
	log.Println("Sending solution...")
	solution := &protocol.Solution{
		Nonce: nonce,
	}

	payload, err := protocol.EncodeSolution(solution)
	if err != nil {
		return "", fmt.Errorf("failed to encode solution: %w", err)
	}

	msg = &protocol.Message{
		Type:    protocol.MessageTypeSolution,
		Payload: payload,
	}

	if err := protocol.WriteMessage(conn, msg); err != nil {
		return "", fmt.Errorf("failed to send solution: %w", err)
	}

	// Receive quote or error
	log.Println("Waiting for quote...")
	msg, err = protocol.ReadMessage(conn)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	if msg.Type == protocol.MessageTypeError {
		errMsg, _ := protocol.DecodeError(msg.Payload)
		return "", fmt.Errorf("server error: %s", errMsg.Message)
	}

	if msg.Type != protocol.MessageTypeQuote {
		return "", fmt.Errorf("unexpected message type: %d", msg.Type)
	}

	quote, err := protocol.DecodeQuote(msg.Payload)
	if err != nil {
		return "", fmt.Errorf("failed to decode quote: %w", err)
	}

	return quote.Text, nil
}