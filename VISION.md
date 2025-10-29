# Forge: Vision & Design Philosophy

> **Convention over configuration for AWS Lambda + Terraform**
>
> Zero config files. Zero lock-in. Zero surprises.

## The Core Theory

**The serverless tooling landscape is broken because it optimizes for the wrong thing.**

### The Wrong Optimization

Most tools optimize for:
- **Feature completeness** (100 options in YAML) - "We support everything!"
- **Framework lock-in** (vendor wants you trapped) - "Use our ecosystem!"
- **Abstraction layers** (hiding complexity = hiding control) - "Trust our magic!"

**This optimization creates perverse incentives:**
- More features → more complexity → harder to understand
- More abstraction → less control → debugging nightmares
- Vendor lock-in → migration friction → teams stuck on outdated tools
- Hidden magic → opaque behavior → production surprises

### The Right Optimization

**Forge optimizes for:**
- **Speed to production** (5 minutes from zero to deployed) - Get value fast
- **Developer autonomy** (you own the infrastructure code) - Full control when needed
- **Exit velocity** (easy to leave Forge behind when you outgrow it) - Zero lock-in
- **Cognitive load reduction** (conventions > configuration) - Less to learn/remember

**This creates a virtuous cycle:**
- Conventions → faster onboarding → productive teams
- Readable output → easy debugging → confident deployments
- Zero lock-in → long-term sustainability → organizational trust
- Functional patterns → predictable behavior → fewer bugs

### The Core Insight: Abstraction vs Indirection

**Bad abstraction (most tools):**
```
Developer intent → YAML config → Mystery plugin → Generated CloudFormation → AWS
                   ↑ You write   ↑ Black box    ↑ Can't see this       ↑ What actually runs
```

**You lose visibility at every layer. Debugging requires understanding ALL the hidden layers.**

**Good abstraction (Forge):**
```
Developer intent → Conventions → Generated Terraform → AWS
                   ↑ File layout ↑ Readable .tf     ↑ What actually runs
```

**Every layer is inspectable. Generated code is the documentation.**

### Omakase for Infrastructure

**Omakase (お任せ)** = "I'll leave it up to you" (Japanese dining)

The chef picks your meal based on expertise. You trust their judgment. But you can always ask for modifications.

**Applied to infrastructure:**
- Forge picks sensible defaults (Go/Python/Node, Function URLs, IAM roles)
- You trust the conventions (5 minutes to production)
- You can always customize (edit the Terraform directly)
- No magic - you see exactly what you're getting

**This is NOT:**
- A framework you're locked into (it's just Terraform)
- An abstraction that hides complexity (generated .tf is readable)
- A black box that generates mysterious CloudFormation (you control the code)

**This IS:**
- A scaffold that generates readable Terraform (inspect anytime)
- A build tool that follows conventions (discoverable patterns)
- A deployment CLI that wraps terraform apply (transparent operations)

## The Problem (Detailed)

### Why Current Tools Fail

**Serverless Framework / SAM:**
```yaml
# serverless.yml - you write this
service: my-app
provider:
  name: aws
  runtime: nodejs20.x
functions:
  api:
    handler: index.handler
    events:
      - http:
          path: /
          method: get
```

**Pain points:**
1. **YAML Hell**: 200-line config files for simple apps
2. **Hidden Magic**: What CloudFormation is it generating? Who knows!
3. **Vendor Lock-In**: Try migrating off Serverless Framework... good luck
4. **Plugin Hell**: Need custom infra? Install 5 plugins, configure each one
5. **No Control**: Can't tweak the generated CloudFormation without hacks
6. **State is Hidden**: Where's my infrastructure? In some abstraction layer
7. **Team Friction**: Junior dev breaks prod because YAML typo wasn't caught
8. **Debugging Nightmare**: Error in CloudFormation? Have fun digging through generated JSON

### Raw Terraform

