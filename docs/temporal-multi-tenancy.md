# Multi-Tenancy in Temporal

## Overview

This document describes how a single Temporal service is shared across multiple projects (tenants) within the organization. Each project operates independently while leveraging a centralized Temporal cluster managed by the platform team.

---

## What is Multi-Tenancy in Temporal?

Multi-tenancy means running **one Temporal server cluster** that serves **multiple independent projects**. Each project gets its own isolated environment (namespace) within the shared Temporal infrastructure.

```
                    ┌──────────────────────────────┐
                    │   Temporal Server (Shared)    │
                    │                               │
                    │  ┌─────────┐  ┌─────────┐    │
                    │  │  NS: A  │  │  NS: B  │    │
                    │  └────┬────┘  └────┬────┘    │
                    │  ┌─────────┐  ┌─────────┐    │
                    │  │  NS: C  │  │  NS: D  │    │
                    │  └────┬────┘  └────┬────┘    │
                    └───────┼────────────┼─────────┘
                            │            │
              ┌─────────────┘            └──────────────┐
              │                                         │
     ┌────────┴────────┐                     ┌──────────┴───────┐
     │   Project A      │                     │   Project B       │
     │   (Own Workers)  │                     │   (Own Workers)   │
     └─────────────────┘                     └──────────────────┘
```

**NS = Namespace**

---

## How Isolation Works

### Namespace Per Project

Every project is assigned a dedicated **namespace**. A namespace in Temporal provides:

| Feature              | Isolation Guarantee                                      |
|----------------------|----------------------------------------------------------|
| Workflow executions  | Project A cannot see or interact with Project B workflows |
| Task queues          | Each project has its own task queues                      |
| Workflow IDs         | Same workflow ID can exist in different namespaces         |
| Search attributes    | Scoped per namespace                                      |
| Retention policy     | Configurable per namespace                                |
| Archival settings    | Configurable per namespace                                |

### What is Shared

- Temporal Server (Frontend, History, Matching, Worker services)
- Database backend (PostgreSQL / MySQL / Cassandra)
- Monitoring and alerting infrastructure
- mTLS certificates and auth policies

### What is Owned by Each Project

