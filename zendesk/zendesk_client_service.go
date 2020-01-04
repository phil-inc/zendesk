package zendesk

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/phil-inc/plib/core/util"
)

// Client describes a client for the Zendesk Core API.
type Client interface {
	WithHeader(name, value string) Client

	AddUserTags(int64, []string) ([]string, error)
	AddTicketTags(int64, []string) ([]string, error)
	BatchUpdateManyTickets([]Ticket) error
	BulkUpdateManyTickets([]int64, *Ticket) error
	CreateIdentity(int64, *UserIdentity) (*UserIdentity, error)
	CreateOrganization(*Organization) (*Organization, error)
	CreateOrganizationMembership(*OrganizationMembership) (*OrganizationMembership, error)
	CreateOrUpdateUser(*User) (*User, error)
	CreateTicket(*Ticket) (*Ticket, error)
	CreateUser(*User) (*User, error)
	DeleteIdentity(int64, int64) error
	DeleteOrganization(int64) error
	DeleteTicket(int64) error
	DeleteUser(int64) (*User, error)
	DeleteOrganizationMembershipByID(int64) error
	ListIdentities(int64) ([]UserIdentity, error)
	ListLocales() ([]Locale, error)
	ListOrganizationMembershipsByUserID(id int64) ([]OrganizationMembership, error)
	ListOrganizations(*ListOptions) ([]Organization, error)
	ListOrganizationUsers(int64, *ListUsersOptions) ([]User, error)
	ListRequestedTickets(int64) ([]Ticket, error)
	ListTicketComments(int64) ([]TicketComment, error)
	ListTicketFields() ([]TicketField, error)
	ListTicketForms() ([]TicketForm, error)
	ListTicketIncidents(int64) ([]Ticket, error)
	ListUsers(*ListUsersOptions) ([]User, error)
	MakeIdentityPrimary(int64, int64) ([]UserIdentity, error)
	SearchUsers(string) ([]User, error)
	ShowIdentity(int64, int64) (*UserIdentity, error)
	ShowLocale(int64) (*Locale, error)
	ShowLocaleByCode(string) (*Locale, error)
	ShowManyUsers([]int64) ([]User, error)
	ShowOrganization(int64) (*Organization, error)
	ShowTicket(int64) (*Ticket, error)
	ShowUser(int64) (*User, error)
	UpdateIdentity(int64, int64, *UserIdentity) (*UserIdentity, error)
	UpdateOrganization(int64, *Organization) (*Organization, error)
	UpdateTicket(int64, *Ticket) (*Ticket, error)
	UpdateUser(int64, *User) (*User, error)
	UploadFile(string, string, io.Reader) (*Upload, error)
	GetAllTickets() ([]Ticket, error)
	GetTicketsIncrementally(int64) ([]Ticket, error)
	GetAllUsers() ([]User, error)
	GetAllTicketMetrics() ([]TicketMetric, error)
	GetTicketMetricsIncrementally([]int64) ([]TicketMetric, error)
	ShowTicketMetric(int64) (*TicketMetric, error)
	GetAllTicketComments([]int64) (map[int64][]TicketComment, error)
	GetUsersIncrementally(int64) ([]User, error)
	GetSatisfactionScores() ([]Score, error)
	GetSatisfactionScoresIncrementally(int64) ([]Score, error)
}

type RequestFunction func(*http.Request) (*http.Response, error)

type MiddlewareFunction func(RequestFunction) RequestFunction

type client struct {
	username string
	password string

	client    *http.Client
	baseURL   *url.URL
	userAgent string
	reqFunc   RequestFunction
	headers   map[string]string
}

// NewEnvClient creates a new Client configured via environment variables.
func NewEnvClient(middleware ...MiddlewareFunction) (Client, error) {
	domain := util.Config("zendesk.domain")
	if domain == "" {
		return nil, errors.New("ZENDESK DOMAIN not found")
	}

	username := util.Config("zendesk.username")
	if username == "" {
		return nil, errors.New("ZENDESK_USERNAME not found")
	}

	password := util.Config("zendesk.password")
	log.Printf("[zendesk_client_service][NewEnvClient]Zendesk config: %s, %s, %s", domain, username, password)
	if password == "" {
		return nil, errors.New("ZENDESK_PASSWORD not found")
	}

	return NewClient(domain, username, password, middleware...)
}

