# Obsygnal-Pulse API Versioning Strategy

## Overview
Obsygnal-Pulse adheres to **Semantic Versioning (SemVer 2.0.0)** and **URL Path Prefix Versioning** for all public REST API endpoints.

## Versioning Rules

1. **Path Prefix**: All public endpoints are rooted under `/api/v1/`.
2. **Backward Compatibility**: Non-breaking changes (new endpoints, optional request parameters, additional response fields) are introduced in-place under `/api/v1/`.
3. **Breaking Changes**: Any breaking modifications to payload schemas, authentication behavior, or endpoint removal will be released under `/api/v2/` with a minimum 6-month deprecation window for `/api/v1/`.
4. **Header Deprecation Signals**: Deprecated endpoints will emit standard HTTP headers:
   - `Deprecation: @<timestamp>`
   - `Sunset: <date>`
   - `Link: <canonical-migration-uri>; rel="successor-version"`

## Deprecation Schedule
| Version | Status | Released | Planned Sunset |
|:---|:---|:---|:---|
| `/api/v1` | **Active** | August 2026 | N/A |
