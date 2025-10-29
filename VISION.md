# Forge: Vision & Philosophy

> **A less opinionated Vercel for AWS Lambda + Terraform**

## What Forge Does (Three Things)

1. **Generate boilerplate code** - Get productive in serverless + Terraform on AWS quickly via approved methods
2. **Build and deploy via convention** - Folder structure determines what gets built and deployed
3. **Ephemeral environments** - Stand up and tear down preview environments from pipelines

**All deployments run in CI/CD pipelines** - keeps deployments in source control, auditable, and reproducible.

## The Problem

Serverless deployment tools force a false choice:

- **Serverless Framework/SAM**: YAML config hell, hidden CloudFormation, vendor lock-in, debugging nightmares
- **Raw Terraform**: 80 lines of boilerplate per Lambda, manual state setup, no deployment patterns
- **Vercel/Netlify**: Amazing DX but locked to their platforms, can't use your AWS account

**Root issue:** Wrong abstraction level—either too much magic (can't customize) or too little help (drowning in boilerplate).

## The Solution

**Forge = Vercel DX + your AWS account + transparent Terraform + zero lock-in**

```
Developer intent → File conventions → Readable Terraform → Your AWS Account
```

Every layer is inspectable. No black boxes. You own the infrastructure.

### What You Get

```bash
# 1. Generate boilerplate (approved patterns for your org)
forge new my-app --runtime=go --auto-state
  → Generates Terraform infrastructure
  → Generates function scaffolding
  → Sets up state backend (S3 + DynamoDB)

# 2. Build via convention (folder structure = intent)
forge build
  → Scans src/functions/*
  → Detects runtimes automatically (Go, Python, Node.js)
  → Builds deployment artifacts

# 3. Deploy from pipeline (never from local)
# In .github/workflows/deploy.yml
forge deploy
  → Runs terraform init/plan/apply
  → Auditable, reproducible, in source control

# 4. Ephemeral PR environments (preview before merge)
# In .github/workflows/pr-preview.yml
forge deploy --namespace=pr-${{ github.event.number }}
  → Isolated AWS resources per PR
  → Test changes before merging

# 5. Cleanup (automatic teardown)
# In .github/workflows/pr-cleanup.yml
forge destroy --namespace=pr-${{ github.event.number }}
  → Tears down preview environment
  → Cost control built-in
```

**Key benefits:**
- **Vercel-like DX** - simple commands, fast feedback, preview environments
- **Your AWS account** - not locked to a platform, full control
- **Zero config files** - conventions determine structure
- **Pipeline-first** - all deploys from CI/CD, never local
- **Readable output** - generated `.tf` files are the documentation
- **Zero lock-in** - edit Terraform directly or stop using Forge anytime

## Core Principles

### 1. Generate Approved Boilerplate
The first job is to **eliminate the boring setup work**:
- **Scaffold Terraform infrastructure** - working templates, not empty files
- **Set up state backend** - S3 + DynamoDB provisioned automatically
- **Generate function code** - hello-world examples that actually deploy
- **Approved patterns** - enterprises can customize templates for their org

Think: `create-react-app` but for serverless infrastructure.

### 2. Convention Over Configuration
The second job is to **infer intent from folder structure**:
- **No config files** - `src/functions/api/main.go` becomes Lambda "api" with Go runtime
- **Auto-discovery** - scan `src/functions/*` to find what needs building
- **Smart defaults** - sensible IAM, logging, monitoring out of the box
- **Exit ramp** - customize generated Terraform when you need to

Think: Rails/Next.js conventions, but for AWS infrastructure.

### 3. Pipeline-First Deployments
The third job is to **make ephemeral environments trivial**:
- **Never deploy from local** - all deployments happen in CI/CD
- **PR preview environments** - `--namespace=pr-123` prefixes all resources
- **Automatic cleanup** - tear down when PR closes
- **Cost control** - namespace tags track per-environment spending
- **Auditable** - all changes tracked in git + Terraform state

Think: Vercel preview deployments, but for your AWS account.

### 4. Transparent, Not Magic
Forge is **less opinionated than Vercel** by design:
- **Generated Terraform is editable** - no black boxes
- **You own the infrastructure** - it's your AWS account
- **Conventions are discoverable** - read the code to see what happens
- **Zero lock-in** - stop using Forge, keep the Terraform

### 5. Production-Grade Code Quality
The tool itself is built to deploy production infrastructure:
- **90% test coverage** - enforced by CI/CD
- **Pure functional programming** - monadic error handling, immutable data
- **Zero linting errors** - `golangci-lint` with strict rules
- **Mutation testing** - ensures tests actually catch bugs

If we're generating your infrastructure code, we better get it right.

## Why This Matters

### The Vercel Comparison

**What Vercel does great:**
- Zero config deployments
- Preview environments per PR
- Fast feedback loop
- Dead simple DX

**Where Vercel falls short (for some teams):**
- Locked to their platform
- Can't use your AWS account
- Limited customization
- Opaque infrastructure

**Forge gives you:**
- ✅ Same great DX (conventions, previews, fast feedback)
- ✅ Your AWS account (full control, no platform lock-in)
- ✅ Transparent infrastructure (editable Terraform)
- ✅ Customizable (approved patterns for your org)

### For Individual Developers
- **10 minutes to production** - `forge new` to deployed Lambda
- **No YAML hell** - conventions over configuration
- **Learn Terraform** properly while using it (readable output)
- **Preview environments** - test before merging

### For Teams
- **PR previews** - isolated AWS resources per pull request
- **Pipeline-driven** - all deploys from CI/CD, auditable
- **Reproducible** - same commands work everywhere
- **No vendor lock-in** - own your infrastructure code

### For Enterprises
- **Approved patterns** - customize boilerplate templates for your org
- **Compliance-ready** - standardized IAM, logging, encryption
- **Cost tracking** - namespace tags enable per-environment attribution
- **Exit strategy** - Terraform remains fully editable and portable

## The Workflow

### Local Development (One-Time Setup)
```bash
# 1. Generate boilerplate
forge new my-app --runtime=go --auto-state
  → Generates infra/ with Terraform
  → Generates src/functions/api/ with hello-world
  → Auto-provisions S3 state backend
  → Creates DynamoDB state lock table
  → Creates .github/workflows/ for CI/CD

# 2. Test build locally
forge build
  → Scans src/functions/*
  → Detects runtimes automatically
  → Builds to .forge/build/*.zip
```

### CI/CD Pipeline (Where Deploys Happen)

**Production deployment** (`.github/workflows/deploy.yml`):
```yaml
on:
  push:
    branches: [main]
jobs:
  deploy:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - run: forge deploy
        # → Builds functions
        # → Runs terraform init/plan/apply
        # → Deploys to production
```

**PR preview environment** (`.github/workflows/pr-preview.yml`):
```yaml
on:
  pull_request:
jobs:
  preview:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - run: forge deploy --namespace=pr-${{ github.event.number }}
        # → Creates isolated AWS resources: my-app-pr-123-*
        # → Comments PR with preview URL
        # → Test changes before merge
```

**PR cleanup** (`.github/workflows/pr-cleanup.yml`):
```yaml
on:
  pull_request:
    types: [closed]
jobs:
  cleanup:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - run: forge destroy --namespace=pr-${{ github.event.number }}
        # → Tears down preview environment
        # → Automatic cost control
```

**Key insight:** Deploy commands run in pipelines, not locally. This ensures all infrastructure changes are:
- ✅ Tracked in git
- ✅ Auditable via CI logs
- ✅ Reproducible
- ✅ Approved via PR process

## Project Structure

**Required:**
```
my-app/
├── infra/              # Terraform infrastructure (source of truth)
│   ├── main.tf         # Define AWS resources explicitly
│   ├── variables.tf    # namespace variable for ephemeral envs
│   ├── backend.tf      # Auto-generated S3 state config
│   └── outputs.tf
```

**Convention (optional but recommended):**
```
└── src/
    └── functions/      # Lambda functions (auto-discovered)
        ├── api/        # Function name = directory name
        │   └── main.go # Runtime detected from entry file
        └── worker/
            └── index.js
```

Forge scans `src/functions/*` to automatically detect:
- **Function names** - directory name (e.g., `api`, `worker`)
- **Runtimes** - detected from entry files:
  - `main.go` or `*.go` → Go (`provided.al2023`)
  - `index.js`, `index.mjs`, `handler.js` → Node.js (`nodejs20.x`)
  - `app.py`, `lambda_function.py`, `handler.py` → Python (`python3.13`)
- **Build targets** - automatically builds to `.forge/build/{name}.zip`

## What Forge Does NOT Do (By Design)

Forge is minimal by design. Developers handle:
- **Dependencies** - `go.mod`, `requirements.txt`, `package.json` (per function)
- **Shared code** - organize as needed, ensure it compiles
- **Secrets** - `.env` files, AWS Secrets Manager, SSM Parameter Store
- **IAM permissions** - define in Terraform
- **API Gateway routing** - define in Terraform
- **VPC configuration** - define in Terraform
- **Environment variables** - define in Terraform Lambda resources
- **Local testing** - use AWS SAM, LocalStack, or similar
- **Logs** - use AWS CloudWatch directly (or `forge logs` when added)
- **Cost management** - tag resources in Terraform

This is intentional—Forge automates the tedious parts while leaving control where it matters.

## Why Serverless.tf Modules

Forge leverages [serverless.tf](https://serverless.tf/) Terraform modules when appropriate:

- **Automate indifferent heavy lifting** - IAM roles, CloudWatch logs, API Gateway wiring
- **Least opinionated approach** - modules are configurable, or drop to raw resources
- **Full control** - swap module for raw resource anytime by editing `.tf`
- **Battle-tested** - used by thousands of production deployments

This is the **automation sweet spot**:
- More opinionated than raw Terraform (faster start)
- Less opinionated than frameworks (easy customization)
- Built on standard Terraform (zero lock-in)

## Comparison

| Feature | Forge | Vercel | Serverless Framework | SAM | Raw Terraform |
|---------|-------|--------|---------------------|-----|---------------|
| **DX** | ⭐⭐⭐⭐⭐ | ⭐⭐⭐⭐⭐ | ⭐⭐⭐ | ⭐⭐ | ⭐ |
| Your AWS account | ✅ Yes | ❌ No | ✅ Yes | ✅ Yes | ✅ Yes |
| Config files | **0** (convention) | 0 | serverless.yml | template.yaml | *.tf |
| Lock-in | **None** | High | High | Medium | None |
| Transparent infra | **Full** (editable .tf) | Opaque | Hidden CloudFormation | Hidden CloudFormation | Full |
| PR previews | **Built-in** (pipeline) | Built-in | Plugin | Manual | Manual |
| State management | **Auto** (S3+DynamoDB) | N/A | N/A | N/A | Manual |
| Boilerplate gen | **Yes** (approved patterns) | No | Limited | Limited | None |
| Learning curve | **Low** | Low | Medium | Medium | High |
| Exit strategy | **Edit .tf** | Migrate | Eject | Switch tools | N/A |
| Cost control | **Namespace tags** | Per-project | Manual | Manual | Manual |

**Forge = Vercel DX + AWS control + Terraform transparency**

## Code Quality Standards (Not Negotiable)

Forge is built to production standards from day 1:

1. **90% minimum test coverage** - aggregate across all packages
2. **Zero linting errors** - enforced by `golangci-lint`
3. **100% test pass rate** - no failures allowed in any test suite
4. **Mutation testing** - ≥80% mutation score for critical packages
5. **Pure functional programming** - monadic error handling, immutable data

**CI/CD enforces these standards** - PRs that violate them are rejected.

This isn't just rigor for its own sake—it ensures Forge is reliable enough to deploy production infrastructure.


