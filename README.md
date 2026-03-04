# k8s-helper

Kubernetes debugging and diagnostics helper. Common kubectl commands wrapped in a friendly CLI.

```bash
go install github.com/Anuar-boop/k8s-helper@latest
```

## Quick Start

```bash
k8s-helper status                      # Cluster overview
k8s-helper health                      # Find unhealthy pods
k8s-helper debug my-pod -n production  # Full pod diagnostics
k8s-helper events -n staging           # Recent events
k8s-helper resources                   # CPU/memory usage
```

## Commands

| Command | Description |
|---------|-------------|
| `status` | Show nodes, pods, and services |
| `health` | Find unhealthy pods and high restart counts |
| `logs <pod>` | View pod logs (supports `-f` and `--tail`) |
| `events` | Show recent cluster events sorted by time |
| `resources` | Show CPU/memory usage for nodes and pods |
| `debug <pod>` | Full diagnostics: describe + logs + events |
| `ingress` | Show ingress routes |
| `secrets` | List secrets (names only, values hidden) |
| `configmaps` | List configmaps |
| `cleanup` | Find completed/failed pods for cleanup |

## Options

| Flag | Description |
|------|-------------|
| `-n <namespace>` | Filter by namespace (default: all) |
| `-c <container>` | Specify container for logs |
| `-f` | Follow log output |
| `--tail <n>` | Number of log lines |
| `--dry-run` | Preview cleanup |

## Features

- One-command cluster overview
- Pod health checker (unhealthy, restart counts)
- Full pod debugging (describe + logs + events)
- Resource usage monitoring
- Cleanup finder for completed/failed pods
- Works with any kubectl-configured cluster
- Zero dependencies (pure Go + kubectl)

## License

MIT
