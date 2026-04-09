import { render, screen, fireEvent } from '@testing-library/react';
import { expect, test, vi, describe } from 'vitest';
import TranslucentGate from './TranslucentGate';
import { ProposalResolution } from '@/types';

const mockSafeProposal: ProposalResolution = {
  proposal: {
    id: 'test-id',
    agentId: 'test-agent',
    description: 'Safe proposal',
    payloadHash: 'hash',
    isSecurityAdjacent: true,
    submittedAtMs: Date.now(),
  },
  verdict: {
    isSafe: true,
    tier: 'SAFETY_TIER_1',
    durationMs: 10,
    policyFingerprint: 'fp',
  },
  resolution: 'SAFE',
};

const mockUnsafeProposal: ProposalResolution = {
  ...mockSafeProposal,
  proposal: { ...mockSafeProposal.proposal, description: 'Unsafe proposal' },
  verdict: {
    ...mockSafeProposal.verdict!,
    isSafe: false,
    errorMessage: 'Safety violation detected',
  },
  resolution: 'UNSAFE',
};

describe('TranslucentGate Safety Invariants', () => {
  test('Approve button is disabled when signature is not checked', () => {
    const onApprove = vi.fn();
    render(<TranslucentGate data={mockSafeProposal} onApprove={onApprove} onDeny={() => {}} />);
    
    const approveBtn = screen.getByText('APPROVE_ACTION');
    expect(approveBtn).toBeDisabled();
  });

  test('Approve button is enabled when safe and signature is checked', () => {
    const onApprove = vi.fn();
    render(<TranslucentGate data={mockSafeProposal} onApprove={onApprove} onDeny={() => {}} />);
    
    const checkbox = screen.getByRole('checkbox');
    fireEvent.click(checkbox);
    
    const approveBtn = screen.getByText('APPROVE_ACTION');
    expect(approveBtn).not.toBeDisabled();
    
    fireEvent.click(approveBtn);
    expect(onApprove).toHaveBeenCalledWith('test-id', 'SIGNED_BY_OPERATOR');
  });

  test('Approve button is PERMANENTLY disabled for UNSAFE verdicts', () => {
    render(<TranslucentGate data={mockUnsafeProposal} onApprove={() => {}} onDeny={() => {}} />);
    
    const checkbox = screen.getByRole('checkbox');
    // Checkbox itself should be disabled if unsafe
    expect(checkbox).toBeDisabled();
    
    const approveBtn = screen.getByText('APPROVE_ACTION');
    expect(approveBtn).toBeDisabled();
  });

  test('Approve button is disabled if Analyst Veto exists', () => {
    const vetoedProposal: ProposalResolution = {
      ...mockSafeProposal,
    analystVerdict: {
      status: 'VETOED',
      date: new Date().toISOString(),
      rationale: 'Strategic veto',
    }
    };
    
    render(<TranslucentGate data={vetoedProposal} onApprove={() => {}} onDeny={() => {}} />);
    
    const checkbox = screen.getByRole('checkbox');
    fireEvent.click(checkbox); // Even if we try to click
    
    const approveBtn = screen.getByText('APPROVE_ACTION');
    expect(approveBtn).toBeDisabled();
  });
});
