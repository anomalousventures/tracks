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
      label: 'CLI Reference',
      items: [
        'cli/overview',
        {
          type: 'category',
          label: 'Commands',
          items: ['cli/commands', 'cli/new', 'cli/version', 'cli/help'],
        },
        'cli/output-modes',
      ],
    },
  ],
};

module.exports = sidebars;