// NewClient creates a new Client.
// You can use either a user email/password combination or an API token.
// For the latter, append /token to the email and use the API token as a password
func NewClient(domain, username, password string, middleware ...MiddlewareFunction) (Client, error) {
	return NewURLClient(fmt.Sprintf("https://%s.zendesk.com", domain), username, password, middleware...)
}

// NewURLClient is like NewClient but accepts an explicit end point instead of a Zendesk domain.
func NewURLClient(endpoint, username, password string, middleware ...MiddlewareFunction) (Client, error) {
	log.Printf("[zendesk_client_service][NewURLClient]Zendesk config: %s, %s, %s", endpoint, username, password)
	baseURL, err := url.Parse(endpoint)
	if err != nil {
		return nil, err
	}

	c := &client{
		baseURL:   baseURL,
		userAgent: "PHIL-Zendesk",
		username:  username,
		password:  password,
		reqFunc:   http.DefaultClient.Do,
		headers:   make(map[string]string),
	}

	if middleware != nil {
		for i := len(middleware) - 1; i >= 0; i-- {
			c.reqFunc = middleware[i](c.reqFunc)
		}
	}

	return c, nil
}

// WithHeader returns an updated client that sends the provided header
// with each subsequent request.
func (c *client) WithHeader(name, value string) Client {
	newClient := *c
	newClient.headers = make(map[string]string)

	for k, v := range c.headers {
		newClient.headers[k] = v
	}

	newClient.headers[name] = value

	return &newClient
}

func (c *client) request(method, endpoint string, headers map[string]string, body io.Reader) (*http.Response, error) {
	rel, err := url.Parse(endpoint)
	if err != nil {
		return nil, err
	}

	url := c.baseURL.ResolveReference(rel)
	req, err := http.NewRequest(method, url.String(), body)
	if err != nil {
		return nil, err
	}

	req.SetBasicAuth(c.username, c.password)
	req.Header.Set("User-Agent", c.userAgent)

	for key, value := range c.headers {
		req.Header.Set(key, value)
	}

	for key, value := range headers {
		req.Header.Set(key, value)
	}

	return c.reqFunc(req)
}

func (c *client) do(method, endpoint string, in, out interface{}) error {
	payload, err := marshall(in)
	if err != nil {
		return err
	}

	headers := map[string]string{}
	if in != nil {
		headers["Content-Type"] = "application/json"
	}

	res, err := c.request(method, endpoint, headers, bytes.NewReader(payload))
	if err != nil {
		return err
	}

	defer res.Body.Close()

	// Retry the request if the retry after header is present. This can happen when we are
	// being rate limited or we failed with a retriable error.
	if res.Header.Get("Retry-After") != "" {
		after, err := strconv.ParseInt(res.Header.Get("Retry-After"), 10, 64)
		if err != nil || after == 0 {
			return unmarshall(res, out)
		}

		time.Sleep(time.Duration(after) * time.Second)

		res, err = c.request(method, endpoint, headers, bytes.NewReader(payload))
		if err != nil {
			return err
		}
		defer res.Body.Close()
	}

	return unmarshall(res, out)
}

func (c *client) get(endpoint string, out interface{}) error {
	return c.do("GET", endpoint, nil, out)
}

