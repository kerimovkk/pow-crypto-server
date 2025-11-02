package protocol

// MessageType represents the type of message in the protocol
type MessageType byte

const (
	// MessageTypeChallengeRequest is sent by client to request a challenge
	MessageTypeChallengeRequest MessageType = 0x01

	// MessageTypeChallengeResponse is sent by server with the challenge
	MessageTypeChallengeResponse MessageType = 0x02

	// MessageTypeSolution is sent by client with the PoW solution
	MessageTypeSolution MessageType = 0x03

	// MessageTypeQuote is sent by server with the wisdom quote
	MessageTypeQuote MessageType = 0x04

	// MessageTypeError is sent by server when an error occurs
	MessageTypeError MessageType = 0x05
)

// Message represents a protocol message
// Format: [1 byte: MessageType][4 bytes: PayloadLength][N bytes: Payload]
type Message struct {
	Type    MessageType
	Payload []byte
}

// ChallengeRequest represents a request for a PoW challenge
type ChallengeRequest struct {
	// Empty payload for now
}

// ChallengeResponse represents the server's challenge
// Payload format: [1 byte: Difficulty][8 bytes: Timestamp][32 bytes: Random Data][N bytes: Client IP]
type ChallengeResponse struct {
	Difficulty int
	Timestamp  int64
	RandomData [32]byte
	ClientIP   string
}

// Solution represents the client's PoW solution
// Payload format: [8 bytes: Nonce]
type Solution struct {
	Nonce uint64
}

// Quote represents a wisdom quote from the server
// Payload format: [N bytes: UTF-8 text]
type Quote struct {
	Text string
}

// Error represents an error message
// Payload format: [2 bytes: Error Code][N bytes: Error Message]
type Error struct {
	Code    uint16
	Message string
}

// Error codes
const (
	ErrorCodeInvalidMessage   uint16 = 1
	ErrorCodeInvalidSolution  uint16 = 2
	ErrorCodeRateLimitExceeded uint16 = 3
	ErrorCodeTimeout          uint16 = 4
	ErrorCodeInternalError    uint16 = 5
)