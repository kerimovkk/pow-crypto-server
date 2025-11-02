package quotes

import (
	"bufio"
	"fmt"
	"math/rand"
	"os"
	"strings"
	"sync"
)

// Manager manages a collection of quotes
type Manager struct {
	quotes []string
	mu     sync.RWMutex
}

// NewManager creates a new quote manager
func NewManager() *Manager {
	return &Manager{
		quotes: make([]string, 0),
	}
}

// LoadFromFile loads quotes from a file (one per line, # for comments)
func (m *Manager) LoadFromFile(path string) error {
	file, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("failed to open file %s: %w", path, err)
	}
	defer file.Close()

	m.mu.Lock()
	defer m.mu.Unlock()
	m.quotes = make([]string, 0)

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		m.quotes = append(m.quotes, line)
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("failed to read quotes file: %w", err)
	}

	if len(m.quotes) == 0 {
		return fmt.Errorf("no quotes found in file")
	}

	return nil
}

// Count returns the number of quotes
func (m *Manager) Count() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.quotes)
}

// GetRandom returns a random quote from the collection
func (m *Manager) GetRandom() (string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if len(m.quotes) == 0 {
		return "", fmt.Errorf("no quotes found in file")
	}

	idx := rand.Intn(len(m.quotes))

	return m.quotes[idx], nil
}
