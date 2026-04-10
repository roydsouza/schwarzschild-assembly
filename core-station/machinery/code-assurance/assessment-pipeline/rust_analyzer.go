package pipeline

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

type RustAnalyzer struct{}

func (r *RustAnalyzer) Name() string { return "Rust-Assurance" }

func (r *RustAnalyzer) Analyze(ctx context.Context, projectRoot string) (Result, error) {
	start := time.Now()
	res := Result{
		CorrectnessScore: 1.0,
		QualityScore:     1.0,
		Findings:         []Finding{},
	}

	// 1. Check for Cargo.toml
	if _, err := os.Stat(filepath.Join(projectRoot, "Cargo.toml")); os.IsNotExist(err) {
		return res, nil // Not a Rust project
	}

	// 2. Run Cargo Clippy (Quality & Correctness)
	// We use -- -D warnings to catch issues, but here we just parse the exit code.
	res.runCargo(ctx, projectRoot, []string{"clippy", "--", "-D", "warnings"}, "Clippy", &res.QualityScore, 0.05)

	// 3. Run Cargo Audit (Security)
	res.runCargo(ctx, projectRoot, []string{"audit"}, "Cargo-Audit", &res.CorrectnessScore, 0.20)

	res.DurationMs = time.Since(start).Milliseconds()
	return res, nil
}

func (r *Result) runCargo(ctx context.Context, dir string, args []string, toolName string, scoreTarget *float32, penalty float32) {
	// Note: result.runTool is not accessible as it's a method of Result in another file but same package.
	// Actually, I should have made runTool a package-level helper or a method on Result.
	// Since I'm in the same package 'pipeline', I can use helper functions defined in any file.
	
	// Implementation follows the same logic as go_analyzer's runTool but specific to cargo.
	cmd := exec.CommandContext(ctx, "cargo", args...)
	cmd.Dir = dir
	
	// Maintain PATH for Apple Silicon Homebrew
	path := os.Getenv("PATH")
	cmd.Env = append(os.Environ(), fmt.Sprintf("PATH=%s:/opt/homebrew/bin:/usr/local/bin", path))

	output, err := cmd.CombinedOutput()
	if err != nil {
		msg := string(output)
		if msg == "" {
			msg = err.Error()
		}

		severity := SeverityWarning
		category := "lint"
		if toolName == "Cargo-Audit" {
			severity = SeverityCritical
			category = "security"
		} else if toolName == "Clippy" {
			severity = SeverityError
		}

		r.Findings = append(r.Findings, Finding{
			Tool:     toolName,
			Severity: severity,
			Message:  fmt.Sprintf("%s findings detected: %s", toolName, msg),
			Category: category,
		})

		// Penalize
		*scoreTarget -= penalty
		if *scoreTarget < 0 {
			*scoreTarget = 0
		}
	}
}
