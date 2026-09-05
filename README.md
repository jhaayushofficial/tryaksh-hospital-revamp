# Tryaksh Hospital & Diagnostics

Tryaksh Hospital is an integrated appointment booking and administration system. It features a modern, mobile-first frontend powered by React and Vite, supported by a highly performant backend built with Go and PostgreSQL.

## Features
- 🚀 **Patient Booking Wizard:** 5-step intuitive flow to book appointments.
- 📱 **WhatsApp OTP Verification:** Passwordless authentication for patients.
- 👨‍⚕️ **Admin Dashboard:** Clinic and Doctor management.
- 📅 **Live Availability:** Accurate slot scheduling and overlapping checks.
- 🔐 **Idempotency:** Prevents double-booking via robust request keys.
- ⚡ **Go + React Stack:** Extremely fast response times and UI rendering.

## Architecture

- `backend/`: Go server using `chi` for routing, `sqlc` for database access, and `pgx` for PostgreSQL pooling.
- `frontend/`: React + TypeScript application built using Vite and TailwindCSS v4.

## Getting Started

### Prerequisites
- Go 1.21+
- Node.js v20+
- PostgreSQL 15+

### Backend Setup
1. Open a terminal in the `backend` folder.
2. Set your environment variables (or let them default):
   ```bash
   export DATABASE_URL="postgres://postgres:postgres@localhost:5432/tryaksh?sslmode=disable"
   export PORT=8080
   ```
3. Run migrations to setup your database (using your preferred schema migration tool, or executing `backend/db/schema.sql` directly).
4. Run the Go server:
   ```bash
   go run cmd/api/main.go
   ```

### Frontend Setup
1. Open a terminal in the `frontend` folder.
2. Install dependencies:
   ```bash
   npm install
   ```
3. Run the development server:
   ```bash
   npm run dev
   ```
4. Access the web app at `http://localhost:5173/`.
5. Access the admin panel at `http://localhost:5173/admin` (default backend admin password is `admin123`).

## License
Private and Confidential. All rights reserved.
