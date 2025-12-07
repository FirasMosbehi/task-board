# Task‑Board 🚀

Task‑Board is a demo/full‑stack sample application built with a DevOps mindset.  
It includes:

- A frontend + backend application
  
- Full CI pipeline (build, test, and push Docker images to GHCR) via GitHub Actions
   
- Docker Compose manifests for local development:
  
  - Build-from-source images
    
  - Pre-built images pulled from GHCR for faster startup
    
- A full observability & monitoring stack (logs / metrics / traces) using:
   
  - OpenTelemetry Collector
    
  - Prometheus (metrics)
    
  - Grafana Loki (logs)
    
  - Grafana Tempo (traces)
    
  - Grafana (dashboard / visualization)  

## 🧰 Technologies  

- Frontend: TypeScript, HTML, CSS, React, Vite
  
- Backend: Go / Gin
    
- Containerization: Docker, Docker Compose
  
- CI/CD: GitHub Actions, GHCR
   
- Observability: OpenTelemetry, Prometheus, Loki, Tempo, Grafana  

## Quick start (local development)  

```bash
# Clone the repo
git clone https://github.com/youruser/task-board.git
cd task-board

# Build everything and run via Docker Compose
make rebuild-local
```

This builds the frontend, backend, and all services (including observability stack) and runs them locally.

Alternatively, to use pre-built images (faster), just:

```bash
make up
```

Then access:

The Task‑Board application (frontend/back) at its configured port (e.g. http://localhost:…)

Grafana dashboard at http://localhost:3000 (or configured port) to inspect logs, metrics, traces

## CI / CD

The project uses GitHub Actions to:

- Build frontend and backend

- Run tests

- Build Docker images

- Push Docker images to GitHub Container Registry (GHCR) on success

This way you get fully reproducible artifacts and can deploy or run them anywhere.

## Observability / Monitoring

The included stack gives you out-of-the-box logging, metrics and tracing. Thanks to OpenTelemetry and Grafana stack, you can:

- Collect application logs → Loki

- Collect metrics → Prometheus

- Collect traces → Tempo

- Visualize logs/metrics/traces via Grafana

- Use this setup as a base for building microservices with production-grade observability.

## Why this project?

Task‑Board is designed to be:

- A learning / reference project for full‑stack + DevOps workflows

- A template / starting base for new projects with containerized services + CI/CD + observability

- A demo to showcase how Docker Compose + GitHub Actions + observability can fit together in a small but realistic app

## Contributing

Feel free to submit issues or pull requests. You can:

- Extend the application (add features to frontend/backend)

- Improve observability configuration (dashboards, alerting rules, etc.)

- Add deployment manifests (Kubernetes, Terraform, etc.)

## License

MIT license