```hcl
# main.tf - you write this
resource "aws_lambda_function" "api" {
  function_name    = "my-app-api"
  role             = aws_iam_role.lambda.arn
  handler          = "index.handler"
  runtime          = "nodejs20.x"
  filename         = "lambda.zip"  # How do I create this?
  source_code_hash = filebase64sha256("lambda.zip")
}

resource "aws_iam_role" "lambda" {
  name = "my-app-lambda-role"
  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Action = "sts:AssumeRole"
      Effect = "Allow"
      Principal = { Service = "lambda.amazonaws.com" }
    }]
  })
}

resource "aws_iam_role_policy_attachment" "lambda_basic" {
  role       = aws_iam_role.lambda.name
  policy_arn = "arn:aws:iam::aws:policy/service-role/AWSLambdaBasicExecutionRole"
}

# ... 50 more lines for a simple Lambda
```

**Pain points:**
1. **Boilerplate Explosion**: 80 lines of HCL for one Lambda
2. **Manual Building**: "How do I zip my Lambda?" (Stack Overflow, here I come)
3. **State Management**: Manually create S3 bucket, configure backend, pray it works
4. **No Patterns**: Every team reinvents "how to structure Lambda projects"
5. **Steep Learning Curve**: Junior devs need 2 weeks to be productive
6. **Copy-Paste Hell**: 5 Lambdas = copy-paste 400 lines, update names everywhere
7. **PR Environments?**: "Just create another workspace, configure variables, hope for the best"

### The Real Problem: Wrong Abstraction Level

**Both approaches fail because:**
- **Too Abstract** (Serverless Framework): You lose control, can't customize
- **Too Concrete** (Raw Terraform): You drown in boilerplate

**What we actually need:**
- **Smart conventions** for 80% case (like Rails)
- **Raw Terraform access** for 20% case (like dropping into SQL in ActiveRecord)
- **Zero magic** in between

## The Forge Theory

### Omakase Infrastructure

**Omakase (お任せ)** = "I'll leave it up to you" (Japanese dining)

The chef picks your meal based on expertise. You trust their judgment. But you can always ask for modifications.

**Applied to infrastructure:**
- Forge picks sensible defaults (Go/Python/Node, Function URLs, IAM roles)
- You trust the conventions (5 minutes to production)
- You can always customize (edit the Terraform directly)

**This is NOT:**
- A framework you're locked into
- An abstraction that hides complexity
- A black box that generates mysterious CloudFormation

**This IS:**
- A scaffold that generates readable Terraform
- A build tool that follows conventions
- A deployment CLI that wraps terraform apply

### Convention Over Configuration (Rails Philosophy)

**Rails proved:**
- `app/models/user.rb` → automatically maps to `users` table
- `app/controllers/users_controller.rb` → automatically routes `/users`
- Zero config files needed for 80% of apps
- You can ALWAYS break conventions when needed

**Forge applies this:**
- `src/functions/api/` → automatically creates Lambda named "api"
- `main.go` → automatically detects Go runtime
- Zero config files for 80% of serverless apps
- You can ALWAYS edit `infra/*.tf` when needed

### Functional Programming = Predictability

**Most deployment tools hide state:**
```javascript
// Serverless Framework - what is this doing?
serverless.deploy()  // 🤷 Magic happens
```

**Forge uses pure functions:**
```go
// PURE: Same input → same output, no surprises
spec := GenerateProjectSpec(ProjectOptions{
    Name: "my-app",
    Runtime: "go",
})
// Now you can inspect spec, test it, reason about it

// IMPURE: Isolated to edges
WriteProject(spec)  // Only NOW do we touch the filesystem
```

**Why this matters:**
- **Testable**: Pure functions are trivial to test
- **Predictable**: No hidden state, no side effects
- **Debuggable**: See exactly what will be generated before writing files
- **Composable**: Functions fit together like Lego blocks

## Real-World Pain Points (Why This Matters)

### The Human Cost of Bad Tooling

