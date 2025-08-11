package hooks

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
	"sync"
	"time"
)

const (
	// MaxCommandLength is the maximum command length before chunking
	MaxCommandLength = 32768
	// DefaultBatchSize for command processing
	DefaultBatchSize = 100
)

// HookManager manages hook execution with command chunking
type HookManager struct {
	memoryStore *MemoryStore // Singleton instance
	config      *HookConfig
	mu          sync.Mutex
	initialized bool
}

// HookConfig contains hook system configuration
type HookConfig struct {
	MaxCommandLength int
	BatchSize        int
	EnableCompression bool
	FallbackMode     string
	Timeout          time.Duration
}

// MemoryStore represents the SQLite memory store
type MemoryStore struct {
	path        string
	initialized bool
	mu          sync.Mutex
}

// NewHookManager creates a new hook manager with default config
func NewHookManager() *HookManager {
	return &HookManager{
		config: &HookConfig{
			MaxCommandLength:  MaxCommandLength,
			BatchSize:         DefaultBatchSize,
			EnableCompression: true,
			FallbackMode:      "direct_execution",
			Timeout:           30 * time.Second,
		},
	}
}

// Initialize ensures single memory store initialization
func (h *HookManager) Initialize() error {
	h.mu.Lock()
	defer h.mu.Unlock()
	
	if h.initialized {
		return nil
	}
	
	// Initialize memory store only once
	if h.memoryStore == nil {
		h.memoryStore = &MemoryStore{
			path:        ".swarm/memory.db",
			initialized: false,
		}
		
		if err := h.memoryStore.Initialize(); err != nil {
			return fmt.Errorf("failed to initialize memory store: %w", err)
		}
	}
	
	h.initialized = true
	return nil
}

// ExecuteWithChunking handles long commands by chunking
func (h *HookManager) ExecuteWithChunking(command string) error {
	// Ensure initialization
	if err := h.Initialize(); err != nil {
		return err
	}
	
	// Check command length
	if len(command) > h.config.MaxCommandLength {
		return h.executeChunked(command)
	}
	
	return h.executeDirect(command)
}

// executeChunked splits long commands into manageable chunks
func (h *HookManager) executeChunked(command string) error {
	// Split command into chunks
	chunks := h.splitCommand(command, h.config.BatchSize)
	
	for i, chunk := range chunks {
		if err := h.executeDirect(chunk); err != nil {
			// Fallback mode on chunk failure
			if h.config.FallbackMode == "direct_execution" {
				return h.executeFallback(chunk)
			}
			return fmt.Errorf("chunk %d failed: %w", i, err)
		}
	}
	
	return nil
}

// executeDirect executes a command directly
func (h *HookManager) executeDirect(command string) error {
	ctx, cancel := context.WithTimeout(context.Background(), h.config.Timeout)
	defer cancel()
	
	cmd := exec.CommandContext(ctx, "sh", "-c", command)
	
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("command failed: %w, output: %s", err, output)
	}
	
	return nil
}

// executeFallback provides fallback execution without hooks
func (h *HookManager) executeFallback(command string) error {
	// Execute without hooks - direct command execution
	cmd := exec.Command("sh", "-c", command)
	return cmd.Run()
}

// splitCommand splits a command into smaller chunks
func (h *HookManager) splitCommand(command string, batchSize int) []string {
	// Simple splitting by lines or arguments
	parts := strings.Split(command, "\n")
	
	var chunks []string
	currentChunk := ""
	
	for _, part := range parts {
		if len(currentChunk)+len(part)+1 > h.config.MaxCommandLength {
			if currentChunk != "" {
				chunks = append(chunks, currentChunk)
			}
			currentChunk = part
		} else {
			if currentChunk != "" {
				currentChunk += "\n"
			}
			currentChunk += part
		}
	}
	
	if currentChunk != "" {
		chunks = append(chunks, currentChunk)
	}
	
	return chunks
}

// Initialize initializes the memory store
func (m *MemoryStore) Initialize() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	if m.initialized {
		return nil
	}
	
	// Create SQLite connection once
	// Implementation details omitted for brevity
	m.initialized = true
	
	return nil
}