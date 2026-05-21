---
layout: home

hero:
  name: "qdoc"
  text: "Query the Docs"
  tagline: Documentation research for AI coding agents. One query, one answer, with citations. No trial-and-error. No wasted tokens.
  actions:
    - theme: brand
      text: Get Started
      link: /guide/getting-started
    - theme: alt
      text: Why qdoc
      link: /guide/why
    - theme: alt
      text: View on GitHub
      link: https://github.com/ibrhr/qdoc

features:
  - icon: "⚡"
    title: 3x Fewer Tokens
    details: qdoc uses ~2-4k tokens per query vs 10-15k+ when agents research docs manually.
  - icon: "📚"
    title: 6 Sources Built In
    details: Go, Python, Next.js, React, FastAPI, Pydantic — plus any local docs directory.
  - icon: "🔌"
    title: 4 Providers
    details: OpenAI, DeepSeek, OpenCode Zen, OpenCode Go — or any OpenAI-compatible API.
  - icon: "<svg xmlns='http://www.w3.org/2000/svg' width='32' height='32' viewBox='0 0 24 24' fill='none' stroke='currentColor' stroke-width='2' stroke-linecap='round' stroke-linejoin='round'><path d='M12 2L2 7l10 5 10-5-10-5z'/><path d='M2 17l10 5 10-5'/><path d='M2 12l10 5 10-5'/></svg>"
    title: Saves Agent Tokens
    details: When your agent researches docs manually, it burns 10-15k tokens per question across multiple rounds of guessing and fetching. qdoc does the research internally and returns a single answer — typically 3-4x fewer tokens.
  - icon: "<svg xmlns='http://www.w3.org/2000/svg' width='32' height='32' viewBox='0 0 24 24' fill='none' stroke='currentColor' stroke-width='2' stroke-linecap='round' stroke-linejoin='round'><path d='M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z'/><polyline points='14 2 14 8 20 8'/><line x1='16' y1='13' x2='8' y2='13'/><line x1='16' y1='17' x2='8' y2='17'/><polyline points='10 9 9 9 8 9'/></svg>"
    title: One-Shot Answers
    details: Ask a question. Get a definitive answer with citations. qdoc runs up to 5 research turns internally — fetching, reading, and deciding which pages matter — then returns a single synthesized response.
  - icon: "<svg xmlns='http://www.w3.org/2000/svg' width='32' height='32' viewBox='0 0 24 24' fill='none' stroke='currentColor' stroke-width='2' stroke-linecap='round' stroke-linejoin='round'><polyline points='16 3 21 3 21 8'/><line x1='4' y1='20' x2='21' y2='3'/><polyline points='21 16 21 21 16 21'/><line x1='15' y1='15' x2='21' y2='21'/><line x1='4' y1='4' x2='9' y2='9'/></svg>"
    title: Headless by Design
    details: Built for agents. <code>--json</code> for structured output with answer, source, and step traces. <code>--no-tui</code> for raw markdown to stdout. Exit code 0 on success, 1 on failure. Zero prompt overhead.
  - icon: "<svg xmlns='http://www.w3.org/2000/svg' width='32' height='32' viewBox='0 0 24 24' fill='none' stroke='currentColor' stroke-width='2' stroke-linecap='round' stroke-linejoin='round'><polyline points='4 17 10 11 4 17'/><line x1='12' y1='15' x2='20' y2='15'/><polyline points='20 17 14 11 20 17'/></svg>"
    title: Terminal TUI
    details: Rich Bubble Tea interface with live streaming responses, scrollable output, and interactive provider/model pickers. The same engine powers both TUI and headless modes — no quality difference.
  - icon: "<svg xmlns='http://www.w3.org/2000/svg' width='32' height='32' viewBox='0 0 24 24' fill='none' stroke='currentColor' stroke-width='2' stroke-linecap='round' stroke-linejoin='round'><path d='M12 20h9'/><path d='M16.5 3.5a2.121 2.121 0 0 1 3 3L7 19l-4 1 1-4L16.5 3.5z'/></svg>"
    title: Any OpenAI-Compatible API
    details: OpenAI, DeepSeek, OpenCode Zen, OpenCode Go — four providers built in. Any OpenAI-compatible endpoint works. Switch providers with one command.
  - icon: "<svg xmlns='http://www.w3.org/2000/svg' width='32' height='32' viewBox='0 0 24 24' fill='none' stroke='currentColor' stroke-width='2' stroke-linecap='round' stroke-linejoin='round'><path d='M22 19a2 2 0 0 1-2 2H4a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h5l2 3h9a2 2 0 0 1 2 2z'/></svg>"
    title: Any Docs Source
    details: "Built-in: Go, Python, Next.js, React, FastAPI, Pydantic. Point qdoc at any local directory of markdown, HTML, reStructuredText, or AsciiDoc files."
---