These aren't just technical problems. They're **organizational friction** that compounds over time:
- Lost productivity (hours per week per developer)
- Degraded morale (frustration with tools)
- Reduced velocity (fear of breaking things)
- Opportunity cost (not shipping features)
- Team turnover (developers leave over tooling pain)

**Every hour spent fighting tools is an hour not spent solving customer problems.**

### Scenario 1: The PR Preview Nightmare

**The Setup:**
You're on a team building a SaaS product. Product manager wants to review UI changes before merging. "Can you give me a preview URL?" they ask.

**With Serverless Framework:**
```bash
# Developer wants to test changes in isolation
$ serverless deploy --stage pr-123

Error: Stage 'pr-123' is not defined in serverless.yml

# Okay, need to add stage config
$ vim serverless.yml
# Add stage: pr-123 configuration
# Copy-paste prod config, change a few values

$ serverless deploy --stage pr-123

Error: CloudFormation stack 'my-app-production' already exists

# Wait, what? I thought I was deploying to pr-123?
# Check serverless.yml... oh, the stack name isn't parameterized
# Fix stack name to include ${opt:stage}

$ serverless deploy --stage pr-123

Error: S3 bucket 'my-app-uploads' already exists

# Of course... static bucket names
# Need to prefix ALL resource names with stage
# Edit serverless.yml, add ${opt:stage} to 12 different places

$ serverless deploy --stage pr-123

Error: Environment variable 'DATABASE_URL' not set

# Right, need to configure env vars for this stage
# Add to serverless.yml under provider.environment.DATABASE_URL
# But how do I get a database URL for pr-123?
# ...do I need to provision a test database too?

# 45 minutes later, finally deployed
# PM: "Thanks! Oh wait, can you change the button color?"
# Developer: *cries internally*
```

**The hidden costs:**
- 45 minutes of developer time ($50-100 in salary cost)
- Context switching kills flow state
- PM waiting blocks product decisions
- Friction discourages PR previews → less feedback → worse product

**With Forge:**
```bash
$ forge deploy --namespace=pr-123

Scanning functions... ✓
Building api... ✓
Deploying to AWS (pr-123)... ✓
Done in 2m 14s

URL: https://pr-123-my-app.lambda-url.us-east-1.on.aws
```

**The win:**
- 2 minutes vs 45 minutes (95% time savings)
- PM gets preview immediately
- Developer stays in flow state
- PR previews become routine, not heroic

**The deeper insight:**
Forge doesn't solve this with more features. It solves it by **eliminating configuration**. The `--namespace` flag just prefixes all resources. No YAML, no stage config, no mental overhead.

### Scenario 2: The "What Did I Deploy?" Problem

**The Setup:**
Security audit. InfoSec asks: "What IAM permissions does your API Lambda have?" You need to answer in 24 hours.

**With Serverless Framework:**
```
Developer: "I need to see what IAM permissions our Lambda has"

# Check serverless.yml
provider:
  iamRoleStatements:
    - Effect: Allow
      Action:
        - dynamodb:*
      Resource: "*"

Developer: "Hmm, dynamodb:* on all resources... that seems broad"
Developer: "But what ACTUAL policy gets created?"

# Check AWS Console → CloudFormation → Stack
# Find stack: my-app-production
# Click Resources tab → 47 resources
# Which one is the IAM role? Ctrl+F "IAM"
# Find: IamRoleLambdaExecution
# Click through to IAM console
# See policy: ServerlessFrameworkPolicy-us-east-1-my-app-production-a3b5c7

# Click policy... 200 lines of JSON
# Includes permissions you didn't specify (framework adds its own)
# Includes inline policies from plugins
# Includes managed policies attached by framework

Developer: "Wait, why do we have s3:PutObject on *?"
Senior: "Oh, the deployment plugin needs that"
Developer: "In production?"
Senior: "Yeah, the framework bundles deployment permissions with runtime permissions"

Developer: "Can I remove it?"
Senior: "Not easily. You'd have to customize the IAM role, which means..."
# Googles "serverless framework custom IAM role"
# Finds 6 different approaches in Stack Overflow
# None of them work with plugins
# 3 hours later: "I still can't separate deployment from runtime permissions"

InfoSec: "We need this resolved before audit"
CTO: "How hard can it be to see IAM permissions?"
```

