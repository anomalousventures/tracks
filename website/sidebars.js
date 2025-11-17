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
      items: ['getting-started/installation', 'getting-started/quickstart'],
    },
    {
      type: 'category',
      label: 'Guides',
      items: [
        'guides/architecture-overview',
        'guides/example-walkthrough',
        'guides/layer-guide',
        'guides/routing-guide',
        'guides/patterns',
        'guides/testing',
      ],
    },
    {
      type: 'category',
      label: 'CLI Reference',
      items: [
        'cli/overview',
        'cli/output-modes',
        {
          type: 'category',
          label: 'Commands',
          items: ['cli/commands', 'cli/new', 'cli/version', 'cli/help'],
        },
      ],
    },
  ],
};

module.exports = sidebars;
