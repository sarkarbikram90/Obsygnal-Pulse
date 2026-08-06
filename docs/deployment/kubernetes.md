# Production Deployment Guide: Managed Kubernetes (GKE / EKS / AKS)

This document provides a comprehensive guide for deploying **Obsygnal-Pulse** via container images to managed Kubernetes environments such as **Google Kubernetes Engine (GKE)**, **Amazon EKS**, **Azure Kubernetes Service (AKS)**, or on-premises OpenShift/Kubernetes clusters.

---

## 🏗️ Architecture Overview

The Kubernetes deployment comprises 5 core components:

| Component | Workload Type | Container Target | Default Port | Description |
| :--- | :--- | :--- | :--- | :--- |
| **API Server** | Deployment (2+ replicas) | `api-server` | 8080 | Core HTTP REST API control plane and health endpoints |
| **Worker** | Deployment (2+ replicas) | `worker` | N/A | Asynchronous detection rule evaluation & ROI calculations |
| **Web UI** | Deployment (2+ replicas) | `web` | 8080 | React 19 single-page dashboard served via unprivileged Nginx |
| **Collector Agent** | DaemonSet (1 per GPU node) | `collector` | N/A | Edge GPU telemetry scraper mounted to `/dev/nvidia*` & `/sys` |
| **Simulator** | Deployment (Optional 1 replica) | `simulator` | N/A | Synthetic GPU metrics generator for staging/demo clusters |

External dependencies required by the cluster:
- **ClickHouse** (OLAP metric storage & audit log lake)
- **Redis** (Sliding-window rate limiter & token cache)
- **NATS JetStream** (Real-time telemetry event bus)

---

## 📦 Container Image Build & Push

All microservice images can be built from the root `Dockerfile` using multi-stage target flags or individual Dockerfiles under `deploy/docker/`.

```bash
# Set your container registry URL
REGISTRY="ghcr.io/sarkarbikram90"
VERSION="v1.0.0"

# Build all 5 container images
docker build --target api-server -t ${REGISTRY}/obsygnal-api-server:${VERSION} .
docker build --target worker     -t ${REGISTRY}/obsygnal-worker:${VERSION} .
docker build --target collector  -t ${REGISTRY}/obsygnal-agent:${VERSION} .
docker build --target web        -t ${REGISTRY}/obsygnal-web:${VERSION} .
docker build --target simulator  -t ${REGISTRY}/obsygnal-simulator:${VERSION} .

# Push to container registry
docker push ${REGISTRY}/obsygnal-api-server:${VERSION}
docker push ${REGISTRY}/obsygnal-worker:${VERSION}
docker push ${REGISTRY}/obsygnal-agent:${VERSION}
docker push ${REGISTRY}/obsygnal-web:${VERSION}
docker push ${REGISTRY}/obsygnal-simulator:${VERSION}
```

---

## 🚀 Option 1: Deploy with Helm (Recommended)

### 1. Prerequisites & Namespace Setup
Ensure `helm` (v3.10+) and `kubectl` are connected to your Kubernetes cluster.

```bash
kubectl create namespace obsygnal-pulse
```

### 2. Configure Production Secrets
Update secrets in `deploy/helm/obsygnal-pulse/values.yaml` or pass them via `--set`:

```bash
helm upgrade --install obsygnal-pulse deploy/helm/obsygnal-pulse \
  --namespace obsygnal-pulse \
  --set global.environment=production \
  --set secrets.jwtSecret="YOUR_SECURE_JWT_SECRET_32_BYTES" \
  --set secrets.clickhouseUser="default" \
  --set secrets.clickhousePassword="YOUR_CLICKHOUSE_PASSWORD" \
  --set config.clickhouseHost="clickhouse.obsygnal-pulse.svc.cluster.local" \
  --set config.redisHost="redis.obsygnal-pulse.svc.cluster.local" \
  --set config.natsUrl="nats://nats.obsygnal-pulse.svc.cluster.local:4222" \
  --set ingress.hosts[0].host="pulse.yourdomain.com"
```

### 3. Verify Rollout
```bash
kubectl get pods,services,ingress -n obsygnal-pulse
```

---

## 🔧 Option 2: Deploy with Kustomize

If deploying without Helm, use the provided Kustomize overlays:

### Staging Deployment
```bash
kubectl apply -k deploy/k8s/overlays/staging
```

### Production Deployment
```bash
kubectl apply -k deploy/k8s/overlays/production
```

---

## 🔒 Security & GPU Node Requirements

1. **GPU DaemonSet Volume Mounts**: The edge collector agent requires host access to `/dev/nvidia*`, `/sys`, and `/proc` to capture hardware telemetry (NVLink bandwidth, MIG instances, XID errors, CUDA allocator).
2. **Node Selector / Tolerations**: Agent pods are scheduled only on GPU-accelerated nodes using `nodeSelector: accelerator: nvidia-gpu` and toleration for `nvidia.com/gpu: NoSchedule`.
3. **Non-Root Execution**: API Server (`UID 1000`), Worker (`UID 1000`), and Web UI (`UID 101`) run as non-root users with strict Kubernetes SecurityContext defaults.

---

## 📊 Health Checks & Observability

- **API Server Health**: `GET /api/v1/health` (liveness) & `GET /api/v1/ready` (readiness)
- **Web UI Health**: `GET /healthz` (Nginx status)
- **Prometheus Metrics**: Exported at `http://api-server.obsygnal-pulse.svc.cluster.local:8080/metrics`