**The deeper problem:**
Generated code is hidden. You can't inspect it. You can't modify it without framework-specific hacks. The **abstraction owns you**.

**With Forge:**
```
Developer: "I need to see what IAM permissions our Lambda has"

# Open infra/main.tf
resource "aws_iam_role" "api_lambda" {
  name = "${var.namespace}my-app-api-lambda"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Action = "sts:AssumeRole"
      Effect = "Allow"
      Principal = { Service = "lambda.amazonaws.com" }
    }]
  })
}

resource "aws_iam_role_policy" "api_lambda" {
  name = "${var.namespace}my-app-api-policy"
  role = aws_iam_role.api_lambda.id

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Effect = "Allow"
      Action = [
        "dynamodb:GetItem",
        "dynamodb:PutItem",
        "dynamodb:Query"
      ]
      Resource = aws_dynamodb_table.main.arn
    }]
  })
}

Developer: "There it is. 20 lines. Exactly what we deployed."
InfoSec: "Can you make GetItem/PutItem/Query more restrictive?"
Developer: "Sure, let me edit lines 27-31"

# Edit Terraform directly
Action = [
  "dynamodb:GetItem",
  "dynamodb:Query"
]

$ forge deploy
# Done in 2 minutes

InfoSec: "Perfect. Can you document this?"
Developer: "It's already documented - the Terraform is the documentation"
```

**The win:**
- **Transparency**: You see exactly what gets deployed
- **No detective work**: infra/main.tf is the source of truth
- **Easy modification**: Edit 3 lines of HCL, redeploy
- **No framework friction**: Just Terraform, universal knowledge

**The deeper insight:**
The generated code IS your documentation. No need for "infrastructure as code" when the code is right there, human-readable, version-controlled.

### Scenario 3: The State Management Surprise

**The Setup:**
Junior developer joins team. First task: "Add a new Lambda function for email notifications."

**With Raw Terraform:**
```
Junior: "Okay, I added the Lambda to main.tf, let me deploy"

$ terraform plan
# Shows plan... looks good
$ terraform apply

Apply complete! Resources: 1 added, 0 changed, 0 destroyed.

Junior: "Done! Deployed the email Lambda"

# 30 minutes later...
Senior: "Why is the production API returning 500s?"
Junior: "Not sure, I only added the email Lambda"
Senior: "Did you configure the remote backend?"
Junior: "What's a remote backend?"

# Investigation reveals...
# Junior ran terraform locally without S3 backend
# Terraform created new LOCAL state file
# New state file thought there were ZERO existing resources
# terraform apply KEPT the email Lambda but had no record of others
# Now two state files diverged: local (1 resource) vs S3 (production with 30 resources)
# Production state thinks resources still exist
# But without managed state, next apply could orphan or delete them

Senior: "We need to terraform import 30 resources back into your state"
Junior: "I don't understand what happened"
Senior: "Terraform state is our source of truth. You created a second source of truth"

# 4 hours of terraform import commands later...
# Still not sure if everything is back in sync
# Team loses confidence in infrastructure
```

**The hidden costs:**
- 4 hours of senior eng time ($200-400)
- Production downtime / customer impact
- Team trust in infrastructure eroded
- Junior developer feels terrible (morale hit)
- Fear of terraform apply going forward

**The root cause:**
State management is **critical** but **invisible**. Teams learn about it the hard way.

