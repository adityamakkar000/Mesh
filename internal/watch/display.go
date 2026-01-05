package watch

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/adityamakkar000/Mesh/internal/ui"
)

// Display renders cluster TPU metrics to stdout
func Display(metrics *ClusterTPUMetrics, clusterName string) {
	clearScreen()
	displayHeader(clusterName)
	displaySummary(metrics)
	displayHostTable(metrics)
}

func clearScreen() {
	fmt.Print("\033[H\033[2J")
}

func displayHeader(clusterName string) {
	timestamp := time.Now().Format("15:04:05")
	ui.Header(fmt.Sprintf("TPU Cluster '%s' - %s", clusterName, timestamp))
	fmt.Println()
}

func displaySummary(m *ClusterTPUMetrics) {
	fmt.Printf("%s%sCluster Summary%s\n", ui.Bold, ui.Cyan, ui.Reset)
	fmt.Printf("  Total Chips:     %s%d%s\n",
		ui.Bold, m.TotalChips, ui.Reset)
	fmt.Printf("  Avg Utilization: %s%.1f%%%s\n",
		utilizationColor(m.AvgUtilization), m.AvgUtilization, ui.Reset)
	fmt.Printf("  Memory Usage:    %.1f / %.1f GB (%s%.1f%%%s)\n",
		m.UsedMemoryGB, m.TotalMemoryGB,
		utilizationColor(m.MemoryUtilizationPct()),
		m.MemoryUtilizationPct(), ui.Reset)
	fmt.Printf("  Active Hosts:    %s%d%s / %d",
		ui.Green, m.ActiveHosts, ui.Reset, m.TotalHosts())

	if m.ErrorHosts > 0 {
		fmt.Printf(" (%s%d error%s%s)",
			ui.Red, m.ErrorHosts, ui.Reset, plural(m.ErrorHosts))
	}
	fmt.Println("\n")
}

func displayHostTable(m *ClusterTPUMetrics) {
	// Table header
	fmt.Printf("%-18s %-6s %-8s %-12s %-15s %s\n",
		"HOST", "CHIPS", "TYPE", "UTIL", "MEMORY", "STATUS")
	fmt.Println(strings.Repeat("─", 80))

	// Sort hosts for consistent display
	hosts := make([]string, 0, len(m.Hosts))
	for host := range m.Hosts {
		hosts = append(hosts, host)
	}
	sort.Strings(hosts)

	// Display each host
	for _, host := range hosts {
		metrics := m.Hosts[host]
		displayHostRow(host, metrics)
	}
}

func displayHostRow(host string, m TPUHostMetrics) {
	if m.Error != "" {
		fmt.Printf("%-18s %s%-6s%s %-8s %-12s %-15s %s%s%s\n",
			host,
			ui.Red, "—", ui.Reset,
			"—",
			"—",
			"—",
			ui.Red, m.Error, ui.Reset)
		return
	}

	utilization := m.AvgUtilization()
	memUsed := m.MemoryUsageGB()
	memTotal := m.TotalMemoryGB()

	status := statusIcon(utilization)

	fmt.Printf("%-18s %-6d %-8s %s%-11.1f%%%s %-15s %s\n",
		host,
		m.ChipCount,
		m.ChipType,
		utilizationColor(utilization),
		utilization,
		ui.Reset,
		fmt.Sprintf("%.1f/%.1f GB", memUsed, memTotal),
		status)
}

func utilizationColor(pct float64) string {
	switch {
	case pct >= 90:
		return ui.Red
	case pct >= 70:
		return ui.Yellow
	case pct >= 40:
		return ui.Green
	default:
		return ui.Cyan
	}
}

func statusIcon(utilization float64) string {
	switch {
	case utilization >= 80:
		return ui.Green + "● active" + ui.Reset
	case utilization >= 10:
		return ui.Yellow + "◐ partial" + ui.Reset
	default:
		return ui.Cyan + "○ idle" + ui.Reset
	}
}

func plural(n int) string {
	if n == 1 {
		return ""
	}
	return "s"
}
