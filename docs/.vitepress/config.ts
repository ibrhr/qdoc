import { defineConfig } from 'vitepress'

export default defineConfig({
  title: 'qdoc',
  description: 'Query the Docs — an LLM-powered CLI for documentation',
  base: '/',
  cleanUrls: true,
  head: [
    ['link', { rel: 'icon', href: '/favicon.svg' }],
  ],
  themeConfig: {
    nav: [
      { text: 'Guide', link: '/guide/getting-started' },
      { text: 'Reference', link: '/reference/cli' },
      { text: 'GitHub', link: 'https://github.com/ibrhr/qdoc' },
    ],
    sidebar: {
      '/guide/': [
        {
          text: 'Guide',
          items: [
            { text: 'Getting Started', link: '/guide/getting-started' },
            { text: 'Installation', link: '/guide/installation' },
            { text: 'Configuration', link: '/guide/configuration' },
            { text: 'Agent Usage', link: '/guide/agent-usage' },
          ],
        },
      ],
      '/reference/': [
        {
          text: 'Reference',
          items: [
            { text: 'CLI Commands', link: '/reference/cli' },
            { text: 'Providers & Models', link: '/reference/providers' },
            { text: 'Doc Sources', link: '/reference/sources' },
          ],
        },
      ],
    },
    socialLinks: [
      { icon: 'github', link: 'https://github.com/ibrhr/qdoc' },
    ],
    footer: {
      message: 'Released under the MIT License.',
    },
  },
})
