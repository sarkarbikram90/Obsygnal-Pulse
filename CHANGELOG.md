# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

---

## [1.0.0] - 2026-08-14

### Added
- **Hexagonal Architecture**: Complete decoupling of core domain from infrastructure using Go ports and adapters pattern.
- **Edge Telemetry Daemon (`obsygnal-agent`)**: High-frequency metric scraper for NVIDIA DCGM Prometheus exporters capturing utilization, VRAM, temperature, power, XID error codes, MIG instances, and NVLink bandwidth.
- **Inefficiency Detection Engine**: 9 deterministic, rule-based detection algorithms:
  - `Underutilized GPU`: Low compute saturation detection.
  - `Idle GPU`: 0% compute utilization and zero running processes.
  - `Low Memory Usage`: Reserved VRAM without utilization.
  - `CPU Starvation`: Host CPU bottleneck with stalled GPU data loading.
  - `Stalled Training`: Hung PyTorch/Ray training loops with allocated VRAM and 0 compute.
  - `Abnormal Resource Spike`: Hardware thermal spikes exceeding safe operating thresholds.
  - `NVIDIA XID Hardware Error`: Hardware faults (XID 31, 43, 62, 79) requiring immediate node isolation.
  - `CUDA Memory Fragmentation`: PyTorch allocator fragmentation (reserved >> allocated).
  - `KV Cache Saturation`: vLLM PagedAttention cache exceeding 90% capacity.
- **Recommendation & ROI Engine**: Multi-cloud GPU pricing database (H200, B200, H100, A100, L40S, L4, RTX 4090, V100) calculating hourly and monthly monetary waste with confidence scores.
- **1-Click Auto-Remediation**: Kubernetes patch manifest generation, Helm commands, and webhook action triggers.
- **Multi-Tenant SaaS Control Plane**:
  - JWT Bearer token authentication with `bcrypt` password hashing.
  - Role-Based Access Control (RBAC) with 5 roles (`viewer`, `operator`, `admin`, `billing`, `owner`).
  - Machine-to-machine API key management (`obsygnal_sec_...`).
  - Usage metering tracking monitored GPU-hours and active nodes against tier limits.
  - SOC2-compliant audit logging for all infrastructure actions.
- **Zero-Dependency Fallback**: Stateless and stateful in-memory repositories allowing zero-dependency local boots and CI testing when ClickHouse/Redis are offline.
- **Data Persistence Adapters**:
  - ClickHouse OLAP adapter for high-throughput time-series telemetry.
  - Redis adapter for sliding-window rate limiting and distributed caching.
  - NATS JetStream event bus integration.
  - Prometheus metrics exporter and Slack/Generic Webhook alert dispatchers.
- **React 19 Web Dashboard**:
  - Single-page application built with Vite, TypeScript, TailwindCSS v4, TanStack Query, and Lucide icons.
  - Real-time Cluster Overview, GPU Heatmap, Metrics Charts, Inefficiencies Inspector, Recommendations Manager, and Billing Quota tracker.
  - Interactive **Executive Fleet ROI Simulator** with custom GPU pricing calculator and executive brief export.
- **Python SDK Client (`obsygnal-pulse`)**: Lightweight Python SDK with PyTorch CUDA memory profiling and background auto-telemetry reporter.
- **Cloud Native Deployment**:
  - Production Helm Chart (`deploy/helm/obsygnal-pulse`) with DaemonSet for GPU nodes.
  - Kustomize overlays for staging and production.
  - Terraform GKE provisioning module.
  - Multi-stage Dockerfile targeting `api-server`, `worker`, `collector`, `web`, and `simulator`.
  - CI/CD workflow with static binary compilation, race detection, and Trivy vulnerability scanning.
