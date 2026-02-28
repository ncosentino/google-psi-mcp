---
description: MCP server that exposes Google PageSpeed Insights Core Web Vitals analysis to AI assistants. Analyze LCP, CLS, FCP, TTFB, and more with Claude, GitHub Copilot, and Cursor.
---

# Google PageSpeed Insights MCP

> **Zero-dependency MCP server for Google PageSpeed Insights Core Web Vitals.**
> Pre-built native binaries for Linux, macOS, and Windows. No Node.js. No Python. No .NET runtime. No Go toolchain. Download one binary and configure your AI tool.

Expose Google PageSpeed Insights analysis directly to AI assistants like Claude, GitHub Copilot, and Cursor via the [Model Context Protocol (MCP)](https://modelcontextprotocol.io). Ask your AI to diagnose Core Web Vitals issues and get actionable, code-level recommendations grounded in real Lighthouse data.

---

## Available Tools

| Tool | Description |
|------|-------------|
| [`analyze_page`](tools/analyze-page/) | Analyze a single URL with Google PageSpeed Insights |
| [`analyze_pages`](tools/analyze-pages/) | Analyze multiple URLs in a single call |

---

## Why This Exists

AI assistants are powerful at diagnosing web performance problems -- but they need real data. This MCP server bridges your AI tool to Google's PageSpeed Insights API v5, giving it:

- **Real Core Web Vitals** (LCP, CLS, FCP, TTFB, TBT, Speed Index) with ratings (good/needs-improvement/poor) per Google's official thresholds
- **Category scores** (performance, SEO, accessibility, best-practices) on a 0-100 scale
- **Prioritized opportunities** with estimated savings
- **Failing audits** with specific descriptions and current values

With this server configured, you can ask: _"Analyze my homepage on mobile and desktop and tell me what's hurting my Core Web Vitals score"_ and get a structured, actionable answer.

!!! warning "PSI calls can be slow"
    The PageSpeed Insights API often takes 5-20+ seconds per URL per strategy. Using `strategy="both"` makes two sequential calls. See the [Troubleshooting](troubleshooting/) page for guidance on avoiding timeouts.

---

## Quick Start

**Three steps:** get an API key → download a binary → configure your AI tool.

[Get Started :material-arrow-right:](getting-started/){ .md-button .md-button--primary }

---

## About

Built by **[Nick Cosentino](https://www.devleader.ca)** (Dev Leader) -- a software engineer and content creator focused on .NET, C#, and software architecture. This server was born out of real work improving Core Web Vitals on [devleader.ca](https://www.devleader.ca) and the desire to use AI assistants effectively during that process.

**Find Nick online:** [Blog](https://www.devleader.ca) · [YouTube](https://www.youtube.com/@devleaderca) · [Newsletter](https://weekly.devleader.ca) · [LinkedIn](https://linkedin.com/in/nickcosentino)

