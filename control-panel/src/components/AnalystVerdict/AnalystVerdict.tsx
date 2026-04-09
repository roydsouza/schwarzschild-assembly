'use client';

import React from 'react';
import ReactMarkdown from 'react-markdown';
import styles from './AnalystVerdict.module.css';
import { AlertShield, CheckCircle, XCircle, Info } from 'lucide-react';

export type VerdictStatus = 'APPROVED' | 'VETOED' | 'PENDING';

interface Props {
  status: VerdictStatus;
  rationale: string;
  date?: string;
}

export default function AnalystVerdict({ status, rationale, date }: Props) {
  const isApproved = status === 'APPROVED';
  const isVetoed = status === 'VETOED';

  return (
    <div className={`${styles.verdict} ${styles[`verdict_${status.toLowerCase()}`]}`}>
      <header className={styles.header}>
        <div className={styles.statusLabel}>
          {isApproved && <CheckCircle size={14} className={styles.icon} />}
          {isVetoed && <XCircle size={14} className={styles.icon} />}
          {!isApproved && !isVetoed && <Info size={14} className={styles.icon} />}
          <span>ANALYST_VERDICT // {status}</span>
        </div>
        {date && <div className={styles.date}>{date}</div>}
      </header>
      
      <div className={styles.content}>
        <ReactMarkdown className={styles.markdown}>
          {rationale}
        </ReactMarkdown>
      </div>
      
      {isVetoed && (
        <footer className={styles.footer}>
          <div className={styles.warning}>
            CRITICAL: Analyst recommends VETO. Approval blocked unless overridden by human justification.
          </div>
        </footer>
      )}
    </div>
  );
}
