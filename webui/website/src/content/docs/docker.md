---
date: '2026-03-06T22:13:54-05:00'
draft: false
title: 'Docker'
description: 'Deploy Hister with Docker Compose, persistent storage, external access, and a reverse proxy.'
---

## Docker Setup

Hister provides official Docker images for both AMD64 and ARM64 architectures.

> **Note on Permissions**: The `latest` image runs as a **non-root user** (UID/GID 1000) by default for better security. Ensure the mounted volume (e.g., `./data`) has the correct permissions. If you need to run as root, use the `ghcr.io/asciimoo/hister:latest-root` image.

### Configuring Hister in a Container

Hister can be fully [configured using environment variables](configuration#environment-variables).
This is the **recommended approach for containerized environments** (Docker, Kubernetes, etc.) as it avoids the need to manage configuration files inside the container or mounted volumes.

If you prefer using a configuration file instead of environment variables, you can generate a default one using Docker:

```bash
docker run --rm ghcr.io/asciimoo/hister:latest create-config > config.yml
```

### Basic Docker Compose

For a simple local setup:

```yaml
services:
  hister:
    image: ghcr.io/asciimoo/hister:latest
    container_name: hister
    user: '1000:1000'
    restart: unless-stopped
    volumes:
      - ./data:/hister/data
    ports:
      - 4433:4433
```

### Docker Compose with External Access

To make Hister accessible from other devices, use the `environment` section in your `compose.yml`:

```yaml
services:
  hister:
    image: ghcr.io/asciimoo/hister:latest
    container_name: hister
    user: '1000:1000'
    restart: unless-stopped
    environment:
      - HISTER__SERVER__ADDRESS=0.0.0.0:4433
      - HISTER__SERVER__BASE_URL=http://192.168.1.100:4433 # Use your actual IP/hostname
    volumes:
      - ./data:/hister/data
    ports:
      - 4433:4433
```

### Docker Compose Behind Reverse Proxy

When running behind a reverse proxy, set the `base_url` to your public domain:

```yaml
services:
  hister:
    image: ghcr.io/asciimoo/hister:latest
    container_name: hister
    user: '1000:1000'
    restart: unless-stopped
    environment:
      - HISTER__SERVER__ADDRESS=0.0.0.0:4433
      - HISTER__SERVER__BASE_URL=https://hister.example.com # Your public URL
    volumes:
      - ./data:/hister/data
    ports:
      - 4433:4433
```
