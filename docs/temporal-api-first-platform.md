# Temporal API-First Platform Architecture

## Overview

This document describes how to run Temporal as a **shared platform service** where any product in the organisation can use Temporal without needing to know Temporal internals. Product teams interact through a clean REST/gRPC API — Temporal is an implementation detail hidden behind the platform.

---

## The Problem

Without an API-first approach, every product team must:

1. Learn the Temporal SDK
2. Manage Temporal client connections
3. Handle connection retries, timeouts, namespaces
4. Know which task queue to target
5. Understand workflow IDs, run IDs, idempotency keys

This doesn't scale across 10+ product teams. You end up with inconsistent implementations, duplicated boilerplate, and no central governance.

---

## Architecture

```
┌──────────────────────────────────────────────────────────┐
│                     Product Teams                         │
│                                                           │
│  Product A          Product B          Product C          │
│  (Orders)           (Inventory)        (Notifications)    │
│      │                  │                   │             │
│      │   HTTP/gRPC      │                   │             │
└──────┼──────────────────┼───────────────────┼─────────────┘
       │                  │                   │
       ▼                  ▼                   ▼
┌──────────────────────────────────────────────────────────┐
│              Temporal Platform API Gateway                 │
│                  (Your team owns this)                     │
│                                                           │
│  ┌─────────────┐ ┌──────────────┐ ┌───────────────────┐  │
│  │ Auth/RBAC   │ │ Rate Limiter │ │ Request Validator │  │
│  └─────────────┘ └──────────────┘ └───────────────────┘  │
│  ┌─────────────┐ ┌──────────────┐ ┌───────────────────┐  │
│  │ Audit Logger│ │ Metrics      │ │ Namespace Router  │  │
│  └─────────────┘ └──────────────┘ └───────────────────┘  │
│                                                           │
│  Temporal Client (single managed connection pool)         │
└───────────────────────────┬───────────────────────────────┘
                            │
                            ▼
┌──────────────────────────────────────────────────────────┐
│                   Temporal Server                          │
│                                                           │
│  Namespace: product-a-prod    ← Product A workflows       │
│  Namespace: product-b-prod    ← Product B workflows       │
│  Namespace: product-c-prod    ← Product C workflows       │
│  Namespace: shared-staging    ← All teams staging          │
└───────────────────────────┬───────────────────────────────┘
                            │
        ┌───────────────────┼───────────────────┐
        ▼                   ▼                   ▼
┌──────────────┐  ┌──────────────┐  ┌──────────────────┐
│ Product A    │  │ Product B    │  │ Product C        │
│ Workers      │  │ Workers      │  │ Workers          │
│ (order-svc)  │  │ (inv-svc)    │  │ (notif-svc)      │
└──────────────┘  └──────────────┘  └──────────────────┘
```

### Key Principle

- **Starting workflows** → goes through the API gateway (product teams call HTTP)
- **Processing workflows** → product teams still run their own workers (they write activities + register them)
- **Querying workflows** → goes through the API gateway

Product teams only need Temporal SDK knowledge for writing workers. All orchestration (start, cancel, signal, query) goes through your API.

---

## API Gateway — Endpoints

### Core Workflow Operations

```
POST   /api/v1/workflows                    Start a new workflow
GET    /api/v1/workflows/{workflowId}       Get workflow status and result
POST   /api/v1/workflows/{workflowId}/cancel    Cancel a running workflow
POST   /api/v1/workflows/{workflowId}/terminate  Terminate a running workflow
POST   /api/v1/workflows/{workflowId}/signal     Send a signal to a workflow
POST   /api/v1/workflows/{workflowId}/query      Query a workflow's state
GET    /api/v1/workflows                    List/search workflows (with filters)
```

### Schedule Operations

```
POST   /api/v1/schedules                    Create a scheduled workflow
GET    /api/v1/schedules/{scheduleId}       Get schedule details
DELETE /api/v1/schedules/{scheduleId}       Delete a schedule
PATCH  /api/v1/schedules/{scheduleId}       Pause/unpause a schedule
```

### Admin/Ops

```
GET    /api/v1/health                       Platform health check
GET    /api/v1/namespaces                   List available namespaces
GET    /api/v1/metrics/{namespace}           Per-namespace usage metrics
```

---

## API Contract — Request/Response Examples

### Start a Workflow

**Request:**
```bash
POST /api/v1/workflows
Authorization: Bearer <product-a-api-key>
Content-Type: application/json

{
  "workflow_type": "OrderWorkflow",
  "workflow_id": "order-ORD-001",
  "task_queue": "order-task-queue",
  "input": {
    "order_id": "ORD-001",
    "customer_id": "CUST-123",
    "item": "Mechanical Keyboard",
    "quantity": 2,
    "price": 79.99
  },
  "options": {
    "execution_timeout": "1h",
    "task_timeout": "30s",
    "id_reuse_policy": "REJECT_DUPLICATE",
    "retry_policy": {
      "max_attempts": 3,
      "initial_interval": "1s",
      "backoff_coefficient": 2.0
    },
    "search_attributes": {
      "CustomerID": "CUST-123",
      "Region": "us-east-1"
    }
  }
}
```

