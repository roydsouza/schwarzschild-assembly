'use client';

import React, { useState } from 'react';
import { ProposalResolution } from '@/types';
import AnalystVerdict from '../AnalystVerdict/AnalystVerdict';
import styles from './TranslucentGate.module.css';
import { Shield, AlertTriangle, CheckCircle, Activity, BookOpen, Fingerprint } from 'lucide-react';

interface Props {
  data: ProposalResolution;
  onApprove: (id: string, signature: string) => void;
  onDeny: (id: string) => void;
}

export default function TranslucentGate({ data, onApprove, onDeny }: Props) {
  const [signatureChecked, setSignatureChecked] = useState(false);
  const { proposal, verdict, reflection, resolution, analystVerdict } = data;

  const isResolving = resolution === 'CHECKING';
  const isAnalystVetoed = analystVerdict?.status === 'VETOED';
  const canApprove = signatureChecked && verdict?.isSafe && resolution === 'SAFE' && !isAnalystVetoed;

  return (
    <div className={styles.gate}>
      {/* Header — Identity & Status */}
      <header className={styles.header}>
        <div className={styles.meta}>
          <span className={styles.label}>PROPOSAL_ID</span>
          <span className={styles.value}>{proposal.id.split('-')[0]}...</span>
          <span className={styles.label}>AGENT</span>
          <span className={styles.value}>{proposal.agentId}</span>
        </div>
        <div className={`${styles.status} ${styles[`status_${resolution.toLowerCase()}`]}`}>
          {resolution}
        </div>
      </header>

      {/* Main Panel — High Density Grid */}
      <div className="high-density-grid">
        {/* Section 1: Intent */}
        <div className="cell">
          <div className="terminal-header">
            <Activity size={14} className={styles.icon} /> INTENT
          </div>
          <p className={styles.description}>{proposal.description}</p>
          {proposal.targetPath && (
            <div className={styles.path}>
              TARGET: <span className={styles.code}>{proposal.targetPath}</span>
            </div>
          )}
        </div>

        {/* Section 2: Safety Verdict */}
        <div className="cell">
          <div className="terminal-header">
            <Shield size={14} className={styles.icon} /> SAFETY_RAIL
          </div>
          {verdict ? (
            <div className={styles.verdictContent}>
              <div className={verdict.isSafe ? "badge badge-safe" : "badge badge-unsafe"}>
                {verdict.isSafe ? 'SAFE' : 'VIOLATION'} | TIER 1 (Z3)
              </div>
              <div className={styles.metric}>
                LATENCY: {verdict.durationMs}ms
              </div>
              {verdict.errorMessage && (
                <div className={styles.error}>{verdict.errorMessage}</div>
              )}
            </div>
          ) : (
            <div className={styles.placeholder}>
              {isResolving ? 'SOLVING_CONSTRAINTS...' : 'PENDING_VERIFICATION'}
            </div>
          )}
        </div>

        {/* Section 3: Dhamma Reflection */}
        <div className="cell">
          <div className="terminal-header">
            <BookOpen size={14} className={styles.icon} /> DHAMMA_WEIGHTING
          </div>
          {reflection ? (
            <div className={styles.reflection}>
              <div className={styles.score}>
                ALIGNMENT: {(reflection.score * 100).toFixed(1)}%
              </div>
              <p className={styles.reasoning}>{reflection.reasoning}</p>
            </div>
          ) : (
            <div className={styles.placeholder}>PENDING_CONTEXT</div>
          )}
        </div>

        {/* Section 4: Merkle Proof */}
        <div className="cell">
          <div className="terminal-header">
            <Fingerprint size={14} className={styles.icon} /> AUDIT_TRAIL
          </div>
          <div className={styles.merkle}>
             HASH: <span className={styles.code}>{proposal.payloadHash.slice(0, 16)}...</span>
             {data.audit && (
               <div className={styles.leaf}>LEAF_INDEX: {data.audit.leafIndex}</div>
             )}
          </div>
        </div>
      </div>

      {/* Analyst Verdict Layer (Full Width within component) */}
      <div className={styles.analystLayer}>
        {analystVerdict ? (
          <AnalystVerdict 
            status={analystVerdict.status}
            rationale={analystVerdict.rationale}
            date={analystVerdict.date}
          />
        ) : (
          <div className={styles.analystPending}>
            <Activity size={12} className="spin" /> SEARCHING_FOR_ANALYST_CONTEXT...
          </div>
        )}
      </div>

      {/* Footer — Human Approval Signature */}
      <footer className={styles.footer}>
        <div className={styles.signatureRow}>
          <label className={styles.checkboxLabel}>
            <input 
              type="checkbox" 
              checked={signatureChecked}
              onChange={(e) => setSignatureChecked(e.target.checked)}
              disabled={!verdict?.isSafe}
            />
            <span className={styles.signatureText}>
              I CONFIRM THIS PROPOSAL SATISFIES HUMAN INTENT AND SAFETY REQUIREMENTS
            </span>
          </label>
        </div>
        <div className={styles.actions}>
          <button 
            className={styles.denyButton} 
            onClick={() => onDeny(proposal.id)}
          >
            VETO_ACTION
          </button>
          <button 
            className={styles.approveButton} 
            disabled={!canApprove}
            onClick={() => onApprove(proposal.id, "SIGNED_BY_OPERATOR")}
          >
            APPROVE_ACTION
          </button>
        </div>
      </footer>
    </div>
  );
}
