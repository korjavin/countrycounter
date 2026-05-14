# Deployment Guide

## Current Portainer Stack Configuration

Update your Portainer stack with the following configuration:

```yaml
version: "3.8"

services:
  tgwebapp:
    image: ghcr.io/korjavin/countrycounter:latest
    container_name: countrycounter
    volumes:
      - cc-data:/app/backend
    networks:
      - vaultwarden_default
    labels:
      - "traefik.enable=true"
      - "traefik.http.routers.countrycounter.rule=Host(`your-domain.com`)"
      - "traefik.http.routers.countrycounter.entrypoints=websecure"
      - "traefik.http.routers.countrycounter.tls.certresolver=myresolver"
      - "io.containers.autoupdate=image"
    restart: unless-stopped
    environment:
      - TELEGRAM_BOT_TOKEN=${TELEGRAM_BOT_TOKEN}
      - DB_PATH=/app/backend/data.db

networks:
  vaultwarden_default:
    external: true

volumes:
  cc-data:
```

## Key Changes Made

1. **Volume Mount**: Uses `cc-data:/app/backend` to match the container's working directory structure
2. **Data Persistence**: The volume holds the SQLite database file (`data.db`) and, during the upgrade window, the legacy `data.json` so it can be auto-imported on first start
3. **Storage Backend**: Switched from a JSON file to SQLite with `pressly/goose` migrations; `DB_PATH` controls the database location (default `/app/backend/data.db`)

## Upgrading from a JSON-storage release

Older releases persisted state in `data.json`. The new release uses SQLite and will auto-import `data.json` exactly once on first start.

1. **Keep the existing volume mounted at `/app/backend`** — both the new `data.db` and the legacy `data.json` need to live in the same directory for the auto-import to find the file.
2. **Pull and start the new image.** On the first start with an empty DB and a present `data.json`, the backend will log `Auto-imported N rows from data.json` and create `data.db` alongside the JSON file. Subsequent starts will log a skip message and ignore the JSON.
3. **Verify a known user**: open the web UI or send `/list` in the Telegram bot and confirm the country list matches what was in `data.json`.
4. **Confidence window**: keep the legacy `data.json` in the volume for one release cycle in case you need to roll back.
5. **Cleanup (after the confidence window)**: delete `data.json` from the volume. The backend will continue to operate from `data.db` only — the auto-import step is a no-op once the DB is non-empty.

If the auto-import fails (corrupt JSON, mid-import DB error), the backend exits at startup rather than running with partial state. Inspect the logs, restore a known-good `data.json`, and restart.

## Environment Variables (Optional Enhancement)

To make the configuration more flexible, you can use environment variables:

1. Set these variables in Portainer's environment section:
   - `DOMAIN=your-domain.com`
   - `NETWORK_NAME=vaultwarden_default`
   - `TELEGRAM_BOT_TOKEN=your_token_here`

2. Use this enhanced stack configuration:

```yaml
version: "3.8"

services:
  tgwebapp:
    image: ghcr.io/korjavin/countrycounter:latest
    container_name: countrycounter
    volumes:
      - cc-data:/app/backend
    networks:
      - ${NETWORK_NAME:-vaultwarden_default}
    labels:
      - "traefik.enable=true"
      - "traefik.http.routers.countrycounter.rule=Host(`${DOMAIN}`)"
      - "traefik.http.routers.countrycounter.entrypoints=websecure"
      - "traefik.http.routers.countrycounter.tls.certresolver=myresolver"
      - "io.containers.autoupdate=image"
    restart: unless-stopped
    environment:
      - TELEGRAM_BOT_TOKEN=${TELEGRAM_BOT_TOKEN}
      - DB_PATH=/app/backend/data.db

networks:
  vaultwarden_default:
    external: true

volumes:
  cc-data:
```

## Automated Deployment (Optional)

The repository now includes automated CI/CD:

1. **GitHub Actions**: Automatically builds and pushes Docker images on every push to master
2. **Webhook Integration**: Can trigger Portainer redeployment via webhook (requires `PORTAINER_REDEPLOY_HOOK` secret)

To enable automated redeployment:
1. Go to your GitHub repository Settings → Secrets and variables → Actions
2. Add `PORTAINER_REDEPLOY_HOOK` secret with your Portainer webhook URL
3. Each push to master will automatically trigger a new deployment

## Manual Deployment

To manually deploy a new version:
1. Pull the latest image: `docker pull ghcr.io/korjavin/countrycounter:latest`
2. Restart the stack in Portainer