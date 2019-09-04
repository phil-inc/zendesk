package zendesk

import (
	"bytes"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/google/go-querystring/query"
)

// User represents a Zendesk user.
//
// Zendesk Core API docs: https://developer.zendesk.com/rest_api/docs/core/users#content
type User struct {
	ID                  int64                  `json:"id,omitempty"`
	URL                 string                 `json:"url,omitempty"`
	Name                string                 `json:"name,omitempty"`
	ExternalID          string                 `json:"external_id,omitempty"`
	Alias               string                 `json:"alias,omitempty"`
	CreatedAt           *time.Time             `json:"created_at,omitempty"`
	UpdatedAt           *time.Time             `json:"updated_at,omitempty"`
	Active              bool                   `json:"active,omitempty"`
	Verified            bool                   `json:"verified,omitempty"`
	Shared              bool                   `json:"shared,omitempty"`
	SharedAgent         bool                   `json:"shared_agent,omitempty"`
	Locale              string                 `json:"locale,omitempty"`
	LocaleID            int64                  `json:"locale_id,omitempty"`
	TimeZone            string                 `json:"time_zone,omitempty"`
	LastLoginAt         *time.Time             `json:"last_login_at,omitempty"`
	Email               string                 `json:"email,omitempty"`
	Phone               string                 `json:"phone,omitempty"`
	Signature           string                 `json:"signature,omitempty"`
	Details             string                 `json:"details,omitempty"`
	Notes               string                 `json:"notes,omitempty"`
	OrganizationID      int64                  `json:"organization_id,omitempty"`
	Role                string                 `json:"role,omitempty"`
	CustomerRoleID      int64                  `json:"custom_role_id,omitempty"`
	Moderator           bool                   `json:"moderator,omitempty"`
	TicketRestriction   string                 `json:"ticket_restriction,omitempty"`
	OnlyPrivateComments bool                   `json:"only_private_comments,omitempty"`
	Tags                []string               `json:"tags,omitempty"`
	RestrictedAgent     bool                   `json:"restricted_agent,omitempty"`
	Suspended           bool                   `json:"suspended,omitempty"`
	UserFields          map[string]interface{} `json:"user_fields,omitempty"`
}

// ShowUser fetches a user by its ID.
//
// Zendesk Core API docs: https://developer.zendesk.com/rest_api/docs/core/users#show-user
func (c *client) ShowUser(id int64) (*User, error) {
	out := new(APIPayload)
	err := c.get(fmt.Sprintf("/api/v2/users/%d.json", id), out)
	return out.User, err
}

func (c *client) ShowManyUsers(ids []int64) ([]User, error) {
	sids := []string{}
	for _, id := range ids {
		sids = append(sids, strconv.FormatInt(id, 10))
	}

	out := new(APIPayload)
	err := c.get(fmt.Sprintf("/api/v2/users/show_many.json?ids=%s", strings.Join(sids, ",")), out)
	return out.Users, err
}

// CreateUser creates a user.
//
// Zendesk Core API docs: https://developer.zendesk.com/rest_api/docs/core/users#create-user
func (c *client) CreateUser(user *User) (*User, error) {
	in := &APIPayload{User: user}
	out := new(APIPayload)
	err := c.post("/api/v2/users.json", in, out)
	return out.User, err
}

// CreateOrUpdateUser creates or updates a user.
//
// Zendesk Core API docs: https://developer.zendesk.com/rest_api/docs/core/users#create-or-update-user
func (c *client) CreateOrUpdateUser(user *User) (*User, error) {
	in := &APIPayload{User: user}
	out := new(APIPayload)
	err := c.post("/api/v2/users/create_or_update.json", in, out)
	return out.User, err
}

// UpdateUser updates a user.
//
// Zendesk Core API docs: https://developer.zendesk.com/rest_api/docs/core/users#update-user
func (c *client) UpdateUser(id int64, user *User) (*User, error) {
	in := &APIPayload{User: user}
	out := new(APIPayload)
	err := c.put(fmt.Sprintf("/api/v2/users/%d.json", id), in, out)
	return out.User, err
}

// DeleteUser deletes an User.
//
// Zendesk Core API docs: https://developer.zendesk.com/rest_api/docs/core/users#delete-user
func (c *client) DeleteUser(id int64) (*User, error) {
	out := new(APIPayload)
	err := c.delete(fmt.Sprintf("/api/v2/users/%d.json", id), out)
	return out.User, err
}

// ListUsersOptions specifies the optional parameters for the list users methods.
type ListUsersOptions struct {
	ListOptions

	Role          []string `url:"role"`
	PermissionSet int64    `url:"permision_set"`
}

