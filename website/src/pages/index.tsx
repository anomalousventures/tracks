import React from 'react';
import clsx from 'clsx';
import Link from '@docusaurus/Link';
import useDocusaurusContext from '@docusaurus/useDocusaurusContext';
import Layout from '@theme/Layout';
import HomepageFeatures from '@site/src/components/HomepageFeatures';

import styles from './index.module.css';

function HomepageHeader() {
  const { siteConfig } = useDocusaurusContext();
  return (
    <header className={clsx('hero hero--primary', styles.heroBanner)}>
      <div className="container">
        <picture>
          <source
            type="image/webp"
            srcSet="https://anomalous-ventures-public-assets.s3.us-west-1.amazonaws.com/logo-256.webp 256w,
                    https://anomalous-ventures-public-assets.s3.us-west-1.amazonaws.com/logo-512.webp 512w"
            sizes="(max-width: 768px) 256px, 300px"
          />
          <img
            src="https://anomalous-ventures-public-assets.s3.us-west-1.amazonaws.com/logo-512.png"
            srcSet="https://anomalous-ventures-public-assets.s3.us-west-1.amazonaws.com/logo-256.png 256w,
                    https://anomalous-ventures-public-assets.s3.us-west-1.amazonaws.com/logo-512.png 512w"
            sizes="(max-width: 768px) 256px, 300px"
            alt="Tracks Logo"
            style={{ width: '300px', marginBottom: '2rem' }}
            loading="lazy"
          />
        </picture>
        <h1 className="hero__title">{siteConfig.title}</h1>
        <p className="hero__subtitle">{siteConfig.tagline}</p>
        <div className={styles.buttons}>
          <Link className="button button--secondary button--lg" to="/docs/intro">
            Get Started - 5min ⏱️
          </Link>
        </div>
      </div>
    </header>
  );
}

export default function Home(): JSX.Element {
  const { siteConfig } = useDocusaurusContext();
  return (
    <Layout
      title={`${siteConfig.title}`}
      description="A Rails-like web framework for Go that generates idiomatic, production-ready applications"
    >
      <HomepageHeader />
      <main>
        <HomepageFeatures />
      </main>
    </Layout>
  );
}
