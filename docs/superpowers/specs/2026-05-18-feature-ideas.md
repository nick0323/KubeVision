# Feature Ideas (Recorded, Not Scheduled)

Recorded 2026-05-18. None prioritized or scheduled.

## 1. Pod Resource Monitoring (CPU/Memory)
- Show CPU/memory usage per pod/node with time-series charts
- Depends on metrics-server in the cluster
- Backend: proxy metrics-server API or use metrics client-go
- Frontend: chart library (e.g., Chart.js, recharts)

## 2. Event Timeline
- Replace current Events table with a visual timeline
- Filter by resource type, severity, time range
- Pure frontend change (backend already has /api/events)

## 3. Helm Release Management
- List installed releases, upgrade, rollback
- Backend: integrate Helm SDK or shell out to helm CLI
- Frontend: dedicated page with release list + action buttons

## 4. Dark/Light Theme Toggle
- Codebase already uses CSS variables — switch theme by overriding variable values
- Pure frontend, minimal effort
- Persist preference in localStorage

## 5. Resource Topology Graph
- Visual graph showing resource relationships (Service→Pod→Deployment→ConfigMap)
- Backend: related_finder already has the logic
- Frontend: SVG/Canvas rendering (e.g., dagre, cytoscape.js)

## 6. Port Forwarding
- In Pod detail page, add a port-forward button
- Backend: WebSocket-based port forwarding tunnel
- Frontend: show generated local proxy URL
