package pow

import (
	"testing"
)

func TestVerify(t *testing.T) {
	challenge, err := GenerateChallenge("192.168.1.1", 4)
	if err != nil {
		t.Fatalf("Failed to generate challenge: %v", err)
	}

	tests := []struct {
		name     string
		nonce    uint64
		clientIP string
		want     bool
	}{
		{
			name:     "Wrong IP",
			nonce:    12345,
			clientIP: "192.168.1.2", // Different IP
			want:     false,
		},
		{
			name:     "Correct IP but wrong nonce (0 zero bits)",
			nonce:    0,
			clientIP: "192.168.1.1",
			want:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Verify(challenge, tt.nonce, tt.clientIP)
			if got != tt.want {
				t.Errorf("Verify() = %v, want %v", got, tt.want)
			}
		})
	}

	// Special test: Find a valid nonce and verify it
	t.Run("Valid solution", func(t *testing.T) {
		// Create a simple challenge with difficulty 4
		simpleChallenge, _ := GenerateChallenge("127.0.0.1", 4)

		// Try to find a valid nonce (brute force, may take a moment)
		var validNonce uint64
		found := false
		for nonce := uint64(0); nonce < 1000000; nonce++ {
			hash := simpleChallenge.ComputeHash(nonce)
			zeroBits := CountLeadingZeroBits(hash)
			if zeroBits >= 4 {
				validNonce = nonce
				found = true
				t.Logf("Found valid nonce: %d with %d zero bits", nonce, zeroBits)
				break
			}
		}

		if !found {
			t.Skip("Could not find valid nonce in 1M iterations")
		}

		// This should return true
		if !Verify(simpleChallenge, validNonce, "127.0.0.1") {
			hash := simpleChallenge.ComputeHash(validNonce)
			zeroBits := CountLeadingZeroBits(hash)
			t.Errorf("Verify() returned false for valid nonce. Nonce has %d zero bits, difficulty is %d",
				zeroBits, simpleChallenge.Difficulty)
		}
	})
}

func TestCountLeadingZeroBits(t *testing.T) {
	tests := []struct {
		name string
		hash [32]byte
		want int
	}{
		{
			name: "All zeros",
			hash: [32]byte{},
			want: 256,
		},
		{
			name: "First byte is 0x80 (10000000)",
			hash: [32]byte{0x80, 0, 0, 0},
			want: 0,
		},
		{
			name: "First byte is 0x00, second is 0x80",
			hash: [32]byte{0x00, 0x80, 0, 0},
			want: 8,
		},
		{
			name: "First byte is 0x01 (00000001)",
			hash: [32]byte{0x01, 0, 0, 0},
			want: 7,
		},
		{
			name: "First byte is 0x0F (00001111)",
			hash: [32]byte{0x0F, 0, 0, 0},
			want: 4,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CountLeadingZeroBits(tt.hash)
			if got != tt.want {
				t.Errorf("CountLeadingZeroBits() = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestSolve(t *testing.T) {
	tests := []struct {
		name       string
		difficulty int
		maxTime    int // seconds
	}{
		{"Easy - 4 bits", 4, 5},
		{"Medium - 8 bits", 8, 10},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			challenge, err := GenerateChallenge("127.0.0.1", tt.difficulty)
			if err != nil {
				t.Fatalf("Failed to generate challenge: %v", err)
			}

			t.Logf("Solving challenge with difficulty %d...", tt.difficulty)
			nonce, err := Solve(challenge)
			if err != nil {
				t.Fatalf("Solve() failed: %v", err)
			}

			t.Logf("Found nonce: %d", nonce)

			// Verify the solution
			if !Verify(challenge, nonce, "127.0.0.1") {
				hash := challenge.ComputeHash(nonce)
				zeroBits := CountLeadingZeroBits(hash)
				t.Errorf("Verify() failed for nonce %d. Hash has %d zero bits, need %d",
					nonce, zeroBits, tt.difficulty)
			}
		})
	}
}
