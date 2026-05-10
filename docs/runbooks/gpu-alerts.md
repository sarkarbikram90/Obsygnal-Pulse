# Runbook: High GPU Idle Rate

## Alert
**Type:** `idle_gpu`  
**Severity:** High  
**Trigger:** GPU utilization < 2% with 0 running processes for > 10 minutes

## Symptoms
- Dashboard shows red heatmap cells
- Idle GPU count increasing in cluster overview
- Estimated waste increasing

## Investigation Steps

### 1. Verify the Alert
```bash
# Check current GPU status via ClickHouse
clickhouse-client --query "
  SELECT gpu_id, avg(utilization), max(running_processes)
  FROM gpu_metrics
  WHERE timestamp > now() - INTERVAL 30 MINUTE
  GROUP BY gpu_id
  HAVING avg(utilization) < 5
  ORDER BY avg(utilization)
"
```

### 2. Check Job Scheduler
```bash
# Are there jobs queued that should be using these GPUs?
kubectl get pods -n ml-workloads -l gpu=required --field-selector=status.phase=Pending
```

### 3. Check Node Health
```bash
# Verify the node is healthy
kubectl describe node <node-name>
kubectl top node <node-name>
```

### 4. Check DCGM Exporter
```bash
# Verify metrics are being reported correctly
curl http://<node-ip>:9400/metrics | grep dcgm_gpu_utilization
```

## Resolution

| Root Cause | Action |
|-----------|--------|
| No jobs scheduled | Release GPU allocation or submit workload |
| Job completed | Clean up lingering GPU reservation |
| Node drain in progress | Wait for drain to complete |
| DCGM reporting 0 incorrectly | Restart dcgm-exporter pod |
| GPU hardware fault | Run `nvidia-smi` diagnostics, may need node replacement |

## Escalation
- **L1:** Platform team — reallocate idle GPUs
- **L2:** ML Ops — investigate job scheduler
- **L3:** Infrastructure — hardware diagnostics

---

# Runbook: Stalled Training Job

## Alert
**Type:** `stalled_training`  
**Severity:** Critical  
**Trigger:** GPU compute < 5% but memory > 30% and processes > 0 for > 15 minutes

## Symptoms
- GPU memory allocated but compute near zero
- Training job making no progress (loss not decreasing)
- Process running but consuming no GPU cycles

## Investigation Steps

### 1. Identify the Job
```bash
clickhouse-client --query "
  SELECT gpu_id, job_id, avg(utilization), avg(memory_utilization), max(running_processes)
  FROM gpu_metrics
  WHERE timestamp > now() - INTERVAL 30 MINUTE
    AND utilization < 5
    AND memory_utilization > 30
  GROUP BY gpu_id, job_id
"
```

### 2. Check Job Logs
```bash
kubectl logs <pod-name> -n ml-workloads --tail=100
```

### 3. Common Causes
- **Deadlock in distributed training** — Check NCCL logs for timeout errors
- **Hung I/O** — Check if the pod is waiting on NFS/S3/GCS
- **Infinite loop in preprocessing** — Check CPU usage on the pod
- **OOM killed subprocess** — Check `dmesg` on the node

## Resolution
1. Kill the stalled job if confirmed hung
2. Restart with checkpoint from last saved state
3. If distributed training deadlock, reduce world size and retry
4. File postmortem for root cause analysis
