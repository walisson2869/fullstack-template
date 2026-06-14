# Fullstack Template

A production-ready fullstack template to skip the boilerplate and focus on building features. Ships with a Go + Gin backend, Next.js 15 frontend, PostgreSQL database, Docker Compose, hot reload, and integration testing — all wired together and ready to go.

## Table of Contents

- [Features](#features)
- [Tech Stack](#tech-stack)
- [Getting Started](#getting-started)
  - [Prerequisites](#prerequisites)
  - [Installation](#installation)
  - [Running the App](#running-the-app)
- [Project Structure](#project-structure)
- [Environment Variables](#environment-variables)
- [Development](#development)
- [Testing](#testing)
- [Contributing](#contributing)
- [License](#license)

## Features

- **Go backend** with Gin, structured into clean `cmd/` and `internal/` layers
- **Next.js frontend** with React 19, TypeScript, and Tailwind CSS 4
- **PostgreSQL** database managed via Docker Compose
- **Hot reload** on both frontend (`next dev`) and backend ([Air](https://github.com/air-verse/air))
- **Integration tests** using [Testcontainers](https://testcontainers.com/) — no mocks, real DB
- **CORS** pre-configured between frontend and backend
- **`.env` support** via `godotenv`
- **Makefile** for common backend tasks

## Tech Stack

| Layer     | Technology                          |
|-----------|-------------------------------------|
| Frontend  | Next.js 16, React 19, TypeScript    |
| Styling   | Tailwind CSS 4                      |
| Backend   | Go, Gin                             |
| Database  | PostgreSQL 16 (via Docker)          |
| ORM/DB    | pgx v5 (standard library style)     |
| Dev Tools | Air (hot reload), pnpm, Docker      |
| Testing   | Go test, Testcontainers             |

## Getting Started

### Prerequisites

Make sure you have the following installed:

- [Go 1.21+](https://go.dev/dl/)
- [Node.js 20+](https://nodejs.org/) and [pnpm](https://pnpm.io/installation)
- [Docker](https://www.docker.com/) and Docker Compose

### Installation

```bash
# Clone the repository
git clone https://github.com/your-username/fullstack-template.git
cd fullstack-template
```

**Backend:**

```bash
cd backend
cp .env .env.local   # adjust values as needed
go mod download
```

**Frontend:**

```bash
cd frontend
pnpm install
```

### Running the App

**1. Start the database:**

```bash
cd backend
make docker-run
```

**2. Start the backend** (in a new terminal):

```bash
cd backend
make watch       # hot reload via Air
# or: make run  # run once, no reload
```

**3. Start the frontend** (in a new terminal):

```bash
cd frontend
pnpm dev
```

The frontend is available at `http://localhost:3000` and the backend API at `http://localhost:8080`.

## Project Structure

```
fullstack-template/
├── backend/
│   ├── cmd/
│   │   └── api/
│   │       └── main.go          # Entry point
│   ├── internal/
│   │   ├── database/
│   │   │   ├── database.go      # DB connection + queries
│   │   │   └── database_test.go # Integration tests
│   │   └── server/
│   │       ├── server.go        # HTTP server setup
│   │       └── routes.go        # Route definitions
│   ├── .air.toml                # Air hot-reload config
│   ├── .env                     # Environment variables (do not commit secrets)
│   ├── docker-compose.yml       # PostgreSQL service
│   ├── go.mod
│   └── Makefile
└── frontend/
    ├── app/
    │   ├── layout.tsx
    │   ├── page.tsx
    │   └── globals.css
    ├── public/
    ├── next.config.ts
    ├── package.json
    └── tsconfig.json
```

## Environment Variables

Copy `backend/.env` and fill in your values. Never commit secrets to source control.

| Variable               | Description                  | Default      |
|------------------------|------------------------------|--------------|
| `PORT`                 | Backend server port          | `8080`       |
| `APP_ENV`              | Environment (`local`/`prod`) | `local`      |
| `BLUEPRINT_DB_HOST`    | Postgres host                | `localhost`  |
| `BLUEPRINT_DB_PORT`    | Postgres port                | `5432`       |
| `BLUEPRINT_DB_DATABASE`| Database name                | `blueprint`  |
| `BLUEPRINT_DB_USERNAME`| Database user                | —            |
| `BLUEPRINT_DB_PASSWORD`| Database password            | —            |
| `BLUEPRINT_DB_SCHEMA`  | Postgres schema              | `public`     |

## Development

### Backend Makefile commands

```bash
make build       # compile the binary
make run         # run without hot reload
make watch       # run with Air hot reload
make docker-run  # start the Postgres container
make docker-down # stop the Postgres container
make test        # run all tests
make itest       # run integration tests only
make clean       # remove compiled binary
```

### Frontend commands

```bash
pnpm dev     # start dev server with hot reload
pnpm build   # production build
pnpm start   # serve production build
pnpm lint    # run ESLint
```

## Testing

Backend tests use [Testcontainers](https://testcontainers.com/) to spin up a real PostgreSQL instance per test run — no mocking the database layer.

```bash
cd backend
make test    # unit + integration tests
make itest   # integration tests only
```

Docker must be running for integration tests.

## Contributing

Contributions are welcome! Please read [CONTRIBUTING.md](CONTRIBUTING.md) to get started.

## License

This project is licensed under the MIT License — see the [LICENSE](LICENSE) file for details.
