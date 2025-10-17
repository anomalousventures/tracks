/**
 * Creating a sidebar enables you to:
 - create an ordered group of docs
 - render a sidebar for each doc of that group
 - provide next/previous navigation

 The sidebars can be generated from the filesystem, or explicitly defined here.

 Create as many sidebars as you want.
 */

// @ts-check

/** @type {import('@docusaurus/plugin-content-docs').SidebarsConfig} */
const sidebars = {
  tutorialSidebar: [
    'intro',
    {
      type: 'category',
      label: 'Getting Started',
      items: [
        'getting-started/installation',
        'getting-started/quick-start',
        'getting-started/first-app',
      ],
    },
    {
      type: 'category',
      label: 'Core Concepts',
      items: [
        'core/architecture',
        'core/database',
        'core/authentication',
        'core/authorization',
        'core/templates',
        'core/routing',
      ],
    },
    {
      type: 'category',
      label: 'Code Generation',
      items: [
        'generation/overview',
        'generation/resources',
        'generation/handlers',
        'generation/services',
      ],
    },
    {
      type: 'category',
      label: 'Testing',
      items: [
        'testing/overview',
        'testing/unit-tests',
        'testing/integration-tests',
        'testing/e2e-tests',
      ],
    },
    {
      type: 'category',
      label: 'Deployment',
      items: [
        'deployment/overview',
        'deployment/docker',
        'deployment/kubernetes',
        'deployment/fly-io',
      ],
    },
  ],
};

module.exports = sidebars;
