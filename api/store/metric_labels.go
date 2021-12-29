package store

import (
	"github.com/prometheus/common/model"
)

type MetricLabels map[string]int

var MetricLabel = MetricLabels{
	//cpu
	"container_cpu_user_seconds_total":   0,
	"container_cpu_system_seconds_total": 0,
	"container_cpu_usage_seconds_total":  0,
	//memory
	"container_memory_cache":           0,
	"container_memory_swap":            0,
	"container_memory_usage_bytes":     0,
	"container_memory_max_usage_bytes": 0,
	//disk
	"container_fs_io_time_seconds_total":          0,
	"container_fs_io_time_weighted_seconds_total": 0,
	"container_fs_writes_bytes_total":             0,
	"container_fs_reads_bytes_total":              0,
	//network
	"container_network_receive_bytes_total":   0,
	"container_network_receive_errors_total":  0,
	"container_network_transmit_bytes_total":  0,
	"container_network_transmit_errors_total": 0,
}

func (l *MetricLabels) IsValid(s model.Sample) bool {
	if _, found := MetricLabel[string(s.Metric["__name__"])]; found {
		return true
	}
	return false
}
