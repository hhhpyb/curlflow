# curlflow

> A modern, desktop-based HTTP client optimized for `curl` workflows.

![License](https://img.shields.io/badge/license-MIT-blue)
![Go](https://img.shields.io/badge/Go-1.23+-00ADD8?logo=go&logoColor=white)
![Vue](https://img.shields.io/badge/Vue-3-4FC08D?logo=vue.js&logoColor=white)
![Wails](https://img.shields.io/badge/Wails-v2-CF1D2B?logo=wails&logoColor=white)

**curlflow** is a powerful cross-platform desktop application that bridges the gap between command-line `curl` usage and graphical HTTP clients. Built with [Wails](https://wails.io/), it combines the performance of Go with a reactive Vue 3 frontend.

![App Screenshot](TODO: Add main application screenshot here)

## ✨ Key Features

- **🔄 Curl Integration Core**
  - **Parse**: Paste complex `curl` commands directly into the UI to decompose them into editable fields (Headers, Body, Params).
  - **Build**: Visual changes to requests automatically regenerate the corresponding `curl` command for easy sharing.
  - Supports flags like `-X`, `-H`, `-d`, `--url`, and more.

- **🚀 Request Management**
  - Execute HTTP requests (GET, POST, PUT, DELETE, etc.).
  - View detailed response metrics: Status, Time, Headers, and Body with syntax highlighting.
  - Organize requests using a file-based system with metadata support (Tags, Status).

- **🌍 Environment Variables**
  - Manage multiple environments (e.g., `Dev`, `Prod`).
  - Use variable substitution in URLs and Headers (e.g., `{{baseUrl}}/api/v1`).
  - Support for path parameter substitution.

- **🔌 OpenAPI / Swagger Sync**
  - Automatically generate and update request collections from remote OpenAPI/Swagger specifications.
  - Keep your local test suite in sync with your API definitions.

- **🛠️ Developer Focused**
  - **Local Storage**: All data is stored locally in JSON files, making it version-control friendly.
  - **Configurable**: Global settings for Proxy, TLS verification, and Timeouts.

## 🛠️ Technology Stack

**Backend:**
- **Go** (v1.23+)
- **Wails** (v2.11.0)
- **Kin-OpenAPI** (OpenAPI support)
- **Cobra** (CLI scaffolding)

**Frontend:**
- **Vue 3** + **TypeScript**
- **Vite** (Build tool)
- **Naive UI** (Component library)
- **Tailwind CSS** (Styling)
- **Pinia** (State Management)
- **Monaco Editor** (Code editing)

## 📦 Installation & Setup

### Prerequisites

Ensure you have the following installed:
- [Go](https://go.dev/) (1.23+)
- [Node.js](https://nodejs.org/) & npm
- [Wails CLI](https://wails.io/docs/gettingstarted/installation)

```bash
go install github.com/wailsapp/wails/v2/cmd/wails@latest
```

### Local Development

1. **Clone the repository**
   ```bash
   git clone https://github.com/yourusername/curlflow.git
   cd curlflow
   ```

2. **Run in Development Mode**
   This starts the Go backend and a local Vite server with hot-reload.
   ```bash
   wails dev
   ```

3. **Access**
   The application will open in a native window. You can also access the frontend via browser at `http://localhost:34115` for UI debugging.

### Building for Production

To build a standalone binary for your OS:

```bash
wails build
```

The output binary will be located in `build/bin/`.

## 📸 Screenshots

| Request Editor | Response View |
|:---:|:---:|
| ![Request](TODO: link to request screenshot) | ![Response](TODO: link to response screenshot) |

## 🤝 Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

1. Fork the Project
2. Create your Feature Branch (`git checkout -b feature/AmazingFeature`)
3. Commit your Changes (`git commit -m 'Add some AmazingFeature'`)
4. Push to the Branch (`git push origin feature/AmazingFeature`)
5. Open a Pull Request

## 📄 License

Distributed under the MIT License. See `LICENSE` for more information.

---
*TODO: Add specific documentation link if available*
*TODO: Add CI/CD build status badge*