/** @type {import('@docusaurus/plugin-content-docs').SidebarsConfig} */
const sidebars = {
  tutorialSidebar: [
    'index',
    {
      type: 'category',
      label: 'Getting Started',
      items: [
        'introduction/concepts',
        'guides/installation',
        'guides/quick-start',
      ],
    },
    {
      type: 'category',
      label: 'Guides',
      items: [
        'guides/generators',
        'guides/media-providers',
      ],
    },
    {
      type: 'category',
      label: 'Examples',
      items: [
        'examples/overview',
        'examples/basic-password',
        'examples/tls-self-signed',
        'examples/aws-secrets-manager',
      ],
    },
    {
      type: 'category',
      label: 'Contributing',
      items: [
        'contributing/process',
      ],
    },
  ],
};

module.exports = sidebars;