// ListOrganizationUsers list the users associated to an organization.
//
// Zendesk Core API docs: https://developer.zendesk.com/rest_api/docs/core/users#list-users
func (c *client) ListOrganizationUsers(id int64, opts *ListUsersOptions) ([]User, error) {
	params, err := query.Values(opts)
	if err != nil {
		return nil, err
	}

	out := new(APIPayload)
	err = c.get(fmt.Sprintf("/api/v2/organizations/%d/users.json?%s", id, params.Encode()), out)
	return out.Users, err
}

// ListUsers list of all users.
//
// Zendesk Core API docs: https://developer.zendesk.com/rest_api/docs/core/users#list-users
func (c *client) ListUsers(opts *ListUsersOptions) ([]User, error) {
	params, err := query.Values(opts)
	if err != nil {
		return nil, err
	}

	out := new(APIPayload)
	err = c.get(fmt.Sprintf("/api/v2/users.json?%s", params.Encode()), out)
	return out.Users, err
}

// SearchUsers searches users by name or email address.
//
// Zendesk Core API docs: https://developer.zendesk.com/rest_api/docs/core/users#search-users
func (c *client) SearchUsers(query string) ([]User, error) {
	out := new(APIPayload)
	err := c.get("/api/v2/users/search.json?query="+query, out)
	return out.Users, err
}

// AddUserTags adds a tag to a user
//
// Zendesk Core API docs: https://developer.zendesk.com/rest_api/docs/core/tags#add-tags
func (c *client) AddUserTags(id int64, tags []string) ([]string, error) {
	in := &APIPayload{Tags: tags}
	out := new(APIPayload)
	err := c.put(fmt.Sprintf("/api/v2/users/%d/tags.json", id), in, out)
	return out.Tags, err
}

// GetUsersIncrementally pull the list of users modified from a specific time point
//
// https://developer.zendesk.com/rest_api/docs/support/incremental_export#incremental-user-export
func (c *client) GetUsersIncrementally(unixTime int64) ([]User, error) {
	users, err := c.getUsersIncrementally(unixTime, nil)
	return users, err
}

func (c *client) getUsersIncrementally(unixTime int64, in interface{}) ([]User, error) {
	result := make([]User, 0)
	payload, err := marshall(in)
	if err != nil {
		return nil, err
	}

	headers := map[string]string{}
	if in != nil {
		headers["Content-Type"] = "application/json"
	}

	apiV2 := "/api/v2/incremental/users.json?start_time="
	url := "https://philhelp.zendesk.com" + apiV2
	apiStartIndex := strings.Index(url, apiV2)
	endpoint := fmt.Sprintf("%s%v", apiV2, unixTime)

	res, err := c.request("GET", endpoint, headers, bytes.NewReader(payload))
	defer res.Body.Close()

	dataPerPage := new(APIPayload)
	if err != nil {
		return nil, err
	}

	currentPage := "emptypage"

	var totalWaitTime int64

	for currentPage != dataPerPage.NextPage {

		// if too many requests(res.StatusCode == 429), delay sending request
		if res.StatusCode == 429 {
			after, err := strconv.ParseInt(res.Header.Get("Retry-After"), 10, 64)
			log.Printf("[ZENDESK] too many requests. Wait for %v seconds\n", after)
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
			result = append(result, dataPerPage.Users...)
			if currentPage == dataPerPage.NextPage {
				break
			}
			currentPage = dataPerPage.NextPage
			log.Printf("[ZENDESK] pulling page: %s\n", currentPage)
		}

		res, _ = c.request("GET", dataPerPage.NextPage[apiStartIndex:], headers, bytes.NewReader(payload))

		dataPerPage = new(APIPayload)
	}
	log.Printf("[ZENDESK] number of records pulled: %v\n", len(result))
	log.Printf("[ZENDESK] total waiting time due to rate limit: %v\n", totalWaitTime)

	return getUniqUsers(result), err
}

// getUniqUsers is to remove the duplicate records due to pagination
// more details can be found int the following link
// https://developer.zendesk.com/rest_api/docs/support/incremental_export#excluding_pagination_duplicates

func getUniqUsers(users []User) []User {
	var Empty struct{}
	keys := make(map[string]struct{})
	result := make([]User, 0)
	for _, user := range users {
		key := fmt.Sprintf("%v %v\n", user.ID, user.UpdatedAt)
		if _, ok := keys[key]; ok {
			continue
		} else {
			keys[key] = Empty
			result = append(result, user)
		}
	}
	return result
}

// GetAllUsers pull the list of all the users
//
// Zendesk Core API docs: https://developer.zendesk.com/rest_api/docs/core/users#list-users

func (c *client) GetAllUsers() ([]User, error) {
	users, err := c.getAllUsers("/api/v2/users.json", nil)
	return users, err
}

