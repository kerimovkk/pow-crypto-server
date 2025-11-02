package pow

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"time"
)

// Challenge represents a Proof of Work challenge
type Challenge struct {
	Data       [32]byte // Random data
	Timestamp  int64    // Unix timestamp
	ClientIP   string   // Client IP address
	Difficulty int      // Number of leading zero bits required
}

// GenerateChallenge creates a new PoW challenge for a client
func GenerateChallenge(clientIP string, difficulty int) (*Challenge, error) {
	var randomData [32]byte
	if _, err := rand.Read(randomData[:]); err != nil {
		return nil, fmt.Errorf("failed to generate random data: %w", err)
	}

	return &Challenge{
		Data:       randomData,
		Timestamp:  time.Now().Unix(),
		ClientIP:   clientIP,
		Difficulty: difficulty,
	}, nil
}

// IsValid checks if the challenge is still valid (not expired)
func (c *Challenge) IsValid(maxAge time.Duration) bool {
	age := time.Since(time.Unix(c.Timestamp, 0))
	return age <= maxAge
}

// String returns a string representation of the challenge for hashing
// Format: base64(random_data):timestamp:client_ip
func (c *Challenge) String() string {
	encoded := base64.StdEncoding.EncodeToString(c.Data[:])
	return fmt.Sprintf("%s:%d:%s", encoded, c.Timestamp, c.ClientIP)
}

// ComputeHash computes SHA-256 hash of challenge + nonce
// Format: SHA256(challenge_string + ":" + nonce)
func (c *Challenge) ComputeHash(nonce uint64) [32]byte {
	data := fmt.Sprintf("%s:%d", c.String(), nonce)
	return sha256.Sum256([]byte(data))
}

// CountLeadingZeroBits counts the number of leading zero bits in a byte array
func CountLeadingZeroBits(hash [32]byte) int {
	count := 0
	for _, b := range hash {
		if b == 0 {
			count += 8
			continue
		}
		// Count leading zeros in this byte
		for i := 7; i >= 0; i-- {
			if (b & (1 << i)) == 0 {
				count++
			} else {
				return count
			}
		}
	}
	return count
}

// Verify checks if a nonce is a valid solution for this challenge
func Verify(c *Challenge, nonce uint64, clientIP string) bool {
	if c.ClientIP != clientIP {
		return false
	}

	// Compute nonce hash
	hash := c.ComputeHash(nonce)

	// Check if difficulty is valid
	d := CountLeadingZeroBits(hash)
	return d >= c.Difficulty
}

// Solve finds a nonce that satisfies the challenge difficulty
func Solve(c *Challenge) (uint64, error) {
	for nonce := range ^uint64(0) {
		hash := c.ComputeHash(nonce)

		if CountLeadingZeroBits(hash) >= c.Difficulty {
			return nonce, nil
		}
	}

	return 0, fmt.Errorf("no solution found")
}
