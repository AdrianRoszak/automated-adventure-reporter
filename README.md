# TrekkingAgentV2 — Automated Adventure Reporter

> ⚠️ **Work in progress** — this project is under active development. Architecture and APIs are subject to change.

Serverless microservices platform for trekking activity processing with AI-powered agentic flow.
Built with **Pulumi (Golang)** for IaC and **Python** for microservices.

## Quickstart

### Prerequisites

- Go 1.20+
- Python 3.14+
- AWS account with credentials configured
- [Pulumi CLI](https://www.pulumi.com/docs/install/) installed and authenticated

### Setup

```bash
# Clone the repository
git clone <repo-url>
cd TrekkingAgentV2

# Configure Pulumi secrets (sensitive values)
pulumi config set --secret telegramToken "your-telegram-token"
pulumi config set --secret deepgramApiKey "your-deepgram-api-key"

# Deploy infrastructure
pulumi stack select dev
pulumi preview
pulumi up
```

## Documentation

| Document | Description |
|----------|-------------|
| [ARCHITECTURE.md](./docs/ARCHITECTURE.md) | Detailed architecture & implementation plan |
| [PULUMI_GUIDE.md](./docs/PULUMI_GUIDE.md) | Pulumi developer guide — commands, workflow, best practices |

## Repository Structure

```
.
├── infrastructure/          # Pulumi IaC (Golang)
├── services/                # Python microservices (Lambda code)
├── libs/                    # Shared Python libraries
├── docs/                    # Project documentation
├── .github/workflows/       # CI/CD pipelines
├── Pulumi.yaml              # Pulumi project settings
└── README.md                # This file
```

## License

Proprietary — all rights reserved.
