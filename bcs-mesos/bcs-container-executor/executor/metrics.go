package executor

import (
	"github.com/prometheus/client_golang/prometheus"
)

// Metrics the scheduler info
var (
	executorSlaveConnection = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: "bkbcs",
		Subsystem: "executor",
		Name:      "slave_connection",
		Help:      "executor slave connection",
	})

	taskgroupReportTotal = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: "bkbcs",
		Subsystem: "executor",
		Name:      "taskgroup_status_report_total",
		Help:      "report taskgroup status total",
	}, []string{"taskgroup"})

	taskgroupAckTotal = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: "bkbcs",
		Subsystem: "executor",
		Name:      "taskgroup_status_ack_total",
		Help:      "ack  taskgroup status total",
	}, []string{"taskgroup"})
)

func init() {
	prometheus.MustRegister(executorSlaveConnection, taskgroupAckTotal, taskgroupReportTotal)
}
