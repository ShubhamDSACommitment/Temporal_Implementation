# Temporal Cost Analysis: Self-Hosted vs Temporal Cloud

## Understanding Actions (Temporal's Billing Unit)

An **action** is any operation Temporal records in workflow history. Every workflow execution generates actions.

### Actions Per Activity Execution

| Action | When It Happens |
|--------|----------------|
| `ActivityTaskScheduled` | Temporal puts task on a task queue |
| `ActivityTaskStarted` | A worker picks it up |
| `ActivityTaskCompleted` | Worker returns the result |
| `WorkflowTaskStarted` | Workflow code resumes to run next step |

**Each activity = ~4 actions**

### Actions Per Workflow Execution

Using our OrderWorkflow as a reference (4 activities):

| Step | Actions |
|------|---------|
| Workflow start | 2 |
| ValidateOrder | 4 |
| ChargePayment (cross-service) | 4 |
| ShipOrder | 4 |
| SendNotification | 4 |
| Workflow complete | 1 |
| **Total per execution** | **~19** |

### What Adds More Actions

- **Retries**: Each retry adds ~3 extra actions per attempt
- **Timers/sleeps**: Each `workflow.Sleep()` = 2 actions (TimerStarted + TimerFired)
- **Signals**: Each signal received = 1-2 actions
- **Child workflows**: Each child has its own action count + 3 actions for parent orchestration
- **Queries**: Free (not recorded in history)

### Quick Formula

```
Actions per workflow ≈ (number_of_activities × 4) + 3
```

---

## Option 1: Temporal Cloud

### Pricing

| Component | Cost |
|-----------|------|
| Actions | ~$25 per 1 million actions |
| Storage | ~$0.042 per GB/month (workflow history retention) |
| Support | Included (premium support extra) |

### What Temporal Manages
- Server infrastructure
- Database (persistence layer)
- Version upgrades and patches
- Scaling and high availability
- Monitoring and alerting
- Multi-region replication (higher tiers)

### What You Manage
- Your workers (deployed in your infra)
- Your workflow and activity code
- API gateway (if using API-first approach)

### Cost Estimates by Scale

| Scale | Actions/Month | Estimated Monthly Cost |
|-------|--------------|----------------------|
| **Small** (1-2 products, low volume) | 1M | ~$25 + storage |
| **Medium** (5 products, moderate volume) | 50M | ~$1,250 + storage |
| **Large** (10+ products, high volume) | 200M | ~$5,000 + storage |
| **Very Large** (org-wide, heavy usage) | 1B | ~$25,000 + storage |

### Example: Order Processing

```
19 actions/order × daily orders × 30 days = monthly actions

1,000 orders/day   →    570K actions/month → ~$15/month
10,000 orders/day  →   5.7M actions/month  → ~$143/month
100,000 orders/day →  57M actions/month    → ~$1,425/month
1,000,000 orders/day → 570M actions/month  → ~$14,250/month
```

---

## Option 2: Self-Hosted (Open Source)

### License
Free — MIT license, no per-action charges.

### Infrastructure Requirements

| Component | Purpose | Minimum |
|-----------|---------|---------|
| Temporal Server | Frontend, History, Matching, Worker services | 4 pods (1 per service) |
| Database | Workflow persistence | PostgreSQL or Cassandra |
| Elasticsearch | Visibility (search/list workflows) | Optional but recommended |
| Kubernetes | Orchestration | 3+ nodes |

### Infrastructure Cost Estimates (Cloud Provider — AWS/GCP)

| Scale | Setup | Estimated Monthly Cost |
|-------|-------|----------------------|
| **Dev/Staging** | 3-node K8s + small Postgres (RDS) | $300–500 |
| **Small Production** | 3-node K8s + HA Postgres | $800–1,500 |
| **Medium Production** | 5-node K8s + HA Postgres + ES | $1,500–3,000 |
| **Large Production** | Multi-node K8s + Cassandra cluster + ES | $5,000–15,000 |
| **Org-Wide (HA, multi-region)** | Multi-region K8s + Cassandra + ES + monitoring | $15,000–40,000 |

### Hidden Costs (Often Overlooked)

| Cost | Estimate |
|------|----------|
| **Dedicated engineer(s)** for Temporal ops | 1-2 engineers' salary |
| **On-call rotation** for Temporal infra | Team time |
| **Upgrade testing** (Temporal releases ~monthly) | 2-4 days/month |
| **Incident response** (DB issues, scaling) | Unpredictable |
| **Monitoring setup** (Grafana, Prometheus, alerts) | Initial setup + maintenance |

---

## Side-by-Side Comparison

| Factor | Self-Hosted | Temporal Cloud |
|--------|------------|----------------|
| **Upfront cost** | High (infra setup + team) | Low (just connect) |
| **Per-action cost** | $0 | ~$25/1M actions |
| **Operational burden** | High | Near zero |
| **Time to production** | Weeks to months | Days |
| **Scaling** | You handle it | Automatic |
| **Upgrades** | Manual, risky | Automatic, zero-downtime |
| **Availability SLA** | You define/maintain | 99.9%+ guaranteed |
| **Data residency** | Full control | Limited to available regions |
| **Multi-tenancy** | You build namespace isolation | Built-in with namespaces |
| **Compliance** | Full audit control | SOC2, HIPAA available |
| **Vendor lock-in risk** | None | Low (can migrate to self-hosted) |
| **Best for** | Large orgs with infra teams | Small-to-mid teams, fast adoption |

---

## Break-Even Analysis

The point where self-hosting becomes cheaper than cloud:

```
Self-hosted monthly cost = Infra + (Engineer salary / 12)

Example:
  Infra: $3,000/month
  1 engineer at $150K/year: $12,500/month
  Total self-hosted: ~$15,500/month

  Break-even with Temporal Cloud:
  $15,500 / $25 per 1M = 620M actions/month

  If you're below 620M actions/month → Cloud is cheaper
  If you're above 620M actions/month → Self-hosted is cheaper
```

Adjust the engineer cost and infra for your org. The break-even point is usually **higher than people expect** because they underestimate operational costs.

---

## Recommendation: Phased Approach

### Phase 1: Start with Temporal Cloud (Month 1-6)
- Get running in days, not months
- Build your API-first gateway on top
- Onboard first 2-3 products
- Collect real usage data (actual actions/month)

### Phase 2: Evaluate (Month 6)
- Review actual cloud costs with real data
- Assess team capacity for self-hosting
- Check if data residency is a hard requirement

### Phase 3: Migrate If Needed (Month 6-12)
- If cloud costs > (infra + engineer costs) → migrate to self-hosted
- Your API gateway doesn't change — just swap the Temporal connection
- Workers don't change — same code, different server address

### Why This Works
- **No wasted time** building infra before proving Temporal's value
- **Real cost data** instead of guesses
- **API-first gateway is portable** — works with either option
- **Low migration risk** — Temporal Cloud and self-hosted use the same APIs

---

## Multi-Product Namespace Strategy

For an org with multiple products, use **Temporal namespaces** to isolate them:

```
Namespace: "product-a-prod"    → Product A workflows
Namespace: "product-b-prod"    → Product B workflows
Namespace: "product-c-prod"    → Product C workflows
Namespace: "shared-staging"    → Shared staging environment
```

Benefits:
- Per-product action tracking (know who's costing what)
- Independent retention policies
- Isolated failure domains
- Per-namespace access control

This works on both Temporal Cloud and self-hosted.
