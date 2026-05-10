# Obsygnal-Pulse API Reference

## Base URL

```
http://localhost:8080/api/v1
```

## Authentication

MVP uses no authentication in development mode. Production requires JWT bearer tokens.

```
Authorization: Bearer <token>
```

---

## Endpoints

### Health & Status

#### `GET /health`
Liveness probe. Returns 200 if the service is running.

**Response:**
```json
{
  "status": "healthy",
  "checks": [
    {"name": "clickhouse", "status": "healthy", "latency_ms": 2},
    {"name": "redis", "status": "healthy", "latency_ms": 1}
  ],
  "version": "0.1.0",
  "uptime_seconds": 3600
}
```

#### `GET /ready`
Readiness probe. Returns 200 only when all dependencies are connected.

---

### Metrics

#### `GET /cluster/overview`
Returns aggregated cluster statistics.

**Query Parameters:**
| Parameter | Type | Required | Default |
|-----------|------|----------|---------|
| cluster_id | string | No | first cluster |

**Response:**
```json
{
  "cluster_id": "sim-cluster-01",
  "total_nodes": 4,
  "healthy_nodes": 4,
  "total_gpus": 16,
  "active_gpus": 12,
  "idle_gpus": 4,
  "avg_utilization": 62.5,
  "avg_memory_usage": 55.3,
  "active_jobs": 8,
  "inefficiency_score": 35,
  "estimated_waste_usd": 12.50,
  "timestamp": "2026-05-09T23:00:00Z"
}
```

#### `GET /gpus/metrics`
Returns GPU metric time series.

**Query Parameters:**
| Parameter | Type | Required |
|-----------|------|----------|
| cluster_id | string | No |
| node_id | string | No |
| gpu_id | string | No |
| from | RFC3339 | No |
| to | RFC3339 | No |
| limit | int | No (default: 100) |

#### `POST /metrics/ingest`
Ingest a batch of GPU and node metrics.

**Request Body:**
```json
{
  "gpu_metrics": [
    {
      "cluster_id": "cluster-01",
      "node_id": "node-01",
      "gpu_id": "gpu-001",
      "timestamp": "2026-05-09T23:00:00Z",
      "utilization": 85.5,
      "memory_utilization": 70.0,
      "temperature": 72.0
    }
  ],
  "source": "dcgm-exporter"
}
```

#### `GET /ingestion/stats`
Returns ingestion pipeline statistics.

**Response:**
```json
{
  "total_received": 150000,
  "total_inserted": 148500,
  "total_dropped": 500,
  "total_duplicates": 1000,
  "buffer_gpu_size": 250,
  "buffer_node_size": 50
}
```

---

### Detection

#### `GET /inefficiencies`
List active inefficiencies.

**Query Parameters:**
| Parameter | Type | Default |
|-----------|------|---------|
| cluster_id | string | — |
| limit | int | 50 |
| offset | int | 0 |

#### `GET /inefficiencies/:id`
Get a specific inefficiency by ID.

---

### Recommendations

#### `GET /recommendations`
List active optimization recommendations.

#### `POST /recommendations/:id/dismiss`
Dismiss a recommendation with a reason.

**Request Body:**
```json
{
  "dismissed_by": "user@example.com",
  "reason": "Already addressed"
}
```

#### `POST /recommendations/:id/apply`
Mark a recommendation as applied.

---

### Alerts

#### `GET /alerts/recent`
List recent alerts with delivery status.

**Query Parameters:**
| Parameter | Type | Default |
|-----------|------|---------|
| cluster_id | string | — |
| limit | int | 20 |

---

## Error Responses

All errors follow this format:

```json
{
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "cluster_id is required",
    "request_id": "abc-123-def"
  }
}
```

| HTTP Code | Error Code | Description |
|-----------|-----------|-------------|
| 400 | VALIDATION_ERROR | Invalid request parameters |
| 401 | UNAUTHORIZED | Missing or invalid JWT token |
| 404 | NOT_FOUND | Resource not found |
| 429 | RATE_LIMITED | Too many requests |
| 500 | INTERNAL_ERROR | Server error |
