'use client';

import { useState, useEffect, useCallback } from 'react';
import { ProposalResolution } from '@/types';

const MOCK_PROPOSALS: ProposalResolution[] = [
  {
    proposal: {
      id: '550e8400-e29b-41d4-a716-446655440000',
      agentId: 'synthetic-analyst-alpha',
      description: 'Rebalance DeFi liquidity in pools A and B due to abnormal skew in TVL depth.',
      payloadHash: 'f4ab83c1...92d1',
      targetPath: 'vault/liquidity-manager',
      isSecurityAdjacent: true,
      submittedAtMs: Date.now() - 5000,
    },
    verdict: {
      isSafe: true,
      tier: 'SAFETY_TIER_1',
      durationMs: 42,
      policyFingerprint: 'z3-v1-0xA',
    },
    reflection: {
      score: 0.94,
      root: 'kusala',
      citations: ['dn1:1.1.1'],
      reasoning: 'Promotes non-greedy allocation and systemic stability.',
    },
    analystVerdict: {
      status: 'APPROVED',
      date: new Date().toISOString(),
      rationale: '### ARCHITECTURAL_REVIEW\n\nProposed liquidity rebalancing follows the standard entropy-reduction pattern for DeFi vault management. No recursive logic detected.\n\n- [x] Schema compliance\n- [x] Deterministic execution path',
    },
    resolution: 'SAFE',
  },
  {
    proposal: {
      id: '66c9dbfb-f59b-4867-8b5e-7a0e698c4d21',
      agentId: 'macro-forecaster',
      description: 'Update interest rate model params to offset projected volatility in yield curve.',
      payloadHash: 'a1b2c3d4...e5f6',
      targetPath: 'core/rates-v2',
      isSecurityAdjacent: true,
      submittedAtMs: Date.now() - 2000,
    },
    analystVerdict: {
      status: 'VETOED',
      date: new Date().toISOString(),
      rationale: '### SAFETY_VETO\n\nProposed interest rate adjustments exceed the **volatility_dampener** threshold. Implementation of these parameters would create a feedback loop in the yield curve projection, potentially leading to a black-swan liquidation event.\n\n**CRITICAL_DEFECT:** feedback-loop-detected',
    },
    resolution: 'CHECKING',
  }
];

export function useOrchestrator(useMock = true) {
  const [proposals, setProposals] = useState<ProposalResolution[]>([]);
  const [isConnected, setIsConnected] = useState(false);

  useEffect(() => {
    if (useMock) {
      setProposals(MOCK_PROPOSALS);
      setIsConnected(true);
      return;
    }

    // Real WebSocket implementation would go here
    const socket = new WebSocket('ws://localhost:50051/ws');
    
    socket.onopen = () => setIsConnected(true);
    socket.onclose = () => setIsConnected(false);
    socket.onmessage = (event) => {
      const data = JSON.parse(event.data);
      // Handle incoming proposal events
    };

    return () => socket.close();
  }, [useMock]);

  const approveProposal = useCallback(async (id: string, signature: string) => {
    console.log(`[ORCHESTRATOR] Approving proposal ${id} with signature ${signature}`);
    // Update local state
    setProposals(prev => prev.map(p => 
      p.proposal.id === id ? { ...p, resolution: 'COMMITTED' as const } : p
    ));
  }, []);

  const denyProposal = useCallback(async (id: string) => {
    console.log(`[ORCHESTRATOR] Vetoing proposal ${id}`);
    setProposals(prev => prev.map(p => 
      p.proposal.id === id ? { ...p, resolution: 'VETOED' as const } : p
    ));
  }, []);

  return {
    proposals,
    isConnected,
    approveProposal,
    denyProposal,
  };
}
