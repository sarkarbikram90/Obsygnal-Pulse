# Frontend Interface (`web/`)

## Overview
The `web/` directory contains the Single Page Application (SPA) that serves as the user-facing dashboard for Obsygnal-Pulse. It visualizes GPU telemetry, highlights detected inefficiencies, and presents actionable optimization recommendations.

## Responsibilities
- Rendering real-time dashboards for cluster health and GPU utilization.
- Providing an interface for users to review, apply, or dismiss infrastructure optimization recommendations.
- Managing client-side state, routing, and API communication.

## Architecture Role
The React application acts as the primary client interacting with the backend REST API (`api-server`). It is entirely decoupled from the backend logic and operates strictly over HTTP JSON payloads.

## Key Components
- `src/components/`: Reusable, atomic UI components (Buttons, Cards, Modals).
- `src/features/`: Domain-specific components (e.g., `RecommendationsList`, `ClusterMap`).
- `src/hooks/`: Custom React hooks for data fetching (e.g., `useMetrics`, `useRecommendations`).
- `src/services/`: API client wrappers orchestrating `fetch` calls to the Go backend.
- `src/pages/`: Top-level route components mapped to specific URLs.

## Dependencies
- **Core Framework**: React 19, TypeScript.
- **Build Tool**: Vite (for lightning-fast HMR and optimized production bundles).
- **Styling**: TailwindCSS v4 (Utility-first CSS).
- **Icons**: Lucide React.
- **Charts**: Recharts (for rendering time-series metric graphs).

## Data Flow
```text
React Component → Custom Hook → API Client Wrapper → HTTP Fetch → Go API Server
```

## Operational Considerations
- **Deployment**: The output of `npm run build` is a collection of static HTML, CSS, and JS files. These should be served via an edge CDN (e.g., Cloudflare Pages, Vercel, AWS CloudFront) or a lightweight web server (e.g., Nginx) rather than being served by the Go backend directly.
- **Performance**: Vite automatically implements code-splitting and asset minification. Ensure images and heavy libraries are lazy-loaded.

## Security Considerations
- **CORS**: The Go backend must be configured to allow Cross-Origin Resource Sharing from the domain hosting this React app.
- **Authentication**: JWT tokens retrieved upon login must be securely stored. In an SPA, `localStorage` is vulnerable to XSS; consider storing JWTs in HttpOnly cookies if the API architecture permits.
- **XSS**: React natively escapes variables to prevent Cross-Site Scripting, but `dangerouslySetInnerHTML` must never be used with user-provided data.

## Testing
- **Unit Tests**: React Testing Library and Vitest should be used for testing individual components and hooks.
- **Expected Coverage**: Focus on complex state logic and API integration points rather than purely presentational components.

## Future Improvements
- Integrate React Query or SWR for more robust data fetching, caching, and background polling.
- Implement WebSocket or Server-Sent Events (SSE) connections for truly real-time metric updates without polling.
