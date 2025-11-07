Deployment checklist and quick-start

1) Build and run locally with Docker Compose

- Ensure `backend/.env` contains your DB and JWT settings (do NOT commit secrets).
- Start the stack:

  docker compose up --build

  This will:
  - start Postgres
  - build and start the backend (reads env from `backend/.env`)

2) Verify

  - Health: `curl http://localhost:8080/health` → `ok`
  - Register/Login/Report: use the provided e2e script `backend/scripts/e2e_test.sh`

3) Production notes

- Use a real secrets store for `JWT_SECRET` (do not keep in repo).
- Add TLS / reverse proxy (nginx) in front of the backend for TLS termination.
- Use an external managed Postgres for production and set `DB_*` envs accordingly.
- If you need token revocation, enable Redis and set `REDIS_ADDR`/`REDIS_PASSWORD` in `backend/.env`. The e2e script will validate revocation when Redis is configured.
