package watch

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"sync"

	"github.com/adityamakkar000/Mesh/internal/parse"
	"github.com/adityamakkar000/Mesh/internal/ssh"
)

const tpuMetricsScript = `python3 -c "
import json
import sys
try:
    from tpu_info import device, metrics
    chips = device.get_local_chips()
    if not chips:
        print(json.dumps({'error': 'no TPU chips found'}))
        sys.exit(1)
    chip_type = chips[0]['type']
    usage = metrics.get_chip_usage(chip_type)
    print(json.dumps({
        'chip_type': chip_type,
        'chip_count': len(chips),
        'chips': usage
    }))
except ImportError as e:
    print(json.dumps({'error': f'tpu-info not installed: {e}'}))
    sys.exit(1)
except Exception as e:
    print(json.dumps({'error': str(e)}))
    sys.exit(1)
"`

type hostResult struct {
	host    string
	metrics TPUHostMetrics
}

// CollectTPUMetrics gathers TPU metrics from all hosts in the cluster
func CollectTPUMetrics(ctx context.Context, cluster parse.NodeConfig) *ClusterTPUMetrics {
	var wg sync.WaitGroup
	resultChan := make(chan hostResult, len(cluster.Hosts))

	for _, host := range cluster.Hosts {
		wg.Add(1)
		go func(h string) {
			defer wg.Done()
			metrics := collectHostMetrics(ctx, cluster, h)
			resultChan <- hostResult{h, metrics}
		}(host)
	}

	go func() {
		wg.Wait()
		close(resultChan)
	}()

	clusterMetrics := NewClusterTPUMetrics()
	for result := range resultChan {
		clusterMetrics.AddHost(result.host, result.metrics)
	}

	clusterMetrics.Finalize()
	return clusterMetrics
}

func collectHostMetrics(ctx context.Context, cluster parse.NodeConfig, host string) TPUHostMetrics {
	client, err := ssh.Connect(ctx, cluster.User, host, cluster.IdentityFile)
	if err != nil {
		return TPUHostMetrics{Error: fmt.Sprintf("connection failed: %v", err)}
	}
	defer client.Close()

	var buf bytes.Buffer
	if err := client.Exec(ctx, tpuMetricsScript, &buf, io.Discard); err != nil {
		return TPUHostMetrics{Error: fmt.Sprintf("exec failed: %v", err)}
	}

	var metrics TPUHostMetrics
	if err := json.Unmarshal(buf.Bytes(), &metrics); err != nil {
		return TPUHostMetrics{Error: fmt.Sprintf("parse failed: %v", err)}
	}

	return metrics
}
