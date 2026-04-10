package gate

import (
	"strings"

	"github.com/rds/aethereum-spine/aethereum-spine/internal/grpc/pb"
)

// Gate implements the Translucent Gate routing logic.
type Gate struct{}

// NewGate creates a new Gate.
func NewGate() *Gate {
	return &Gate{}
}

// Route returns true if the proposal requires human approval.
// Security-adjacent targets: safety-rail/, merkle-log/ schema, auth logic,
// proto/ contracts, or CLAUDE.md.
func (g *Gate) Route(p *pb.ActionProposal) bool {
	if p.IsSecurityAdjacent {
		return true
	}

	if p.TargetPath == "" {
		return false
	}

	target := p.TargetPath
	return strings.HasPrefix(target, "core-station/security/") ||
		strings.HasPrefix(target, "proto/") ||
		strings.Contains(target, "merkle") ||
		strings.HasSuffix(target, "CLAUDE.md")
}
