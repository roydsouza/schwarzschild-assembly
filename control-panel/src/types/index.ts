export type SafetyTier = 'SAFETY_TIER_UNSPECIFIED' | 'SAFETY_TIER_1' | 'SAFETY_TIER_2';

export type ResolutionState = 'RECEIVED' | 'CHECKING' | 'SAFE' | 'UNSAFE' | 'VETOED' | 'COMMITTED';

export interface ActionProposal {
  id: string;
  agentId: string;
  description: string;
  payloadHash: string;
  targetPath?: string;
  isSecurityAdjacent: boolean;
  submittedAtMs: number;
  payload?: any;
}

export interface SafetyVerdict {
  isSafe: boolean;
  tier: SafetyTier;
  durationMs: number;
  policyFingerprint: string;
  proofBytes?: string; // base64
  errorMessage?: string;
  violations?: Record<string, string>;
}

export interface DhammaReflection {
  score: number;
  root: 'kusala' | 'akusala' | 'neutral';
  citations: string[];
  reasoning: string;
}

export interface FitnessImpact {
  metricName: string;
  delta: number;
  isRegressive: boolean;
}

export interface MerkleNote {
  leafIndex: number;
  rootHash: string;
  timestamp: string;
}

export interface ProposalResolution {
  proposal: ActionProposal;
  verdict?: SafetyVerdict;
  reflection?: DhammaReflection;
  fitness?: FitnessImpact[];
  resolution: ResolutionState;
  audit?: MerkleNote;
  analystVerdict?: {
    status: 'APPROVED' | 'VETOED' | 'PENDING';
    rationale: string;
    date: string;
  };
}