func (c *client) getAll(endpoint string, in interface{}) ([]Ticket, error) {
	result := make([]Ticket, 0)
	payload, err := marshall(in)
	if err != nil {
		return nil, err
	}

	headers := map[string]string{}
	if in != nil {
		headers["Content-Type"] = "application/json"
	}

	res, err := c.request("GET", endpoint, headers, bytes.NewReader(payload))
	dataPerPage := new(APIPayload)
	if err != nil {
		return nil, err
	}

	apiV2 := "/api/v2/"
	fieldName := strings.Split(endpoint[len(apiV2):], ".")[0]
	defer res.Body.Close()

	err = unmarshall(res, dataPerPage)

	apiStartIndex := strings.Index(dataPerPage.NextPage, apiV2)
	currentPage := endpoint

	var totalWaitTime int64
	for currentPage != "" {
		// if too many requests(res.StatusCode == 429), delay sending request
		if res.StatusCode == 429 {
			after, err := strconv.ParseInt(res.Header.Get("Retry-After"), 10, 64)
			log.Printf("[zendesk_client_service][getAll] too many requests. Wait for %v seconds\n", after)
			totalWaitTime += after
			if err != nil {
				return nil, err
			}
			time.Sleep(time.Duration(after) * time.Second)
		} else {
			if fieldName == "tickets" {
				result = append(result, dataPerPage.Tickets...)
			}
			currentPage = dataPerPage.NextPage
			log.Printf("[zendesk_client_service][getAll] pulling page: %s\n", currentPage)
		}
		res, _ = c.request("GET", dataPerPage.NextPage[apiStartIndex:], headers, bytes.NewReader(payload))
		dataPerPage = new(APIPayload)
		err = unmarshall(res, dataPerPage)
		if err != nil {
			return nil, err
		}
	}
	log.Printf("[zendesk_client_service][getAll] number of records pulled: %v\n", len(result))
	log.Printf("[zendesk_client_service][getAll] total waiting time due to rate limit: %v\n", totalWaitTime)

	return result, err
}

func (c *client) getOneByOne(in interface{}) ([]Ticket, error) {
	endpointPrefix := "/api/v2/tickets/"
	endpointPostfix := ".json"
	result := make([]Ticket, 0)
	payload, err := marshall(in)
	if err != nil {
		return nil, err
	}

	headers := map[string]string{}
	if in != nil {
		headers["Content-Type"] = "application/json"
	}
	record := new(APIPayload)

	// currently we can manually set the starting and ending IDs for data pulling
	// because memory may reach its limit if the dataset is too large
	// ideally, we want to load data to database in batches on the fly
	// instead of loading the entire chunk
	startID := 1
	endID := 10000
	ticketID := startID // start
	endpoint := fmt.Sprintf("%s%v%s", endpointPrefix, ticketID, endpointPostfix)
	res, err := c.request("GET", endpoint, headers, bytes.NewReader(payload))
	defer res.Body.Close()

	var totalWaitTime int64
	for ticketID < endID {
		log.Printf("[zendesk_client_service][getOneByOne] currently extracting: %s\n", endpoint)

		// handle page not found
		if res.StatusCode == 404 {
			log.Printf("[zendesk_client_service][getOneByOne] 404 not found: %s\n", endpoint)
			// handle too many requests (rate limit)
		} else if res.StatusCode == 429 {
			after, err := strconv.ParseInt(res.Header.Get("Retry-After"), 10, 64)
			log.Printf("[zendesk_client_service][getOneByOne] too many requests. Wait for %v seconds\n", after)
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

			result = append(result, *record.Ticket)
		}

		record = new(APIPayload)
		ticketID++
		endpoint = fmt.Sprintf("%s%v%s", endpointPrefix, ticketID, endpointPostfix)
		res, _ = c.request("GET", endpoint, headers, bytes.NewReader(payload))
	}

	log.Printf("[zendesk_client_service][getOneByOne] number of records pulled: %v\n", len(result))
	log.Printf("[zendesk_client_service][getOneByOne] total waiting time due to rate limit: %v\n", totalWaitTime)
	return result, nil
}

func (c *client) post(endpoint string, in, out interface{}) error {
	return c.do("POST", endpoint, in, out)
}

func (c *client) put(endpoint string, in, out interface{}) error {
	return c.do("PUT", endpoint, in, out)
}

func (c *client) delete(endpoint string, out interface{}) error {
	return c.do("DELETE", endpoint, nil, out)
}

func marshall(in interface{}) ([]byte, error) {
	if in == nil {
		return nil, nil
	}

	return json.Marshal(in)
}

func unmarshall(res *http.Response, out interface{}) error {
	if res.StatusCode < 200 || res.StatusCode >= 300 {
		apierr := new(APIError)
		apierr.Response = res
		if err := json.NewDecoder(res.Body).Decode(apierr); err != nil {
			apierr.Type = "Unknown"
			apierr.Description = "Oops! Something went wrong when parsing the error response."
		}
		return apierr
	}

	if out != nil {
		return json.NewDecoder(res.Body).Decode(out)
	}

	return nil
}

