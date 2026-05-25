# TrekkingAgentV2 - Architecture & Implementation Plan

## 📋 Overview

Serverless microservices platform for trekking activity processing with AI-powered agentic flow.
Built with **Pulumi (Golang)** for IaC and **Python** for microservices.

## 🎯 Strategy: "Adopt & Extend"

**Golden Rules:**
1. **Existing AWS resources are imported** into Pulumi, never recreated
2. **New components** (LangChain agents) are created by Pulumi from scratch
3. **Changes to existing resources** are additive only (e.g., new IAM permissions, new API endpoints)
4. **All existing resources** are protected with `--protect` flag after import

## 🏗️ Repository Structure

```
TrekkingAgentV2/
├── infrastructure/                    # Pulumi IaC (Golang)
│   ├── go.mod / go.sum
│   ├── Pulumi.yaml
│   ├── cmd/
│   │   └── pulumi-app/
│   │       └── main.go               # Entry point Pulumi
│   ├── pkg/
│   │   ├── stacks/
│   │   │   ├── dev.go
│   │   │   ├── staging.go
│   │   │   └── prod.go
│   │   ├── modules/
│   │   │   ├── networking/           # VPC, subnets, security groups
│   │   │   ├── compute/              # Lambda functions (import + new)
│   │   │   ├── api/                  # API Gateway (import + new endpoints)
│   │   │   ├── storage/              # S3 buckets (import + new)
│   │   │   ├── database/             # DynamoDB tables (import + new)
│   │   │   ├── messaging/            # SNS topics (import + new)
│   │   │   └── iam/                  # IAM roles & policies
│   │   └── config/
│   │       └── config.go             # Typed configuration
│   ├── tests/
│   │   └── ...                       # Unit tests for Pulumi modules
│   └── Makefile
│
├── services/                          # Python microservices (Lambda code)
│   ├── strava-webhook/               # EXISTING: Strava -> DynamoDB -> SNS
│   ├── telegram-voice/               # EXISTING: Telegram voice -> S3 -> Deepgram
│   ├── agent-orchestrator/           # NEW: LangChain agent orchestration
│   └── report-generator/             # NEW: Report generation agent
│
├── libs/                              # Shared Python libraries
│   ├── trekking-agent-core/          # LangChain agents, tools, chains
│   └── aws-utils/                    # Shared AWS helpers
│
├── .github/
│   └── workflows/
│       ├── ci.yml                    # Lint, test, build
│       └── deploy.yml                # Pulumi deploy (multi-stack)
│
├── docs/
│   ├── architecture.md
│   ├── adr/                          # Architecture Decision Records
│   └── api-specs/
│
├── .gitignore
├── ARCHITECTURE.md                   # This file
└── README.md
```

## 🧩 Pulumi Modules (Golang)

### 1. Networking Module (`pkg/modules/networking/`)
- VPC with public/private subnets
- VPC Endpoints for Lambda (DynamoDB, S3, SNS) - no NAT Gateway (cost saving)
- Security Groups for Lambda functions

### 2. IAM Module (`pkg/modules/iam/`)
- **NEW roles** for new Lambda functions (agent-orchestrator, report-generator)
- **Additional policies** granting new Lambdas access to existing resources
- Principle of Least Privilege - granular permissions per Lambda
- Resource-based policies on existing S3 buckets, SNS topics, DynamoDB tables

### 3. Database Module (`pkg/modules/database/`)
- **EXISTING (import):** `Activities` table
- **NEW:** `AgentSessions` table - LangChain agent sessions
- **NEW:** `Reports` table - generated reports metadata
- All with auto-scaling (on-demand capacity)
- TTL for old data

### 4. Storage Module (`pkg/modules/storage/`)
- **EXISTING (import):** `voice-uploads-{env}` bucket
- **EXISTING (import):** `transcriptions-{env}` bucket
- **NEW:** `reports-{env}` bucket - generated reports
- All with encryption at rest (SSE-S3 or KMS)
- Lifecycle policies for archiving
- Block Public Access enabled

### 5. Compute Module (`pkg/modules/compute/`)
- **EXISTING (import):** `strava-webhook` Lambda
- **EXISTING (import):** `telegram-voice` Lambda
- **EXISTING (import):** `deepgram-transcriber` Lambda
- **EXISTING (import):** `notification-sender` Lambda
- **NEW:** `agent-orchestrator` Lambda - main LangChain orchestrator
- **NEW:** `report-generator` Lambda - report generation
- ARM64/Graviton for cost savings
- Provisioned Concurrency for critical paths
- Reserved Concurrency for throttling protection
- Lambda Powertools for observability

### 6. API Module (`pkg/modules/api/`)
- **EXISTING (import):** API Gateway (HTTP API)
- **NEW endpoints:**
  - `POST /agent/process` → agent-orchestrator Lambda
  - `GET /reports/{id}` → report-generator Lambda
- WAF for protection
- CORS configured