**Response (202 Accepted):**
```json
{
  "workflow_id": "order-ORD-001",
  "run_id": "a1b2c3d4-e5f6-7890-abcd-ef1234567890",
  "namespace": "product-a-prod",
  "status": "RUNNING",
  "started_at": "2026-03-19T10:30:00Z"
}
```

### Get Workflow Status

**Request:**
```bash
GET /api/v1/workflows/order-ORD-001
Authorization: Bearer <product-a-api-key>
```

**Response:**
```json
{
  "workflow_id": "order-ORD-001",
  "run_id": "a1b2c3d4-e5f6-7890-abcd-ef1234567890",
  "workflow_type": "OrderWorkflow",
  "namespace": "product-a-prod",
  "status": "COMPLETED",
  "started_at": "2026-03-19T10:30:00Z",
  "completed_at": "2026-03-19T10:30:05Z",
  "result": {
    "order_id": "ORD-001",
    "status": "COMPLETED",
    "total_charged": 159.98,
    "tracking_number": "TRACK-48291"
  }
}
```

### Send a Signal

**Request:**
```bash
POST /api/v1/workflows/order-ORD-001/signal
Authorization: Bearer <product-a-api-key>
Content-Type: application/json

{
  "signal_name": "cancel_order",
  "input": {
    "reason": "customer_request",
    "refund": true
  }
}
```

**Response (200 OK):**
```json
{
  "workflow_id": "order-ORD-001",
  "signal_name": "cancel_order",
  "delivered": true
}
```

### List/Search Workflows

**Request:**
```bash
GET /api/v1/workflows?status=FAILED&workflow_type=OrderWorkflow&start_time_after=2026-03-18T00:00:00Z
Authorization: Bearer <product-a-api-key>
```

**Response:**
```json
{
  "workflows": [
    {
      "workflow_id": "order-ORD-042",
      "workflow_type": "OrderWorkflow",
      "status": "FAILED",
      "started_at": "2026-03-18T14:22:00Z",
      "error": "payment failed: card declined"
    }
  ],
  "next_page_token": "eyJsYXN0X..."
}
```

---

## Multi-Tenancy with Namespaces

Each product gets its own Temporal namespace, providing full isolation.

### Namespace Strategy

```
product-a-dev        → Product A development
product-a-staging    → Product A staging
product-a-prod       → Product A production

product-b-dev        → Product B development
product-b-staging    → Product B staging
product-b-prod       → Product B production

platform-internal    → Platform team's own workflows (monitoring, cleanup)
```

### What Namespaces Isolate

| Concern | Isolated? |
|---------|-----------|
| Workflow history | ✅ Product A can't see Product B's workflows |
| Task queues | ✅ Same task queue name in different namespaces = separate queues |
| Search attributes | ✅ Each namespace has its own |
| Retention policy | ✅ Product A can keep 30 days, Product B keeps 7 days |
| Rate limits | ✅ One product hitting limits doesn't affect others |
| Action counting | ✅ Track per-product costs |

### How the API Gateway Routes to Namespaces

The API key determines which namespace to use. Product teams never specify the namespace directly.

```
API Key: pk_product_a_xxxx  →  routes to namespace "product-a-prod"
API Key: pk_product_b_xxxx  →  routes to namespace "product-b-prod"
API Key: sk_product_a_xxxx  →  routes to namespace "product-a-staging"
```

---

## Authentication & Authorization (RBAC)

### API Key Scopes

Each product team gets API keys with specific permissions:

```json
{
  "api_key": "pk_product_a_xxxx",
  "product": "product-a",
  "namespace": "product-a-prod",
  "permissions": [
    "workflow:start",
    "workflow:read",
    "workflow:signal",
    "workflow:cancel"
  ],
  "rate_limit": 1000,
  "allowed_workflow_types": [
    "OrderWorkflow",
    "RefundWorkflow"
  ],
  "allowed_task_queues": [
    "order-task-queue",
    "payment-task-queue"
  ]
}
```

### Permission Matrix

| Role | Start | Read | Signal | Cancel | Terminate | List | Admin |
|------|-------|------|--------|--------|-----------|------|-------|
| **product-dev** | ✅ | ✅ | ✅ | ✅ | ❌ | ✅ | ❌ |
| **product-prod** | ✅ | ✅ | ✅ | ✅ | ❌ | ✅ | ❌ |
| **product-admin** | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ❌ |
| **platform-admin** | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |

