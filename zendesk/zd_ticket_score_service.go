package zendesk

import (
	"bytes"
	"fmt"
	"log"
	"strconv"
	"time"
)

// TicketMetric represents a Zendesk ticket satisfaction rating.
//
// Zendesk Core API docs: https://developer.zendesk.com/rest_api/docs/support/satisfaction_ratings

type Score struct {
	ID          int64      `json:"id,omitempty"`
	AssigneeID  int64      `json:"assignee_id,omitempty`
	GroupID     int64      `json:"group_id,omitempty"`
	RequesterID int64      `json:"requester_id,omitempty"`
	TicketID    int64      `json:"ticket_id,omitempty"`
	Score       string     `json:"score,omitempty"`
	CreatedAt   *time.Time `json:"created_at,omitempty"`
	UpdatedAt   *time.Time `json:"updated_at,omitempty"`
}

// GetSatisfactionScores pull the list of all the scores
// due to memory limit, we need to pull by page
//
// Zendesk Core API docs: https://developer.zendesk.com/rest_api/docs/support/satisfaction_ratings

func (c *client) GetSatisfactionScores() ([]Score, error) {
	scores, err := c.getSatisfactionScores("/api/v2/satisfaction_ratings.json?page=", nil)
	return scores, err
}

func (c *client) getSatisfactionScores(endpoint string, in interface{}) ([]Score, error) {
	// startingPageNumber will be adjusted while pulling
	startingPageNumber := 1

	result := make([]Score, 0)
	payload, err := marshall(in)
	if err != nil {
		return nil, err
	}

	headers := map[string]string{}
	if in != nil {
		headers["Content-Type"] = "applications/json"
	}

	currentPage := fmt.Sprintf("%s%v", endpoint, startingPageNumber)
	res, err := c.request("GET", currentPage, headers, bytes.NewReader(payload))
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	// numberOfPages will be customized when pulling data
	numberOfPages := 50
	count := 1

	// APIPayload defined the fields received from Zendesk
	dataPerPage := new(APIPayload)
	err = unmarshall(res, dataPerPage)
	if err != nil {
		return nil, err
	}

	var totalWaitTime int64
	for count < numberOfPages && currentPage != "" {
		// if too many requests(res.StatusCode == 429), delay sending request
		if res.StatusCode == 429 {
			after, err := strconv.ParseInt(res.Header.Get("Retry-After"), 10, 64)
			if err != nil {
				return nil, err
			}

			log.Printf("[zd_ticket_score_service][getSatisfactionScores] too many requests. Wait for %v seconds\n", after)
			totalWaitTime += after
			time.Sleep(time.Duration(after) * time.Second)
		} else {
			result = append(result, dataPerPage.SatisfactionRatings...)
			currentPage = dataPerPage.NextPage
			log.Printf("[zd_ticket_score_service][getSatisfactionScores] pulling page: %s\n", currentPage)
		}

		currentPage = fmt.Sprintf("%s%v", endpoint, startingPageNumber+count)
		count++
		res, _ = c.request("GET", currentPage, headers, bytes.NewReader(payload))
		dataPerPage = new(APIPayload)
		err = unmarshall(res, dataPerPage)
		if err != nil {
			return nil, err
		}
	}

	log.Printf("[zd_ticket_score_service][getSatisfactionScores] number of records pulled: %v\n", len(result))
	log.Printf("[zd_ticket_score_service][getSatisfactionScores] total waiting time due to rate limit: %v\n", totalWaitTime)

	return result, err
}
