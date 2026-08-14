# Infrastructure & Deployment (`deploy/`)

## Overview
The `deploy/` directory houses all Infrastructure-as-Code (IaC), containerization definitions, Helm charts, and orchestration manifests necessary to provision, deploy, and scale Obsygnal-Pulse across cloud environments (GKE, EKS, AKS) and bare-metal GPU clusters.

---

## Directory Structure

```
deploy/
├── helm/                     # Production-grade Helm charts
│   └── obsygnal-pulse/       # Main chart with API Server, Worker, Agent DaemonSet, Web UI
│       ├── Chart.yaml
│       ├── values.yaml
│       └── templates/
│           ├── deployment-api.yaml
│           ├── deployment-worker.yaml
│           ├── deployment-web.yaml
│           ├── daemonset.yaml        # GPU node collector with NVIDIA host mounts
│           ├── ingress.yaml
│           └── secrets.yaml
├── k8s/                      # Kustomize base and environment overlays
│   ├── base/
│   └── overlays/
│       ├── staging/
│       └── production/
├── terraform/                # Terraform cloud provisioning modules
│   └── gke/                  # Google Kubernetes Engine with NVIDIA GPU node pools
└── README.md
```

---

## Key Deployment Targets

### 1. Production Helm Chart (`deploy/helm/obsygnal-pulse`)
The official Helm chart installs all Obsygnal-Pulse components onto Kubernetes:
- **API Control Plane**: Stateless Deployment with Horizontal Pod Autoscaling (HPA) and liveness/readiness probes.
- **Background Worker**: Asynchronous detection and recommendation worker.
- **React Web UI**: Static frontend served via hardened, non-root Nginx container.
- **Edge Telemetry DaemonSet**: Scheduled exclusively on GPU-accelerated nodes (`accelerator: nvidia-gpu` nodeSelector) with host access to `/dev/nvidia*`, `/sys`, and `/proc`.

#### Installation:
```bash
helm upgrade --install obsygnal-pulse deploy/helm/obsygnal-pulse \
  --namespace obsygnal-system \
  --create-namespace \
  --set global.environment="production" \
  --set secrets.jwtSecret="YOUR_PRODUCTION_JWT_SECRET_32_BYTES" \
  --set secrets.clickhousePassword="YOUR_CLICKHOUSE_PASSWORD" \
  --set config.clickhouseHost="clickhouse.obsygnal-system.svc.cluster.local" \
  --set config.redisHost="redis.obsygnal-system.svc.cluster.local"
```

### 2. Kustomize Overlays (`deploy/k8s/`)
For GitOps pipelines (ArgoCD, Flux) without Helm:
- **Staging**: `kubectl apply -k deploy/k8s/overlays/staging`
- **Production**: `kubectl apply -k deploy/k8s/overlays/production`

### 3. Terraform Cloud Modules (`deploy/terraform/`)
- **GKE Module (`deploy/terraform/gke/`)**: Automates creation of a VPC, GKE Standard cluster, and an autoscaling NVIDIA H100 / A100 GPU node pool with NVIDIA GPU Operator drivers pre-installed.

### 4. Local Development Stack (`docker-compose.yml`)
Spins up ClickHouse, Redis, NATS JetStream, Prometheus, and Grafana:
```bash
make docker-up
```

---

## Container Security & Hardening

All container images are built using multi-stage builds in the root [Dockerfile](../Dockerfile):
- **Minimal Attack Surface**: Uses Alpine / Distroless runtime images.
- **Non-Root Execution**: API Server (`UID 1000`), Worker (`UID 1000`), and Web UI (`UID 101`) execute as unprivileged users.
- **Automated Vulnerability Scanning**: Trivy scans images on every commit in the CI/CD pipeline.
- **Zero Hardcoded Secrets**: All sensitive tokens and secrets are injected exclusively via Kubernetes Secrets or environment variables.