### Why This Matters

- Product A can't start workflows in Product B's namespace
- Dev API keys can't hit production namespaces
- Only platform admins can terminate workflows (terminate = force kill, loses history)
- Audit log captures who did what

---

## Rate Limiting & Quotas

### Per-Product Limits

```yaml
product-a:
  workflows_per_second: 100
  concurrent_workflows: 50000
  max_payload_size: 2MB
  max_workflow_duration: 24h
  signal_per_second: 200

product-b:
  workflows_per_second: 50
  concurrent_workflows: 10000
  max_payload_size: 1MB
  max_workflow_duration: 1h
  signal_per_second: 100
```

### Why Rate Limiting at the Gateway

Temporal itself has rate limits, but if one product floods Temporal, it affects everyone. The gateway enforces per-product limits **before** requests reach Temporal.

```
Product A (100 req/s) ──┐
                        ├──► Gateway (enforces per-product limits) ──► Temporal
Product B (200 req/s) ──┘    "Product B: 50 passed, 150 rejected (429)"
```

---

## Audit Logging

Every API call is logged for compliance and debugging:

```json
{
  "timestamp": "2026-03-19T10:30:00Z",
  "product": "product-a",
  "api_key_id": "pk_product_a_xxxx",
  "action": "workflow:start",
  "namespace": "product-a-prod",
  "workflow_id": "order-ORD-001",
  "workflow_type": "OrderWorkflow",
  "source_ip": "10.0.1.42",
  "status": 202,
  "latency_ms": 45
}
```

Useful for:
- **Cost attribution**: which product is generating how many actions
- **Debugging**: who started that runaway workflow
- **Compliance**: full audit trail of workflow operations
- **Alerting**: spike detection per product

---

## Onboarding a New Product Team

### What the Platform Team Does (One-Time)

1. Create a Temporal namespace for the product
2. Generate API keys (dev, staging, prod)
3. Set rate limits and quotas
4. Configure retention policy

### What the Product Team Does

1. **Write their workflow and activity code** (they need Temporal SDK only for this)
2. **Deploy their workers** (listens on their task queue in their namespace)
3. **Call the platform API** to start workflows — no Temporal SDK needed for this part

### Product Team's Minimal Knowledge

| Must Know | Don't Need to Know |
|-----------|-------------------|
| How to write a workflow function | How to connect to Temporal server |
| How to write activities | Namespace configuration |
| How to deploy a worker | Connection pooling, retries |
| Their API endpoint and key | Temporal server address |
| Their task queue name | Cluster topology |

### Onboarding Checklist

```
□ Namespace created: {product}-{env}
□ API keys generated and shared securely
□ Rate limits configured
□ Retention policy set
□ Product team has worker template/example
□ Task queue naming convention agreed (e.g., {product}-{service}-task-queue)
□ Workflow type naming convention agreed (e.g., {Product}{Action}Workflow)
□ Monitoring dashboard created for the namespace
□ Alerting rules configured (failed workflows, high latency)
□ Product team has access to Temporal UI (read-only, their namespace only)
```

---

## Observability

### Metrics the Platform Collects

**Per Namespace / Per Product:**

| Metric | Purpose |
|--------|---------|
| `workflows_started_total` | Volume tracking, cost attribution |
| `workflows_completed_total` | Success rate |
| `workflows_failed_total` | Error rate, alerting |
| `workflow_duration_seconds` | Performance |
| `activities_scheduled_total` | Action count estimation (for cloud cost) |
| `api_requests_total` | Gateway load |
| `api_request_duration_seconds` | Gateway latency |
| `rate_limit_rejections_total` | Capacity planning |

### Dashboards

Each product gets a dashboard showing:

```
┌─────────────────────────────────────────────────────┐
│  Product A — Temporal Dashboard                      │
│                                                      │
│  Active Workflows: 1,234    Failed (24h): 12         │
│  Avg Duration: 4.2s         P99 Duration: 12.8s      │
│                                                      │
│  ┌─────────────────────────────────┐                 │
│  │  Workflows Started (7 days)     │                 │
│  │  ████████████████████           │                 │
│  │  ██████████████████████████     │                 │
│  │  █████████████████              │                 │
│  └─────────────────────────────────┘                 │
│                                                      │
│  Estimated Temporal Cloud Cost: $342/month            │
│  Actions This Month: 13.7M                           │
└─────────────────────────────────────────────────────┘
```

### Alerting Rules

