package watch

import "fmt"

// TPUChipMetrics represents metrics for a single TPU chip
type TPUChipMetrics struct {
	DeviceID     int     `json:"device_id"`
	MemoryUsage  int64   `json:"memory_usage"`
	TotalMemory  int64   `json:"total_memory"`
	DutyCyclePct float64 `json:"duty_cycle_pct"`
	ChipType     string  `json:"chip_type"`
}

// TPUHostMetrics represents aggregated metrics for all chips on a single host
type TPUHostMetrics struct {
	ChipType  string           `json:"chip_type"`
	ChipCount int              `json:"chip_count"`
	Chips     []TPUChipMetrics `json:"chips"`
	Error     string           `json:"error,omitempty"`
}

// MemoryUsageGB returns memory usage in gigabytes
func (m *TPUHostMetrics) MemoryUsageGB() float64 {
	var total int64
	for _, chip := range m.Chips {
		total += chip.MemoryUsage
	}
	return float64(total) / 1e9
}

// TotalMemoryGB returns total memory in gigabytes
func (m *TPUHostMetrics) TotalMemoryGB() float64 {
	var total int64
	for _, chip := range m.Chips {
		total += chip.TotalMemory
	}
	return float64(total) / 1e9
}

// AvgUtilization returns average duty cycle across all chips
func (m *TPUHostMetrics) AvgUtilization() float64 {
	if len(m.Chips) == 0 {
		return 0
	}
	var total float64
	for _, chip := range m.Chips {
		total += chip.DutyCyclePct
	}
	return total / float64(len(m.Chips))
}

// ClusterTPUMetrics represents aggregated TPU metrics across the entire cluster
type ClusterTPUMetrics struct {
	Hosts          map[string]TPUHostMetrics
	TotalChips     int
	TotalMemoryGB  float64
	UsedMemoryGB   float64
	AvgUtilization float64
	ActiveHosts    int
	ErrorHosts     int
}

// NewClusterTPUMetrics creates a new cluster metrics aggregator
func NewClusterTPUMetrics() *ClusterTPUMetrics {
	return &ClusterTPUMetrics{
		Hosts: make(map[string]TPUHostMetrics),
	}
}

// AddHost adds a host's metrics and updates cluster aggregates
func (c *ClusterTPUMetrics) AddHost(host string, metrics TPUHostMetrics) {
	c.Hosts[host] = metrics

	if metrics.Error != "" {
		c.ErrorHosts++
		return
	}

	c.ActiveHosts++
	c.TotalChips += metrics.ChipCount
	c.TotalMemoryGB += metrics.TotalMemoryGB()
	c.UsedMemoryGB += metrics.MemoryUsageGB()

	for _, chip := range metrics.Chips {
		c.AvgUtilization += chip.DutyCyclePct
	}
}

// Finalize computes final averages
func (c *ClusterTPUMetrics) Finalize() {
	if c.TotalChips > 0 {
		c.AvgUtilization /= float64(c.TotalChips)
	}
}

// MemoryUtilizationPct returns memory utilization as percentage
func (c *ClusterTPUMetrics) MemoryUtilizationPct() float64 {
	if c.TotalMemoryGB == 0 {
		return 0
	}
	return 100 * c.UsedMemoryGB / c.TotalMemoryGB
}

// TotalHosts returns total number of hosts
func (c *ClusterTPUMetrics) TotalHosts() int {
	return len(c.Hosts)
}

// String returns a summary string
func (c *ClusterTPUMetrics) String() string {
	return fmt.Sprintf("Cluster: %d chips, %.1f%% util, %d/%d hosts active",
		c.TotalChips, c.AvgUtilization, c.ActiveHosts, c.TotalHosts())
}
