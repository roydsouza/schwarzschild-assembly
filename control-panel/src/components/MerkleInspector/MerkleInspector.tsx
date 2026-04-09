'use client';

import React from 'react';
import styles from './MerkleInspector.module.css';
import { Fingerprint, Box, ChevronRight, Binary } from 'lucide-react';

interface LogEntry {
  index: number;
  hash: string;
  type: string;
  timestamp: string;
}

const MOCK_ENTRIES: LogEntry[] = [
  { index: 124, hash: 'f4ab83c192d1...33e4', type: 'ACTION_SAFE', timestamp: '14:22:01' },
  { index: 123, hash: 'a1b2c3d4e5f6...77a1', type: 'FACTORY_SYNC', timestamp: '14:15:30' },
  { index: 122, hash: '99e32a184bc2...dd0f', type: 'ACTION_SAFE', timestamp: '14:02:12' },
  { index: 121, hash: '88c1b0d2e5f3...aac9', type: 'POLICY_UPDATE', timestamp: '13:58:45' },
];

export default function MerkleInspector() {
  return (
    <div className={styles.container}>
      <header className="terminal-header">
        <Fingerprint size={12} style={{marginRight: '6px'}} /> AUDIT_TRAIL_BROWSER
      </header>

      <div className={styles.treeVisual}>
        <div className={styles.rootNode}>
          <div className={styles.nodeLabel}>ROOT_HASH (STH)</div>
          <div className={styles.nodeValue}>0xA4F2...1E9D</div>
        </div>
        <div className={styles.connector}></div>
      </div>

      <div className={styles.log}>
        <div className={styles.logHeader}>
          <span>INDEX</span>
          <span>EVENT_TYPE</span>
          <span>COMMIT_HASH</span>
        </div>
        
        {MOCK_ENTRIES.map(entry => (
          <div key={entry.index} className={styles.logEntry}>
            <span className={styles.index}>{entry.index}</span>
            <span className={styles.type}>
              <Binary size={10} style={{marginRight: '4px'}} />
              {entry.type}
            </span>
            <span className={styles.hash}>{entry.hash}</span>
            <ChevronRight size={12} className={styles.chevron} />
          </div>
        ))}
      </div>

      <footer className={styles.footer}>
        <div className={styles.treeStats}>
          SIZE: 125 LEAVES | CONSISTENCY: VERIFIED [RFC6962]
        </div>
      </footer>
    </div>
  );
}
