package pipeline

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

type GoAnalyzer struct{}

func (g *GoAnalyzer) Name() string { return "Go-Assurance" }

func (g *GoAnalyzer) Analyze(ctx context.Context, projectRoot string) (Result, error) {
	start := time.Now()
	res := Result{
		CorrectnessScore: 1.0,
		QualityScore:     1.0,
		Findings:         []Finding{},
	}

	// 1. Check for Go module
	if _, err := os.Stat(filepath.Join(projectRoot, "go.mod")); os.IsNotExist(err) {
		return res, nil // Not a Go project
	}

	// 2. Run Go Vet (Standard Correctness)
	res.runTool(ctx, projectRoot, "go", []string{"vet", "./..."}, "Vet", &res.CorrectnessScore, 0.05)

	// 3. Run Staticcheck (Deep Quality)
	res.runTool(ctx, projectRoot, "staticcheck", []string{"./..."}, "Staticcheck", &res.QualityScore, 0.02)

	// 4. Run Gocyclo (Complexity)
	res.runTool(ctx, projectRoot, "gocyclo", []string{"-over", "10", "."}, "Gocyclo", &res.QualityScore, 0.05)

	// 5. Run Govulncheck (Security)
	res.runTool(ctx, projectRoot, "govulncheck", []string{"./..."}, "Govulncheck", &res.CorrectnessScore, 0.20)

	res.DurationMs = time.Since(start).Milliseconds()
	return res, nil
}

func (r *Result) runTool(ctx context.Context, dir, command string, args []string, toolName string, scoreTarget *float32, penalty float32) {
	cmd := exec.CommandContext(ctx, command, args...)
	cmd.Dir = dir
	
	// Ensure standard PATHs are available for these tools
	// On this environment, we may need to inject /usr/local/go/bin and ~/go/bin
	path := os.Getenv("PATH")
	cmd.Env = append(os.Environ(), fmt.Sprintf("PATH=%s:/usr/local/go/bin:%s/go/bin", path, os.Getenv("HOME")))

	output, err := cmd.CombinedOutput()
	if err != nil {
		// Tool reported issues or failed to run
		msg := string(output)
		if msg == "" {
			msg = err.Error()
		}

		severity := SeverityWarning
		if toolName == "Govulncheck" || toolName == "Vet" {
			severity = SeverityError
		}

		r.Findings = append(r.Findings, Finding{
			Tool:     toolName,
			Severity: severity,
			Message:  fmt.Sprintf("%s findings detected or execution failed.", toolName),
			Category: "static-analysis",
		})

		// Apply penalty
		*scoreTarget -= penalty
		if *scoreTarget < 0 {
			*scoreTarget = 0
		}
	}
}
