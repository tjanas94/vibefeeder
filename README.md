# VibeFeeder

VibeFeeder is a web application designed to simplify online content consumption. It aggregates articles from user-provided RSS feeds and uses AI to generate concise, daily summaries. This tool is perfect for individuals who want to stay informed across various topics but lack the time to browse through multiple sources individually.

## Table of Contents

- [Project Description](#vibefeeder)
- [Tech Stack](#tech-stack)
- [Getting Started Locally](#getting-started-locally)
- [Available Scripts](#available-scripts)
- [Project Scope (MVP)](#project-scope-mvp)
- [Project Status](#project-status)
- [License](#license)

## Tech Stack

The project is built with a modern tech stack focused on performance and developer experience.

- **Backend:**
  - **Language:** [Go](https://go.dev/) (v1.25)
  - **Framework:** [Echo](https://echo.labstack.com/) for routing and middleware.
  - **Templating:** [Templ](https://templ.guide/) for server-side rendered HTML.
- **Frontend:**
  - **Frameworks:** [htmx](https://htmx.org/) for dynamic interactions and [Alpine.js](https://alpinejs.dev/) for component state management.
  - **Styling:** [Tailwind CSS](https://tailwindcss.com/) (v4) with the [DaisyUI](https://daisyui.com/) component library.
  - **Runtime:** [Node.js](https://nodejs.org/) (v22)
- **Database & Auth:**
  - [Supabase](https://supabase.com/) (PostgreSQL) for the database and user authentication.
- **AI Integration:**
  - [OpenRouter](https://openrouter.ai/) to interact with various large language models for summary generation.
- **Tooling & DevOps:**
  - **Task Runner:** [Go-Task](https://taskfile.dev/) for automating development and build tasks.
  - **CI/CD:** [GitHub Actions](https://github.com/features/actions).
  - **Hosting:** [Hetzner Cloud](https://www.hetzner.com/cloud) via Docker.

## Getting Started Locally

To set up and run the project on your local machine, follow these steps.

### Prerequisites

Make sure you have the following tools installed:

- [Go](https://go.dev/doc/install) (version 1.25)
- [Node.js](https://nodejs.org/en/download) (version 22)
- [Go-Task](https://taskfile.dev/installation/)

### Installation & Setup

1.  **Clone the repository:**

    ```sh
    git clone https://github.com/tjanas94/vibefeeder.git
    cd vibefeeder
    ```

2.  **Install dependencies:**

    ```sh
    task install-deps
    ```

3.  **Set up environment variables:**
    Create a `.env` file in the root of the project and add the necessary environment variables for database connections and API keys.

4.  **Run the development server:**
    This command starts the development server with hot-reloading for Go, Templ, and CSS changes.
    ```sh
    task dev
    ```

The application should now be running on your local server.

## Available Scripts

This project uses `Taskfile.yml` to define and run scripts. Below are the most common tasks.

| Command             | Description                                               |
| ------------------- | --------------------------------------------------------- |
| `task install-deps` | Install project dependencies.                             |
| `task dev`          | Run the complete development environment with hot-reload. |
| `task build`        | Create a production-ready build of the application.       |
| `task rebuild`      | Clean all build artifacts and then run a new build.       |
| `task run`          | Run the compiled binary from the `dist/` directory.       |
| `task clean`        | Remove all build artifacts and generated files.           |
| `task lint`         | Run all available linters (Go, Prettier).                 |
| `task fmt`          | Format all code in the project (Go, Templ, others).       |

## Project Scope (MVP)

The scope for the Minimum Viable Product (MVP) is focused on delivering the core functionality.

### Key Features

- **User Management:** Secure user registration, login, and logout.
- **RSS Feed Management:** Ability to add, edit, delete, and view a list of RSS feeds with their status.
- **Content Aggregation:** The system automatically fetches new articles from all active feeds.
- **On-Demand Summaries:** Users can generate a single, consolidated summary from articles published in the last 24 hours.

## Project Status

This project is currently in the **MVP development phase**.

## License

This project is distributed under the **MIT License**. See the `LICENSE` file for more information.
