package server

import (
	"fmt"
	"log"
	"net"
	"time"

	"github.com/kerimovkk/pow-server/internal/pow"
	"github.com/kerimovkk/pow-server/internal/protocol"
)

// Error codes
const (
	ErrorCodeInvalidMessage    = 1
	ErrorCodeInvalidSolution   = 2
	ErrorCodeRateLimitExceeded = 3
	ErrorCodeTimeout           = 4
	ErrorCodeInternalError     = 5
)

// sendError sends an error message to the client
func (s *Server) sendError(conn net.Conn, code uint16, message string) {
	conn.SetWriteDeadline(time.Now().Add(s.config.WriteTimeout))

	errMsg := &protocol.Error{
		Code:    code,
		Message: message,
	}

	payload, err := protocol.EncodeError(errMsg)
	if err != nil {
		log.Printf("Failed to encode error: %v", err)
		return
	}

	msg := &protocol.Message{
		Type:    protocol.MessageTypeError,
		Payload: payload,
	}

	if err := protocol.WriteMessage(conn, msg); err != nil {
		log.Printf("Failed to send error: %v", err)
	}
}

// handleChallengeResponse implements the PoW challenge-response protocol
func (s *Server) handleChallengeResponse(conn net.Conn, challenge *pow.Challenge) error {
	clientIP := conn.RemoteAddr().(*net.TCPAddr).IP.String()

	// Read challenge request from client
	conn.SetReadDeadline(time.Now().Add(s.config.ReadTimeout))
	msg, err := protocol.ReadMessage(conn)
	if err != nil {
		return fmt.Errorf("failed to read challenge request: %w", err)
	}

	if msg.Type != protocol.MessageTypeChallengeRequest {
		s.sendError(conn, ErrorCodeInvalidMessage, "Expected challenge request")
		return fmt.Errorf("unexpected message type: %d", msg.Type)
	}

	log.Printf("Received challenge request from %s", clientIP)

	// Send challenge to client
	challengeResp := &protocol.ChallengeResponse{
		Difficulty: challenge.Difficulty,
		Timestamp:  challenge.Timestamp,
		RandomData: challenge.Data,
		ClientIP:   clientIP,
	}
	buf, err := protocol.EncodeChallengeResponse(challengeResp)
	if err != nil {
		return fmt.Errorf("failed to encode challenge response: %w", err)
	}

	msg = &protocol.Message{
		Type:    protocol.MessageTypeChallengeResponse,
		Payload: buf,
	}

	err = protocol.WriteMessage(conn, msg)
	if err != nil {
		return fmt.Errorf("failed to send challenge response to the client: %w", err)
	}

	// Read solution from client
	msg, err = protocol.ReadMessage(conn)
	if err != nil {
		return fmt.Errorf("failed to read solution: %w", err)
	}
	if msg.Type != protocol.MessageTypeSolution {
		s.sendError(conn, ErrorCodeInvalidMessage, "Expected solution")
		return fmt.Errorf("unexpected message type")
	}

	solution, err := protocol.DecodeSolution(msg.Payload)
	if err != nil {
		return fmt.Errorf("failed to decode solution: %w", err)
	}

	// Verify solution
	if !pow.Verify(challenge, solution.Nonce, clientIP) {
		s.sendError(conn, ErrorCodeInvalidSolution, "Invalid solution")
		return fmt.Errorf("invalid PoW solution")
	}

	// Send quote or error
	quote, err := s.quotes.GetRandom()
	if err != nil {
		s.sendError(conn, ErrorCodeInternalError, "No quotes available")
		return fmt.Errorf("failed to get quote: %w", err)
	}

	quoteMsg := &protocol.Quote{Text: quote}
	payload, err := protocol.EncodeQuote(quoteMsg)
	if err != nil {
		return fmt.Errorf("failed to encode quote: %w", err)
	}

	msg = &protocol.Message{
		Type:    protocol.MessageTypeQuote,
		Payload: payload,
	}
	conn.SetWriteDeadline(time.Now().Add(s.config.WriteTimeout))

	return protocol.WriteMessage(conn, msg)
}
