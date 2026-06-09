# Go TUI Exercises: HTTP Client (`/004_http_client`)

Build a **HTTP Client Harness** that lets you quickly fetch and preview the content of any public URL in a terminal-friendly format.

- **Concepts introduced:** Network programming with `net/http`, concurrent request handling with channels, dynamic layout management with Bubble Tea, and HTTP response formatting.
- **Exercise Details:** For complete implementation, see the main.go file.

---

## Project Summary

Create a TUI application that:
- Displays a clean interface with a URL input field
- Shows a loading spinner when fetching URLs
- Displays formatted response data
- Handles network errors gracefully
- Supports common web formats (HTML, plain text)

Press `Q` or `Esc` to exit the application.