func (c *client) getAllUsers(endpoint string, in interface{}) ([]User, error) {
	result := make([]User, 0)
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
			log.Printf("[ZENDESK] too many requests. Wait for %v seconds\n", after)
			totalWaitTime += after
			if err != nil {
				return nil, err
			}
			time.Sleep(time.Duration(after) * time.Second)
		} else {
			if fieldName == "users" {
				result = append(result, dataPerPage.Users...)
			}
			currentPage = dataPerPage.NextPage
			log.Printf("[ZENDESK] pulling page: %s\n", currentPage)
		}
		res, _ = c.request("GET", dataPerPage.NextPage[apiStartIndex:], headers, bytes.NewReader(payload))
		dataPerPage = new(APIPayload)
		err = unmarshall(res, dataPerPage)
		if err != nil {
			return nil, err
		}
	}
	log.Printf("[ZENDESK] number of records pulled: %v\n", len(result))
	log.Printf("[ZENDESK] total waiting time due to rate limit: %v\n", totalWaitTime)

	return result, err
}

//UpdateEndUser updates the info of one end user
//
// Zendesk Core API docs: https://developer.zendesk.com/rest_api/docs/core/end_user#update-user
func (c *client) UpdateEndUser(id int64, user *User) (*User, error) {
	out := new(APIPayload)
	in := &APIPayload{User: user}
	err := c.put(fmt.Sprintf("/api/v2/end_users/%d.json", id), in, out)
	return user, err
}

// UserIdentity represents a Zendesk user identity.
//
// Zendesk Core API docs: https://developer.zendesk.com/rest_api/docs/core/user_identities
type UserIdentity struct {
	ID                 int64      `json:"id,omitempty"`
	URL                string     `json:"url,omitempty"`
	UserID             int64      `json:"user_id,omitempty"`
	Type               string     `json:"type,omitempty"`
	Value              string     `json:"value,omitempty"`
	Verified           bool       `json:"verified,omitempty"`
	Primary            bool       `json:"primary,omitempty"`
	CreatedAt          *time.Time `json:"created_at,omitempty"`
	UpdatedAt          *time.Time `json:"updated_at,omitempty"`
	UndeliverableCount int64      `json:"undeliverable_count,omitempty"`
	DeliverableState   string     `json:"deliverable_state,omitempty"`
}

// ListIdentities lists all user identities.
//
// Zendesk Core API docs: https://developer.zendesk.com/rest_api/docs/core/user_identities#list-identities
func (c *client) ListIdentities(userID int64) ([]UserIdentity, error) {
	out := new(APIPayload)
	err := c.get(fmt.Sprintf("/api/v2/users/%d/identities.json", userID), out)
	return out.Identities, err
}

// ShowIdentity fetches a user identity by its ID and user ID.
//
// Zendesk Core API docs: https://developer.zendesk.com/rest_api/docs/core/user_identities#show-identity
func (c *client) ShowIdentity(userID, id int64) (*UserIdentity, error) {
	out := new(APIPayload)
	err := c.get(fmt.Sprintf("/api/v2/users/%d/identities/%d.json", userID, id), out)
	return out.Identity, err
}

// CreateIdentity creates a user identity.
//
// Zendesk Core API docs: https://developer.zendesk.com/rest_api/docs/core/user_identities#create-identity
func (c *client) CreateIdentity(userID int64, identity *UserIdentity) (*UserIdentity, error) {
	in := &APIPayload{Identity: identity}
	out := new(APIPayload)
	err := c.post(fmt.Sprintf("/api/v2/users/%d/identities.json", userID), in, out)
	return out.Identity, err
}

// UpdateIdentity updates the value and verified status of a user identity.
//
// Zendesk Core API docs: https://developer.zendesk.com/rest_api/docs/core/user_identities#update-identity
func (c *client) UpdateIdentity(userID, id int64, identity *UserIdentity) (*UserIdentity, error) {
	in := &APIPayload{Identity: identity}
	out := new(APIPayload)
	err := c.put(fmt.Sprintf("/api/v2/users/%d/identities/%d.json", userID, id), in, out)
	return out.Identity, err
}

// DeleteIdentity deletes a user identity.
//
// Zendesk Core API docs: https://developer.zendesk.com/rest_api/docs/core/user_identities#delete-identity
func (c *client) DeleteIdentity(userID, id int64) error {
	return c.delete(fmt.Sprintf("/api/v2/users/%d/identities/%d.json", userID, id), nil)
}

func (c *client) MakeIdentityPrimary(userID, id int64) ([]UserIdentity, error) {
	out := new(APIPayload)
	err := c.put(fmt.Sprintf("/api/v2/end_users/%d/identities/%d/make_primary.json", userID, id), nil, out)
	return out.Identities, err
}
