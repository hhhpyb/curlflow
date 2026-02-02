# Project Context: curlflow

## Overview
**curlflow** is a desktop application built using **Wails**, which combines a **Go** backend with a **Vue 3 + TypeScript** frontend. It appears to be based on the official Wails Vue-TS template.

## Technology Stack

### Backend
*   **Language:** Go
*   **Framework:** [Wails](https://wails.io/) (v2)
*   **Key Files:**
    *   `main.go`: Application entry point.
    *   `app.go`: Application lifecycle logic and exposed Go methods.

### Frontend
*   **Framework:** Vue 3
*   **Language:** TypeScript
*   **Bundler:** Vite
*   **UI Library:** Naive UI (`naive-ui`, `@vicons/ionicons5`)
*   **Styling:** Tailwind CSS (`tailwindcss`, `postcss`, `autoprefixer`)
*   **Editor:** Monaco Editor (`@guolao/vue-monaco-editor`)
*   **Key Files:**
    *   `frontend/src/main.ts`: Frontend entry point.
    *   `frontend/src/App.vue`: Main Vue component.
    *   `frontend/wailsjs/`: Auto-generated bindings for calling Go methods from JS/TS.

## Building and Running

### Prerequisites
*   Go (1.18+)
*   Node.js & npm
*   Wails CLI (`go install github.com/wailsapp/wails/v2/cmd/wails@latest`)

### Development
To run the application in live development mode (with hot reloading):
```bash
wails dev
```
*   This starts a Vite dev server for the frontend.
*   A dedicated dev server for browser access usually runs on `http://localhost:34115`.

### Production Build
To build a redistributable application:
```bash
wails build
```
The output binary will be located in `build/bin/`.

## Project Structure
*   `app.go`: Contains the `App` struct and methods exposed to the frontend.
*   `main.go`: Sets up the Wails application options and starts the app.
*   `frontend/`: Source code for the Vue application.
    *   `src/`: Components, assets, and styles.
    *   `wailsjs/`: Generated Go bindings.
*   `build/`: Build artifacts and configuration for different platforms (Windows, macOS).
*   `wails.json`: Wails project configuration.

## Development Conventions
*   **Frontend-Backend Communication:** Methods defined on the `App` struct in `app.go` are exposed to the frontend via the Wails runtime. These are automatically generated into `frontend/wailsjs/`.
*   **Styling:** Utility-first CSS using Tailwind CSS.
*   **Components:** Naive UI is used for ready-made UI components.