### 7. Messaging Module (`pkg/modules/messaging/`)
- **EXISTING (import):** `activity-events` SNS topic
- **NEW:** `agent-events` SNS topic
- **NEW:** `report-ready` SNS topic
- **NEW:** EventBridge bus for agent communication
- DLQ (Dead Letter Queue) for each Lambda

## 🤖 Agentic Flow with LangChain

### Architecture

```
[Strava Webhook] ──→ [Activity Agent] ──→ [Context Agent] ──→ [Report Agent] ──→ [Report]
                         │                     │                     │
                         ▼                     ▼                     ▼
                    [DynamoDB]           [S3 - Transcriptions]   [S3 - Reports]
                         │                     │                     │
                         └──────────┬──────────┘                     │
                                    ▼                                │
                              [Orchestrator Agent] ──────────────────┘
                                    │
                                    ▼
                              [SNS - Notification]
```

### Agent Roles (LangChain + LangGraph)

1. **Activity Agent** - Fetches Strava activity data, analyzes (distance, elevation, time)
2. **Context Agent** - Merges activity data with voice transcriptions (Deepgram), creates context
3. **Orchestrator Agent** - Manages flow, decides when to trigger report
4. **Report Agent** - Generates narrative report using LLM + prompt engineering

### Communication
- Each agent = separate Lambda (loose coupling, independent scaling)
- Communication via **EventBridge** (async) or **Step Functions** (sync)
- **LangGraph** for agent graph definition
- **Prompt templates** in Secrets Manager or S3
- **LLM** - OpenAI / Anthropic / local (configurable via Secrets Manager)

## 🔒 Security

1. **Secrets Management:** AWS Secrets Manager for:
   - Strava API keys (Client ID, Client Secret, Access Token)
   - Telegram Bot Token
   - Deepgram API Key
   - LLM API Keys (OpenAI, Anthropic)

2. **Encryption:**
   - S3: SSE-S3 or KMS
   - DynamoDB: encryption at rest
   - Lambda env vars: no secrets, only Secrets Manager references
   - API Gateway: TLS 1.2+

3. **IAM:**
   - Each Lambda has **dedicated role** with minimal permissions
   - Resource-based policies on S3, SNS, DynamoDB
   - **No modification** of existing Lambda roles - only additive policies

4. **Network:**
   - Lambda in VPC only if needed
   - VPC Endpoints for AWS services (no NAT Gateway)
   - API Gateway public but WAF-protected

## 📈 Scalability

- **Lambda:** concurrency limits, reserved concurrency, provisioned concurrency
- **DynamoDB:** on-demand capacity mode (auto-scaling)
- **S3:** infinite scalability
- **SNS:** automatic scaling
- **API Gateway:** throttling and quota limits
- **EventBridge:** event retry and DLQ

## 🧪 TDD (Test-Driven Development)

### Pulumi (Golang):
- `pulumi/pulumi-go` - unit tests for modules
- `github.com/stretchr/testify/assert` - assertions
- Tests for each module: networking, iam, database, storage, compute, api, messaging

### Python microservices:
- `pytest` + `pytest-mock`
- `moto` (mock AWS services)
- `pytest-benchmark` for performance tests
- Unit tests for each LangChain agent

## 📦 Implementation Phases

### Phase 1: Foundation
- [ ] Git flow setup (done)
- [ ] ARCHITECTURE.md (this file)
- [ ] .gitignore
- [ ] Pulumi project initialization (Golang)
- [ ] Basic project structure

### Phase 2: Import Existing Resources
- [ ] Inventory existing AWS resources
- [ ] `pulumi import` for each resource with `--protect`
- [ ] Verify imports with `pulumi preview`

### Phase 3: New Infrastructure Modules
- [ ] Database module (AgentSessions, Reports tables)
- [ ] Storage module (Reports bucket)
- [ ] Compute module (agent-orchestrator, report-generator Lambdas)
- [ ] API module (new endpoints)
- [ ] Messaging module (new SNS topics, EventBridge)
- [ ] IAM module (new roles and policies)

### Phase 4: Agent Implementation (Python)
- [ ] Shared library: trekking-agent-core
- [ ] Activity Agent
- [ ] Context Agent
- [ ] Orchestrator Agent
- [ ] Report Agent
- [ ] Integration tests

### Phase 5: CI/CD
- [ ] GitHub Actions CI pipeline
- [ ] GitHub Actions deploy pipeline
- [ ] Multi-stack deployment (dev/staging/prod)

### Phase 6: Documentation
- [ ] Architecture Decision Records (ADRs)
- [ ] API specifications
- [ ] README with setup instructions
- [ ] Deployment guide

## 🚀 Git Flow Strategy

```
main ────────●─────────────────────────●──
              \                       /
develop ──────●──●──●──●──●──●──●──●──●──
               \  \     \  \     \
feature/       ●   ●     ●  ●     ●
import-*           agent-*    ci-cd
```

- `main` - production-ready code
- `develop` - integration branch
- `feature/import-*` - branches for importing existing resources
- `feature/agent-*` - branches for new agent implementations
- `feature/ci-cd` - CI/CD pipeline setup
