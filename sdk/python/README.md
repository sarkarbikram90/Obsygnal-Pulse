# 🐍 Obsygnal-Pulse Python SDK (`obsygnal-pulse`)

The official Python client library for **Obsygnal-Pulse** — providing real-time GPU telemetry, PyTorch CUDA memory allocator profiling, and automated efficiency reporting for distributed AI workloads.

---

## 📦 Installation

```bash
pip install obsygnal-pulse
```

---

## ⚡ Quickstart

### 1. PyTorch Training Loop Auto-Profiling

Add `obsygnal.init()` at the start of your training script. The SDK spawns a lightweight, non-blocking background thread that captures GPU utilization and PyTorch CUDA allocator memory metrics (`allocated`, `reserved`, and `device name`):

```python
import obsygnal
import torch
import torch.nn as nn

# Initialize the auto-profiler (runs in background thread)
client = obsygnal.init(
    api_key="obsygnal_sec_your_collector_key",
    endpoint="http://localhost:8080",
    cluster_id="us-east-prod-h100"
)

# Standard PyTorch model training
model = nn.Linear(1024, 1024).cuda()
optimizer = torch.optim.Adam(model.parameters())

for epoch in range(10):
    for batch in dataloader:
        optimizer.zero_grad()
        loss = model(batch.cuda()).sum()
        loss.backward()
        optimizer.step()

# Gracefully stop the background reporter when done
client.stop()
```

---

## 🔧 Manual Step Telemetry

If you prefer explicit control over when telemetry is reported (e.g. per epoch or evaluation phase):

```python
from obsygnal import ObsygnalClient

client = ObsygnalClient(
    api_key="obsygnal_sec_your_collector_key",
    endpoint="http://localhost:8080",
    cluster_id="us-east-prod-h100"
)

for epoch in range(num_epochs):
    train()
    # Explicitly report step completion
    client.report_step(step_name=f"epoch_{epoch}")
```

---

## 🛡️ Telemetry Captured

When PyTorch is loaded and CUDA is available, the SDK automatically extracts:
- `torch.cuda.memory_allocated()`: Actual memory actively used by tensors.
- `torch.cuda.memory_reserved()`: Total memory held by PyTorch's caching allocator.
- `torch.cuda.get_device_name()`: GPU hardware identifier.
- Process count and active execution state.

If PyTorch is not available, it safely falls back to standard process reporting without crashing the host workload.
