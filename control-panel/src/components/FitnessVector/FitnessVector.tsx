'use client';

import React from 'react';
import styles from './FitnessVector.module.css';
import { Activity, Shield, Hash, Book, Clock, CreditCard } from 'lucide-react';

export interface Metric {
  name: string;
  value: number;
  unit: string;
  direction: 'higher' | 'lower';
  threshold: number;
  weight: number;
}

const MOCK_METRICS: Metric[] = [
  { name: 'SAFETY_COMPLIANCE', value: 1.0, unit: 'violation/day', direction: 'lower', threshold: 0.1, weight: 0.30 },
  { name: 'AUDIT_INTEGRITY', value: 1.0, unit: 'fail/day', direction: 'lower', threshold: 0.1, weight: 0.25 },
  { name: 'DHAMMA_ALIGNMENT', value: 0.94, unit: 'score', direction: 'higher', threshold: 0.6, weight: 0.15 },
  { name: 'SYSTEM_LATENCY', value: 42, unit: 'ms', direction: 'lower', threshold: 100, weight: 0.20 },
  { name: 'OPERATIONAL_COST', value: 12.40, unit: 'usd/day', direction: 'lower', threshold: 15.0, weight: 0.10 },
];

export default function FitnessVector() {
  const getStatus = (m: Metric) => {
    const isGood = m.direction === 'higher' ? m.value >= m.threshold : m.value <= m.threshold;
    // For this simple UI, we'll just check if it's "close" to threshold for amber
    const margin = m.direction === 'higher' ? (m.value - m.threshold) / m.threshold : (m.threshold - m.value) / m.threshold;
    
    if (!isGood) return 'unsafe';
    if (margin < 0.1) return 'warn';
    return 'safe';
  };

  const getIcon = (name: string) => {
    switch (name) {
      case 'SAFETY_COMPLIANCE': return <Shield size={12} />;
      case 'AUDIT_INTEGRITY': return <Hash size={12} />;
      case 'DHAMMA_ALIGNMENT': return <Book size={12} />;
      case 'SYSTEM_LATENCY': return <Clock size={12} />;
      case 'OPERATIONAL_COST': return <CreditCard size={12} />;
      default: return <Activity size={12} />;
    }
  };

  return (
    <div className={styles.container}>
      <header className="terminal-header">
        <Activity size={12} style={{marginRight: '6px'}} /> GLOBAL_FITNESS_VECTOR
      </header>
      
      <div className={styles.grid}>
        {MOCK_METRICS.map(m => {
          const status = getStatus(m);
          return (
            <div key={m.name} className={styles.row}>
              <div className={styles.label}>
                {getIcon(m.name)}
                <span>{m.name}</span>
              </div>
              <div className={`${styles.value} ${styles[`value_${status}`]}`}>
                {m.value.toFixed(m.unit === 'usd/day' ? 2 : 2)}
                <span className={styles.unit}>{m.unit}</span>
              </div>
            </div>
          );
        })}
      </div>

      <div className={styles.aggregate}>
        <div className={styles.aggregateLabel}>WEIGHTED_FITNESS_SUM</div>
        <div className={styles.aggregateValue}>0.982</div>
      </div>
    </div>
  );
}
