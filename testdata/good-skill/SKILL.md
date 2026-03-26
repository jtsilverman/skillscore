---
name: deploy-app
description: Deploys the application to production with zero-downtime rolling updates. Use when the user asks to deploy, push to production, or ship a release.
disable-model-invocation: true
---

# Deploy Application

Deploy the application using a zero-downtime rolling update strategy.

## Prerequisites

Ensure the following before deploying:
- All tests pass locally
- The current branch is up to date with main

## Workflow

1. Run the test suite to verify nothing is broken
2. Build the production artifacts
3. Push to the deployment target
4. Verify the deployment succeeded with a health check
5. If health check fails, automatically roll back

## Usage

```bash
python scripts/deploy.py --env production --verify
```

## Configuration

The deploy script reads from `deploy.yaml` in the project root:

```yaml
target: production
health_check_url: /api/health
rollback_on_failure: true
```

## Rollback

If anything goes wrong, run:

```bash
python scripts/deploy.py --rollback
```

This restores the previous version within 30 seconds.

## Additional resources

- For environment-specific settings, see [reference/environments.md](reference/environments.md)
- For deployment history, see [reference/changelog.md](reference/changelog.md)
