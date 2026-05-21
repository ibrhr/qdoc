import { defineConfig } from 'vitepress'

const version = typeof process !== 'undefined' ? (process.env.QDOC_VERSION || 'dev') : 'dev';

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
      { text: 'Why qdoc', link: '/guide/why' },
      { text: 'Agent Usage', link: '/guide/agent-usage' },
      { text: 'Reference', link: '/reference/providers' },
      { text: 'GitHub', link: 'https://github.com/ibrhr/qdoc' },
    ],
    sidebar: {
      '/guide/': [
        {
          text: 'Guide',
          items: [
            { text: 'Getting Started', link: '/guide/getting-started' },
            { text: 'Why qdoc', link: '/guide/why' },
            { text: 'Installation', link: '/guide/installation' },
            { text: 'Configuration', link: '/guide/configuration' },
            { text: 'Agent Usage', link: '/guide/agent-usage' },
            { text: 'FAQ', link: '/guide/faq' },
            { text: 'Changelog', link: '/guide/changelog' },
            { text: 'Contributing', link: '/guide/contributing' },
          ],
        },
      ],
      '/reference/': [
        {
          text: 'Reference',
          items: [
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
      message: `Released under the MIT License. qdoc v${version}`,
    },
  },
})
