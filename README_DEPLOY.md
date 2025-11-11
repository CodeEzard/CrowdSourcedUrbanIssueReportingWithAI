Deployment quick-start

Option A) One-service deploy (recommended): backend serves frontend

- The backend image already contains frontend/ and serves it as same-origin. This avoids CORS and cookie headaches.

Cloud (Render) using Docker

1) Create a managed Postgres on Render (or your provider) and note the Database URL.
2) Create a new Web Service → use “Deploy an existing Dockerfile”.
   - Repository: this repo
   - Dockerfile path: backend/Dockerfile
   - Build context directory: .
   - Instance: pick a plan
   - Port: 8080
3) Add Environment Variables
   - PORT=8080
   - FRONTEND_DIR=/app/frontend
   - DATABASE_DSN=postgres://USER:PASS@HOST:5432/DB?sslmode=require
   - JWT_SECRET=change_me
   - ALLOWED_ORIGIN=  (leave empty for same-origin)
   - ML_API_URL=https://urgency-api-latest.onrender.com/predict
   - IMAGE_CLASSIFICATION_API_URL=https://issue-classification-api.onrender.com/predict
4) Deploy. After healthy, open the service URL — it serves the UI and API.

One-click with render.yaml

- This repo includes render.yaml. On Render, choose “New +” → “Blueprint” and point to your repo. It will:
   - Create a free Postgres instance (urban-civic-db)
   - Create a Web Service using backend/Dockerfile with context .
   - Wire DATABASE_DSN automatically via fromDatabase
   - Generate a JWT_SECRET
   - Set ML API URLs
   - After deploy, open the web service URL.

Cloud (Railway / Fly.io / Others)

- Railway: New Project → From GitHub → Configure service as Docker build.
  - Set Dockerfile path backend/Dockerfile and context .
  - Add the same environment variables as above.
- Fly.io: fly launch --no-deploy, set builders to use Dockerfile, add secrets, fly deploy.

Option B) VM with Docker Compose

1) Copy repo to the VM.
2) Create backend/.env from backend/env.sample and set real values.
3) Run:

   docker compose up --build -d

   Notes:
   - We fixed compose to build with root context so the Dockerfile can COPY go.mod and frontend/.
   - The app will listen on http://SERVER:8080.

Verification

- Health: curl http://HOST:8080/health → ok
- UI: open http://HOST:8080/
- Auth: register/login in the UI; the backend sets an HttpOnly cookie for authenticated actions (comments/upvotes/report).

Production notes

- Secrets: use provider secret management for JWT_SECRET and DB URL.
- TLS: prefer your cloud’s automatic TLS. If self-hosting, terminate TLS with a reverse proxy (nginx/Caddy) in front of the service.
- Database: use managed Postgres; for Compose, the included db service is for dev only.
- Redis (optional): set REDIS_ADDR/REDIS_PASSWORD to enable token revocation.
- CORS: if you later host the frontend separately, set ALLOWED_ORIGIN to that origin and ensure client requests send credentials when needed.