- Workflow definitions (code)
- Activity implementations (code)
- Worker processes (deployed within the project's own infrastructure)
- Business logic and payload schemas

---

## Namespace Naming Convention

Each namespace follows a consistent naming pattern:

```
<project-name>-<environment>
```

Examples:

| Project        | Environment | Namespace               |
|----------------|-------------|-------------------------|
| Order Service  | Production  | `order-service-prod`    |
| Order Service  | Staging     | `order-service-staging` |
| Payment Service| Production  | `payment-service-prod`  |
| Notification   | Production  | `notification-prod`     |

---

## Onboarding a New Project (Tenant)

### Step 1 — Register Namespace

```bash
# Platform team creates the namespace
tctl namespace register order-service-prod \
  --retention 30 \
  --description "Order Service production workflows"
```

### Step 2 — Project Team Connects Workers

```typescript
// order-service/src/worker.ts
import { NativeConnection, Worker } from '@temporalio/worker';
import * as activities from './activities';

async function run() {
  const connection = await NativeConnection.connect({
    address: 'temporal.internal.company.com:7233', // shared Temporal address
  });

  const worker = await Worker.create({
    connection,
    namespace: 'order-service-prod',               // project's own namespace
    taskQueue: 'order-tasks',
    workflowsPath: require.resolve('./workflows'),
    activities,
  });

  await worker.run();
}

run();
```

### Step 3 — Project Team Starts Workflows

```typescript
// order-service/src/client.ts
import { Client, Connection } from '@temporalio/client';

async function startOrder(orderData: OrderInput) {
  const connection = await Connection.connect({
    address: 'temporal.internal.company.com:7233',
  });

  const client = new Client({
    connection,
    namespace: 'order-service-prod',
  });

  const result = await client.workflow.start('orderWorkflow', {
    taskQueue: 'order-tasks',
    workflowId: `order-${orderData.orderId}`,
    args: [orderData],
  });

  return result;
}
```

---

## Cross-Project Communication

When Project A needs to trigger a workflow in Project B:

```typescript
// Project A calling into Project B's namespace
const projectBClient = new Client({
  connection,
  namespace: 'payment-service-prod',  // target project namespace
});

await projectBClient.workflow.start('processPayment', {
  taskQueue: 'payment-tasks',
  workflowId: `payment-${orderId}`,
  args: [paymentData],
});
```

Alternatively, use **signals** to communicate between running workflows across namespaces.

> **Note**: Cross-namespace calls should be governed by access policies. Not every project should be able to call every other project.

---

## Resource Management

### Rate Limiting Per Namespace

Temporal supports per-namespace rate limits to prevent one noisy tenant from affecting others:

```yaml
# Temporal server dynamic config
frontend.namespaceRPS:
  - value: 1000
    constraints:
      namespace: "order-service-prod"
  - value: 500
    constraints:
      namespace: "notification-prod"
```

### Worker Scaling Per Project

Each project scales its own workers independently:

| Project          | Worker Replicas | Reason                        |
|------------------|-----------------|-------------------------------|
| Order Service    | 10              | High volume, critical path    |
| Notification     | 3               | Lower volume, async           |
| Reporting        | 2               | Batch processing, off-peak    |

---

## Monitoring

### Per-Tenant Metrics

Temporal emits metrics tagged with namespace, allowing per-project dashboards:

```
temporal_workflow_completed{namespace="order-service-prod"}
temporal_workflow_failed{namespace="payment-service-prod"}
temporal_activity_execution_latency{namespace="notification-prod"}
```

### Recommended Alerts Per Namespace

- Workflow failure rate > threshold
- Workflow execution latency > SLA
- Worker task queue backlog growing
- Schedule-to-start latency (indicates worker starvation)

---

## Security

### Authentication & Authorization

| Layer            | Mechanism                                              |
|------------------|--------------------------------------------------------|
| Transport        | mTLS between workers and Temporal server               |
| Namespace access | Authorizer plugin — each project gets certs/tokens scoped to their namespace |
| Payload          | Custom Codec Server for encryption at rest             |

### Example: Namespace-Scoped Access

```
Project A certs  → can access: order-service-prod, order-service-staging
Project B certs  → can access: payment-service-prod, payment-service-staging
Platform team    → can access: all namespaces
```

---

## Infrastructure Layout

```
┌─────────────────────────────────────────────────────────┐
│                   Platform Team Manages                  │
│                                                          │
│  ┌────────────┐  ┌────────────┐  ┌────────────────────┐ │
│  │  Temporal   │  │  Temporal   │  │   PostgreSQL /     │ │
│  │  Frontend   │  │  History    │  │   Cassandra DB     │ │
│  │  Service    │  │  Service    │  │   (shared)         │ │
│  └────────────┘  └────────────┘  └────────────────────┘ │
│  ┌────────────┐  ┌────────────┐  ┌────────────────────┐ │
│  │  Temporal   │  │  Temporal   │  │   Elasticsearch /  │ │
│  │  Matching   │  │  Worker     │  │   OpenSearch       │ │
│  │  Service    │  │  Service    │  │   (visibility)     │ │
│  └────────────┘  └────────────┘  └────────────────────┘ │
│                                                          │
│  ┌────────────────────────────────────────────────────┐  │
│  │  Temporal Web UI  •  Grafana  •  Prometheus        │  │
│  └────────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────┘

┌──────────────┐  ┌──────────────┐  ┌──────────────┐
│  Project A   │  │  Project B   │  │  Project C   │
│  Workers     │  │  Workers     │  │  Workers     │
│  (k8s pods)  │  │  (k8s pods)  │  │  (k8s pods)  │
└──────────────┘  └──────────────┘  └──────────────┘
   Each project deploys and scales its own workers
```

---

## Summary

| Aspect                  | Approach                                           |
|-------------------------|----------------------------------------------------|
| Deployment              | Single shared Temporal cluster                     |
| Tenant isolation        | One namespace per project per environment           |
| Worker ownership        | Each project runs its own workers                   |
| Code ownership          | Each project owns its workflows and activities      |
| Rate limiting           | Per-namespace RPS limits                            |
| Monitoring              | Namespace-tagged metrics, per-project dashboards    |
| Security                | mTLS + namespace-scoped authorization               |
| Cross-project calls     | Allowed via explicit namespace targeting with policy |
| Scaling                 | Each project scales workers independently           |
| Platform responsibility | Server infra, DB, upgrades, namespace provisioning  |
