'use client';

import React from 'react';
import TranslucentGate from '@/components/TranslucentGate/TranslucentGate';
import FitnessVector from '@/components/FitnessVector/FitnessVector';
import MerkleInspector from '@/components/MerkleInspector/MerkleInspector';
import { useOrchestrator } from '@/hooks/useOrchestrator';
import styles from './page.module.css';

export default function Home() {
  const { proposals, isConnected, approveProposal, denyProposal } = useOrchestrator(true);

  return (
    <main className={styles.main}>
      <header className={styles.pageHeader}>
        <div className={styles.titleGroup}>
          <h1>SATI-CENTRAL // TRANSLUCENT_GATE</h1>
          <div className={styles.tagline}>HUMAN_IN_THE_LOOP_SAFETY_INTERFACE</div>
        </div>
        <div className={styles.connectionStatus}>
          <span className={isConnected ? styles.dot_online : styles.dot_offline}></span>
          {isConnected ? 'ROOT_SPINE_ONLINE' : 'ORCHESTRATOR_OFFLINE'}
        </div>
      </header>

      <section className={styles.queue}>
        <div className={styles.queueHeader}>
          PENDING_PROPOSALS ({proposals.filter(p => p.resolution !== 'COMMITTED' && p.resolution !== 'VETOED').length})
        </div>
        
        <div className={styles.list}>
          {proposals.map((p) => (
            <TranslucentGate 
              key={p.proposal.id}
              data={p}
              onApprove={approveProposal}
              onDeny={denyProposal}
            />
          ))}
          
          {proposals.length === 0 && (
            <div className={styles.empty}>
              NO_PENDING_PROPOSALS_IN_QUEUE
            </div>
          )}
        </div>
      </section>

      <aside className={styles.vectorSidebar}>
        <FitnessVector />
        <div style={{ marginTop: '16px' }}>
          <MerkleInspector />
        </div>
      </aside>
    </main>
  );
}