**With Forge:**
```bash
$ forge new my-app --auto-state

Creating project my-app...
✓ Created infra/ directory
✓ Created src/functions/
✓ Provisioning S3 bucket for Terraform state...
  S3 Bucket: forge-state-my-app
  DynamoDB Table: forge_locks_my_app
✓ Generated backend.tf

Next steps:
  cd my-app
  forge deploy
```

**Junior dev's first deploy:**
```bash
$ cd my-app
$ forge deploy

Initializing Terraform...
Backend: S3 bucket forge-state-my-app
State locking: DynamoDB table forge_locks_my_app

Deploy complete! All team members will see this state.
```

**The win:**
- Remote state configured from day 1 (no way to mess this up)
- State locking prevents concurrent applies (prevents race conditions)
- Team collaboration works automatically
- Junior dev can't accidentally overwrite prod

**The deeper insight:**
Best practices should be the **default path**, not something you have to learn to configure. Forge makes the right thing the easy thing.

### Scenario 4: The Debugging Hell

**With Serverless Framework:**
```
Error: Failed to create CloudFormation stack
Caused by: Error creating Lambda function
Caused by: InvalidParameterValueException: The role defined for the function cannot be assumed by Lambda.

Developer: "What role? I didn't define a role!"
*Googles for 30 minutes*
*Finds it's generated by the framework*
*Can't see the actual IAM policy*
*Tries random fixes from Stack Overflow*
*Still broken after 2 hours*
```

**With Forge:**
```
Error: terraform apply failed
Error: Error creating Lambda function: InvalidParameterValueException:
The role defined for the function cannot be assumed by Lambda.

Developer: *Opens infra/main.tf*
Developer: *Sees the exact IAM role definition*
Developer: *Spots the typo in assume_role_policy*
Developer: *Fixes it in 30 seconds*
```

### Scenario 5: The Migration Disaster

**Trying to leave Serverless Framework:**
```
Manager: "We need to migrate to Terraform for compliance"
Team: "Our entire app is in Serverless Framework"
Manager: "Can't you export it?"
Team: "There's no clean export path"
Manager: "Rewrite it then"
Team: *Spends 3 months rewriting everything*
```

**With Forge:**
```
Manager: "Can we customize this more?"
Team: "Sure, it's just Terraform"
*Team edits infra/*.tf files directly*
*Optionally stops using forge commands*
*Zero migration needed - it was Terraform all along*
```

### Scenario 6: The Onboarding Disaster

**New dev joins team using Serverless Framework:**
```
Week 1: Learn YAML config format
Week 2: Learn CloudFormation (needed for debugging)
Week 3: Learn all the plugins and their configs
Week 4: Learn the framework's quirks and workarounds
Week 5: Finally productive (maybe)
```

**New dev joins team using Forge:**
```
Day 1:
  - See src/functions/api/ - "Oh, that's the API function"
  - Open infra/main.tf - "Oh, that's the Lambda resource"
  - Run forge build - "Oh, it built my code"
  - Run forge deploy - "Oh, it deployed it"
  - Productive by lunch
```

## The Developer Productivity Tax

### Quantifying the Hidden Cost

Let's do the math on a 5-person team over one year:

**With Traditional Serverless Tools (Serverless Framework, SAM):**

| Activity | Hours/Week/Dev | Annual Cost (5 devs @ $150k) |
|----------|---------------|-------------------------------|
| Fighting YAML configs | 2 | $36,000 |
| Debugging framework issues | 3 | $54,000 |
| PR environment setup friction | 2 | $36,000 |
| State management issues | 1 | $18,000 |
| Onboarding new devs | 0.5 | $9,000 |
| Migration/upgrade pain | 1 | $18,000 |
| **Total** | **9.5 hrs/wk** | **$171,000/year** |

**That's 25% of team capacity lost to tooling friction.**

**With Forge:**