```yaml
# Platform-wide
- alert: TemporalServerDown
  condition: temporal_health != 1
  severity: critical
  notify: platform-oncall

# Per product
- alert: HighWorkflowFailureRate
  condition: workflow_failure_rate > 5%
  for: 10m
  severity: warning
  notify: product-team-slack

- alert: WorkflowStuck
  condition: workflow_running_duration > configured_max_duration
  severity: warning
  notify: product-team-slack

- alert: RateLimitBreached
  condition: rate_limit_rejections > 0
  for: 5m
  severity: info
  notify: product-team-slack
```

---

## Deployment Architecture

### Production Setup

```
┌─────────────────────────────────────────────────────────┐
│                    Kubernetes Cluster                     │
│                                                          │
│  ┌──────────────────────────────────────────────┐        │
│  │  Platform Namespace                           │        │
│  │                                               │        │
│  │  ┌─────────────────┐  ┌────────────────────┐ │        │
│  │  │ API Gateway      │  │ Temporal Server    │ │        │
│  │  │ (2+ replicas)    │  │ (4 services)       │ │        │
│  │  │ + Load Balancer  │  │ Frontend           │ │        │
│  │  └─────────────────┘  │ History             │ │        │
│  │                        │ Matching            │ │        │
│  │  ┌─────────────────┐  │ Worker              │ │        │
│  │  │ PostgreSQL (HA)  │  └────────────────────┘ │        │
│  │  │ or Cassandra     │                         │        │
│  │  └─────────────────┘  ┌────────────────────┐ │        │
│  │                        │ Elasticsearch      │ │        │
│  │  ┌─────────────────┐  └────────────────────┘ │        │
│  │  │ Temporal UI      │                         │        │
│  │  │ (internal only)  │                         │        │
│  │  └─────────────────┘                          │        │
│  └──────────────────────────────────────────────┘        │
│                                                          │
│  ┌────────────────┐ ┌────────────────┐ ┌──────────────┐ │
│  │ Product A NS   │ │ Product B NS   │ │ Product C NS │ │
│  │ order-worker   │ │ inv-worker     │ │ notif-worker │ │
│  │ payment-worker │ │ catalog-worker │ │              │ │
│  └────────────────┘ └────────────────┘ └──────────────┘ │
└─────────────────────────────────────────────────────────┘
```

### What the Platform Team Deploys
- Temporal server (4 internal services)
- PostgreSQL / Cassandra
- Elasticsearch
- API Gateway (the service described in this doc)
- Temporal UI (internal access only)
- Monitoring stack (Prometheus + Grafana)

### What Product Teams Deploy
- Their workers only
- Workers connect to Temporal through internal service DNS
- Workers are in their own Kubernetes namespaces

---

## Migration Path from Current Setup

### Where We Are Now

```
order-service  ──► direct SDK connection ──► Temporal
payment-service ──► direct SDK connection ──► Temporal
```

### Step 1: Build the API Gateway

Add the gateway service alongside existing direct connections. Both paths work simultaneously.

```
order-service  ──► direct SDK (still works)  ──► Temporal
new-product    ──► API Gateway ──────────────► Temporal
```

### Step 2: Migrate Workflow Starters

Move the workflow-starting code from `order-service/cmd/starter/` to API calls. Workers stay the same.

```
order-starter  ──► API Gateway ──► Temporal
order-worker   ──► direct SDK ──► Temporal (worker still uses SDK directly)
payment-worker ──► direct SDK ──► Temporal
```

### Step 3: Onboard New Products

New products only interact through the API. They never see Temporal directly.

### What Doesn't Change

- Workers always use the Temporal SDK directly (this is by design)
- Activity and workflow code stays the same
- Cross-service payload passing (like OrderWorkflow → ChargePayment) stays the same
- The API gateway only handles orchestration commands (start, cancel, signal, query)

---

## FAQ

**Q: Do workers still connect directly to Temporal?**
Yes. The API gateway handles orchestration (start/cancel/signal/query). Workers always connect to Temporal directly via the SDK — this is how Temporal is designed. The gateway doesn't sit in the activity execution path.

**Q: What if the API gateway goes down?**
Running workflows continue unaffected — workers talk directly to Temporal. Only new workflow starts, signals, and queries are impacted. Run 2+ gateway replicas behind a load balancer.

**Q: Can we use Temporal Cloud with this setup?**
Yes. The gateway just changes its connection string from `temporal-server:7233` to `your-namespace.tmprl.cloud:7233`. Everything else stays the same.

**Q: How do we handle large payloads?**
Temporal has a ~2MB payload limit. For larger data, store the payload in S3/GCS and pass the reference URL through Temporal. The gateway can enforce this automatically.

**Q: Can product teams still use Temporal UI?**
Yes. Give them read-only access scoped to their namespace. They can see their workflows, inspect history, and debug — but can't modify other products' workflows.

**Q: How do we track costs per product?**
Each product has its own namespace. Count actions per namespace. The gateway's audit log also tracks API calls per product. Both data sources together give you accurate cost attribution.
