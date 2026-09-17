# Cheapest Cloud Deployment for LiftIQ

## Context

LiftIQ currently runs only via `make demo-up` on a laptop (7 processes + Docker
TimescaleDB + local Ollama). The goal is a **single always-on cloud demo** at the
lowest possible cost. Two decisions are locked in from discussion:

- **Hosting:** one small VPS (Hetzner CX22, ~€4/mo) running `docker compose`.
  Self-host TimescaleDB in a container — free managed Postgres (Supabase/Neon)
  ships `pgvector` but **not** the TimescaleDB extension, which the ingestor
  migration hard-requires (`create_hypertable` / `add_dimension`).
- **RAG:** keep the knowledge base but **swap Ollama for a managed, OpenAI-compatible
  LLM API** via a thin adapter. Self-hosting `llama3.2` would need an 8 GB+ box
  (4–8× the cost); a managed API is pennies-per-query with zero idle cost.

Why a plain VPS beats a scale-to-zero PaaS: the ingestor is a continuous 5-second
poller and the compliance engine only trusts telemetry inside a 5-minute stale
window, so the data services must stay warm 24/7. Sleeping tiers give no savings.

### What already exists vs. what's missing

| Piece | State |
|---|---|
| `elevator-simulator`, `telemetry-ingestor`, `compliance-engine` | Have Dockerfiles, in `deploy/docker-compose.yml` ✅ |
| `report-generator` | **No Dockerfile**, not in compose; needs headless Chromium ❌ |
| `knowledge-base` | **No Dockerfile**, not in compose; hard-wired to Ollama ❌ |
| `liftiq-web` (static SPA) | Not in compose; prod build has no `/api/*` router (vite proxy is dev-only) ❌ |
| DB image `timescaledb:latest-pg16` | **Lacks pgvector** — knowledge base can't `CREATE EXTENSION vector` ❌ |
| `deploy/env/.env.prod(.example)` | Missing entirely ❌ |

## Target architecture (one Hetzner CX22, 2 vCPU / 4 GB)

```
                Internet
                   │  :80/:443 (only public ports)
              ┌────▼─────┐   Caddy: serves web/dist + auto-HTTPS
              │  caddy   │   /api/compliance/* → compliance:8080
              └──┬───┬───┘   /api/reports/*    → report:8082
                 │   │       /api/knowledge/*   → knowledge:8085
   ┌─────────────┘   └──────────────┬───────────────┐
   ▼            ▼            ▼        ▼               ▼
compliance   report    knowledge  simulator      (internal docker net)
 :8080        :8082      :8085       :8000
   │            (chrome)   │  └── OpenAI-compatible LLM API (external, HTTPS)
   │                       │
   └──────────┬────────────┴──────────── ingestor (poller, no port)
              ▼
        timescaledb-ha (pgvector + timescaledb)  ──vol── timescaledb_data
```

Footprint fits 4 GB comfortably: DB ~1 GB, report/Chrome ~512 MB peak, everything
else <128 MB each, Caddy ~32 MB.

## Cost

| Item | Monthly |
|---|---|
| Hetzner CX22 (2 vCPU / 4 GB / 40 GB) | ~€3.79–4.50 |
| Domain (for HTTPS) — or free DuckDNS/sslip.io | ~$1 (or $0) |
| LLM API (gpt-4o-mini + text-embedding-3-small) | cents — seed is one-time; queries are pennies |
| **Total** | **~$5/month** |

## Work items

### 1. Managed-LLM adapter (knowledge-base) — the RAG swap

Ollama is cleanly abstracted behind two interfaces, so this is additive:
`embed.Embedder` (`knowledge-base/internal/embed/embedder.go`) and
`llm.Generator` (`knowledge-base/internal/llm/generator.go`).

- **New** `knowledge-base/internal/embed/openai.go` — implements `Embedder` via
  `POST {baseURL}/v1/embeddings` (`{model, input, dimensions: 768}`). Requesting
  **768 dims** keeps the existing `vector(768)` pgvector column
  (`knowledge-base/internal/store/pgx.go:33`) unchanged — no schema migration.
- **New** `knowledge-base/internal/llm/openai.go` — implements `Generator` via
  `POST {baseURL}/v1/chat/completions`, mirroring the prompt/context assembly in
  the existing `internal/llm/ollama.go`.
- **Edit** `knowledge-base/internal/config/config.go` — add `LLM_PROVIDER`
  (`ollama` default | `openai`), `OPENAI_BASE_URL` (default
  `https://api.openai.com`), `OPENAI_API_KEY`, and embed/chat model names.
- **Edit wiring only** in `knowledge-base/cmd/knowledged/main.go` (~lines 53–54)
  and `knowledge-base/cmd/ingest/main.go` (~line 103) to select the impl by
  `LLM_PROVIDER`. Keep Ollama as the local-dev default so `make demo-up` is
  unaffected.

> Endpoint is any OpenAI-compatible host — OpenAI, Together, DeepInfra, Groq,
> OpenRouter. (Anthropic-native alternative: Claude for generation + Voyage for
> embeddings; skipped here because it's two providers and no OpenAI-compatible
> embeddings endpoint — more moving parts than the cheapest path needs.)

### 2. Two new Dockerfiles

