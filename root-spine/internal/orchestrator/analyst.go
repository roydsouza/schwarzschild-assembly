package orchestrator

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"go.uber.org/zap"
)

// VerdictState represents the outcome of an automated analysis.
type VerdictState string

const (
	VerdictApproved VerdictState = "APPROVED"
	VerdictVetoed   VerdictState = "VETOED"
	VerdictPending  VerdictState = "PENDING"
)

// AnalystVerdict represents a parsed verdict from Claude Code.
type AnalystVerdict struct {
	ID        string
	State     VerdictState
	Rationale string
	FilePath  string
	ModTime   time.Time
}

// VerdictManager monitors the analyst-verdicts/ directory.
type VerdictManager struct {
	logger      *zap.Logger
	verdictsDir string
	verdicts    map[string]AnalystVerdict
	mu          sync.RWMutex
}

// NewVerdictManager creates a new verdict manager.
func NewVerdictManager(logger *zap.Logger, dir string) *VerdictManager {
	return &VerdictManager{
		logger:      logger,
		verdictsDir: dir,
		verdicts:    make(map[string]AnalystVerdict),
	}
}

// Start initiates the polling loop.
func (m *VerdictManager) Start(interval time.Duration) {
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for range ticker.C {
			m.poll()
		}
	}()
}

// GetVerdict retrieves the latest verdict for a given artifact/proposal.
func (m *VerdictManager) GetVerdict(artifactID string) (AnalystVerdict, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	v, ok := m.verdicts[artifactID]
	return v, ok
}

func (m *VerdictManager) poll() {
	files, err := filepath.Glob(filepath.Join(m.verdictsDir, "*.md"))
	if err != nil {
		m.logger.Error("failed to glob verdicts", zap.Error(err))
		return
	}

	for _, file := range files {
		if filepath.Base(file) == "README.md" {
			continue
		}

		info, err := os.Stat(file)
		if err != nil {
			continue
		}

		// Check if we already processed this file and it hasn't changed
		// (Actually, multiple files might target the same artifactID. We want the latest.)
		
		content, err := os.ReadFile(file)
		if err != nil {
			continue
		}

		v, err := m.parseVerdict(string(content), file, info.ModTime())
		if err != nil {
			m.logger.Warn("failed to parse verdict", zap.String("file", file), zap.Error(err))
			continue
		}

		m.mu.Lock()
		// Only update if it's newer than what we have
		if existing, ok := m.verdicts[v.ID]; !ok || v.ModTime.After(existing.ModTime) {
			m.verdicts[v.ID] = v
			m.logger.Info("ingested analyst verdict", zap.String("id", v.ID), zap.String("state", string(v.State)))
		}
		m.mu.Unlock()
	}
}

// parseVerdict extracts ID and state from the markdown verdict.
func (m *VerdictManager) parseVerdict(content, path string, modTime time.Time) (AnalystVerdict, error) {
	lines := strings.Split(content, "\n")
	v := AnalystVerdict{
		Rationale: content,
		FilePath:  path,
		ModTime:   modTime,
		State:     VerdictPending,
	}

	for _, line := range lines {
		if strings.HasPrefix(line, "**Artifact:**") {
			v.ID = strings.TrimSpace(strings.TrimPrefix(line, "**Artifact:**"))
		}
		if strings.HasPrefix(line, "**Verdict:**") {
			stateStr := strings.TrimSpace(strings.TrimPrefix(line, "**Verdict:**"))
			if strings.Contains(stateStr, "APPROVED") {
				v.State = VerdictApproved
			} else if strings.Contains(stateStr, "VETOED") {
				v.State = VerdictVetoed
			}
		}
	}

	if v.ID == "" {
		return v, fmt.Errorf("missing artifact ID in verdict")
	}

	return v, nil
}
