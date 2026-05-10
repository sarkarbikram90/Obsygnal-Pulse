# ADR-001: Use ClickHouse for Analytical Storage

**Date:** 2026-05-09  
**Status:** Accepted  
**Deciders:** Engineering Team

## Context

Obsygnal-Pulse needs to store high-volume GPU telemetry metrics (10K+ data points/second at scale) and perform analytical queries (aggregations, time-range filters, percentile calculations) efficiently.

## Decision

Use **ClickHouse** as the primary analytical storage engine instead of PostgreSQL, TimescaleDB, or InfluxDB.

## Rationale

| Factor | ClickHouse | PostgreSQL | TimescaleDB | InfluxDB |
|--------|-----------|------------|-------------|----------|
| Analytical query speed | 10-100x faster | Baseline | 3-10x | 2-5x |
| Compression ratio | 10-20x | 2-3x | 5-8x | 5-10x |
| Insert throughput | 1M+ rows/sec | 10K rows/sec | 50K rows/sec | 100K rows/sec |
| Columnar storage | Native | No | No | Partial |
| SQL compatibility | Yes (subset) | Full | Full | InfluxQL/Flux |
| Operational complexity | Medium | Low | Medium | Medium |

Key reasons:
1. **Column-oriented storage** — GPU metrics are wide (30+ columns) but queries typically touch 3-5 columns. Columnar access is dramatically faster.
2. **MergeTree engine** — Built-in partitioning, TTL, and materialized views eliminate the need for external ETL.
3. **Compression** — LZ4 + delta encoding achieves 10-20x compression on time-series data.
4. **Insert performance** — Batch inserts at 1M+ rows/second handles our projected throughput.

## Consequences

- **Positive:** Sub-second analytical queries on billions of rows, built-in TTL, native materialized views
- **Negative:** No transactions, eventual consistency on merges, smaller ecosystem than PostgreSQL
- **Mitigation:** Use Redis for transactional state; ClickHouse for read-heavy analytics only

## Alternatives Considered

- **PostgreSQL + pg_partman:** Simpler ops but 10x slower on analytical queries
- **TimescaleDB:** Good middle ground but still row-oriented under the hood
- **InfluxDB:** Purpose-built for time-series but weaker SQL and ecosystem
