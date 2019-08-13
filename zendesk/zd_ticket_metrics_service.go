package zendesk

import (
	"fmt"
	"time"
)

// TicketMetric represents a Zendesk TicketMetric.
//
// Zendesk Core API docs: https://developer.zendesk.com/rest_api/docs/core/ticket_metrics

type TicketMetric struct {
	ID                   int64      `json:"id,omitempty"`
	TicketID             int64      `json:"ticket_id,omitempty"`
	URL                  string     `json:"url,omitempty"`
	GroupStations        int64      `json:"group_stations,omitempty"`
	AssigneeStations     int64      `json:"assignee_stations,omitempty"`
	Reopens              int64      `json:"reopens,omitempty"`
	Replies              int64      `json:"replies,omitempty"`
	AssigneeUpdatedAt    *time.Time `json:"assignee_updated_at,omitempty"`
	RequesterUpdatedAt   *time.Time `json:"requester_updated_at,omitempty"`
	StatusUpdatedAt      *time.Time `json:"status_updated_at,omitempty"`
	InitiallyAssigneeAt  *time.Time `json:"initially_assigned_at,omitempty"`
	AssignedAt           *time.Time `json:"assigned_at,omitempty"`
	SolvedAt             *time.Time `json:"solved_at,omitempty"`
	LatestCommentAddedAt *time.Time `json:"latest_comment_added_at,omitempty"`
	FirstResolutionTime  Object     `json:"first_resolution_time_in_minutes,omitempty"`
	ReplyTime            Object     `json:"reply_time_in_minutes,omitempty"`
	FullResolutionTime   Object     `json:"full_resolution_time_in_minutes,omitempty"`
	AgentWaitTime        Object     `json:"agent_wait_time_in_minutes,omitempty"`
	RequesterWaitTime    Object     `json:"requester_wait_time_in_minutes,omitempty"`
	CreatedAt            *time.Time `json:"created_at,omitempty"`
	UpdatedAt            *time.Time `json:"updated_at,omitempty"`
}

type Object struct {
	Calendar int64 `json:"calendar"`
	Business int64 `json:"business"`
}

func (c *client) GetAllTicketMetrics() ([]TicketMetric, error) {
	out := new(APIPayload)
	err := c.get("/api/v2/ticket_metrics.json", out)
	return out.TicketMetrics, err
}

func (c *client) ShowTicketMetric(id int64) (*TicketMetric, error) {
	out := new(APIPayload)
	err := c.get(fmt.Sprintf("/api/v2/ticket_metrics/%d.json", id), out)
	return out.TicketMetric, err
}
