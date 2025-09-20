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

networks:
  vaultwarden_default:
    external: true

volumes:
  cc-data:
```

## Key Changes Made

1. **Volume Mount**: Changed from `cc-data:/backend` to `cc-data:/app/backend` to match the container's working directory structure
2. **Data Persistence**: The volume ensures that `data.json` file persists between container restarts

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