package zendesk

import (
	"bytes"
	"fmt"
	"log"
	"strconv"
	"time"
)

// TicketComment represents a Zendesk Ticket Comment.
//
// Zendesk Core API docs: https://developer.zendesk.com/rest_api/docs/core/ticket_comments

type TicketComment struct {
	ID          int64        `json:"id,omitempty"`
	Type        string       `json:"type,omitempty"`
	Body        string       `json:"body,omitempty"`
	HTMLBody    string       `json:"html_body,omitempty"`
	PlainBody   string       `json:"plain_body,omitempty"`
	Public      bool         `json:"public"`
	AuthorID    int64        `json:"author_id,omitempty"`
	Attachments []Attachment `json:"attachments,omitempty"`
	Via         *Via         `json:"via,omitempty"`
	MetaData    interface{}  `json:"metadata,omitempty"`
	CreatedAt   *time.Time   `json:"created_at,omitempty"`
	Uploads     []string     `json:"uploads,omitempty"`
}

// Attachment represents a Zendesk attachment for tickets and forum posts.
//
// Zendesk Core API docs: https://developer.zendesk.com/rest_api/docs/core/attachments

type Attachment struct {
	ID          int64       `json:"id"`
	FileName    string      `json:"file_name"`
	ContentURL  string      `json:"content_url"`
	ContentType string      `json:"content_type"`
	Size        int64       `json:"size"`
	Inline      bool        `json:"inline,omitempty"`
	Thumbnails  []Thumbnail `json:"thumbnails"`
}

type Thumbnail struct {
	ID          int    `json:"id"`
	FileName    string `json:"file_name"`
	ContentURL  string `json:"content_url"`
	ContentType string `json:"content_type"`
	Size        int    `json:"size"`
}

type Via struct {
	Channel *string `json:"channel"`
	Source  *Flow   `json:"source"`
}
type Flow struct {
	To   *ToObject   `json:"to"`
	From *FromObject `json:"from"`
	Rel  string      `json:"rel"`
}

type ToObject struct {
}

type FromObject struct {
	ID    int    `json:"id"`
	Title string `json:"title"`
}

func (c *client) ListTicketComments(id int64) ([]TicketComment, error) {
	out := new(APIPayload)
	err := c.get(fmt.Sprintf("/api/v2/tickets/%d/comments.json", id), out)
	return out.Comments, err
}

func (c *client) GetAllTicketComments(ticketIDs []int64) (map[int64][]TicketComment, error) {
	log.Printf("[ZENDESK] Start GetAllTicketComments")
	ticketCommentsMap, err := c.getTicketCommentsOneByOne(nil, ticketIDs)
	if err != nil {
		return nil, err
	}
	log.Printf("[ZENDESK] number of ticket comments: %v", len(ticketCommentsMap))
	log.Printf("[ZENDESK] End GetAllTicketComments")
	return ticketCommentsMap, nil
}

// getTicketCommentOneByOne return a map with ticket id as the key and
// an array of ticket comments as its value
func (c *client) getTicketCommentsOneByOne(in interface{}, ticketIDs []int64) (map[int64][]TicketComment, error) {
	log.Printf("[ZENDESK] Start getTicketCommentsOneByOne")
	endpointPrefix := "/api/v2/tickets/"
	endpointPostfix := "/comments.json"

	result := make(map[int64][]TicketComment)
	payload, err := marshall(in)
	if err != nil {
		return nil, err
	}

	headers := map[string]string{}
	if in != nil {
		headers["Content-Type"] = "application/json"
	}
	record := new(APIPayload)

	numTickets := len(ticketIDs)
	if numTickets == 0 {
		return result, nil
	}
	log.Printf("[ZENDESK] numTickets: %v", numTickets)

	endpoint := fmt.Sprintf("%s%v%s", endpointPrefix, ticketIDs[0], endpointPostfix)
	res, err := c.request("GET", endpoint, headers, bytes.NewReader(payload))
	defer res.Body.Close()

	var totalWaitTime int64
	log.Printf("[ZENDESK] Start for loop in getTicketCommentsOneByOne")
	for ticketInd := 1; ticketInd < numTickets; ticketInd++ {
		// handle page not found
		if res.StatusCode == 404 {
			log.Printf("[ZENDESK] 404 not found: %s\n", endpoint)
			// handle too many requests (rate limit)
		} else if res.StatusCode == 429 {
			after, err := strconv.ParseInt(res.Header.Get("Retry-After"), 10, 64)
			log.Printf("[ZENDESK] too many requests. Wait for %v seconds\n", after)
			totalWaitTime += after
			if err != nil {
				return nil, err
			}
			time.Sleep(time.Duration(after) * time.Second)
			continue
		} else {
			err = unmarshall(res, record)
			if err != nil {
				return nil, err
			}
			result[ticketIDs[ticketInd-1]] = record.Comments
		}

		record = new(APIPayload)
		endpoint = fmt.Sprintf("%s%v%s", endpointPrefix, ticketIDs[ticketInd], endpointPostfix)
		res, _ = c.request("GET", endpoint, headers, bytes.NewReader(payload))
	}

	log.Printf("[ZENDESK] number of records pulled: %v\n", len(result))
	log.Printf("[ZENDESK] total waiting time due to rate limit: %v\n", totalWaitTime)
	return result, nil
}
