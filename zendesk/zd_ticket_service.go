package zendesk

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/google/go-querystring/query"
)

// Ticket represents a Zendesk Ticket.
//
// Zendesk Core API docs: https://developer.zendesk.com/rest_api/docs/core/tickets
type Ticket struct {
	ID                 int64          `json:"id,omitempty"`
	URL                string         `json:"url,omitempty"`
	ExternalID         string         `json:"external_id,omitempty"`
	Type               string         `json:"type,omitempty"`
	Subject            string         `json:"subject,omitempty"`
	RawSubject         string         `json:"raw_subject,omitempty"`
	Description        string         `json:"description,omitempty"`
	Priority           string         `json:"priority,omitempty"`
	Comment            *TicketComment `json:"comment,omitempty"`
	Status             string         `json:"status,omitempty"`
	Recipient          string         `json:"recipient,omitempty"`
	RequesterID        int64          `json:"requester_id,omitempty"`
	Requester          *User          `json:"requester,omitempty"`
	SubmitterID        int64          `json:"submitter_id,omitempty"`
	AssigneeID         int64          `json:"assignee_id,omitempty"`
	OrganizationID     int64          `json:"organization_id,omitempty"`
	GroupID            int64          `json:"group_id,omitempty"`
	CollaboratorIDs    []int64        `json:"collaborator_ids,omitempty"`
	EmailCCIDs         []int64        `json:"email_cc_ids,omitempty"`
	FollowerIDs        []int64        `json:"follower_ids,omitempty"`
	ForumTopicID       int64          `json:"forum_topic_id,omitempty"`
	ProblemID          int64          `json:"problem_id,omitempty"`
	HasIncidents       bool           `json:"has_incidents,omitempty"`
	DueAt              *time.Time     `json:"due_at,omitempty"`
	Tags               []string       `json:"tags,omitempty"`
	Via                *Via           `json:"via,omitempty"`
	CreatedAt          *time.Time     `json:"created_at,omitempty"`
	UpdatedAt          *time.Time     `json:"updated_at,omitempty"`
	CustomFields       []CustomField  `json:"custom_fields,omitempty"`
	SatisfactionRating *SAT           `json:"satisfaction_rating,omitempty"`
	BrandID            int64          `json:"brand_id,omitempty"`
	TicketFormID       int64          `json:"ticket_form_id,omitempty"`
	FollowupSourceID   int64          `json:"via_followup_source_id,omitempty"`
	IsPublic           bool           `json:"is_public"`
	AdditionalTags     []string       `json:"additional_tags,omitempty"`
	RemoveTags         []string       `json:"remove_tags,omitempty"`
}

type SAT struct {
	ID      int64  `json:"id"`
	Score   string `json:"score"`
	Comment string `json:"comment"`
}

type CustomField struct {
	ID    int64       `json:"id"`
	Value interface{} `json:"value"`
}

func (c *client) ShowTicket(id int64) (*Ticket, error) {
	out := new(APIPayload)
	err := c.get(fmt.Sprintf("/api/v2/tickets/%d.json", id), out)
	return out.Ticket, err
}

/*  The implementation below only works for no pagination case.

func (c *client) GetAllTickets() ([]Ticket, error) {
	out := new(APIPayload)
	err := c.get("/api/v2/tickets.json", out)
	return out.Tickets, err
}
*/

func (c *client) GetAllTickets() ([]Ticket, error) {
	tickets, err := c.getOneByOne(nil)
	return tickets, err
}

// GetTicketsIncrementally pull the list of tickets modified from a specific time point
//
// https://developer.zendesk.com/rest_api/docs/support/incremental_export
func (c *client) GetTicketsIncrementally(unixTime int64) ([]Ticket, error) {
	log.Printf("[zd_ticket_service][GetTicketsIncrementally] Start GetTicketsIncrementally")
	log.Printf("[zd_ticket_service][GetTicketsIncrementally] %s, %s", c.username, c.password)
	tickets, err := c.getTicketsIncrementally(unixTime, nil)
	log.Printf("[zd_ticket_service][GetTicketsIncrementally] Number of tickets: %v", len(tickets))
	return tickets, err
}

