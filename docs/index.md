---
layout: home

hero:
  name: "qdoc"
  text: "Query the Docs"
  tagline: An LLM-powered CLI that reads documentation so you don't have to
  actions:
    - theme: brand
      text: Get Started
      link: /guide/getting-started
    - theme: alt
      text: View on GitHub
      link: https://github.com/ibrhr/qdoc

features:
  - icon: 🧠
    title: LLM-Powered Research
    details: qdoc fetches documentation indexes, uses an LLM to pick relevant pages, reads them, and synthesizes a definitive answer — all in one shot.
  - icon: 🔗
    title: Multi-Turn Reading
    details: The LLM can request multiple pages in parallel, drill deeper into specific sections, and iterate until it has a complete answer.
  - icon: 🤖
    title: Agent-Friendly
    details: Use <code>--json</code> or <code>--no-tui</code> for headless mode. Perfect for CI pipelines and AI coding agents.
  - icon: 🖥️
    title: Terminal-First
    details: Built with Bubble Tea. Interactive provider/model selection, live streaming responses, and a rich TUI for every query.
  - icon: 🔌
    title: Multi-Provider
    details: OpenAI, DeepSeek, and Opencode gateways. Just set an API key and go.
  - icon: 📚
    title: Any Docs Source
    details: Built-in sources for Go and FastAPI docs. Point it at any local directory of markdown files too.
---
