package zendesk

import (
	"fmt"
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
	Public      bool         `json:"public,omitempty"`
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
	AttachmentID int64       `json:"id"`
	FileName     string      `json:"file_name"`
	ContentURL   string      `json:"content_url"`
	ContentType  string      `json:"content_type"`
	Size         int64       `json:"size"`
	Inline       bool        `json:"inline,omitempty"`
	Thumbnails   []Thumbnail `json:"thumbnails"`
}

type Thumbnail struct {
	ThumbnailID int    `json:"id"`
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