func (c *client) getTicketsIncrementally(unixTime int64, in interface{}) ([]Ticket, error) {
	log.Printf("[zd_ticket_service][getTicketsIncrementally] Start getTicketsIncrementally")
	log.Printf("[zd_ticket_service][getTicketsIncrementally] %s, %s", c.username, c.password)
	result := make([]Ticket, 0)
	payload, err := marshall(in)
	if err != nil {
		return nil, err
	}

	headers := map[string]string{}
	if in != nil {
		headers["Content-Type"] = "application/json"
	}

	apiV2 := "/api/v2/incremental/tickets.json?start_time="
	url := "https://philhelp.zendesk.com" + apiV2
	apiStartIndex := strings.Index(url, apiV2)
	endpoint := fmt.Sprintf("%s%v", apiV2, unixTime)

	res, err := c.request("GET", endpoint, headers, bytes.NewReader(payload))
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	dataPerPage := new(APIPayload)
	currentPage := "emptypage"
	var totalWaitTime int64
	log.Printf("[zd_ticket_service][getTicketsIncrementally] Start for loop in getTicketsIncrementally")
	for currentPage != dataPerPage.NextPage {
		// if too many requests(res.StatusCode == 429), delay sending request
		if res.StatusCode == 429 {
			after, err := strconv.ParseInt(res.Header.Get("Retry-After"), 10, 64)
			log.Printf("[zd_ticket_service][getTicketsIncrementally] too many requests. Wait for %v seconds\n", after)
			totalWaitTime += after
			if err != nil {
				return nil, err
			}
			time.Sleep(time.Duration(after) * time.Second)
			dataPerPage.NextPage = currentPage
		} else {
			err = unmarshall(res, dataPerPage)
			if err != nil {
				return nil, err
			}
			result = append(result, dataPerPage.Tickets...)
			if currentPage == dataPerPage.NextPage {
				break
			}
			currentPage = dataPerPage.NextPage
		}

		res, _ = c.request("GET", dataPerPage.NextPage[apiStartIndex:], headers, bytes.NewReader(payload))

		dataPerPage = new(APIPayload)
	}
	log.Printf("[zd_ticket_service][getTicketsIncrementally] number of records pulled: %v\n", len(result))
	log.Printf("[zd_ticket_service][getTicketsIncrementally] total waiting time due to rate limit: %v\n", totalWaitTime)

	return getUniqTickets(result), err
}

// getUniqTickets is to remove the duplicate records due to pagination
// more details can be found int the following link
// https://developer.zendesk.com/rest_api/docs/support/incremental_export#excluding_pagination_duplicates

func getUniqTickets(tickets []Ticket) []Ticket {
	var Empty struct{}
	keys := make(map[int64]struct{})
	result := make([]Ticket, 0)
	for _, ticket := range tickets {
		if _, ok := keys[ticket.ID]; ok {
			continue
		} else {
			keys[ticket.ID] = Empty
			result = append(result, ticket)
		}
	}
	return result
}

func (c *client) CreateTicket(ticket *Ticket) (*Ticket, error) {
	in := &APIPayload{Ticket: ticket}
	out := new(APIPayload)
	err := c.post("/api/v2/tickets.json", in, out)
	return out.Ticket, err
}

func (c *client) UpdateTicket(id int64, ticket *Ticket) (*Ticket, error) {
	in := &APIPayload{Ticket: ticket}
	out := new(APIPayload)
	err := c.put(fmt.Sprintf("/api/v2/tickets/%d.json", id), in, out)
	return out.Ticket, err
}

func (c *client) BatchUpdateManyTickets(tickets []Ticket) error {
	in := &APIPayload{Tickets: tickets}
	out := new(APIPayload)
	err := c.put("/api/v2/tickets/update_many.json", in, out)
	return err
}

func (c *client) BulkUpdateManyTickets(ids []int64, ticket *Ticket) error {
	parsed := []string{}
	for _, id := range ids {
		parsed = append(parsed, strconv.FormatInt(id, 10))
	}

	in := &APIPayload{Ticket: ticket}
	out := new(APIPayload)
	err := c.put(fmt.Sprintf("/api/v2/tickets/update_many.json?ids=%s", strings.Join(parsed, ",")), in, out)
	return err
}

func (c *client) ListRequestedTickets(userID int64) ([]Ticket, error) {
	out := new(APIPayload)
	err := c.get(fmt.Sprintf("/api/v2/users/%d/tickets/requested.json", userID), out)
	return out.Tickets, err
}

// ListTicketIncidents list all incidents related to the problem
func (c *client) ListTicketIncidents(problemID int64) ([]Ticket, error) {
	out := new(APIPayload)
	err := c.get(fmt.Sprintf("/api/v2/tickets/%d/incidents.json", problemID), out)

	return out.Tickets, err
}

// DeleteTickets deletes a Ticket.
//
// Zendesk Core API docs: https://developer.zendesk.com/rest_api/docs/core/tickets#delete-ticket
func (c *client) DeleteTicket(id int64) error {
	return c.delete(fmt.Sprintf("/api/v2/tickets/%d.json", id), nil)
}

// Upload represents a Zendesk file upload.
type Upload struct {
	Token       string       `json:"token"`
	Attachment  *Attachment  `json:"attachment"`
	Attachments []Attachment `json:"attachments"`
}

// ShowAttachment fetches an attachment by its ID.
//
// Zendesk Core API docs: https://developer.zendesk.com/rest_api/docs/core/attachments#getting-attachments
func (c *client) ShowAttachment(id int64) (*Attachment, error) {
	out := new(APIPayload)
	err := c.get(fmt.Sprintf("/api/v2/tickets/%d.json", id), out)
	return out.Attachment, err
}

