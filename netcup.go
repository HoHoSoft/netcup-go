package netcup

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

const endpoint = "https://ccp.netcup.net/run/webservice/servers/endpoint.php?JSON"

// The Client is the entrypoint to the Netcup API
type Client struct {
	customerNumber int
	apiKey         string
	sessionID      string

	endpoint   string
	httpClient *http.Client

	domainName string
	records    []Record
}

// A Record is a domain DNS record
type Record struct {
	ID           string `json:"id"`
	Hostname     string `json:"hostname"`
	Type         string `json:"type"`
	Priority     string `json:"priority"`
	Destination  string `json:"destination"`
	DeleteRecord bool   `json:"deleterecord"`
	State        string `json:"state"`
}

// requestBody for all messages sent to the API
type requestBody struct {
	Action string      `json:"action"`
	Param  interface{} `json:"param"`
}

// responseBody sent by the API
type responseBody struct {
	ServerRequestID string           `json:"serverrequestid"`
	ClientRequestID string           `json:"clientrequestid"`
	Action          string           `json:"action"`
	Status          string           `json:"status"`
	StatusCode      int              `json:"statuscode"`
	ShortMessage    string           `json:"shortmessage"`
	LongMessage     string           `json:"longmessage"`
	ResponseData    *json.RawMessage `json:"responsedata"`
}

// NewClient returns a new client for the Netcup CCP API
func NewClient(customerNumber int, apiKey string) *Client {
	c := &Client{
		customerNumber: customerNumber,
		apiKey:         apiKey,
		endpoint:       endpoint,
		httpClient:     &http.Client{},
	}

	return c
}

func (c *Client) request(action string, param interface{}) (*json.RawMessage, error) {
	if c.sessionID == "" && action != "login" {
		return nil, fmt.Errorf("no session ID. Make sure to login first")
	}


	paramMap := map[string]interface{}{}

	if param != nil {
		// convert param to map
		paramBytes, err := json.Marshal(param)
		if err != nil {
			return nil, err
		}
		err = json.Unmarshal(paramBytes, &paramMap)
		if err != nil {
			return nil, err
		}
	}
	
	// Add common request data
	paramMap["customernumber"] = c.customerNumber
	paramMap["apikey"] = c.apiKey
	if c.sessionID != "" {
		paramMap["apisessionid"] = c.sessionID
	}

	request := &requestBody{Action: action, Param: paramMap}
	buf, err := json.Marshal(request)
	if err != nil {
		return nil, err
	}

	response, err := c.httpClient.Post(c.endpoint, "application/json", bytes.NewBuffer(buf))
	if err != nil {
		return nil, err
	}

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	responseBody := responseBody{}
	err = json.Unmarshal(body, &responseBody)
	if err != nil {
		return nil, err
	}

	if responseBody.StatusCode != 2000 {
		return nil, fmt.Errorf(`Request "%s" failed: %s`, action, responseBody.LongMessage)
	}

	return responseBody.ResponseData, nil
}

// Login before calling any other actions
func (c *Client) Login(apiPassword string) error {
	param := struct {
		APIPassword string `json:"apipassword"`
	}{apiPassword}

	buf, err := c.request("login", param)
	if err != nil {
		return err
	}

	responseData := struct {
		APISessionID string `json:"apisessionid"`
	}{}

	err = json.Unmarshal(*buf, &responseData)
	if err != nil {
		return err
	}

	c.sessionID = responseData.APISessionID
	return nil
}

// Logout to close the session
func (c *Client) Logout() error {
	_, err := c.request("logout", nil)
	if err != nil {
		return err
	}

	c.sessionID = ""
	return nil
}

// GetRecords of a domain
func (c *Client) GetRecords(domainname string) ([]Record, error) {
	param := struct {
		DomainName string `json:"domainname"`
	}{domainname}

	buf, err := c.request("infoDnsRecords", param)
	if err != nil {
		return nil, err
	}

	responseData := struct {
		DNSRecords []Record `json:"dnsrecords"`
	}{}

	err = json.Unmarshal(*buf, &responseData)
	if err != nil {
		return nil, err
	}

	c.records = responseData.DNSRecords

	return responseData.DNSRecords, nil
}

// UpdateRecords of a domain
func (c *Client) UpdateRecords(domainname string, records []Record) ([]Record, error) {
	type recordSet struct {
		Records []Record `json:"dnsrecords"`
	}
	param := struct {
		DomainName string `json:"domainname"`
		DNSRecordSet recordSet `json:"dnsrecordset"`

	}{domainname, recordSet{records}}

	buf, err := c.request("updateDnsRecords", param)
	if err != nil {
		return nil, err
	}

	responseData := struct {
		DNSRecords []Record `json:"dnsrecords"`
	}{}

	err = json.Unmarshal(*buf, &responseData)
	if err != nil {
		return nil, err
	}

	c.records = responseData.DNSRecords

	return responseData.DNSRecords, nil
}
