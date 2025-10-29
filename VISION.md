# Forge: Vision & Philosophy

> **Convention over configuration for AWS Lambda + Terraform**

- **Serverless Framework/SAM**: YAML config hell, hidden CloudFormation, vendor lock-in, debugging nightmares
- **Raw Terraform**: 80 lines of boilerplate per Lambda, manual state setup, no deployment patterns

**Root issue:** Wrong abstraction level—either too much magic (can't customize) or too little help (drowning in boilerplate).

## The Solution

**Forge = Rails conventions + transparent Terraform + zero lock-in**

```
Developer intent → File conventions → Readable Terraform → AWS
```

Every layer is inspectable. No black boxes.

### What You Get

```bash
# 1. Create project (auto-provisions state backend)
forge new my-app --auto-state

# 2. Build (auto-discovers from src/functions/*)
forge build

# 3. Deploy
forge deploy

# 4. PR preview (isolated namespace)
forge deploy --namespace=pr-123
```

**Key benefits:**
- Zero config files (conventions determine structure)
- Generated Terraform is readable and editable
- S3 state + DynamoDB locking from day 1
- PR environments: `--namespace` prefixes all resources
- Exit anytime (just use the Terraform directly)

## Core Principles

1. **Convention over configuration** - `src/functions/api/main.go` automatically becomes Lambda named "api" with Go runtime
2. **Transparent output** - Generated `.tf` files are the documentation
3. **Productive Right Away**: We can get you up and running within 10 min to deploy the TF and App code in a Dev environment well generating all the boilerplate we can.
4. **Bake in Needed Standardization**: There is always a push and pull between deeveloper autonomy and Corporate guardrails for compliance and security. We try to thrad tha tline as much as possible with core patterns of building or deploying. 
5. Sane defaults, we are opinionated but leave it up to you to change what is needed or leave when needed. 
6. **Zero lock-in** - Edit Terraform directly or stop using Forge entirely

## Why Serverless.tf

Forge leverages [serverless.tf](https://serverless.tf/) Terraform modules when appropriate:

- **Automate indifferent heavy lifting** (IAM roles, CloudWatch logs, API Gateway)
- **Least opinionated approach** (modules are configurable, or drop to raw resources)
- **Full control** (swap module for raw resource anytime by editing .tf)
- **Battle-tested** (used by thousands of deployments)

This is the **automation sweet spot**:
- More opinionated than raw Terraform (faster start)
- Less opinionated than frameworks (easy customization)
- Built on standard Terraform (zero lock-in)


