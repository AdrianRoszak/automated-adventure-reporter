# TrekkingAgentV2 — Automated Adventure Reporter

Serverless microservices platform for trekking activity processing with AI-powered agentic flow.
Built with **Pulumi (Golang)** for IaC and **Python** for microservices.

## Architecture Overview

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

## Repository Structure

```
.
├── infrastructure/          # Pulumi IaC (Golang)
│   ├── main.go             # Pulumi program
│   ├── go.mod / go.sum     # Go module dependencies
│   └── ...
├── services/               # Python microservices (Lambda code)
│   ├── strava-webhook/     # Strava → DynamoDB → SNS
│   ├── telegram-voice/     # Telegram voice → S3 → Deepgram
│   ├── agent-orchestrator/ # LangChain agent orchestration
│   └── report-generator/   # Report generation agent
├── libs/                   # Shared Python libraries
├── .github/workflows/      # CI/CD pipelines
├── Pulumi.yaml             # Pulumi project settings
├── ARCHITECTURE.md         # Detailed architecture & implementation plan
└── README.md               # This file
```

## Prerequisites

- Go 1.20 or later
- Python 3.14+
- An AWS account with credentials configured
- Pulumi CLI installed and authenticated
- Pulumi secrets configured (see below)

## Getting Started

### 1. Clone & install dependencies

```bash
git clone <repo-url>
cd TrekkingAgentV2
```

### 2. Configure Pulumi secrets

```bash
# Set the Telegram bot token (sensitive)
pulumi config set --secret telegramToken "your-telegram-token"

# Set the Deepgram API key (sensitive)
pulumi config set --secret deepgramApiKey "your-deepgram-api-key"
```

### 3. Deploy infrastructure

```bash
pulumi stack select dev   # or staging / prod
pulumi preview            # review changes
pulumi up                 # apply changes
```

## Security

- **No secrets in code** — all API keys and tokens are stored as Pulumi encrypted secrets
- **IAM least privilege** — each Lambda has a dedicated role with minimal permissions
- **Encryption at rest** — S3 (SSE-S3/KMS), DynamoDB (encryption at rest)
- **TLS 1.2+** — API Gateway enforces secure connections
- **Block Public Access** — enabled on all S3 buckets

## Documentation

- [ARCHITECTURE.md](./ARCHITECTURE.md) — detailed architecture & implementation plan
- [PULUMI_GUIDE.md](./PULUMI_GUIDE.md) — Pulumi developer guide (commands, workflow, best practices)

## License

Proprietary — all rights reserved.
