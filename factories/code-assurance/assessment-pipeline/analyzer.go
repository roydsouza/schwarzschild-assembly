package pipeline

import (
	"context"
)

// AssessmentSeverity defines the impact of a finding.
type AssessmentSeverity string

const (
	SeverityCritical AssessmentSeverity = "CRITICAL"
	SeverityError    AssessmentSeverity = "ERROR"
	SeverityWarning  AssessmentSeverity = "WARNING"
	SeverityInfo     AssessmentSeverity = "INFO"
)

// Finding represents a single issue discovered by an analyzer.
type Finding struct {
	Tool     string             `json:"tool"`
	Severity AssessmentSeverity `json:"severity"`
	Message  string             `json:"message"`
	Location string             `json:"location"` // file:line or path
	Category string             `json:"category"`
}

// Result is the outcome of a single analyzer's run.
type Result struct {
	CorrectnessScore float32   `json:"correctness_score"` // 0.0 to 1.0
	QualityScore     float32   `json:"quality_score"`     // 0.0 to 1.0
	Findings         []Finding `json:"findings"`
	DurationMs       int64     `json:"duration_ms"`
}

// CodeQualityAssessment matches the TypeScript interface CodeQualityPanel expects.
type CodeQualityAssessment struct {
	ArtifactId               string          `json:"artifactId"`
	LintViolations           []LintViolation `json:"lintViolations"`
	MaxCyclomaticComplexity  int             `json:"maxCyclomaticComplexity"`
	MeanCyclomaticComplexity float64         `json:"meanCyclomaticComplexity"`
	CveCritical              int             `json:"cveCritical"`
	CveHigh                  int             `json:"cveHigh"`
	TestCoveragePercent      float64         `json:"testCoveragePercent"`
	DuplicationPercent       float64         `json:"duplicationPercent"`
	OverallStatus            string          `json:"overallStatus"` // 'pass' | 'warn' | 'block'
	AssessedAtMs             int64           `json:"assessedAtMs"`
}

type LintViolation struct {
	Level string `json:"level"` // 'error' | 'warning'
	Count int    `json:"count"`
	Tool  string `json:"tool"`
}

// MapToAssessment converts a raw Result into the typed CodeQualityAssessment.
func (r *Result) MapToAssessment(artifactId string) CodeQualityAssessment {
	assessment := CodeQualityAssessment{
		ArtifactId:    artifactId,
		AssessedAtMs:  0, // Will be set by caller or time.Now()
		OverallStatus: "pass",
	}

	toolViolations := make(map[string]map[string]int) // tool -> level -> count

	for _, f := range r.Findings {
		if f.Category == "security" {
			if f.Severity == SeverityCritical {
				assessment.CveCritical++
			} else if f.Severity == SeverityError {
				assessment.CveHigh++
			}
		}

		// Track violations for linting tools
		level := "warning"
		if f.Severity == SeverityError || f.Severity == SeverityCritical {
			level = "error"
		}

		if _, ok := toolViolations[f.Tool]; !ok {
			toolViolations[f.Tool] = make(map[string]int)
		}
		toolViolations[f.Tool][level]++
	}

	for tool, levels := range toolViolations {
		for level, count := range levels {
			assessment.LintViolations = append(assessment.LintViolations, LintViolation{
				Level: level,
				Count: count,
				Tool:  tool,
			})
		}
	}

	// Decision Logic (Rule 1 thresholds)
	if r.CorrectnessScore < 0.90 || assessment.CveCritical > 0 {
		assessment.OverallStatus = "block"
	} else if r.QualityScore < 0.85 {
		assessment.OverallStatus = "warn"
	}

	return assessment
}

// Analyzer defines the contract for language or tool-specific analysis.
type Analyzer interface {
	Name() string
	Analyze(ctx context.Context, projectRoot string) (Result, error)
}

// Aggregator orchestrates multiple analyzers and merges their results.
type Aggregator struct {
	analyzers []Analyzer
}

func NewAggregator(analyzers ...Analyzer) *Aggregator {
	return &Aggregator{analyzers: analyzers}
}

func (a *Aggregator) Run(ctx context.Context, projectRoot string) (Result, error) {
	var combined Result
	combined.CorrectnessScore = 1.0
	combined.QualityScore = 1.0

	for _, analyzer := range a.analyzers {
		res, err := analyzer.Analyze(ctx, projectRoot)
		if err != nil {
			// Log error but continue with other analyzers
			combined.Findings = append(combined.Findings, Finding{
				Tool:     analyzer.Name(),
				Severity: SeverityError,
				Message:  "Analyzer failed execution: " + err.Error(),
			})
			continue
		}

		// Merged scores (geometric mean or minimum? Minimum is safer for security).
		if res.CorrectnessScore < combined.CorrectnessScore {
			combined.CorrectnessScore = res.CorrectnessScore
		}
		if res.QualityScore < combined.QualityScore {
			combined.QualityScore = res.QualityScore
		}

		combined.Findings = append(combined.Findings, res.Findings...)
		combined.DurationMs += res.DurationMs
	}

	return combined, nil
}
