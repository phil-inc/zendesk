package zendesk

import (
	"bytes"
	"fmt"
	"log"
	"net/url"
	"strconv"
	"strings"
	"time"
)

type Call struct {
	AgentID                      int         `json:"agent_id"`
	CallCharge                   string      `json:"call_charge"`
	CallRecordingConsent         string      `json:"call_recording_consent"`
	CallRecordingConsentAction   string      `json:"call_recording_consent_action"`
	CallRecordingConsentKeypress string      `json:"call_recording_consent_keypress"`
	Callback                     bool        `json:"callback"`
	CallbackSource               interface{} `json:"callback_source"`
	CompletionStatus             string      `json:"completion_status"`
	ConsultationTime             int         `json:"consultation_time"`
	CreatedAt                    time.Time   `json:"created_at"`
	CustomerID                   int         `json:"customer_id"`
	CustomerRequestedVoicemail   bool        `json:"customer_requested_voicemail"`
	DefaultGroup                 bool        `json:"default_group"`
	Direction                    string      `json:"direction"`
	Duration                     int         `json:"duration"`
	ExceededQueueWaitTime        bool        `json:"exceeded_queue_wait_time"`
	HoldTime                     int         `json:"hold_time"`
	ID                           int         `json:"id"`
	IvrAction                    interface{} `json:"ivr_action"`
	IvrDestinationGroupName      interface{} `json:"ivr_destination_group_name"`
	IvrHops                      interface{} `json:"ivr_hops"`
	IvrRoutedTo                  interface{} `json:"ivr_routed_to"`
	IvrTimeSpent                 interface{} `json:"ivr_time_spent"`
	Line                         string      `json:"line"`
	LineID                       int         `json:"line_id"`
	MinutesBilled                int         `json:"minutes_billed"`
	NotRecordingTime             int         `json:"not_recording_time"`
	OutsideBusinessHours         bool        `json:"outside_business_hours"`
	Overflowed                   bool        `json:"overflowed"`
	OverflowedTo                 interface{} `json:"overflowed_to"`
	PhoneNumber                  string      `json:"phone_number"`
	PhoneNumberID                int         `json:"phone_number_id"`
	QualityIssues                []string    `json:"quality_issues"`
	RecordingControlInteractions int         `json:"recording_control_interactions"`
	RecordingTime                int         `json:"recording_time"`
	TalkTime                     int         `json:"talk_time"`
	TicketID                     int         `json:"ticket_id"`
	TimeToAnswer                 int         `json:"time_to_answer"`
	UpdatedAt                    time.Time   `json:"updated_at"`
	Voicemail                    bool        `json:"voicemail"`
	WaitTime                     int         `json:"wait_time"`
	WrapUpTime                   int         `json:"wrap_up_time"`
}

type CallLeg struct {
	AgentID          int         `json:"agent_id"`
	AvailableVia     interface{} `json:"available_via"`
	CallCharge       string      `json:"call_charge"`
	CallID           int         `json:"call_id"`
	CompletionStatus string      `json:"completion_status"`
	ConferenceFrom   interface{} `json:"conference_from"`
	ConferenceTime   interface{} `json:"conference_time"`
	ConferenceTo     interface{} `json:"conference_to"`
	ConsultationFrom interface{} `json:"consultation_from"`
	ConsultationTime interface{} `json:"consultation_time"`
	ConsultationTo   interface{} `json:"consultation_to"`
	CreatedAt        time.Time   `json:"created_at"`
	Duration         int         `json:"duration"`
	ForwardedTo      interface{} `json:"forwarded_to"`
	HoldTime         int         `json:"hold_time"`
	ID               int         `json:"id"`
	MinutesBilled    int         `json:"minutes_billed"`
	QualityIssues    []string    `json:"quality_issues"`
	TalkTime         int         `json:"talk_time"`
	TransferredFrom  interface{} `json:"transferred_from"`
	TransferredTo    interface{} `json:"transferred_to"`
	Type             string      `json:"type"`
	UpdatedAt        time.Time   `json:"updated_at"`
	UserID           int         `json:"user_id"`
	WrapUpTime       interface{} `json:"wrap_up_time"`
}

//https://developer.zendesk.com/api-reference/voice/talk-api/incremental_exports/#incremental-call-legs-export
func (c *client) GetCallLegIncrementally(unixTime int64) ([]CallLeg, error) {
	log.Printf("[zd_ticket_service][GetCallLegsIncrementally] Start GetCallLegsIncrementally")
	callLegs, err := c.getCallLegsIncrementally(unixTime, nil)
	log.Printf("[zd_ticket_service][GetTicketsIncrementally] Number of CallLegs: %v", len(callLegs))
	return callLegs, err
}

func (c *client) getCallLegsIncrementally(unixTime int64, in interface{}) ([]CallLeg, error) {
	log.Printf("[zd_ticket_service][getCallLegsIncrementally] Start getCallLegsIncrementally")
	result := make([]CallLeg, 0)
	payload, err := marshall(in)
	if err != nil {
		return nil, err
	}

	headers := map[string]string{}
	if in != nil {
		headers["Content-Type"] = "application/json"
	}

	apiV2 := "/api/v2/channels/voice/stats/incremental/legs?start_time="
	rel, err := url.Parse(apiV2)
	if err != nil {
		return nil, err
	}
	url := c.baseURL.ResolveReference(rel)
	apiStartIndex := strings.Index(url.String(), apiV2)
	endpoint := fmt.Sprintf("%s%v", apiV2, unixTime)

	res, err := c.request("GET", endpoint, headers, bytes.NewReader(payload))
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	dataPerPage := new(APIPayload)
	currentPage := "emptypage"
	var totalWaitTime int64
	log.Printf("[zd_ticket_service][getCallLegsIncrementally] Start for loop in getCallLegsIncrementally")
	for currentPage != dataPerPage.NextPage {
		// if too many requests(res.StatusCode == 429), delay sending request
		if res.StatusCode == 429 {
			after, err := strconv.ParseInt(res.Header.Get("Retry-After"), 10, 64)
			log.Printf("[zd_ticket_service][getCallLegsIncrementally] too many requests. Wait for %v seconds\n", after)
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
			result = append(result, dataPerPage.CallLegs...)
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

	return result, err
}