- **New** `report-generator/Dockerfile` — multi-stage: `golang:1.25-alpine` build
  → final `alpine:3.21` with `apk add chromium nss freetype harfbuzz ca-certificates
  ttf-freefont`. Set `CHROMEDP` to the system Chromium path (chromedp launches the
  browser it finds on PATH). Give it a **512 MB** limit (Chrome peaks during render).
- **New** `knowledge-base/Dockerfile` — multi-stage build of both `knowledged` and
  `ingest` binaries → `alpine:3.21`. Env: `DATABASE_URL`, `LLM_PROVIDER=openai`,
  `OPENAI_*`, model names. ~128 MB limit.

### 3. Web app + reverse proxy (Caddy)

The prod static build calls relative `/api/compliance`, `/api/reports`,
`/api/knowledge` — the vite proxy (`liftiq-web/vite.config.ts`) is dev/preview-only
and vanishes in a static build. Caddy replaces it and adds free TLS.

- **New** `liftiq-web/Dockerfile` — multi-stage: `node:20` runs `npm ci && npm run
  build` → `caddy:2-alpine` serving `/srv` with a `Caddyfile`.
- **New** `liftiq-web/Caddyfile` — serve SPA (with `try_files` fallback to
  `index.html`) and reverse-proxy, stripping the prefix to match today's rewrite:
  ```
  handle_path /api/compliance/*  { reverse_proxy compliance:8080 }
  handle_path /api/reports/*     { reverse_proxy report:8082 }
  handle_path /api/knowledge/*   { reverse_proxy knowledge:8085 }
  handle { try_files {path} /index.html; file_server }
  ```
  Set the site address to the domain for automatic HTTPS (or `:80` for IP-only).

### 4. Extend compose to the full stack

- **Edit** `deploy/docker-compose.yml`:
  - Switch DB image `timescale/timescaledb:latest-pg16` →
    **`timescale/timescaledb-ha:pg16`** (bundles pgvector). Use a fresh volume.
  - Add `report`, `knowledge`, and `caddy` services (build contexts
    `../report-generator`, `../knowledge-base`, `../liftiq-web`), with
    `DATABASE_URL` for knowledge, `depends_on` DB health, and health checks.
- **Edit** `deploy/docker-compose.prod.yml`:
  - Add resource limits: report `0.5cpu/512M`, knowledge `0.25cpu/128M`,
    caddy `0.25cpu/64M`.
  - **Remove the public `8000`/`8080` port maps** — only Caddy publishes `80/443`;
    everything else stays on the internal docker network.

### 5. Secrets / env

- **New** `deploy/env/.env.prod.example` and (uncommitted) `.env.prod` with:
  `DB_PASSWORD` (strong), `OPENAI_API_KEY`, `OPENAI_BASE_URL`, `LLM_PROVIDER=openai`,
  embed/chat model names, `SITE_ADDRESS` (domain) for Caddy. `.env.prod` is already
  gitignored per repo convention.

### 6. Provision the VPS

1. Create Hetzner CX22 (Ubuntu 24.04). Firewall: allow 22/80/443 only.
2. Install Docker + compose plugin; `git clone` the repo.
3. Point a DNS `A` record at the box IP (or use DuckDNS) so Caddy can issue TLS.
4. `cp deploy/env/.env.prod.example deploy/env/.env.prod` and fill secrets.
5. Bring up:
   ```
   docker compose -f deploy/docker-compose.yml -f deploy/docker-compose.prod.yml \
     --env-file deploy/env/.env.prod up --build -d
   ```
6. Seed the knowledge base once (uses the managed API now):
   ```
   docker compose ... run --rm knowledge ./ingest --seed testdata/smartrise_c4_seed.json
   ```

## Verification (end-to-end, against the live domain)

1. `curl https://<domain>/api/compliance/units` → 3 units;
   `/api/reports/health` and `/api/knowledge/health` → ok.
2. Load `https://<domain>/` → Scan screen lists ELV-001/002/003.
3. Run an inspection: open ELV-003, confirm the 20 ASME rules render with
   live pass/fail (proves simulator→ingestor→DB→compliance is flowing).
4. Generate a PDF from the summary screen (proves report-generator + Chromium).
5. Ask a knowledge query (e.g. a Smartrise fault code) → cited answer
   (proves the OpenAI-compatible embed+LLM adapter + pgvector).
6. Fault-injection smoke test: `POST /elevators/ELV-003/fault
   {"fault":"door_motor_degradation"}` on the simulator, wait ~1 min, refresh
   compliance → door-close-force rule trends toward fail. Clear the fault after.
7. Reboot the VPS → `restart: always` brings the whole stack back; data persists
   in `timescaledb_data`.

## Notes / risks

- **Chromium in Alpine** is the fiddliest step; if fonts/sandbox cause trouble,
  fall back to Debian-slim + `google-chrome-stable`, or the
  `chromedp/headless-shell` image with chromedp in remote-allocator mode.
- **Simulator stays deployed** as the data source (stand-in for real BMS). Phase 3
  would later repoint `SIMULATOR_URL` at a real BACnet gateway — no topology change.
- **Embedding dimension:** pinning the managed embeddings to 768 dims avoids any
  pgvector schema change; if a provider can't emit 768, make the `vector(N)` column
  and `Dimension()` configurable and re-seed.
- Keep `DB_PASSWORD` and `OPENAI_API_KEY` only in `.env.prod` (never committed);
  DB port is not exposed publicly.
