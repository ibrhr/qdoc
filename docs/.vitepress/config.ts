import { defineConfig } from 'vitepress'

export default defineConfig({
  title: 'qdoc — Agent for Agents',
  description: 'Documentation research for AI coding agents. One LLM call, one answer. No trial-and-error. No wasted tokens.',
  base: '/',
  cleanUrls: true,
  head: [
    ['link', { rel: 'icon', href: '/favicon.svg' }],
    ['meta', { property: 'og:title', content: 'qdoc — Agent for Agents' }],
    ['meta', { property: 'og:description', content: 'Documentation research for AI coding agents. One LLM call, one answer.' }],
  ],
  themeConfig: {
    nav: [
      { text: 'Guide', link: '/guide/getting-started' },
      { text: 'Agent Usage', link: '/guide/agent-usage' },
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
