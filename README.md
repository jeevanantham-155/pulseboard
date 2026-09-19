# Live Polling

A production-oriented live polling application built as two separate applications:

- `frontend`: React and Vite client
- `backend`: Go and Gin API, MongoDB persistence, Redis Pub/Sub, and WebSocket delivery

## Current Status

The core live polling flow is implemented: MongoDB-backed authentication and polls, Redis result events, poll-scoped WebSockets, public voting, and a React dashboard/public voting experience.

## Architecture

The browser talks to the Go backend in two ways:

1. REST requests handle signup, login, poll management, public poll reads, and votes.
2. A WebSocket connection subscribes the browser to results for one public poll.

MongoDB is the source of truth for users, polls, and votes. Redis Pub/Sub distributes result events between backend instances. Each backend instance subscribes to result events and broadcasts only the events for the matching poll to its connected WebSocket clients.

## How Real-Time Updates Work

`Vote -> Go API -> MongoDB/Redis -> Redis Pub/Sub -> Go subscriber -> WebSocket -> React clients`

The completed flow will validate a vote in Gin, persist it in MongoDB, update the appropriate Redis state, publish a poll result event, receive that event through a long-lived Redis subscriber, and broadcast a `poll_results_updated` message to every WebSocket client watching that poll. React will apply the result without refreshing the page.

## Features

- Account signup and login with bcrypt password hashing and JWT access tokens
- Protected poll creation, management, and deletion
- Public shareable poll URLs
- Server-side vote validation and duplicate-vote prevention per browser/device
- Live result counts, percentages, and progress bars
- Responsive dashboard and public voting experience
- Configurable server-enforced timers with live poll closure events
- Owner-only PDF result exports generated from database results
- Persisted light/dark theme with system preference fallback

## Technology Stack

- Frontend: React, Vite, JavaScript, React Router, Axios, responsive CSS
- Backend: Go, Gin, REST, WebSocket, JWT, bcrypt
- Data: MongoDB for persistent application data
- Realtime: Redis Pub/Sub for cross-instance result propagation

## Folder Structure

```text
backend/
  config/ controllers/ middleware/ models/ repository/
  routes/ services/ utils/ websocket/ main.go go.mod
frontend/
  src/
    components/ context/ hooks/ pages/ services/ utils/
    App.jsx main.jsx styles.css
```

## Environment Variables

Copy `backend/.env.example` to `backend/.env` and `frontend/.env.example` to `frontend/.env`.

Backend variables:

- `PORT`: HTTP server port, normally `8080`
- `MONGODB_URI`: MongoDB connection string
- `MONGODB_DATABASE`: database name
- `REDIS_URL`: Redis connection string
- `JWT_SECRET`: long random signing secret
- `FRONTEND_URL`: allowed browser origin

Frontend variables:

- `VITE_API_URL`: backend API base URL, normally `http://localhost:8080`
- `VITE_WS_URL`: backend WebSocket base URL, normally `ws://localhost:8080`

## Local Setup

### Prerequisites

- Node.js 20 or newer
- Go 1.22 or newer
- MongoDB 7 or a MongoDB Atlas deployment
- Redis 7 or a managed Redis service

### Backend

```powershell
cd backend
go mod download
go run .
```

The backend loads `backend/.env` for local development when present. In hosted environments, provide the same values through the platform's environment configuration. Startup fails unless `MONGODB_URI` and `REDIS_URL` are present and both cloud services successfully respond to a ping.

### Frontend

```powershell
cd frontend
npm install
npm run dev
```

The Vite development server normally runs at `http://localhost:5173` and the API at `http://localhost:8080`.

### Docker

Docker mode runs the frontend, Go backend, MongoDB, and Redis containers together. It uses Docker-internal MongoDB and Redis URLs, so `backend/.env` is not required for this local Docker mode. Run from the project root:

```powershell
docker compose up --build
```

Open `http://localhost:5173`. Stop the stack with `Ctrl+C`, or use `docker compose down` from another terminal.

Docker stores MongoDB data in the named `mongodb_data` volume. Remove that volume only when you intentionally want to erase local Docker data:

```powershell
docker compose down -v
```

## API and WebSocket Documentation

The database infrastructure health endpoint is available now. It performs live pings and never returns credentials:

```text
GET /api/health
200 { "status": "ok", "mongodb": "connected", "redis": "connected" }
503 { "status": "error", "mongodb": "disconnected", "redis": "connected" }
```

Core endpoints:

```text
POST /api/auth/signup
POST /api/auth/login
POST /api/polls
GET  /api/polls
GET  /api/polls/:id
DELETE /api/polls/:id
GET  /api/public/polls/:id
POST /api/public/polls/:id/vote
GET  /api/polls/:id/results/pdf
GET  /ws/polls/:pollId
```

Votes require an `X-Voter-Key` browser/device key. The frontend generates and stores this key in the browser so the MongoDB unique `(pollId, voterKey)` index prevents obvious repeat votes from that browser. This is basic duplicate protection, not fraud prevention.


## Deployment

The frontend can be deployed to a static hosting provider, and the Go backend can be deployed as a container or managed service. Configure the frontend API/WebSocket URLs to point at the deployed backend, and provide hosted MongoDB and Redis connection URLs through server environment variables. Production deployment instructions, CORS configuration, TLS requirements, and health checks will be documented with the completed implementation.

## Security Decisions

- Passwords are hashed with bcrypt and never persisted in plain text.
- JWT signing secrets and database credentials are read only from environment variables.
- Poll management endpoints require authentication and ownership checks.
- Public endpoints expose only the poll data required for voting and results.
- All important validation is performed on the backend.

## Known Limitations

The database infrastructure is complete, but user, poll, vote, Redis Pub/Sub, and browser WebSocket business flows are not implemented yet.
