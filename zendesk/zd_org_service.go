package zendesk

import (
	"fmt"
	"time"

	"github.com/google/go-querystring/query"
)

// Organization represents a Zendesk organization.
//
// Zendesk Core API docs: https://developer.zendesk.com/rest_api/docs/core/organizations
type Organization struct {
	ID                 int64                  `json:"id,omitempty"`
	URL                string                 `json:"url,omitempty"`
	ExternalID         string                 `json:"external_id,omitempty"`
	Name               string                 `json:"name,omitempty"`
	CreatedAt          *time.Time             `json:"created_at,omitempty"`
	UpdatedAt          *time.Time             `json:"updated_at,omitempty"`
	DomainNames        []string               `json:"domain_names,omitempty"`
	Details            string                 `json:"details,omitempty"`
	Notes              string                 `json:"notes,omitempty"`
	GroupID            int64                  `json:"group_id,omitempty"`
	SharedTickets      bool                   `json:"shared_tickets,omitempty"`
	SharedComments     bool                   `json:"shared_comments,omitempty"`
	OrganizationFields map[string]interface{} `json:"organization_fields,omitempty"`
}

// ShowOrganization fetches an organization by its ID.
//
// Zendesk Core API docs: https://developer.zendesk.com/rest_api/docs/core/organizations#show-organization
func (c *client) ShowOrganization(id int64) (*Organization, error) {
	out := new(APIPayload)
	err := c.get(fmt.Sprintf("/api/v2/organizations/%d.json", id), out)
	return out.Organization, err
}

// CreateOrganization creates an organization.
//
// Zendesk Core API docs: https://developer.zendesk.com/rest_api/docs/core/organizations#create-organization
func (c *client) CreateOrganization(org *Organization) (*Organization, error) {
	in := &APIPayload{Organization: org}
	out := new(APIPayload)
	err := c.post("/api/v2/organizations.json", in, out)
	return out.Organization, err
}

// UpdateOrganization updates an organization.
//
// Zendesk Core API docs: https://developer.zendesk.com/rest_api/docs/core/organizations#update-organization
func (c *client) UpdateOrganization(id int64, org *Organization) (*Organization, error) {
	in := &APIPayload{Organization: org}
	out := new(APIPayload)
	err := c.put(fmt.Sprintf("/api/v2/organizations/%d.json", id), in, out)
	return out.Organization, err
}

// ListOrganizations list all organizations.
//
// Zendesk Core API docs: https://developer.zendesk.com/rest_api/docs/core/organizations#list-organizations
func (c *client) ListOrganizations(opts *ListOptions) ([]Organization, error) {
	params, err := query.Values(opts)
	if err != nil {
		return nil, err
	}

	out := new(APIPayload)
	err = c.get("/api/v2/organizations.json?"+params.Encode(), out)
	return out.Organizations, err
}

// DeleteOrganization deletes an Organization.
//
// Zendesk Core API docs: https://developer.zendesk.com/rest_api/docs/core/organizations#delete-organization
func (c *client) DeleteOrganization(id int64) error {
	return c.delete(fmt.Sprintf("/api/v2/organizations/%d.json", id), nil)
}

// OrganizationMembership represents a Zendesk association between an org and a user.
//
// Zendesk Core API docs: https://developer.zendesk.com/rest_api/docs/core/organization_memberships
type OrganizationMembership struct {
	ID             int64      `json:"id,omitempty"`
	URL            string     `json:"url,omitempty"`
	UserID         int64      `json:"user_id,omitempty"`
	OrganizationID int64      `json:"organization_id,omitempty"`
	Default        bool       `json:"default,omitempty"`
	CreatedAt      *time.Time `json:"created_at,omitempty"`
	UpdatedAt      *time.Time `json:"updated_at,omitempty"`
}

// CreateOrganizationMembership creates an organization membership.
//
// Zendesk Core API docs: https://developer.zendesk.com/rest_api/docs/core/organization_memberships#create-membership
func (c *client) CreateOrganizationMembership(orgMembership *OrganizationMembership) (*OrganizationMembership, error) {
	in := &APIPayload{OrganizationMembership: orgMembership}
	out := new(APIPayload)
	err := c.post("/api/v2/organization_memberships.json", in, out)
	return out.OrganizationMembership, err
}

// ListOrganizationMembershipsByUserID returns all organization memberships for a specific user
//
// Zendesk Core API docs: https://developer.zendesk.com/rest_api/docs/core/organization_memberships#list-memberships
func (c *client) ListOrganizationMembershipsByUserID(id int64) ([]OrganizationMembership, error) {
	out := new(APIPayload)
	err := c.get(fmt.Sprintf("/api/v2/users/%d/organization_memberships.json", id), out)
	return out.OrganizationMemberships, err
}

// DeleteOrganizationMembership removes an organization membership
//
// Zendesk Core API docs: https://developer.zendesk.com/rest_api/docs/core/organization_memberships#delete-membership
func (c *client) DeleteOrganizationMembershipByID(id int64) error {
	return c.delete(fmt.Sprintf("/api/v2/organization_memberships/%d.json", id), nil)
}