| Activity | Hours/Week/Dev | Annual Cost (5 devs @ $150k) |
|----------|---------------|-------------------------------|
| Conventions just work | 0 | $0 |
| Edit Terraform directly | 0.5 | $9,000 |
| PR environments (--namespace) | 0 | $0 |
| State auto-configured | 0 | $0 |
| Onboarding (conventions) | 0.1 | $1,800 |
| Zero migration needed | 0 | $0 |
| **Total** | **0.6 hrs/wk** | **$10,800/year** |

**Savings: $160,000/year for a 5-person team.**

But the real cost isn't just money. It's:
- **Opportunity cost**: Features you didn't ship
- **Morale cost**: Developers frustrated with tools quit
- **Velocity cost**: Fear of infrastructure slows down iteration
- **Innovation cost**: No time to experiment when fighting tools

### The Compound Effect

**Year 1:**
- Team fights tools 9.5 hrs/week
- Ships 25% fewer features
- Competitors ship faster

**Year 2:**
- Team still fighting same tools
- Technical debt accumulates (YAML sprawl)
- Senior devs leave (burnout from framework friction)

**Year 3:**
- Junior devs inherit unmaintainable YAML
- Migration to raw Terraform starts (6-month project)
- Product development stalls

**With Forge:**
- Year 1: Ship features fast, own infrastructure
- Year 2: Infrastructure scales with team (just add .tf files)
- Year 3: No migration needed (it was always Terraform)

### The Real Question

**Not:** "Can we afford to use Forge?"

**But:** "Can we afford NOT to?"

Every week spent fighting serverless frameworks is a week your competitors spend shipping features.

## The Forge Solution (Detailed)

### Why Serverless.tf Matters

**The Problem with Reinventing the Wheel:**

Most serverless tools either:
1. Lock you into their abstractions (Serverless Framework)
2. Make you write everything from scratch (Raw Terraform)

**Forge takes a third path: Leverage proven Terraform modules**

