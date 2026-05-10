# Obsygnal-Pulse Architecture Overview

## System Design

Obsygnal-Pulse is built as a **modular monolith** following **hexagonal architecture** (ports and adapters) principles. This gives us the clean domain boundaries of microservices with the operational simplicity of a single deployment unit.

## Layers

```
┌─────────────────────────────────────────┐
│             React Dashboard              │
├─────────────────────────────────────────┤
│           API Gateway (Gin)              │
│  ┌──────┐ ┌───────┐ ┌──────────────┐   │
│  │Middleware│ │Handlers│ │   Routes   │   │
├──┴──────┴─┴───────┴─┴──────────────┴───┤
│         Application Services             │
│  ┌─────────┐ ┌──────────┐ ┌─────────┐  │
│  │Ingestion│ │Detection │ │Alerting │   │
│  │         │ │          │ │         │   │
│  └─────────┘ └──────────┘ └─────────┘  │
│  ┌──────────────┐ ┌──────────────────┐  │
│  │Recommendation│ │   Analytics      │  │
│  └──────────────┘ └──────────────────┘  │
├─────────────────────────────────────────┤
│            Domain Models                 │
│  metrics | detection | recommendation   │
│  jobs    | alerting  | cluster           │
├─────────────────────────────────────────┤
│         Ports (Interfaces)               │
│  MetricsRepository | CacheStore          │
│  InefficiencyRepo  | RecommendationRepo  │
│  AlertRepository   | Notifier            │
├─────────────────────────────────────────┤
│         Adapters (Implementations)       │
│  ClickHouse | Redis | Slack | Webhook    │
│  Kubernetes | Prometheus                 │
└─────────────────────────────────────────┘
```

## Key Design Decisions

1. **Hexagonal Architecture**: Domain logic has zero dependencies on infrastructure. All external systems are accessed through port interfaces.

2. **Rule-Based Detection**: Detection rules are pure functions that transform metrics into inefficiencies. This is deterministic, testable, and interpretable — essential properties for operational tooling.

3. **Event-Driven Pipeline**: Detection → Recommendation → Alerting flows through in-process callbacks. This will migrate to Redis Streams when we need horizontal scaling.

4. **ClickHouse for Analytics**: Column-oriented storage with MergeTree engines provides 10-100x query performance vs PostgreSQL for our analytical workload.

## Data Model

```
GPU Telemetry → Ingestion (validate, dedup, buffer) → ClickHouse
                                                         ↓
Detection Engine (periodic eval) ← Query recent metrics ←┘
         ↓
Inefficiency → Recommendation Engine → Recommendations
         ↓
Alerting Service → Slack / Webhook
```

## Scaling Strategy

| Component | Scaling Approach |
|-----------|-----------------|
| API Server | Horizontal (stateless) |
| Ingestion | Horizontal (partition by cluster) |
| Detection | Vertical (single evaluator per cluster) |
| ClickHouse | Distributed tables + sharding |
| Redis | Sentinel / Cluster mode |
