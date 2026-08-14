# Security Policy

The Obsygnal-Pulse team takes the security of our software, user data, and infrastructure extremely seriously. We appreciate the efforts of security researchers and developers in responsibly disclosing vulnerabilities to us.

---

## Supported Versions

Security updates and critical patches are actively provided for the following versions:

| Version | Supported | Notes |
| :--- | :---: | :--- |
| `1.x` / `main` | :white_check_mark: | Active development & current release |
| `< 1.0.0` | :x: | Pre-release / unsupported |

---

## Reporting a Vulnerability

If you discover a security vulnerability in Obsygnal-Pulse, **please do not disclose it publicly or open a public GitHub issue.**

### How to Report Privately:
1. **Email**: Send detailed vulnerability reports to [security@obsygnal.io](mailto:security@obsygnal.io).
2. **GitHub Private Advisory**: You can also report vulnerabilities via [GitHub Security Advisories](https://github.com/sarkarbikram90/Obsygnal-Pulse/security/advisories/new).

### What to Include in Your Report:
To help us triage and resolve the issue quickly, please provide:
- A descriptive summary of the vulnerability.
- Affected component(s) (`api-server`, `collector`, `worker`, `auth`, `web`, Helm chart, etc.).
- Step-by-step reproduction steps or a minimal proof-of-concept (PoC).
- Potential impact and severity assessment (e.g., privilege escalation, denial-of-service, data exposure).
- Suggested remediation or mitigation, if known.

---

## Response & Disclosure SLA

We adhere to a coordinated vulnerability disclosure process:

- **Initial Acknowledgment**: Within **24–48 hours** of report receipt.
- **Triage & Validation**: Within **7 business days** with severity confirmation.
- **Remediation & Patch**: Critical severity issues will be patched and released within **14–30 business days**.
- **Public Disclosure**: Coordinated after a patched release is published and users have had adequate time to upgrade.

---

## Security Best Practices for Deployments

When deploying Obsygnal-Pulse in production environments:
1. **Never commit secrets**: Set `JWT_SECRET`, database passwords, and API keys via environment variables or Kubernetes Secrets.
2. **Enforce Non-Development Mode**: Ensure `ENVIRONMENT=production` in production so that strict JWT validation and CORS controls are active.
3. **Use TLS Everywhere**: Encrypt all traffic between edge collector daemons, the control plane, ClickHouse, and Redis.
4. **Least-Privilege RBAC**: Assign granular roles (`viewer`, `operator`, `admin`, `billing`, `owner`) to users and rotate collector API keys periodically.
