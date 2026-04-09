'use client';

import { useState, useEffect, useCallback, useRef } from 'react';
import { io, Socket } from 'socket.io-client';
import { OrchestratorClient } from '@/types/OrchestratorServiceClientPb';
import { ApprovalRequest, VetoRequest, Empty } from '@/types/orchestrator_pb';
import { 
  ProposalResolution, 
  ActionProposal, 
  SafetyVerdict, 
  ResolutionState,
  MerkleNote 
} from '@/types';

// Root Spine targets
const SOCKET_URL = 'http://localhost:8080';
const GRPC_WEB_URL = 'http://localhost:8081';

export function useOrchestrator(useMock = false) {
  const [proposals, setProposals] = useState<ProposalResolution[]>([]);
  const [isConnected, setIsConnected] = useState(false);
  const [error, setError] = useState<string | null>(null);
  
  // Persistence refs for networking
  const socketRef = useRef<Socket | null>(null);
  const grpcRef = useRef<OrchestratorClient | null>(null);

  useEffect(() => {
    if (useMock) {
      setIsConnected(true);
      return;
    }

    // 1. Initialize gRPC-Web client
    grpcRef.current = new OrchestratorClient(GRPC_WEB_URL);

    // 2. Initialize Socket.IO connection
    const socket = io(SOCKET_URL, {
      transports: ['websocket'],
      reconnectionAttempts: 5,
    });
    socketRef.current = socket;

    socket.on('connect', () => {
      setIsConnected(true);
      setError(null);
      console.log('[CONTROL_PANEL] Connected to Root Spine');
    });

    socket.on('disconnect', () => {
      setIsConnected(false);
      console.log('[CONTROL_PANEL] Disconnected from Root Spine');
    });

    socket.on('connect_error', (err: Error) => {
      setError(`Connection failed: ${err.message}`);
      setIsConnected(false);
    });

    // 3. Handle live verification events
    socket.on('verification_event', (raw: string) => {
      try {
        const event = JSON.parse(raw);
        handleIncomingEvent(event);
      } catch (e) {
        console.error('Failed to parse verification event', e);
      }
    });

    return () => {
      socket.disconnect();
    };
  }, [useMock]);

  /**
   * handleIncomingEvent updates the proposal queue based on real-time signals.
   */
  const handleIncomingEvent = (event: any) => {
    setProposals(prev => {
      const existing = prev.find(p => p.proposal.id === event.proposal_id);
      
      // Map PB event types to UI resolution states
      let newState: ResolutionState = 'RECEIVED';
      if (event.event_type === 1) newState = 'SAFE'; // VERIFICATION_SAFE
      if (event.event_type === 2) newState = 'UNSAFE'; // VERIFICATION_UNSAFE
      if (event.event_type === 3) newState = 'CHECKING'; // GATE_PENDING
      if (event.event_type === 4) newState = 'COMMITTED'; // GATE_APPROVED
      if (event.event_type === 5) newState = 'VETOED'; // GATE_DENIED

      if (!existing) {
        // This is a new proposal we haven't seen yet
        const newProposal: ProposalResolution = {
          proposal: {
            id: event.proposal_id,
            agentId: 'unknown', // Would be enriched by a separate fetch if needed
            description: 'Live proposal from Root Spine...',
            payloadHash: '',
            isSecurityAdjacent: event.event_type === 3,
            submittedAtMs: event.timestamp_ms || Date.now(),
          },
          resolution: newState,
        };
        return [newProposal, ...prev];
      }

      // Update existing proposal
      return prev.map(p => {
        if (p.proposal.id === event.proposal_id) {
          return {
            ...p,
            resolution: newState,
            // If it's a safe result, attach the proof
            verdict: event.safe_result ? {
              isSafe: true,
              tier: 'SAFETY_TIER_1',
              durationMs: event.safe_result.duration_ms,
              policyFingerprint: '',
              proofBytes: event.safe_result.proof_certificate_hex,
            } : p.verdict
          };
        }
        return p;
      });
    });
  };

  /**
   * approveProposal sends the human approval signature to the Translucent Gate.
   */
  const approveProposal = useCallback(async (id: string, signature: string) => {
    if (!grpcRef.current) return;

    console.log(`[ORCHESTRATOR] Signing approval for ${id}`);
    
    const req = new ApprovalRequest();
    req.setProposalId(id);
    req.setApprovalSignature(signature);

    try {
      const proof = await grpcRef.current.approveAction(req, {});
      console.log(`[ORCHESTRATOR] Proposal ${id} committed to Merkle log`, proof.toObject());
      
      setProposals(prev => prev.map(p => 
        p.proposal.id === id ? { ...p, resolution: 'COMMITTED' as const } : p
      ));
    } catch (err: any) {
      console.error('ApproveAction failed', err);
      setError(`Approval failed: ${err.message}`);
    }
  }, []);

  /**
   * denyProposal sends a manual veto to the Root Spine.
   */
  const denyProposal = useCallback(async (id: string) => {
    if (!grpcRef.current) return;

    console.log(`[ORCHESTRATOR] Vetoing proposal ${id}`);
    
    const req = new VetoRequest();
    req.setProposalId(id);
    req.setVetoedBy('OPERATOR');
    req.setRationale('Manual veto from Control Panel');

    try {
      await grpcRef.current.vetoAction(req, {});
      setProposals(prev => prev.map(p => 
        p.proposal.id === id ? { ...p, resolution: 'VETOED' as const } : p
      ));
    } catch (err: any) {
      console.error('VetoAction failed', err);
      setError(`Veto failed: ${err.message}`);
    }
  }, []);

  return {
    proposals,
    isConnected,
    error,
    approveProposal,
    denyProposal,
  };
}
