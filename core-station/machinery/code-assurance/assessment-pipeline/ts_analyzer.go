package pipeline

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

type TSAnalyzer struct{}

func (t *TSAnalyzer) Name() string { return "TypeScript-Assurance" }

func (t *TSAnalyzer) Analyze(ctx context.Context, projectRoot string) (Result, error) {
	start := time.Now()
	res := Result{
		CorrectnessScore: 1.0,
		QualityScore:     1.0,
		Findings:         []Finding{},
	}

	// 1. Check for package.json
	if _, err := os.Stat(filepath.Join(projectRoot, "package.json")); os.IsNotExist(err) {
		return res, nil // Not a TS project
	}

	// 2. Run TSC check (Standard Correctness)
	// We use npx tsc --noEmit
	cmd := exec.CommandContext(ctx, "npx", "tsc", "--noEmit")
	cmd.Dir = projectRoot
	output, err := cmd.CombinedOutput()
	if err != nil {
		res.CorrectnessScore = 0.0 // TS errors are non-negotiable for correctness
		res.Findings = append(res.Findings, Finding{
			Tool:     "tsc",
			Severity: SeverityError,
			Message:  "TypeScript compilation errors detected: " + string(output),
			Category: "compilation",
		})
	}

	// 3. Optional: Run Vitest (Correctness)
	// Only if vitest is configured
	if _, err := os.Stat(filepath.Join(projectRoot, "vitest.config.ts")); err == nil {
		testCmd := exec.CommandContext(ctx, "npx", "vitest", "run")
		testCmd.Dir = projectRoot
		if testErr := testCmd.Run(); testErr != nil {
			res.CorrectnessScore -= 0.2
			if res.CorrectnessScore < 0 { res.CorrectnessScore = 0 }
			res.Findings = append(res.Findings, Finding{
				Tool:     "vitest",
				Severity: SeverityError,
				Message:  "Frontend unit tests failed.",
				Category: "testing",
			})
		}
	}

	res.DurationMs = time.Since(start).Milliseconds()
	return res, nil
}
