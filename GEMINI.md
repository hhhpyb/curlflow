# Project Context: curlflow

## Overview
**curlflow** is a desktop-based HTTP client application built using **Wails**. It combines a **Go** backend with a **Vue 3 + TypeScript** frontend. The application is designed to help developers manage HTTP requests, specifically focusing on parsing and building `curl` commands, managing environment variables, and syncing with OpenAPI (Swagger) specifications.

## Technology Stack

### Backend
*   **Language:** Go (v1.23+)
*   **Framework:** [Wails](https://wails.io/) (v2.11.0)
*   **CLI Framework:** [Cobra](https://github.com/spf13/cobra)
*   **OpenAPI Support:** [kin-openapi](https://github.com/getkin/kin-openapi)
*   **Key Libraries:** `google/uuid`, `go-shellwords` (for curl parsing).

### Frontend
*   **Framework:** Vue 3
*   **State Management:** Pinia
*   **Language:** TypeScript
*   **Bundler:** Vite
*   **UI Library:** Naive UI
*   **Icons:** `@vicons/ionicons5`
*   **Styling:** Tailwind CSS
*   **Editor:** Monaco Editor (`@guolao/vue-monaco-editor`)

## Key Features
*   **Curl Integration:** Parse complex `curl` commands into a structured UI and reconstruct `curl` commands from the UI. Handles flags like `-X`, `-H`, `-d`, `--url`.
*   **Request Management:** Execute HTTP requests (GET, POST, etc.) and view detailed responses (status, time, headers, body).
*   **Environment Variables:** Manage multiple environments and use variables `{{var}}` in requests. Supports path parameters `{key}` substitution.
*   **Storage:** File-based storage for requests and settings.
    *   **Request Files:** Stored as JSON with `_meta` (ID, status, tags) and `data` (request details).
    *   **Config:** stored in `.curlflow/` or root of workspace.
*   **OpenAPI Syncing:** Automatically generate and update request files from a remote Swagger/OpenAPI URL.
*   **Global Settings:** Configure proxy, insecure TLS, and request timeouts. Saved in user config directory (`~/.config/curlflow` or equivalent).

## Building and Running

### Prerequisites
*   Go (1.23+)
*   Node.js & npm
*   Wails CLI (`go install github.com/wailsapp/wails/v2/cmd/wails@latest`)

### Development
To run the application in live development mode (with hot reloading):
```bash
wails dev
```
*   This starts a Vite dev server for the frontend and a Wails dev server for the backend.
*   Browser-based development (with Go bindings) is available at `http://localhost:34115`.

### Production Build
To build a redistributable application:
```bash
wails build
```
The output binary will be located in `build/bin/`.

## Project Structure

### Backend (`/`)
*   `main.go`: Entry point, Wails configuration.
*   `app.go`: Core application logic and methods exposed to the frontend (Wails Bindings: `App`).
*   `internal/domain/`: Shared data models (`HttpRequest`, `HttpResponse`, `RequestFile`).
*   `internal/parser/`: Logic for parsing and building `curl` commands.
*   `internal/runner/`: Service for executing HTTP requests.
*   `internal/storage/`: Service for file system operations (loading/saving requests, selecting directories).
*   `internal/syncer/`: Service for synchronizing requests with OpenAPI specifications.

### Frontend (`/frontend`)
*   `src/App.vue`: Root Vue component.
*   `src/components/`: UI components (e.g., `MainLayout`, `RequestPanel`, `ResponsePanel`, `Sidebar`).
*   `src/stores/`: Pinia stores for state management:
    *   `request.ts`: Manages the current request/response, file tree logic, and Wails bridge calls.
    *   `env.ts`: Manages environment variables and active environment.
    *   `settings.ts`: Manages global application settings.
*   `wailsjs/`: Auto-generated Go bindings for frontend-backend communication.

## Development Conventions
*   **Exposing Go Methods:** Methods added to the `App` struct in `app.go` are automatically available in the frontend via `window.go.main.App`.
*   **Styling:** Utility-first CSS using Tailwind CSS.
*   **UI Components:** Use Naive UI components for consistency.
*   **State Management:** Use Pinia stores for complex UI states and to encapsulate Wails bridge calls.
*   **File Format:** Saved requests are stored as JSON with a specific structure including metadata (`_meta`) and request data (`data`).

## Memories
- 始终以中文回答用户。 (Always answer the user in Chinese.)
