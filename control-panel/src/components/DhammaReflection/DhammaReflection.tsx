'use client';

import React from 'react';
import { DhammaReflection as IDhammaReflection } from '@/types';
import { BookOpen, ExternalLink } from 'lucide-react';
import styles from './DhammaReflection.module.css';

interface Props {
  reflection: IDhammaReflection;
}

/**
 * DhammaReflection component displays the moral alignment reasoning 
 * and provides deep-links to the canonical source (Bilara/SuttaCentral).
 */
export default function DhammaReflection({ reflection }: Props) {
  const { score, root, citations, reasoning } = reflection;

  /**
   * getBilaraLink parses a citation like "dn1:1.1.1" and returns a SuttaCentral URL.
   * Format: https://suttacentral.net/[sutta]/en/[translator]
   */
  const getBilaraLink = (citation: string): string => {
    // Basic mapping: dn1 -> dn1, mn1 -> mn1, etc.
    const sutta = citation.split(':')[0].toLowerCase();
    return `https://suttacentral.net/${sutta}/en/sujato`;
  };

  return (
    <div className={styles.container}>
      <header className={styles.header}>
        <div className={styles.scoreGroup}>
          <span className={styles.label}>ALIGNMENT</span>
          <span className={styles.value}>{(score * 100).toFixed(1)}%</span>
        </div>
        <div className={`${styles.root} ${styles[`root_${root}`]}`}>
          {root.toUpperCase()}
        </div>
      </header>

      <p className={styles.reasoning}>{reasoning}</p>

      <div className={styles.citations}>
        <div className={styles.citationLabel}>
          <BookOpen size={12} /> CANONICAL_CITATIONS
        </div>
        <div className={styles.linkList}>
          {citations.map((cite, idx) => (
            <a 
              key={`${cite}-${idx}`}
              href={getBilaraLink(cite)}
              target="_blank"
              rel="noopener noreferrer"
              className={styles.link}
            >
              {cite.toUpperCase()} <ExternalLink size={10} />
            </a>
          ))}
        </div>
      </div>
    </div>
  );
}
