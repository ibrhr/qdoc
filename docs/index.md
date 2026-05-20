---
layout: home

hero:
  name: "qdoc"
  text: "Agent for Agents"
  tagline: Documentation research for AI coding agents. One call, one answer. No trial-and-error. No wasted tokens.
  actions:
    - theme: brand
      text: Get Started
      link: /guide/getting-started
    - theme: alt
      text: Agent Integration
      link: /guide/agent-usage
    - theme: alt
      text: View on GitHub
      link: https://github.com/ibrhr/qdoc

features:
  - icon: "<svg xmlns='http://www.w3.org/2000/svg' width='32' height='32' viewBox='0 0 24 24' fill='none' stroke='currentColor' stroke-width='2' stroke-linecap='round' stroke-linejoin='round'><path d='M12 2L2 7l10 5 10-5-10-5z'/><path d='M2 17l10 5 10-5'/><path d='M2 12l10 5 10-5'/></svg>"
    title: Saves Agent Tokens
    details: Agents don't need to guess which doc pages are relevant. qdoc passes the full doc index to an LLM, which selects exactly the right pages. No wasted fetches. No irrelevant content. The LLM reads the right pages on the first pass.
  - icon: "<svg xmlns='http://www.w3.org/2000/svg' width='32' height='32' viewBox='0 0 24 24' fill='none' stroke='currentColor' stroke-width='2' stroke-linecap='round' stroke-linejoin='round'><path d='M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z'/><polyline points='14 2 14 8 20 8'/><line x1='16' y1='13' x2='8' y2='13'/><line x1='16' y1='17' x2='8' y2='17'/><polyline points='10 9 9 9 8 9'/></svg>"
    title: One-Shot Answers
    details: Ask a question. Get a single, definitive answer with citations. No back-and-forth. No clarifying questions. qdoc does up to 5 turns of doc research internally, then returns the synthesized result as one response.
  - icon: "<svg xmlns='http://www.w3.org/2000/svg' width='32' height='32' viewBox='0 0 24 24' fill='none' stroke='currentColor' stroke-width='2' stroke-linecap='round' stroke-linejoin='round'><polyline points='16 3 21 3 21 8'/><line x1='4' y1='20' x2='21' y2='3'/><polyline points='21 16 21 21 16 21'/><line x1='15' y1='15' x2='21' y2='21'/><line x1='4' y1='4' x2='9' y2='9'/></svg>"
    title: Headless by Design
    details: Built for agents. <code>--json</code> for structured JSON with answer, source, and step traces. <code>--no-tui</code> for raw markdown to stdout. Exit code 0 on success, 1 on failure. Zero prompt overhead.
  - icon: "<svg xmlns='http://www.w3.org/2000/svg' width='32' height='32' viewBox='0 0 24 24' fill='none' stroke='currentColor' stroke-width='2' stroke-linecap='round' stroke-linejoin='round'><polyline points='4 17 10 11 4 17'/><line x1='12' y1='15' x2='20' y2='15'/><polyline points='20 17 14 11 20 17'/></svg>"
    title: Terminal TUI (for humans)
    details: Rich Bubble Tea v2 interface with live streaming responses, scrollable output, interactive provider and model pickers. The same engine powers both the TUI and headless modes — no quality difference.
  - icon: "<svg xmlns='http://www.w3.org/2000/svg' width='32' height='32' viewBox='0 0 24 24' fill='none' stroke='currentColor' stroke-width='2' stroke-linecap='round' stroke-linejoin='round'><path d='M12 20h9'/><path d='M16.5 3.5a2.121 2.121 0 0 1 3 3L7 19l-4 1 1-4L16.5 3.5z'/></svg>"
    title: Any OpenAI-Compatible API
    details: OpenAI, DeepSeek, Opencode Zen, Opencode Go — four providers built in. Any OpenAI-compatible API endpoint works. Switch providers with one command. Env var or config file. Your call.
  - icon: "<svg xmlns='http://www.w3.org/2000/svg' width='32' height='32' viewBox='0 0 24 24' fill='none' stroke='currentColor' stroke-width='2' stroke-linecap='round' stroke-linejoin='round'><path d='M22 19a2 2 0 0 1-2 2H4a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h5l2 3h9a2 2 0 0 1 2 2z'/></svg>"
    title: Any Docs Source
    details: "Built-in: Go standard library docs and FastAPI framework docs. Point qdoc at any local directory of markdown, HTML, reStructuredText, or AsciiDoc files. Extensible with custom sources."
---
