# Temporal Workflow Designer - Complete Project Walkthrough

This document explains everything that happens in this project — from `docker compose up` to a user designing and running a workflow.

---

## Table of Contents

1. [Architecture Overview](#architecture-overview)
2. [Infrastructure Startup](#infrastructure-startup)
3. [Service Startup](#service-startup)
4. [Activity Registration](#activity-registration)
5. [User Opens the UI](#user-opens-the-ui)
6. [Designing a Workflow](#designing-a-workflow)
7. [Saving a Workflow](#saving-a-workflow)
8. [Running a Workflow](#running-a-workflow)
9. [Dynamic Workflow Execution](#dynamic-workflow-execution)
10. [Polling for Results](#polling-for-results)
11. [Heartbeat and Expiry](#heartbeat-and-expiry)
12. [Project Structure](#project-structure)

---

## Architecture Overview

```
┌──────────────────────────────────────────────────────────────────────┐
│                         BROWSER (Vue 3 SPA)                          │
│  ActivityPalette  |  Canvas (Vue Flow)  |  PropertiesPanel           │
│  ContextPad       |  WorkflowToolbar    |  Saved Workflows Page      │
└──────────┬───────────────────┬────────────────────┬──────────────────┘
           │                   │                    │
           │ GET /api/         │ POST /api/         │ POST /api/
           │ activities        │ definitions        │ workflows/start
           ▼                   ▼                    ▼
┌──────────────────────────────────────────────────────────────────────┐
│                     temporal-gateway (:8090)                          │
│                                                                      │
│  Roles:                                                              │
│  1. Static file server (serves the Vue SPA)                          │
│  2. REST API (HTTP ↔ Temporal/PostgreSQL bridge)                     │
│  3. Dynamic Workflow engine (worker on platform-task-queue)           │
│  4. Activity Registry (stores service registrations)                 │
└──────┬──────────┬────────────────────────┬───────────────────────────┘
       │          │                        │
       ▼          ▼                        ▼
  PostgreSQL    PostgreSQL           Temporal Server (:7233)
  activity_     workflow_            (workflow orchestration)
  registry      definitions                │
                                           │ routes activities by
                                           │ task_queue + activity_name
                                    ┌──────┴──────┐
                                    ▼             ▼
                             order-service   payment-service
                             (order-task-    (payment-task-
                              queue)          queue)
```

### Services

| Service | Port | Purpose |
|---------|------|---------|
| **postgresql** | 5432 | Database for Temporal internals + workflow designer data |
| **temporal** | 7233 | Temporal Server (workflow orchestration engine, gRPC) |
| **temporal-ui** | 8080 | Temporal's built-in Web UI (for debugging workflows) |
| **temporal-gateway** | 8090 | Our REST API + workflow engine + UI server |
| **order-service** | 8081 | Worker for order activities (validate, ship, notify) |
| **payment-service** | 8082 | Worker for payment activities (charge, refund) |

---

## Infrastructure Startup

When you run `docker compose up -d --build`, Docker starts services in dependency order:

### Step 1: PostgreSQL starts

```
postgresql starts
       ↓
Creates default database: "temporal" (user: temporal, password: temporal)
       ↓
Runs init-scripts/create-databases.sh:
  → CREATE DATABASE workflow_designer (if not exists)
  → GRANT ALL to temporal user
       ↓
Two databases now exist:
  - "temporal"           → used by Temporal Server for its internal state
  - "workflow_designer"  → used by temporal-gateway for our app data
```

### Step 2: Temporal Server starts

```
temporal (temporalio/auto-setup) starts
       ↓
Connects to PostgreSQL (database: "temporal")
       ↓
Auto-creates its internal tables (executions, tasks, namespaces, etc.)
       ↓
gRPC server listening on port 7233
       ↓
All workers and clients connect here
```

### Step 3: Temporal UI starts

```
temporal-ui starts
       ↓
Connects to Temporal Server at temporal:7233
       ↓
Web UI available at http://localhost:8080
       ↓
Shows namespaces, workflows, activity history (useful for debugging)
```

---

## Service Startup

### Step 4: temporal-gateway starts

```
temporal-gateway starts (cmd/server/main.go)
       ↓
1. Loads config:
   - TEMPORAL_ADDRESS=temporal:7233
   - DATABASE_URL=postgres://temporal:temporal@postgresql:5432/workflow_designer
   - PORT=8090
   - STATIC_DIR=/static (Vue SPA build output)
       ↓
2. Connects to PostgreSQL (workflow_designer database)
       ↓
3. Runs migrations — creates tables:
   - workflow_definitions (stores saved workflow designs)
   - activity_registry (stores registered activities from services)
       ↓
4. Connects to Temporal Server
       ↓
5. Starts a Temporal Worker on "platform-task-queue":
   - Registers DynamicWorkflow (the engine that executes user-designed workflows)
       ↓
6. Starts HTTP server on :8090 with routes:
   - GET  /api/definitions          → list saved workflows
   - GET  /api/definitions/{id}     → get one workflow
   - POST /api/definitions          → save new workflow
   - PUT  /api/definitions/{id}     → update workflow
   - DELETE /api/definitions/{id}   → delete workflow
   - POST /api/workflows/start      → execute a workflow
   - GET  /api/workflows/{id}/status → check execution status
   - POST /api/activities/register  → services register activities
   - GET  /api/activities           → list registered activities
   - DELETE /api/activities/service/{name} → deregister a service
   - GET  /                         → serves Vue SPA
       ↓
Gateway is ready. Waiting for requests.
```

### Step 5: order-service starts

```
order-service starts (cmd/worker/main.go)
       ↓
1. Loads config:
   - TEMPORAL_ADDRESS=temporal:7233
   - GATEWAY_URL=http://temporal-gateway:8090
       ↓
2. Connects to Temporal Server (with retry, up to 30 attempts)
       ↓
3. Creates a Temporal Worker on "order-task-queue"
   Registers:
   - OrderWorkflow (the hardcoded order processing workflow)
   - ValidateOrder activity
   - ShipOrder activity
   - SendNotification activity
       ↓
4. Registers activities with the gateway (see next section)
       ↓
5. Starts heartbeat goroutine (re-registers every 2 minutes)
       ↓
6. Starts health check server on :8081 (/healthz endpoint)
       ↓
Worker is running. Listening for tasks on "order-task-queue".
```

### Step 6: payment-service starts

```
payment-service starts (cmd/worker/main.go)
       ↓
Same pattern as order-service:
1. Connects to Temporal
2. Creates worker on "payment-task-queue"
   Registers:
   - ChargePayment activity (explicit name via RegisterActivityWithOptions)
   - RefundPayment activity (explicit name via RegisterActivityWithOptions)
3. Registers activities with gateway
4. Starts heartbeat
5. Health check on :8082
```

---

## Activity Registration

When order-service and payment-service start, they tell the gateway what activities they provide. This is so the UI palette knows what's available to drag onto the canvas.

### Registration Request (order-service example)

```
POST http://temporal-gateway:8090/api/activities/register

{
  "service_name": "order-service",
  "activities": [
    {
      "activity_name": "ValidateOrder",
      "display_name": "Validate Order",
      "description": "Validates order details and inventory",
      "task_queue": "order-task-queue",
      "category": "Order",
      "input_schema": [
        { "name": "order_id", "type": "string", "required": true, "description": "Order identifier" },
        { "name": "customer_id", "type": "string", "required": true, "description": "Customer identifier" },
        { "name": "item", "type": "string", "required": true, "description": "Item name" },
        { "name": "quantity", "type": "number", "required": true, "description": "Quantity ordered" },
        { "name": "price", "type": "number", "required": true, "description": "Item price" }
      ],
      "default_mapping": {
        "order_id": "$input.order_id",
        "customer_id": "$input.customer_id",
        "item": "$input.item",
        "quantity": "$input.quantity",
        "price": "$input.price"
      }
    },
    // ... ShipOrder, SendNotification
  ]
}
```

### What happens on the gateway

```
RegistryHandler.Register()
       ↓
Decodes JSON body
       ↓
store.BulkUpsertActivities() — runs in a DB transaction:
  INSERT INTO activity_registry (...) VALUES (...)
  ON CONFLICT (activity_name, task_queue)
  DO UPDATE SET ... last_heartbeat = NOW()
       ↓
Response: { "status": "registered" }
```

### After both services register

The `activity_registry` table contains:

| activity_name | task_queue | service_name | category | last_heartbeat |
|---|---|---|---|---|
| ValidateOrder | order-task-queue | order-service | Order | 2026-03-24 10:00:00 |
| ShipOrder | order-task-queue | order-service | Order | 2026-03-24 10:00:00 |
| SendNotification | order-task-queue | order-service | Order | 2026-03-24 10:00:00 |
| ChargePayment | payment-task-queue | payment-service | Payment | 2026-03-24 10:00:01 |
| RefundPayment | payment-task-queue | payment-service | Payment | 2026-03-24 10:00:01 |

---

## User Opens the UI

```
User opens http://localhost:8090
       ↓
temporal-gateway's spaHandler serves index.html + JS/CSS bundles
       ↓
Vue 3 app boots:
  - App.vue renders (navbar + RouterView)
  - Default route "/" loads the Designer page
       ↓
Designer page renders three panels:
  ┌─────────────┬──────────────────┬──────────────┐
  │  Activity   │                  │  Properties  │
  │  Palette    │   Canvas         │  Panel       │
  │  (left)     │   (center)       │  (right)     │
  └─────────────┴──────────────────┴──────────────┘
       ↓
ActivityPalette.vue mounts → useActivityRegistry().fetchActivities()
       ↓
GET /api/activities
       ↓
Gateway: store.ListActivities()
  → SELECT * FROM activity_registry WHERE last_heartbeat > NOW() - INTERVAL '5 minutes'
       ↓
Response:
{
  "categories": [
    {
      "name": "Order",
      "activities": [
        { "activity_name": "ValidateOrder", "display_name": "Validate Order", ... },
        { "activity_name": "ShipOrder", "display_name": "Ship Order", ... },
        { "activity_name": "SendNotification", "display_name": "Send Notification", ... }
      ]
    },
    {
      "name": "Payment",
      "activities": [
        { "activity_name": "ChargePayment", "display_name": "Charge Payment", ... },
        { "activity_name": "RefundPayment", "display_name": "Refund Payment", ... }
      ]
    }
  ]
}
       ↓
Palette renders:
  Events
    ├── Start Event
    └── End Event
  Order
    ├── Validate Order     (order-task-queue)
    ├── Ship Order         (order-task-queue)
    └── Send Notification  (order-task-queue)
  Payment
    ├── Charge Payment     (payment-task-queue)
    └── Refund Payment     (payment-task-queue)
```

---

## Designing a Workflow

Everything in this phase is **client-side only**. No API calls. All state lives in the Pinia store (`workflow.ts`).

### Dragging an activity from palette to canvas

```
User drags "Validate Order" from palette
       ↓
ActivityPalette.vue: onDragStart(event, item)
  Sets drag data:
    application/node-kind     = "activity"
    application/display-name  = "Validate Order"
    application/activity-name = "ValidateOrder"
    application/task-queue    = "order-task-queue"
    application/input-mapping = '{"order_id":"$input.order_id",...}'
       ↓
User drops on canvas
       ↓
useDragAndDrop composable: handles the drop event
  Reads drag data → creates a new node:
  {
    id: "step-ValidateOrder-1",
    type: "activity",
    position: { x: 300, y: 200 },
    data: {
      label: "Validate Order",
      activityName: "ValidateOrder",
      taskQueue: "order-task-queue",
      timeoutSeconds: 30,
      retryPolicy: { maxAttempts: 3, initialIntervalSec: 1, backoffCoefficient: 2.0 },
      inputMapping: { "order_id": "$input.order_id", ... }
    }
  }
       ↓
store.addNode(node) → node appears on canvas
```

### Connecting nodes with edges

```
User drags from one node's handle to another
       ↓
Vue Flow fires connect event
       ↓
store.addEdge({ id: "e-step1-step2", source: "step1", target: "step2" })
       ↓
An arrow appears: step1 → step2
This edge means: step2 depends_on step1 (step2 runs after step1 completes)
```

### Editing node properties

```
User clicks a node on the canvas
       ↓
store.selectNode(nodeId)
       ↓
PropertiesPanel.vue shows three tabs:

General tab:
  - Display Name: "Validate Order"
  - Activity Name: "ValidateOrder"
  - Task Queue: dropdown (order-task-queue / payment-task-queue)

Configuration tab:
  - Timeout: 30 seconds
  - Retry Policy: max 3 attempts, 1s initial interval, 2.0 backoff

Input/Output tab:
  - Expected Fields (from inputSchema — read-only hints):
      order_id    string  REQUIRED
      customer_id string  REQUIRED
      item        string  REQUIRED
      quantity    number  REQUIRED
      price       number  REQUIRED
  - Input Mapping (editable key-value table):
      order_id    → $input.order_id
      customer_id → $input.customer_id
      item        → $input.item
      quantity    → $input.quantity
      price       → $input.price
```

### Using the Context Pad

```
User clicks a node → ContextPad appears to the right of the node
       ↓
Two buttons:
  [+] Append node  → opens a menu of available activities
  [🗑] Delete node  → removes the node and its edges
       ↓
User clicks [+] → selects "Charge Payment"
       ↓
appendNode(item):
  1. Creates a new activity node positioned 250px to the right
  2. Automatically creates an edge from selected node → new node
  3. Pushes to undo history
```

---

## Saving a Workflow

```
User clicks "Save" button in WorkflowToolbar
       ↓
handleSave()
       ↓
exportDefinition() — converts canvas state to WorkflowDefinition:
  1. store.exportSteps():
     - Filters out event nodes (startEvent, endEvent)
     - For each activity node, creates a StepDefinition:
       {
         id: "step-ValidateOrder-1",
         name: "Validate Order",
         activity_name: "ValidateOrder",
         task_queue: "order-task-queue",
         timeout_seconds: 30,
         retry_policy: { max_attempts: 3, initial_interval_sec: 1, backoff_coefficient: 2.0 },
         input_mapping: { "order_id": "$input.order_id", ... },
         depends_on: ["step-StartEvent-0"]   ← derived from edges
       }
     - depends_on is built by looking at incoming edges to each node
  2. Wraps steps in:
     {
       id: "",
       name: "My Order Flow",
       description: "",
       version: 1,
       steps: [ ... ]
     }
       ↓
First save: POST /api/definitions
  → Gateway generates UUID, stores in workflow_definitions table
  → Returns { id: "abc-123", ... }
  → store.workflowId = "abc-123"

Subsequent saves: PUT /api/definitions/abc-123
  → Gateway updates the existing row
```

### What's stored in PostgreSQL

```sql
-- workflow_definitions table
id:          "abc-123"
name:        "My Order Flow"
description: ""
version:     1
definition:  '{                              -- JSONB column
  "steps": [
    {
      "id": "step-ValidateOrder-1",
      "name": "Validate Order",
      "activity_name": "ValidateOrder",
      "task_queue": "order-task-queue",
      "timeout_seconds": 30,
      "input_mapping": { "order_id": "$input.order_id", ... },
      "depends_on": []
    },
    {
      "id": "step-ChargePayment-2",
      "activity_name": "ChargePayment",
      "task_queue": "payment-task-queue",
      "input_mapping": { "order_id": "$input.order_id", "amount": "$input.price" },
      "depends_on": ["step-ValidateOrder-1"]
    }
  ]
}'
```

---

## Running a Workflow

```
User clicks "Run" → Run dialog opens
       ↓
User enters JSON input:
{
  "order_id": "ORD-001",
  "customer_id": "CUST-001",
  "item": "Widget",
  "quantity": 2,
  "price": 49.99
}
       ↓
User clicks "Execute"
       ↓
handleRun():
  1. Auto-saves the current canvas (so DB has latest version)
  2. POST /api/workflows/start
     {
       "definition_id": "abc-123",
       "input": { "order_id": "ORD-001", ... }
     }
       ↓
execution.go: Start()
  1. Fetches definition from DB: store.Get("abc-123")
  2. Creates workflowID: "dynamic-My-Order-Flow-a1b2c3d4"
  3. Wraps into DynamicWorkflowInput:
     {
       "definition": { ... full WorkflowDefinition ... },
       "input": { "order_id": "ORD-001", ... }
     }
  4. Calls Temporal:
     ExecuteWorkflow(
       ID: "dynamic-My-Order-Flow-a1b2c3d4",
       TaskQueue: "platform-task-queue",
       Workflow: "DynamicWorkflow",
       Input: dwInput
     )
       ↓
Temporal Server:
  - Creates a workflow execution record
  - Places a task on "platform-task-queue"
       ↓
Response to frontend:
  { "workflow_id": "dynamic-My-Order-Flow-a1b2c3d4", "run_id": "..." }
```

---

## Dynamic Workflow Execution

The temporal-gateway's worker picks up the task from `platform-task-queue` and runs `DynamicWorkflow`.

### How DynamicWorkflow works (DAG execution)

```go
// dynamic_workflow.go — simplified logic

func DynamicWorkflow(ctx, dwInput) {
    steps     = dwInput.Definition.Steps
    input     = dwInput.Input
    completed = {}
    results   = {}

    // Keep looping until all steps are done
    while len(completed) < len(steps) {

        // Find steps whose dependencies are ALL satisfied
        ready = []
        for each step in steps:
            if step not in completed:
                if ALL of step.depends_on are in completed:
                    ready.append(step)

        // Execute all ready steps IN PARALLEL
        for each step in ready:
            activityInput = resolveInputMapping(step.input_mapping, input, results)
            futures[step] = ExecuteActivity(step.activity_name, activityInput,
                                            TaskQueue: step.task_queue)

        // Wait for all parallel steps
        for each step in ready:
            results[step.id] = futures[step].Get()
            completed[step.id] = true
    }

    return results
}
```

### Concrete example execution

Given this workflow: `ValidateOrder → ChargePayment → ShipOrder → SendNotification`

```
ROUND 1: Check dependencies
  ValidateOrder: depends_on=[]              → READY
  ChargePayment: depends_on=[ValidateOrder] → NOT READY (ValidateOrder not done)
  ShipOrder:     depends_on=[ChargePayment] → NOT READY
  SendNotif:     depends_on=[ShipOrder]     → NOT READY
       ↓
Execute ValidateOrder:
  1. Resolve input mapping:
       "order_id": "$input.order_id" → "ORD-001"
       "quantity": "$input.quantity"  → 2
       "price":    "$input.price"    → 49.99
  2. Temporal: ExecuteActivity("ValidateOrder", input, TaskQueue: "order-task-queue")
  3. Temporal routes to order-service worker
  4. order-service runs ValidateOrder() → returns { "status": "VALIDATED", "order_id": "ORD-001" }
  5. stepResults["step-ValidateOrder-1"] = { "status": "VALIDATED", ... }

ROUND 2:
  ChargePayment: depends_on=[ValidateOrder] → READY (ValidateOrder is done)
  ShipOrder:     depends_on=[ChargePayment] → NOT READY
       ↓
Execute ChargePayment:
  1. Resolve: "amount": "$input.price" → 49.99
  2. Temporal: ExecuteActivity("ChargePayment", input, TaskQueue: "payment-task-queue")
  3. Temporal routes to payment-service worker
  4. payment-service runs ChargePayment() → returns { "transaction_id": "TXN-42", "amount_charged": 49.99 }

ROUND 3:
  ShipOrder: depends_on=[ChargePayment] → READY
       ↓
Execute ShipOrder:
  1. Temporal routes to order-service
  2. Returns { "tracking_number": "TRACK-12345", "carrier": "FedEx" }

ROUND 4:
  SendNotification: depends_on=[ShipOrder] → READY
       ↓
Execute SendNotification:
  1. Returns { "status": "SENT", "order_id": "ORD-001" }

ALL STEPS DONE → return merged results
```

### Input Mapping Expressions

The input mapping system supports three expression types:

| Expression | Resolves to | Example |
|---|---|---|
| `$input.field` | Workflow input value | `$input.order_id` → `"ORD-001"` |
| `$steps.stepId.field` | A prior step's result | `$steps.step-ChargePayment-2.transaction_id` → `"TXN-42"` |
| Literal | The string itself | `"Order update"` → `"Order update"` |

This allows **chaining data between steps** — e.g., the notification step can reference the tracking number from the ship step.

### Parallel Execution

If two steps have the same dependency, they run in parallel:

```
          ┌─── ChargePayment ───┐
Validate ─┤                     ├─── ShipOrder
          └─── SendNotification ─┘
```

In this case, ChargePayment and SendNotification both depend only on ValidateOrder, so they execute simultaneously in Round 2.

---

## Polling for Results

```
After starting the workflow, the frontend polls every 2 seconds:

GET /api/workflows/dynamic-My-Order-Flow-a1b2c3d4/status
       ↓
execution.go: Status()
  → Temporal: DescribeWorkflowExecution(workflowId)
       ↓
Response cycle:
  { "status": "Running" }                  ← steps still executing
  { "status": "Running" }                  ← waiting...
  { "status": "Completed", "result": {     ← all done!
      "step-ValidateOrder-1": { "status": "VALIDATED" },
      "step-ChargePayment-2": { "transaction_id": "TXN-42", "amount_charged": 49.99 },
      "step-ShipOrder-3": { "tracking_number": "TRACK-12345", "carrier": "FedEx" },
      "step-SendNotification-4": { "status": "SENT" },
      "_input": { "order_id": "ORD-001", ... }
  }}
       ↓
UI shows: "Workflow completed! Result: { ... }"
```

If a step fails:
```
{ "status": "Failed", "error": "step step-ChargePayment-2 (ChargePayment) failed: invalid payment amount" }
```

The frontend stops polling and shows the error.

---

## Heartbeat and Expiry

### How services stay alive in the registry

```
Every 2 minutes:
  order-service   → POST /api/activities/register → updates last_heartbeat = NOW()
  payment-service → POST /api/activities/register → updates last_heartbeat = NOW()
```

### What happens when a service goes down

```
1. order-service stops (docker compose stop order-service)
2. Heartbeat goroutine stops
3. last_heartbeat freezes at the time it stopped
4. After 5 minutes, the gateway's ListActivities() query filters it out:
     WHERE last_heartbeat > NOW() - INTERVAL '5 minutes'
5. User refreshes the UI → Order activities disappear from palette
6. Payment activities still appear (payment-service is still running)
```

### Graceful shutdown

```
order-service receives SIGTERM
       ↓
close(heartbeatStop)  → stops the heartbeat goroutine
healthServer.Shutdown() → stops health check
w.Stop()               → stops the Temporal worker
       ↓
Activities naturally expire after 5 minutes
```

---

## Project Structure

```
Temporial_Implementation/
├── shared/                          # Shared Go module (used by all services)
│   ├── go.mod                       # Module: temporal-shared
│   ├── types.go                     # OrderInput, PaymentRequest, ActivityRegistration, etc.
│   ├── taskqueues.go                # Task queue constants (order-task-queue, etc.)
│   ├── workflow_definition.go       # WorkflowDefinition, StepDefinition structs
│   └── registry_client.go           # RegisterWithGateway(), StartHeartbeat()
│
├── temporal-gateway/                # Central hub: API + workflow engine + UI server
│   ├── Dockerfile
│   ├── go.mod
│   ├── cmd/server/main.go           # Entry point: connects DB, Temporal, starts HTTP
│   └── internal/
│       ├── config/config.go          # Environment config
│       ├── store/postgres.go         # PostgreSQL: migrations, CRUD for definitions + registry
│       ├── api/
│       │   ├── router.go             # HTTP routes + CORS + SPA handler
│       │   ├── definitions.go        # CRUD handlers for workflow definitions
│       │   ├── execution.go          # Start workflow + poll status
│       │   └── registry.go           # Register/List/Deregister activity endpoints
│       └── workflow/
│           └── dynamic_workflow.go   # DynamicWorkflow: DAG execution engine
│
├── order-service/                   # Order processing worker
│   ├── Dockerfile
│   ├── go.mod
│   ├── cmd/worker/main.go           # Entry: Temporal worker + activity registration + heartbeat
│   └── internal/
│       ├── config/config.go          # Config with GatewayURL
│       ├── activities/
│       │   └── order_activities.go   # ValidateOrder, ShipOrder, SendNotification
│       └── workflow/
│           └── order_workflow.go     # Hardcoded OrderWorkflow (alternative to dynamic)
│
├── payment-service/                 # Payment processing worker
│   ├── Dockerfile
│   ├── go.mod
│   ├── cmd/worker/main.go           # Entry: Temporal worker + registration + heartbeat
│   └── internal/
│       ├── config/config.go
│       └── activities/
│           └── payment_activities.go # ChargePayment, RefundPayment
│
├── workflow-designer/               # Vue 3 frontend (SPA)
│   ├── src/
│   │   ├── App.vue                   # Root component with navigation
│   │   ├── api/client.ts             # Axios HTTP client for all API calls
│   │   ├── types/workflow.ts         # TypeScript types (PaletteItem, WorkflowDefinition, etc.)
│   │   ├── stores/workflow.ts        # Pinia store (nodes, edges, selection)
│   │   ├── composables/
│   │   │   ├── useActivityRegistry.ts # Fetches + caches activities from API (singleton)
│   │   │   ├── useHistory.ts          # Undo/redo
│   │   │   ├── useDragAndDrop.ts      # Palette → canvas drag handling
│   │   │   ├── useAutoLayout.ts       # Auto-arrange nodes
│   │   │   ├── useWorkflowExport.ts   # Canvas → WorkflowDefinition conversion
│   │   │   └── useKeyboardShortcuts.ts
│   │   └── components/
│   │       ├── sidebar/
│   │       │   └── ActivityPalette.vue # Left panel: draggable activities
│   │       ├── canvas/
│   │       │   └── ContextPad.vue      # Floating menu on selected node
│   │       ├── properties/
│   │       │   └── PropertiesPanel.vue # Right panel: edit node properties
│   │       └── toolbar/
│   │           └── WorkflowToolbar.vue # Top bar: save, run, undo, redo
│   └── ...
│
├── docker-compose.yml               # All services orchestration
├── init-scripts/
│   └── create-databases.sh          # Creates workflow_designer DB
└── dynamicconfig/
    └── development-sql.yaml         # Temporal server config
```

---

## Database Tables

### workflow_definitions (stores saved workflow designs)

| Column | Type | Description |
|--------|------|-------------|
| id | TEXT PK | UUID generated on creation |
| name | TEXT | Workflow name |
| description | TEXT | User description |
| version | INT | Version number |
| definition | JSONB | Full step definitions (the workflow blueprint) |
| created_at | TIMESTAMPTZ | Creation timestamp |
| updated_at | TIMESTAMPTZ | Last update timestamp |

### activity_registry (stores registered activities from services)

| Column | Type | Description |
|--------|------|-------------|
| activity_name | TEXT (PK) | e.g., "ValidateOrder" |
| task_queue | TEXT (PK) | e.g., "order-task-queue" |
| display_name | TEXT | e.g., "Validate Order" |
| description | TEXT | Human-readable description |
| service_name | TEXT | e.g., "order-service" |
| category | TEXT | e.g., "Order" — used for palette grouping |
| input_schema | JSONB | Expected input fields with types |
| default_mapping | JSONB | Default input mapping expressions |
| registered_at | TIMESTAMPTZ | First registration time |
| last_heartbeat | TIMESTAMPTZ | Last heartbeat (used for expiry) |

---

## Key Concepts

### Task Queues
Task queues are how Temporal routes work to the right service. Each service listens on its own queue:
- `order-task-queue` → order-service picks up tasks
- `payment-task-queue` → payment-service picks up tasks
- `platform-task-queue` → temporal-gateway picks up DynamicWorkflow executions

### Dynamic vs Hardcoded Workflows
This project has **two ways** to run workflows:
1. **DynamicWorkflow** — user designs in the UI, saved as JSON, executed by the gateway's engine
2. **OrderWorkflow** — hardcoded Go function in order-service (used for direct Temporal API calls)

The dynamic approach is the main feature — it lets non-developers design workflows visually.

### Activity Registration Pattern
Instead of hardcoding available activities in the frontend, services **self-register** on startup. This means:
- Adding a new microservice automatically makes its activities appear in the UI
- No frontend code changes needed for new services
- Activities disappear from the palette when a service goes down (heartbeat expiry)
