package telemetry

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// JobsExecutedTotal counts total jobs executed
	JobsExecutedTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "jobs_executed_total",
			Help: "Total number of jobs executed",
		},
		[]string{"tenant_id", "project_id", "status"},
	)
	
	// JobsFailedTotal counts failed jobs
	JobsFailedTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "jobs_failed_total",
			Help: "Total number of failed jobs",
		},
		[]string{"tenant_id", "project_id"},
	)
	
	// SchedulerWakeupsTotal counts scheduler wakeups
	SchedulerWakeupsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "scheduler_wakeups_total",
			Help: "Total number of scheduler wakeups",
		},
		[]string{"tenant_id", "project_id"},
	)
	
	// ActiveAgents counts active agents
	ActiveAgents = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "active_agents",
			Help: "Number of active agents",
		},
		[]string{"tenant_id", "project_id"},
	)
)
