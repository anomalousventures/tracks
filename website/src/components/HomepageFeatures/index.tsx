import React from 'react';
import clsx from 'clsx';
import styles from './styles.module.css';

type FeatureItem = {
  title: string;
  emoji: string;
  description: JSX.Element;
};

const FeatureList: FeatureItem[] = [
  {
    title: 'Rapid Development',
    emoji: '🚀',
    description: (
      <>
        Generate complete CRUD resources with a single command. Interactive TUI
        for project setup. Live reload with Air. AI-powered development via MCP.
      </>
    ),
  },
  {
    title: 'Type-Safe Everything',
    emoji: '🔒',
    description: (
      <>
        Type-safe templates with templ, type-safe SQL with SQLC. Catch errors
        at compile time, not runtime. Production-ready code from day one.
      </>
    ),
  },
  {
    title: 'Security First',
    emoji: '🛡️',
    description: (
      <>
        Built-in authentication (magic links, OTP, OAuth). RBAC authorization
        with Casbin. Security headers configured by default. Input validation included.
      </>
    ),
  },
];

function Feature({title, emoji, description}: FeatureItem) {
  return (
    <div className={clsx('col col--4')}>
      <div className="text--center">
        <span style={{fontSize: '4rem'}}>{emoji}</span>
      </div>
      <div className="text--center padding-horiz--md">
        <h3>{title}</h3>
        <p>{description}</p>
      </div>
    </div>
  );
}

export default function HomepageFeatures(): JSX.Element {
  return (
    <section className={styles.features}>
      <div className="container">
        <div className="row">
          {FeatureList.map((props, idx) => (
            <Feature key={idx} {...props} />
          ))}
        </div>
      </div>
    </section>
  );
}
