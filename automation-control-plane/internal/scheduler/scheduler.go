package scheduler

import (
	"context"
	"fmt"
	"time"

	"github.com/automation-platform/control-plane/internal/centrifugo"
	"github.com/automation-platform/control-plane/internal/store/mysql"
	"github.com/automation-platform/control-plane/internal/store/redis"
)

// Scheduler schedules jobs to agents
type Scheduler struct {
	mysqlStore    *mysql.Store
	redisStore    *redis.Store
	centrifugo    *centrifugo.Client
	presence      *redis.AgentPresence
	maxConcurrent int
	fanOutCap     int
}

// Config holds scheduler configuration
type Config struct {
	MySQLStore    *mysql.Store
	RedisStore    *redis.Store
	Centrifugo    *centrifugo.Client
	MaxConcurrent int // Max concurrent jobs per agent
	FanOutCap     int // Max agents to notify per job
}

// NewScheduler creates a new scheduler
func NewScheduler(cfg Config) *Scheduler {
	return &Scheduler{
		mysqlStore:    cfg.MySQLStore,
		redisStore:    cfg.RedisStore,
		centrifugo:    cfg.Centrifugo,
		presence:      redis.NewAgentPresence(cfg.RedisStore),
		maxConcurrent: cfg.MaxConcurrent,
		fanOutCap:     cfg.FanOutCap,
	}
}

// JobTarget represents job targeting configuration
type JobTarget struct {
	Type  string            `json:"type"` // agent, label, query, any
	Value string            `json:"value,omitempty"`
	Labels map[string]string `json:"labels,omitempty"`
}

// Schedule schedules a job to eligible agents
func (s *Scheduler) Schedule(ctx context.Context, job *mysql.Job, target JobTarget) error {
	// 1. Resolve eligible agents from Redis presence
	agents, err := s.presence.ListAgents(ctx, job.TenantID, job.ProjectID)
	if err != nil {
		return fmt.Errorf("failed to list agents: %w", err)
	}

	if len(agents) == 0 {
		return fmt.Errorf("no agents available for tenant %s, project %s", job.TenantID, job.ProjectID)
	}

	// 2. Filter by state and capabilities
	eligible := s.filterIdle(ctx, job.TenantID, job.ProjectID, agents)

	// 3. Apply targeting rules
	candidates := s.applyTarget(target, eligible, job.TenantID, job.ProjectID)

	// 4. Apply backpressure & fan-out caps
	selected := s.limitFanout(candidates, s.fanOutCap)

	if len(selected) == 0 {
		return fmt.Errorf("no eligible agents after filtering")
	}

	// 5. Notify via Centrifugo
	message := centrifugo.JobAvailableMessage{
		Type:  "job_available",
		JobID: job.JobID,
	}

	for _, agentID := range selected {
		channel := fmt.Sprintf("agents.%s.%s", job.TenantID, agentID)
		if err := s.centrifugo.Publish(ctx, channel, message); err != nil {
			// Log error but continue with other agents
			fmt.Printf("Failed to notify agent %s: %v\n", agentID, err)
		}
	}

	return nil
}

// filterIdle filters agents that are idle (not currently executing jobs)
func (s *Scheduler) filterIdle(ctx context.Context, tenantID, projectID string, agentIDs []string) []string {
	// Check Redis for active job counts per agent
	// For simplicity, we assume agents with presence are idle
	// In production, you'd track active job counts
	idle := make([]string, 0)
	
	for _, agentID := range agentIDs {
		present, err := s.presence.IsAgentPresent(ctx, tenantID, projectID, agentID)
		if err == nil && present {
			// Check if agent has active jobs (simplified - would check actual job count)
			// For now, assume present agents are idle
			idle = append(idle, agentID)
		}
	}
	
	return idle
}

// applyTarget applies targeting rules to filter agents
func (s *Scheduler) applyTarget(target JobTarget, agents []string, tenantID, projectID string) []string {
	switch target.Type {
	case "agent":
		// Target specific agent
		for _, agentID := range agents {
			if agentID == target.Value {
				return []string{agentID}
			}
		}
		return []string{}
	
	case "label":
		// Target agents with specific labels
		// TODO: Implement label matching from agent metadata in Redis
		// For now, return all agents
		return agents
	
	case "query":
		// Target agents matching a query
		// TODO: Implement query parsing and matching
		return agents
	
	case "any":
		// Target any available agent
		return agents
	
	default:
		return agents
	}
}

// limitFanout limits the number of agents to notify
func (s *Scheduler) limitFanout(agents []string, cap int) []string {
	if len(agents) <= cap {
		return agents
	}
	return agents[:cap]
}

// CleanupExpiredLeases cleans up expired job leases
func (s *Scheduler) CleanupExpiredLeases(ctx context.Context) error {
	// Query for jobs with expired leases
	// This would be called periodically by a background goroutine
	// Implementation would update jobs with expired leases back to 'pending'
	// For now, this is a placeholder
	return nil
}

// Start starts the scheduler background processes
func (s *Scheduler) Start(ctx context.Context) error {
	// Start background goroutine for lease cleanup
	go func() {
		ticker := time.NewTicker(1 * time.Minute)
		defer ticker.Stop()
		
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				if err := s.CleanupExpiredLeases(ctx); err != nil {
					fmt.Printf("Lease cleanup error: %v\n", err)
				}
			}
		}
	}()
	
	return nil
}
