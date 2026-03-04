// k8s-helper — Kubernetes debugging and diagnostics helper
//
// Common Kubernetes debugging commands wrapped in a friendly CLI.
// No dependencies beyond kubectl.

package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

const version = "1.0.0"

func kubectl(args ...string) (string, error) {
	cmd := exec.Command("kubectl", args...)
	out, err := cmd.CombinedOutput()
	return string(out), err
}

func kubectlPrint(args ...string) {
	cmd := exec.Command("kubectl", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Run()
}

func cmdStatus(namespace string) {
	fmt.Printf("\n  \x1b[1mCluster Status\x1b[0m\n\n")

	// Nodes
	fmt.Printf("  \x1b[36mNodes:\x1b[0m\n")
	kubectlPrint("get", "nodes", "-o", "wide")

	// Pods
	fmt.Printf("\n  \x1b[36mPods:\x1b[0m\n")
	if namespace == "" {
		kubectlPrint("get", "pods", "--all-namespaces", "-o", "wide")
	} else {
		kubectlPrint("get", "pods", "-n", namespace, "-o", "wide")
	}

	// Services
	fmt.Printf("\n  \x1b[36mServices:\x1b[0m\n")
	if namespace == "" {
		kubectlPrint("get", "svc", "--all-namespaces")
	} else {
		kubectlPrint("get", "svc", "-n", namespace)
	}
}

func cmdPodHealth(namespace string) {
	fmt.Printf("\n  \x1b[1mPod Health Check\x1b[0m\n\n")

	args := []string{"get", "pods", "-o", "json"}
	if namespace != "" {
		args = append(args, "-n", namespace)
	} else {
		args = append(args, "--all-namespaces")
	}

	out, err := kubectl(args...)
	if err != nil {
		fmt.Fprintf(os.Stderr, "  Error: %s\n", out)
		return
	}

	// Show unhealthy pods
	fmt.Printf("  \x1b[36mUnhealthy Pods:\x1b[0m\n")
	filterArgs := []string{"get", "pods", "--field-selector", "status.phase!=Running,status.phase!=Succeeded"}
	if namespace != "" {
		filterArgs = append(filterArgs, "-n", namespace)
	} else {
		filterArgs = append(filterArgs, "--all-namespaces")
	}
	kubectlPrint(filterArgs...)

	// Show restart counts
	fmt.Printf("\n  \x1b[36mPods with restarts:\x1b[0m\n")
	restartArgs := []string{"get", "pods", "-o", "custom-columns=NAMESPACE:.metadata.namespace,NAME:.metadata.name,RESTARTS:.status.containerStatuses[*].restartCount,STATUS:.status.phase"}
	if namespace != "" {
		restartArgs = append(restartArgs, "-n", namespace)
	} else {
		restartArgs = append(restartArgs, "--all-namespaces")
	}
	kubectlPrint(restartArgs...)
}

func cmdLogs(podName, namespace, container string, follow bool, lines int) {
	args := []string{"logs", podName}
	if namespace != "" {
		args = append(args, "-n", namespace)
	}
	if container != "" {
		args = append(args, "-c", container)
	}
	if follow {
		args = append(args, "-f")
	}
	if lines > 0 {
		args = append(args, "--tail", fmt.Sprintf("%d", lines))
	}
	kubectlPrint(args...)
}

func cmdEvents(namespace string) {
	fmt.Printf("\n  \x1b[1mRecent Events\x1b[0m\n\n")

	args := []string{"get", "events", "--sort-by=.lastTimestamp"}
	if namespace != "" {
		args = append(args, "-n", namespace)
	} else {
		args = append(args, "--all-namespaces")
	}
	kubectlPrint(args...)
}

func cmdResources(namespace string) {
	fmt.Printf("\n  \x1b[1mResource Usage\x1b[0m\n\n")

	fmt.Printf("  \x1b[36mNode Resources:\x1b[0m\n")
	kubectlPrint("top", "nodes")

	fmt.Printf("\n  \x1b[36mPod Resources:\x1b[0m\n")
	if namespace != "" {
		kubectlPrint("top", "pods", "-n", namespace)
	} else {
		kubectlPrint("top", "pods", "--all-namespaces")
	}
}

func cmdDebug(podName, namespace string) {
	fmt.Printf("\n  \x1b[1mDebugging pod: %s\x1b[0m\n\n", podName)

	args := func(extra ...string) []string {
		base := []string{}
		if namespace != "" {
			base = append(base, "-n", namespace)
		}
		return append(base, extra...)
	}

	// Describe
	fmt.Printf("  \x1b[36mPod Description:\x1b[0m\n")
	kubectlPrint(append([]string{"describe", "pod", podName}, args()...)...)

	// Logs
	fmt.Printf("\n  \x1b[36mRecent Logs (last 50 lines):\x1b[0m\n")
	kubectlPrint(append([]string{"logs", podName, "--tail=50"}, args()...)...)

	// Events for this pod
	fmt.Printf("\n  \x1b[36mRelated Events:\x1b[0m\n")
	kubectlPrint(append([]string{"get", "events", "--field-selector", fmt.Sprintf("involvedObject.name=%s", podName)}, args()...)...)
}

func cmdIngress(namespace string) {
	fmt.Printf("\n  \x1b[1mIngress Routes\x1b[0m\n\n")

	args := []string{"get", "ingress", "-o", "wide"}
	if namespace != "" {
		args = append(args, "-n", namespace)
	} else {
		args = append(args, "--all-namespaces")
	}
	kubectlPrint(args...)
}

func cmdSecrets(namespace string) {
	fmt.Printf("\n  \x1b[1mSecrets\x1b[0m (names only — values not shown)\n\n")

	args := []string{"get", "secrets"}
	if namespace != "" {
		args = append(args, "-n", namespace)
	} else {
		args = append(args, "--all-namespaces")
	}
	kubectlPrint(args...)
}

func cmdConfigMaps(namespace string) {
	fmt.Printf("\n  \x1b[1mConfigMaps\x1b[0m\n\n")

	args := []string{"get", "configmaps"}
	if namespace != "" {
		args = append(args, "-n", namespace)
	} else {
		args = append(args, "--all-namespaces")
	}
	kubectlPrint(args...)
}

func cmdCleanup(namespace string, dryRun bool) {
	fmt.Printf("\n  \x1b[1mCleanup\x1b[0m\n\n")

	// Completed pods
	fmt.Printf("  \x1b[36mCompleted Pods:\x1b[0m\n")
	args := []string{"get", "pods", "--field-selector", "status.phase=Succeeded"}
	if namespace != "" {
		args = append(args, "-n", namespace)
	} else {
		args = append(args, "--all-namespaces")
	}
	kubectlPrint(args...)

	// Failed pods
	fmt.Printf("\n  \x1b[36mFailed Pods:\x1b[0m\n")
	args = []string{"get", "pods", "--field-selector", "status.phase=Failed"}
	if namespace != "" {
		args = append(args, "-n", namespace)
	} else {
		args = append(args, "--all-namespaces")
	}
	kubectlPrint(args...)

	if !dryRun {
		fmt.Printf("\n  \x1b[33mUse --dry-run to preview. Not deleting without confirmation.\x1b[0m\n")
	}
}

func printUsage() {
	fmt.Printf(`
  k8s-helper v%s — Kubernetes debugging helper

  Usage:
    k8s-helper <command> [options]

  Commands:
    status           Show cluster status (nodes, pods, services)
    health           Check pod health (unhealthy, restarts)
    logs <pod>       Show pod logs
    events           Show recent events
    resources        Show resource usage (CPU, memory)
    debug <pod>      Full pod debug info (describe + logs + events)
    ingress          Show ingress routes
    secrets          List secrets (names only)
    configmaps       List configmaps
    cleanup          Find completed/failed pods for cleanup

  Options:
    -n <namespace>   Kubernetes namespace (default: all)
    -c <container>   Container name (for logs)
    -f               Follow logs
    --tail <n>       Number of log lines (default: all)
    --dry-run        Preview cleanup without deleting

  Examples:
    k8s-helper status
    k8s-helper health -n production
    k8s-helper logs my-pod -n default -f
    k8s-helper debug my-pod -n staging
    k8s-helper events -n kube-system
    k8s-helper resources
`, version)
}

func main() {
	args := os.Args[1:]

	if len(args) == 0 || args[0] == "--help" || args[0] == "help" {
		printUsage()
		return
	}

	if args[0] == "--version" {
		fmt.Println("k8s-helper", version)
		return
	}

	// Check kubectl
	if _, err := exec.LookPath("kubectl"); err != nil {
		fmt.Fprintln(os.Stderr, "  Error: kubectl not found in PATH")
		os.Exit(1)
	}

	command := args[0]
	namespace := ""
	container := ""
	follow := false
	tailLines := 0
	dryRun := false
	var positional []string

	for i := 1; i < len(args); i++ {
		switch args[i] {
		case "-n":
			if i+1 < len(args) {
				namespace = args[i+1]
				i++
			}
		case "-c":
			if i+1 < len(args) {
				container = args[i+1]
				i++
			}
		case "-f":
			follow = true
		case "--tail":
			if i+1 < len(args) {
				fmt.Sscanf(args[i+1], "%d", &tailLines)
				i++
			}
		case "--dry-run":
			dryRun = true
		default:
			if !strings.HasPrefix(args[i], "-") {
				positional = append(positional, args[i])
			}
		}
	}

	switch command {
	case "status":
		cmdStatus(namespace)
	case "health":
		cmdPodHealth(namespace)
	case "logs":
		if len(positional) == 0 {
			fmt.Fprintln(os.Stderr, "  Error: specify a pod name")
			os.Exit(1)
		}
		cmdLogs(positional[0], namespace, container, follow, tailLines)
	case "events":
		cmdEvents(namespace)
	case "resources", "top":
		cmdResources(namespace)
	case "debug":
		if len(positional) == 0 {
			fmt.Fprintln(os.Stderr, "  Error: specify a pod name")
			os.Exit(1)
		}
		cmdDebug(positional[0], namespace)
	case "ingress":
		cmdIngress(namespace)
	case "secrets":
		cmdSecrets(namespace)
	case "configmaps", "cm":
		cmdConfigMaps(namespace)
	case "cleanup":
		cmdCleanup(namespace, dryRun)
	default:
		fmt.Fprintf(os.Stderr, "  Unknown command: %s\n", command)
		os.Exit(1)
	}

	fmt.Println()
}
