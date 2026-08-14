# Domain Layer (`internal/domain/`)

## Overview
The `domain` package is the core of Obsygnal-Pulse. It defines all business models, entities, enums, and domain validation methods.

---

## Strict Isolation Rules

1. **Zero External Dependencies**: The `domain` package imports only standard library packages (`time`, `strings`, etc.) and `internal/pkg/errors`. It never imports databases, HTTP routers, or third-party frameworks.
2. **Deterministic Domain Logic**: All entity methods (e.g. `GPUMetric.IsIdle()`, `GPUMetric.IsOverheating()`) are pure, deterministic functions.

---

## Domain Subpackages

| Subpackage | Models & Entities | Key Responsibilities |
| :--- | :--- | :--- |
| `metrics/` | `GPUMetric`, `NodeMetric`, `ProcessMetric` | Hardware primitives: utilization, VRAM bytes, PyTorch reserved VRAM, temperature, power, XID error codes, NVLink bandwidth, MIG IDs. |
| `detection/` | `Inefficiency`, `InefficiencyType`, `Severity` | Inefficiency entity classification across all 9 detection rules. |
| `recommendation/` | `Recommendation`, `ActionType`, `ROIModel` | Cost-optimization recommendations, hourly/monthly dollar savings calculations, and confidence scores. |
| `auth/` | `User`, `Organization`, `OrgMember`, `Role`, `APIKey`, `AuditEvent`, `Usage` | Multi-tenant identity, RBAC permission matrices, machine keys, billing metering, and SOC2 audit logs. |
| `cluster/` | `Cluster`, `Node`, `ClusterOverview` | Physical and virtual cluster topology, healthy node counts, and aggregated utilization. |
| `jobs/` | `Job`, `JobStatus`, `WorkloadType` | ML training and inference workload tracking (PyTorch, Ray, vLLM). |
| `alerting/` | `Alert`, `AlertChannel`, `AlertSeverity` | Notification dispatch routing, severity levels, and delivery tracking. |

---

## Testing Domain Models

```bash
go test -v ./internal/domain/...
```
All domain packages maintain 100% deterministic test coverage.
