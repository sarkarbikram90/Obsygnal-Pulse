# Infrastructure & Deployment (`deploy/`)

## Overview
The `deploy/` directory houses all Infrastructure-as-Code (IaC), containerization configurations, and orchestration manifests necessary to provision, deploy, and scale Obsygnal-Pulse in a production environment.

## Responsibilities
- Defining repeatable, deterministic environments.
- Packaging Go binaries and React static assets into lightweight, secure Docker images.
- Providing local development orchestration via Docker Compose.
- Defining Kubernetes manifests for production scheduling, auto-scaling, and networking.

## Architecture Role
This folder does not contain application code. It defines the physical (or virtual) topology in which the application runs. It orchestrates the API Server, Workers, Redis, and ClickHouse.

## Key Components
- `docker/`: Contains optimized, multi-stage `Dockerfiles` for building lightweight Alpine-based container images for all services.
- `docker-compose.yml`: The local development stack, spinning up ClickHouse, Redis, the API, the Worker, and the Simulator via a single command (`make docker-up`).
- `k8s/`: (WIP) Kubernetes manifests including Deployments, Services, ConfigMaps, and Secrets.
- `terraform/`: (WIP) HCL files for provisioning managed cloud infrastructure (e.g., AWS ElastiCache for Redis, managed ClickHouse, EKS clusters).

## Dependencies
- Requires Docker, Docker Compose, Kubernetes (kubectl), and Terraform.

## Operational Considerations
- **Resource Limits**: Production deployments must strictly define CPU and Memory requests/limits to prevent noisy neighbor issues on shared clusters.
- **Scaling**: 
  - API servers should use Horizontal Pod Autoscalers (HPA) scaling on CPU or HTTP request latency.
  - Workers should use KEDA (Kubernetes Event-driven Autoscaling) scaled based on Redis queue depth.
- **State Management**: While stateless API/Worker pods can be freely killed, ClickHouse and Redis are stateful. Production deployments should utilize managed services rather than raw StatefulSets where possible.

## Security Considerations
- **Images**: Dockerfiles must use minimal distroless or Alpine bases. Run all containers as non-root users.
- **Secrets**: Passwords and JWT secrets must be injected into the container environment via Kubernetes Secrets or external secret managers (e.g., HashiCorp Vault), never baked into images.

## Testing
- **Local Validation**: Infrastructure configs are validated locally using `make docker-up` and running E2E integration tests against the resulting stack.

## Future Improvements
- Add Helm charts (`deploy/helm/`) to parameterize Kubernetes deployments across `dev`, `staging`, and `prod` environments.
- Integrate container vulnerability scanning (e.g., Trivy) into the CI pipeline.
