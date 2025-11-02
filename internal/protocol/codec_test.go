package protocol

import (
	"testing"
)

func TestEncodeDecodeError(t *testing.T) {
	tests := []struct {
		name    string
		errCode uint16
		errMsg  string
	}{
		{"With message", ErrorCodeInvalidSolution, "Invalid solution provided"},
		{"Empty message", ErrorCodeTimeout, ""},
		{"Long message", ErrorCodeInternalError, "A very long error message with many details about what went wrong in the system"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create error
			original := &Error{
				Code:    tt.errCode,
				Message: tt.errMsg,
			}

			// Encode
			encoded, err := EncodeError(original)
			if err != nil {
				t.Fatalf("EncodeError failed: %v", err)
			}

			t.Logf("Encoded %d bytes: %v", len(encoded), encoded)

			// Decode
			decoded, err := DecodeError(encoded)
			if err != nil {
				t.Fatalf("DecodeError failed: %v", err)
			}

			// Verify
			if decoded.Code != original.Code {
				t.Errorf("Code mismatch: got %d, want %d", decoded.Code, original.Code)
			}
			if decoded.Message != original.Message {
				t.Errorf("Message mismatch: got %q, want %q", decoded.Message, original.Message)
			}
		})
	}
}

func TestDecodeSolution(t *testing.T) {
	tests := []struct {
		name    string
		nonce   uint64
		wantErr bool
	}{
		{"Small nonce", 42, false},
		{"Large nonce", 18446744073709551615, false},
		{"Zero", 0, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			original := &Solution{Nonce: tt.nonce}

			encoded, err := EncodeSolution(original)
			if err != nil {
				t.Fatalf("EncodeSolution failed: %v", err)
			}

			decoded, err := DecodeSolution(encoded)
			if (err != nil) != tt.wantErr {
				t.Errorf("DecodeSolution error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && decoded.Nonce != original.Nonce {
				t.Errorf("Nonce mismatch: got %d, want %d", decoded.Nonce, original.Nonce)
			}
		})
	}
}

func TestEncodeDecode_ChallengeResponse(t *testing.T) {
	tests := []struct {
		name       string
		difficulty int
		timestamp  int64
		randomData [32]byte
		clientIP   string
	}{
		{"Standard challenge", 20, 1234567890, [32]byte{1, 2, 3, 4, 5}, "192.168.1.1"},
		{"Zero difficulty", 0, 0, [32]byte{}, "127.0.0.1"},
		{"Max difficulty", 255, 9999999999, [32]byte{255, 255, 255, 255}, "10.0.0.1"},
		{"Empty IP", 20, 1234567890, [32]byte{1, 2, 3}, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			original := &ChallengeResponse{
				Difficulty: tt.difficulty,
				Timestamp:  tt.timestamp,
				RandomData: tt.randomData,
				ClientIP:   tt.clientIP,
			}

			// Encode
			encoded, err := EncodeChallengeResponse(original)
			if err != nil {
				t.Fatalf("EncodeChallengeResponse failed: %v", err)
			}

			t.Logf("Encoded %d bytes", len(encoded))

			// Decode
			decoded, err := DecodeChallengeResponse(encoded)
			if err != nil {
				t.Fatalf("DecodeChallengeResponse failed: %v", err)
			}

			// Verify
			if decoded.Difficulty != original.Difficulty {
				t.Errorf("Difficulty mismatch: got %d, want %d", decoded.Difficulty, original.Difficulty)
			}
			if decoded.Timestamp != original.Timestamp {
				t.Errorf("Timestamp mismatch: got %d, want %d", decoded.Timestamp, original.Timestamp)
			}
			if decoded.RandomData != original.RandomData {
				t.Errorf("RandomData mismatch")
			}
			if decoded.ClientIP != original.ClientIP {
				t.Errorf("ClientIP mismatch: got %q, want %q", decoded.ClientIP, original.ClientIP)
			}
		})
	}
}