// APIPayload represents the payload of an API call.
type APIPayload struct {
	Attachment              *Attachment              `json:"attachment"`
	Attachments             []Attachment             `json:"attachments"`
	Comment                 *TicketComment           `json:"comment,omitempty"`
	Comments                []TicketComment          `json:"comments,omitempty"`
	Identity                *UserIdentity            `json:"identity,omitempty"`
	Identities              []UserIdentity           `json:"identities,omitempty"`
	Locale                  *Locale                  `json:"locale,omitempty"`
	Locales                 []Locale                 `json:"locales,omitempty"`
	Organization            *Organization            `json:"organization,omitempty"`
	OrganizationMembership  *OrganizationMembership  `json:"organization_membership,omitempty"`
	OrganizationMemberships []OrganizationMembership `json:"organization_memberships,omitempty"`
	Organizations           []Organization           `json:"organizations,omitempty"`
	Tags                    []string                 `json:"tags,omitempty"`
	Ticket                  *Ticket                  `json:"ticket,omitempty"`
	TicketField             *TicketField             `json:"ticket_field,omitempty"`
	TicketFields            []TicketField            `json:"ticket_fields,omitempty"`
	Tickets                 []Ticket                 `json:"tickets,omitempty"`
	Upload                  *Upload                  `json:"upload,omitempty"`
	User                    *User                    `json:"user,omitempty"`
	Users                   []User                   `json:"users,omitempty"`
	TicketForm              *TicketForm              `json:"ticket_form,omitempty"`
	TicketForms             []TicketForm             `json:"ticket_forms,omitempty"`
	TicketMetric            *TicketMetric            `json:"ticket_metric,omitempty"`
	TicketMetrics           []TicketMetric           `json:"ticket_metrics,omitempty"`
	NextPage                string                   `json:"next_page,omitempty"`
	SatisfactionRating      Score                    `json:"satisfaction_rating,omitempty"`
	SatisfactionRatings     []Score                  `json:"satisfaction_ratings,omitempty"`
}

// APIError represents an error response returnted by the API.
type APIError struct {
	Response *http.Response

	Type        string                       `json:"error,omitmepty"`
	Description string                       `json:"description,omitempty"`
	Details     map[string][]*APIErrorDetail `json:"details,omitempty"`
}

func (e *APIError) Error() string {
	msg := fmt.Sprintf("%v %v: %d", e.Response.Request.Method, e.Response.Request.URL, e.Response.StatusCode)

	if e.Type != "" {
		msg = fmt.Sprintf("%s %v", msg, e.Type)
	}

	if e.Description != "" {
		msg = fmt.Sprintf("%s: %v", msg, e.Description)
	}

	if e.Details != nil {
		msg = fmt.Sprintf("%s: %+v", msg, e.Details)
	}

	return msg
}

// APIErrorDetail represents a detail about an APIError.
type APIErrorDetail struct {
	Type        string `json:"error,omitempty"`
	Description string `json:"description,omitempty"`
}

func (e *APIErrorDetail) Error() string {
	msg := ""

	if e.Type != "" {
		msg = e.Type + ": "
	}

	if e.Description != "" {
		msg += e.Description
	}

	return msg
}

// Bool is a helper function that returns a pointer to the bool value b.
func Bool(b bool) *bool {
	p := b
	return &p
}

// Int is a helper function that returns a pointer to the int value i.
func Int(i int64) *int64 {
	p := i
	return &p
}

// String is a helper function that returns a pointer to the string value s.
func String(s string) *string {
	p := s
	return &p
}

// ListOptions specifies the optional parameters for the list methods that support pagination.
//
// Zendesk Core API doscs: https://developer.zendesk.com/rest_api/docs/core/introduction#pagination
type ListOptions struct {
	// Sets the page of results to retrieve.
	Page int `url:"page,omitempty"`
	// Sets the number of results to include per page.
	PerPage int `url:"per_page,omitempty"`
}