// UploadFile uploads a file as a io.Reader.
//
// Zendesk Core API docs: https://developer.zendesk.com/rest_api/docs/core/attachments#uploading-files
func (c *client) UploadFile(filename string, token string, filecontent io.Reader) (*Upload, error) {
	params, err := query.Values(struct {
		Filename string `url:"filename"`
		Token    string `url:"token,omitempty"`
	}{filename, token})
	if err != nil {
		return nil, err
	}

	headers := map[string]string{
		"Content-Type": "application/binary",
	}

	res, err := c.request("POST", fmt.Sprintf("/api/v2/uploads.json?%s", params.Encode()), headers, filecontent)
	if err != nil {
		return nil, err
	}

	out := new(APIPayload)
	err = unmarshall(res, out)
	return out.Upload, err
}

type TicketForm struct {
	URL                string     `json:"url,omitempty"`
	ID                 int64      `json:"id,omitempty"`
	Name               string     `json:"name,omitempty"`
	RawName            string     `json:"raw_name,omitempty"`
	DisplayName        string     `json:"display_name,omitempty"`
	RawDisplayName     string     `json:"raw_display_name,omitempty"`
	EndUserVisible     bool       `json:"end_user_visible,omitempty"`
	Position           int64      `json:"position,omitempty"`
	TicketFieldIDs     []int64    `json:"ticket_field_ids,omitempty"`
	Active             bool       `json:"active,omitempty"`
	Default            bool       `json:"default,omitempty"`
	CreatedAt          *time.Time `json:"created_at,omitempty"`
	UpdatedAt          *time.Time `json:"updated_at,omitempty"`
	InAllBrands        bool       `json:"in_all_brands,omitempty"`
	RestrictedBrandIDs []int64    `json:"restricted_brand_ids,omitempty"`
}

func (c *client) ListTicketForms() ([]TicketForm, error) {
	out := new(APIPayload)
	err := c.get(fmt.Sprintf("/api/v2/ticket_forms.json"), out)
	return out.TicketForms, err
}

type TicketField struct {
	ID                  int64               `json:"id,omitempty"`
	Type                TicketFieldType     `json:"type,omitempty"`
	Title               string              `json:"title,omitempty"`
	Description         string              `json:"description,omitempty"`
	Position            int64               `json:"position,omitempty"`
	Active              bool                `json:"active,omitempty"`
	Required            bool                `json:"required,omitempty"`
	RegexpForValidation string              `json:"regexp_for_validation,omitempty"`
	VisibleInPortal     bool                `json:"visible_in_portal,omitempty"`
	EditableInPortal    bool                `json:"editable_in_portal,omitempty"`
	RequiredInPortal    bool                `json:"required_in_portal,omitempty"`
	CreatedAt           *time.Time          `json:"created_at,omitempty"`
	UpdatedAt           *time.Time          `json:"updated_at,omitempty"`
	SystemFieldOptions  []SystemFieldOption `json:"system_field_options,omitempty"`
	CustomFieldOptions  []CustomFieldOption `json:"custom_field_options,omitempty"`
}
type SystemFieldOption struct {
	Name  string `json:"name,omitempty"`
	Value string `json:"value,omitempty"`
}
type CustomFieldOption struct {
	ID      int64  `json:"id,omitempty"`
	Name    string `json:"name,omitempty"`
	RawName string `json:"raw_name,omitempty"`
	Value   string `json:"value,omitempty"`
	Default bool   `json:"default,omitempty"`
}

// ListTicketFields list all availbale custom ticket fields
func (c *client) ListTicketFields() ([]TicketField, error) {
	out := new(APIPayload)
	err := c.get(fmt.Sprintf("/api/v2/ticket_fields.json"), out)

	return out.TicketFields, err
}

type TicketFieldType string

const (
	// System field types
	SubjectType     TicketFieldType = "subject"
	DescriptionType TicketFieldType = "description"
	StatusType      TicketFieldType = "status"
	TicketType      TicketFieldType = "tickettype"
	PriorityType    TicketFieldType = "priority"
	GroupType       TicketFieldType = "group"
	AssigneeType    TicketFieldType = "assignee"

	// Customed field types
	TextType     TicketFieldType = "text"
	TextAreaType TicketFieldType = "textarea"
	CheckBoxType TicketFieldType = "checkbox"
	DateType     TicketFieldType = "date"
	IntegerType  TicketFieldType = "integer"
	DecimalType  TicketFieldType = "decimal"
	RegExpType   TicketFieldType = "regexp"
	TaggerType   TicketFieldType = "tagger"
)

func (c *client) AddTicketTags(id int64, tags []string) ([]string, error) {
	in := &APIPayload{Tags: tags}
	out := new(APIPayload)
	err := c.put(fmt.Sprintf("/api/v2/tickets/%d/tags.json", id), in, out)

	return out.Tags, err
}
