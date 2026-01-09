package telemetry

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// AgentJobsExecutedTotal counts total jobs executed by agent
	AgentJobsExecutedTotal = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "agent_jobs_executed_total",
			Help: "Total number of jobs executed by agent",
		},
	)
	
	// AgentJobsFailedTotal counts failed jobs
	AgentJobsFailedTotal = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "agent_jobs_failed_total",
			Help: "Total number of failed jobs",
		},
	)
)
