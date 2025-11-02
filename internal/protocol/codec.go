package protocol

import (
	"encoding/binary"
	"fmt"
	"io"
)

// WriteMessage writes a message to the writer
// Message format: [1 byte: MessageType][4 bytes: PayloadLength][N bytes: Payload]
func WriteMessage(w io.Writer, msg *Message) error {
	// Write message type (1 byte)
	if err := binary.Write(w, binary.BigEndian, msg.Type); err != nil {
		return fmt.Errorf("failed to write message type: %w", err)
	}

	// Write payload length (4 bytes)
	payloadLen := uint32(len(msg.Payload))
	if err := binary.Write(w, binary.BigEndian, payloadLen); err != nil {
		return fmt.Errorf("failed to write payload length: %w", err)
	}

	// Write payload
	if payloadLen > 0 {
		if _, err := w.Write(msg.Payload); err != nil {
			return fmt.Errorf("failed to write payload: %w", err)
		}
	}

	return nil
}

// ReadMessage reads a message from the reader
func ReadMessage(r io.Reader) (*Message, error) {
	msg := &Message{}

	// Read message type (1 byte)
	if err := binary.Read(r, binary.BigEndian, &msg.Type); err != nil {
		return nil, fmt.Errorf("failed to read message type: %w", err)
	}

	// Read payload length (4 bytes)
	var payloadLen uint32
	if err := binary.Read(r, binary.BigEndian, &payloadLen); err != nil {
		return nil, fmt.Errorf("failed to read payload length: %w", err)
	}

	// Validate payload length (max 1MB to prevent memory attacks)
	if payloadLen > 1024*1024 {
		return nil, fmt.Errorf("payload too large: %d bytes", payloadLen)
	}

	// Read payload
	if payloadLen > 0 {
		msg.Payload = make([]byte, payloadLen)
		if _, err := io.ReadFull(r, msg.Payload); err != nil {
			return nil, fmt.Errorf("failed to read payload: %w", err)
		}
	}

	return msg, nil
}

// EncodeChallengeResponse encodes a ChallengeResponse into bytes
// Payload format: [1 byte: Difficulty][8 bytes: Timestamp][32 bytes: Random Data][N bytes: Client IP]
func EncodeChallengeResponse(cr *ChallengeResponse) ([]byte, error) {
	clientIPBytes := []byte(cr.ClientIP)
	buf := make([]byte, 41+len(clientIPBytes))

	// Encode difficulty (1 byte)
	buf[0] = byte(cr.Difficulty)

	// Encode timestamp (8 bytes)
	binary.BigEndian.PutUint64(buf[1:9], uint64(cr.Timestamp))

	// Encode random data (32 bytes)
	copy(buf[9:41], cr.RandomData[:])

	// Encode client IP (N bytes)
	copy(buf[41:], clientIPBytes)

	return buf, nil
}

// DecodeChallengeResponse decodes bytes into a ChallengeResponse
func DecodeChallengeResponse(payload []byte) (*ChallengeResponse, error) {
	if len(payload) < 41 {
		return nil, fmt.Errorf("invalid payload length: expected at least 41, got %d", len(payload))
	}

	cr := &ChallengeResponse{}

	// Decode difficulty (1 byte)
	cr.Difficulty = int(payload[0])

	// Decode timestamp (8 bytes)
	cr.Timestamp = int64(binary.BigEndian.Uint64(payload[1:9]))

	// Decode random data (32 bytes)
	copy(cr.RandomData[:], payload[9:41])

	// Decode client IP (N bytes)
	if len(payload) > 41 {
		cr.ClientIP = string(payload[41:])
	}

	return cr, nil
}

// EncodeSolution encodes a Solution into bytes
// Payload format: [8 bytes: Nonce]
func EncodeSolution(sol *Solution) ([]byte, error) {
	buf := make([]byte, 8)

	// Encode nonce (8 bytes) - uint64 BigEndian
	binary.BigEndian.PutUint64(buf, sol.Nonce)

	return buf, nil
}

// DecodeSolution decodes bytes into a Solution
func DecodeSolution(payload []byte) (*Solution, error) {
	if len(payload) != 8 {
		return nil, fmt.Errorf("invalid payload length: expected 8, got %d", len(payload))
	}

	// Decode nonce (8 bytes) - BigEndian uint64
	nonce := binary.BigEndian.Uint64(payload)

	return &Solution{Nonce: nonce}, nil
}

// EncodeQuote encodes a Quote into bytes
// Payload format: [N bytes: UTF-8 text]
func EncodeQuote(q *Quote) ([]byte, error) {
	return []byte(q.Text), nil
}

// DecodeQuote decodes bytes into a Quote
func DecodeQuote(payload []byte) (*Quote, error) {
	return &Quote{Text: string(payload)}, nil
}

// EncodeError encodes an Error into bytes
// Payload format: [2 bytes: Error Code][N bytes: Error Message]
func EncodeError(e *Error) ([]byte, error) {
	buf := make([]byte, 2+len(e.Message))

	// Encode error code (2 bytes) - uint16 BigEndian
	binary.BigEndian.PutUint16(buf[0:2], e.Code)

	// Encode error message (variable length) - UTF-8 string
	copy(buf[2:], []byte(e.Message))

	return buf, nil
}

// DecodeError decodes bytes into an Error
func DecodeError(payload []byte) (*Error, error) {
	if len(payload) < 2 {
		return nil, fmt.Errorf("invalid payload length: expected at least 2, got %d", len(payload))
	}

	e := &Error{}

	// Decode error code (2 bytes) - BigEndian uint16
	e.Code = binary.BigEndian.Uint16(payload[0:2])

	// Decode error message (remaining bytes) - UTF-8 string
	if len(payload) > 2 {
		e.Message = string(payload[2:])
	}

	return e, nil
}