[Serverless.tf](https://serverless.tf/) provides production-ready Terraform modules for AWS serverless:
- Lambda functions with proper IAM roles
- API Gateway v2 (HTTP API) with intelligent defaults
- DynamoDB tables with best-practice configurations
- SQS queues, SNS topics, S3 buckets
- Battle-tested by thousands of deployments

**How Forge Uses Serverless.tf:**

```hcl
# Forge generates this (using serverless.tf modules when appropriate)
module "lambda_api" {
  source = "terraform-aws-modules/lambda/aws"

  function_name = "${var.namespace}my-app-api"
  handler       = "bootstrap"
  runtime       = "provided.al2023"

  # ... sensible defaults from serverless.tf
}

# But you can ALWAYS drop down to raw resources
resource "aws_lambda_function" "custom" {
  # Full control when you need it
}
```

**The Forge Philosophy on Modules:**

1. **Use proven modules** (serverless.tf) for 80% case
2. **Zero magic** - generated code is readable
3. **Easy exit ramp** - replace module with raw resource anytime
4. **No lock-in** - modules are just Terraform

**Why This Matters:**

- **Automate indifferent heavy lifting**: State backends, IAM roles, CloudWatch logs
- **Least opinionated way possible**: Modules are configurable, or use raw resources
- **Developer always has control**: Edit the .tf file, swap module for resource
- **Get productive fast**: Defaults work, customize when needed
- **Eject at any point**: Remove Forge, keep the Terraform

This is the **automation sweet spot**:
- More opinionated than raw Terraform (faster to start)
- Less opinionated than frameworks (easier to customize)
- Built on standard Terraform (zero lock-in)

### Core Philosophy

1. **Omakase (Convention Over Configuration)**
   - Inspired by Ruby on Rails / DHH
   - Zero config files (`forge.yaml`, `serverless.yml`, etc.)
   - Smart defaults that just work (powered by serverless.tf modules)
   - Exit ramp: customize Terraform directly

2. **Functional Programming**
   - Pure functions, immutable data
   - Railway-oriented programming (Either monad)
   - No hidden state, predictable behavior
   - Testable and composable

3. **Minimal Magic**
   - No black boxes
   - Generated Terraform is readable/editable
   - Conventions are discoverable
   - Developer owns the infrastructure
   - Modules available but optional

4. **Production-Ready**
   - Ephemeral PR environments built-in
   - State management automated
   - Multi-region support
   - Cost tracking via tags
   - Battle-tested modules from serverless.tf

## How It Works

### Convention-Based Discovery

Forge discovers Lambda functions by scanning your project:

```
my-app/
├── infra/              # Terraform (you own this)
│   ├── provider.tf
│   ├── variables.tf    # namespace for ephemeral envs
│   ├── main.tf         # Lambda resources
│   └── backend.tf      # S3 state (auto-generated)
└── src/
    └── functions/      # Forge scans this
        ├── api/        # Function name = directory
        │   └── main.go # Runtime detected automatically
        └── worker/
            └── index.js
```

**Conventions:**
- Function name = directory name
- Runtime = detected from entry file:
  - `main.go` → Go (provided.al2023)
  - `index.js` → Node.js (nodejs20.x)
  - `app.py` → Python (python3.13)
- Build output = `.forge/build/{name}.zip`

**No configuration needed!**

### The Workflow

```bash
# 1. Create project (one command)
forge new my-app --runtime=go --auto-state

# Generated:
# - infra/ with Terraform config
# - src/functions/api/ with hello-world
# - S3 bucket for state (optional)
# - DynamoDB table for locking (optional)

# 2. Build (auto-discovery)
forge build
# Scans src/functions/*, detects runtimes, builds all

# 3. Deploy
forge deploy
# Builds + runs terraform apply

# 4. PR Preview (ephemeral environment)
forge deploy --namespace=pr-123
# Everything gets pr-123- prefix
# Isolated AWS resources
# Separate Terraform state

# 5. Cleanup
forge destroy --namespace=pr-123
```

### Ephemeral PR Environments

**GitHub Actions example:**

```yaml
# .github/workflows/pr-preview.yml
name: PR Preview
on: pull_request

jobs:
  deploy:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - name: Deploy preview
        run: forge deploy --namespace=pr-${{ github.event.number }}
      - name: Comment URL
        uses: actions/github-script@v7
        with:
          script: |
            github.rest.issues.createComment({
              issue_number: context.issue.number,
              owner: context.repo.owner,
              repo: context.repo.repo,
              body: '🚀 Preview deployed: pr-${{ github.event.number }}'
            })
```

**On PR close:**

```yaml
# .github/workflows/pr-cleanup.yml
on:
  pull_request:
    types: [closed]

jobs:
  cleanup:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - run: forge destroy --namespace=pr-${{ github.event.number }}
```

**Zero configuration. Just works.**

## Comparison Table

| Feature | Forge | Serverless Framework | SAM | Terraform |
|---------|-------|---------------------|-----|-----------|
| Config files | **0** (convention) | serverless.yml | template.yaml | *.tf |
| Lock-in | **None** | High | Medium | None |
| Terraform access | **Full control** | Hidden | None | Full |
| PR previews | **Built-in** | Plugin | Manual | Manual |
| State management | **Automated** | N/A | N/A | Manual |
| Learning curve | **Low** | Medium | Medium | High |
| Exit strategy | **Edit .tf directly** | Eject/rewrite | Switch tools | N/A |
| Multi-runtime | **Go/Python/Node** | Many | Many | Any |
| Custom infra | **Just add .tf** | Plugins | None | Native |

## Feature Roadmap

### ✅ Phase 1: Foundation (DONE)
- Convention-based discovery
- Multi-runtime builds (Go/Python/Node.js)
- `forge build`, `forge deploy`
- Namespace support for PR envs
- Pure functional architecture

### 🚧 Phase 2: Production Ready (IN PROGRESS)
- **`forge new --auto-state`** (current focus)
- S3 bucket auto-provisioning
- DynamoDB state locking
- backend.tf generation
- AWS credential detection

### 📋 Phase 3: Developer Experience
- Interactive TUI (bubbletea)
- `forge logs --namespace=pr-123`
- `forge list` (show all namespaces)
- Watch mode (`forge watch`)
- Cost tracking per namespace

### 📋 Phase 4: CI/CD Integration
- GitHub Actions workflow generation
- GitLab CI templates
- Deployment status on PRs
- Automatic PR cleanup
- Cost estimation comments

### 📋 Phase 5: Advanced Features
- Lambda Layers for shared deps
- API Gateway routing helpers
- DynamoDB auto-provisioning
- SQS/SNS wiring
- Custom domains (Route53)
- VPC helpers

### 📋 Phase 6: Observability
- Cost dashboard
- Performance metrics
- Error tracking
- Rollback support
- Drift detection

## Technical Architecture

### Functional Programming Principles

**Pure Core, Imperative Shell:**

```go
// PURE: Calculation (no I/O)
func GenerateBackendTF(cfg StateConfig) string {
    // Same inputs → same output
    // Testable, predictable
}

// IMPERATIVE: Action (I/O at edges)
func WriteBackendTF(path string, content string) error {
    // Side effects isolated
}

// COMPOSITION
func SetupProject(opts ProjectOpts) Either[error, Project] {
    spec := GenerateProjectSpec(opts)  // Pure
    return WriteProject(spec)          // Impure
}
```

**Railway-Oriented Programming:**

```go
// Either monad for error handling
result := pipeline.New(
    ScanFunctions(),      // Either[error, State]
    BuildFunctions(),     // Either[error, State]
    DeployTerraform(),    // Either[error, State]
).Run(ctx, initialState)

// Automatic short-circuit on first error
```

**Immutable Data:**

```go
type Function struct {
    Name    string  // No setters
    Runtime string  // Pure data
    Path    string  // Immutable
}

// Transformations return new values
func WithRuntime(f Function, rt string) Function {
    return Function{
        Name: f.Name,
        Runtime: rt,  // New value, original unchanged
        Path: f.Path,
    }
}
```

### Code Quality Standards

- **90% test coverage minimum** (enforced in CI)
- **Zero linting errors** (golangci-lint)
- **100% test pass rate**
- **Mutation testing** for critical paths
- Pure functions are easy to test!

## Why This Matters

### For Developers
- **5-minute setup** from zero to deployed
- **No YAML wrestling**
- **Full Terraform control** when you need it
- **PR previews without extra work**

### For Teams
- **No vendor lock-in**
- **Consistent patterns** across projects
- **Easy onboarding** (conventions, not config)
- **Cost visibility** per PR/env

### For Organizations
- **Reduced cloud costs** (ephemeral envs)
- **Faster iteration** (instant previews)
- **Better security** (IAM in Terraform)
- **Audit trail** (git + Terraform state)

## Getting Started

```bash
# Install (when published)
brew install forge  # or: go install github.com/lewis/forge@latest

# Create project
forge new my-app --runtime=go --auto-state

# Deploy
cd my-app
forge deploy

# Done! Lambda deployed with:
# - Function URL for testing
# - Terraform state in S3
# - Full IAM permissions
# - Ready for PR previews
```

## Exit Ramp Strategy

**Don't like Forge's conventions?**

1. **Customize Terraform**: Edit `infra/*.tf` files directly
2. **Custom builds**: Run `go build`, `npm install` manually
3. **Organize differently**: Move functions out of `src/functions/`
4. **Use other tools**: Generated Terraform works with any tool
5. **Remove Forge entirely**: Just keep the `.tf` files

**Zero lock-in. You own the infrastructure code.**

## Contributing

See `CLAUDE.md` for:
- Functional programming patterns
- Code quality standards
- Architecture decisions
- Development workflow

## License

MIT

---

**Built with ❤️ using functional programming principles**

*No black boxes. No magic. No surprises.*
