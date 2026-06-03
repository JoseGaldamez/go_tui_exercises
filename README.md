# Go TUI Exercises: CLI Harness Development (Gemini/Antigravity Style)

Welcome to the **Go TUI Exercises** repository! This project is a curated, step-by-step learning journey designed to teach you how to build a highly responsive, interactive, and beautifully styled command-line interface (CLI) execution **Harness** from scratch in Go.

Inspired by elite developer tools like the **Gemini CLI** and **Antigravity CLI**, these exercises guide you through constructing high-performance terminal environments using **Bubble Tea** (the industry-standard Go TUI framework) and **Lip Gloss** (a terminal layout and styling engine).

## Repository Goal

The primary goal of this repository is to act as a hands-on workbook for mastering terminal user interface (TUI) engineering in Go.

Through a series of progressive, bite-sized exercises, you will learn to build custom terminal controls, handle low-level input/output loops, manage byte/rune state buffers securely, and design state-of-the-art visual panels without relying on heavy off-the-shelf widgets.

---

## Exercises Roadmap

This repository is structured sequentially. As the collection grows, each exercise will build upon the last to construct increasingly complex terminal applications.

### [001] Echo REPL & Command History (`/001_echo_repl_history`)

Build the core foundation of an interactive harness by implementing a custom character-by-character text input buffer and a scrollable history log.

- **Concepts introduced:** The Elm Architecture (Model-Update-View), basic Lip Gloss layout boxes, UTF-8 safe backspacing, and a lightweight slash-command router (`/pop`, `/shift`, `/clean`, `/delete`).
- **Exercise Details:** For complete architecture details, implementation steps, and safety guides, see the project-specific [001 README](file:///Users/josegaldamez/Documents/golang/go_tui_exercises/001_echo_repl_history/README.md).

---

### [002] Textinput REPL & Command History (`/002_textinput_repl_history`)

Refactor the CLI Harness to replace manual rune accumulation and custom backspacing/cursor rendering with the official `github.com/charmbracelet/bubbles/textinput` component.

- **Concepts introduced:** Bubble Tea sub-components, delegation of state and event loops, custom component styling (text, prompt, cursor), and standard text editing controls (cursor navigation, backspacing, and keyboard shortcuts).
- **Exercise Details:** For complete architecture details, refactor comparison, and setup, see the project-specific [002 README](file:///Users/josegaldamez/Documents/golang/go_tui_exercises/002_textinput_repl_history/README.md).

---

_Crafted with ❤️ as an interactive learning module for building state-of-the-art developer CLI interfaces in Go._
