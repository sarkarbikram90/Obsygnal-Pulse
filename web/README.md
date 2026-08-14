# Frontend Interface (`web/`)

## Overview
The `web/` directory contains the Single Page Application (SPA) dashboard for **Obsygnal-Pulse**. Built with **React 19**, **TypeScript**, and **TailwindCSS v4**, it provides AI infrastructure teams with real-time GPU cluster visibility, bottleneck triage, ROI cost calculations, and 1-click remediation workflows.

---

## Tech Stack & Dependencies

| Layer | Technology | Purpose |
| :--- | :--- | :--- |
| **Framework** | React 19.x & React DOM | Declarative component UI |
| **Language** | TypeScript 5.8+ | Strict type safety across API models and UI props |
| **Build Tool** | Vite 6.x | Lightning-fast HMR and optimized production bundling |
| **State & Data Fetching** | TanStack Query v5 (`@tanstack/react-query`) | Server state caching, background polling, and optimistic updates |
| **Styling** | TailwindCSS v4 (`@tailwindcss/vite`) | Modern, utility-first CSS design system in dark theme |
| **Icons** | Lucide React | Clean, consistent developer UI iconography |
| **Charts & Heatmaps** | Recharts & Custom Canvas | Real-time GPU telemetry time-series and node heatmaps |
| **Routing** | React Router v7 | Nested route layout and protected view navigation |

---

## Dashboard Feature Modules

```
src/
├── components/           # Reusable UI primitives & layout shells
│   ├── Layout.tsx        # App header, sidebar navigation, tenant selector, waste ticker
│   ├── ClusterOverview.tsx # Cluster health score, active GPUs, total compute
│   ├── GPUHeatmap.tsx    # Live tile map showing temperature/utilization per GPU
│   ├── InefficiencyPanel.tsx # Detected bottlenecks with severity badges
│   ├── RecommendationPanel.tsx # Actionable ROI savings & 1-click auto-fix modal
│   ├── MetricsChart.tsx  # Time-series graphs for compute, VRAM, and power
│   ├── APIKeyManager.tsx # Machine agent API key generation & rotation
│   └── AuthModal.tsx     # Tenant login and registration modal
├── features/             # Route-level views
│   ├── cluster/Dashboard.tsx     # Main executive cluster monitoring hub
│   ├── cluster/Inefficiencies.tsx# Filterable bottleneck inspection view
│   ├── cluster/Nodes.tsx         # Node-level hardware inventory & status
│   ├── recommendations/          # Optimization task list with patch generator
│   ├── executive/FleetSimulator.tsx # Interactive GPU fleet cost & ROI calculator
│   ├── billing/BillingView.tsx   # Monitored GPU-hours & tier usage metering
│   ├── alerts/Alerts.tsx         # Incident dispatch history and channel status
│   └── jobs/Jobs.tsx             # Active ML training workload tracker
├── context/              # React Auth and Workspace Context
│   └── AuthContext.tsx   # JWT storage, login/signup/logout, and tenant scoping
└── services/api.ts       # Type-safe HTTP fetch client connecting to Go API Server
```

---

## Getting Started Locally

### 1. Install Dependencies
```bash
cd web
npm install
```

### 2. Run Local Development Server (Vite)
```bash
npm run dev
```
The dashboard runs at [http://localhost:5173](http://localhost:5173). Requests to `/api/v1` are proxied to the Go API Server at `http://localhost:8080`.

### 3. Typecheck & Production Build
```bash
# Typecheck TypeScript files
npx tsc --noEmit

# Compile production bundle to dist/
npm run build
```

---

## Production Deployment

The production build generates optimized static HTML/CSS/JS in `web/dist/`. In production, this bundle is served by an unprivileged Nginx container (see `Dockerfile` `target: web`) or deployed to any edge CDN (Cloudflare Pages, Vercel, S3/CloudFront).
