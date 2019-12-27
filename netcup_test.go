package netcup

import (
	"io/ioutil"
	"reflect"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

var (
	// mux is the HTTP request multiplexer used with the test server.
	mux *http.ServeMux

	// client is the dnspod client being tested.
	client *Client

	// server is a test HTTP server used to provide mock API responses.
	server *httptest.Server
)

// This method of testing http client APIs is borrowed from
// Will Norris's work in go-github @ https://github.com/google/go-github
func setup() {
	mux = http.NewServeMux()
	server = httptest.NewServer(mux)

	client = NewClient(1234, "key")
	client.endpoint = server.URL
}

func teardown() {
	server.Close()
}

func TestLogin(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		want := `{"action":"login","param":{"apikey":"key","apipassword":"password","customernumber":1234}}`
		body, err := ioutil.ReadAll(r.Body)

		if err != nil || string(body) != want {
			t.Error("Client did not send correct login request.")
		}

		fmt.Fprint(w, `{
			"serverrequestid": "",
			"clientrequestid": "",
			"action": "login",
			"status": "success",
			"statuscode": 2000,
			"shortmessage": "Login successful",
			"longmessage": "Session has been created successful.",
			"responsedata": {
			  "apisessionid": "thisisasessionid"
			}
		  }`)
	})

	err := client.Login("password")
	if err != nil {
		t.Error(err)
	}

	if client.sessionID != "thisisasessionid" {
		t.Error("Client did not get correct session ID.")
	}
}

func TestLogout(t *testing.T) {
	setup()
	defer teardown()

	// fake login
	client.sessionID = "thisisasessionid"

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		want := `{"action":"logout","param":{"apikey":"key","apisessionid":"thisisasessionid","customernumber":1234}}`
		body, err := ioutil.ReadAll(r.Body)

		if err != nil || string(body) != want {
			t.Error("Client did not send correct logout request.")
		}

		fmt.Fprint(w, `{
			"serverrequestid": "",
			"clientrequestid": "",
			"action": "logout",
			"status": "success",
			"statuscode": 2000,
			"shortmessage": "Logout successful",
			"longmessage": "Session has been terminated successful.",
			"responsedata": ""
		  }`)
	})

	err := client.Logout()
	if err != nil {
		t.Error(err)
	}

	if client.sessionID != "" {
		t.Error("Client did not unset session ID.")
	}
}

func TestInfoDNSRecords(t *testing.T) {
	setup()
	defer teardown()

	// fake login
	client.sessionID = "thisisasessionid"

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		want := `{"action":"infoDnsRecords","param":{"apikey":"key","apisessionid":"thisisasessionid","customernumber":1234,"domainname":"example.com"}}`
		body, err := ioutil.ReadAll(r.Body)

		if err != nil || string(body) != want {
			t.Error("Client did not send correct infoDnsRecords request.")
		}

		fmt.Fprint(w, `{
			"serverrequestid": "",
			"clientrequestid": "",
			"action": "infoDnsRecords",
			"status": "success",
			"statuscode": 2000,
			"shortmessage": "DNS records found",
			"longmessage": "DNS Records for this zone were found.",
			"responsedata": {
			  "dnsrecords": [
				{
				  "id": "123451",
				  "hostname": "@",
				  "type": "A",
				  "priority": "0",
				  "destination": "127.0.0.1",
				  "deleterecord": false,
				  "state": "yes"
				},
				{
				  "id": "123452",
				  "hostname": "@",
				  "type": "MX",
				  "priority": "10",
				  "destination": "mail.example.com",
				  "deleterecord": false,
				  "state": "yes"
				},
				{
				  "id": "123453",
				  "hostname": "mail",
				  "type": "AAAA",
				  "priority": "0",
				  "destination": "1234:5678:90ab:cdef::1",
				  "deleterecord": false,
				  "state": "yes"
				},
				{
				  "id": "123454",
				  "hostname": "mail",
				  "type": "A",
				  "priority": "0",
				  "destination": "127.0.0.1",
				  "deleterecord": false,
				  "state": "yes"
				},
				{
				  "id": "123455",
				  "hostname": "www",
				  "type": "CNAME",
				  "priority": "0",
				  "destination": "@",
				  "deleterecord": false,
				  "state": "yes"
				}
			  ]
			}
		  }`)
	})

	records, err := client.GetRecords("example.com")
	if err != nil {
		t.Error(err)
	}

	wantRecords := []Record{
		Record{ID: "123451", Hostname: "@", Type: "A", Priority: "0", Destination: "127.0.0.1", DeleteRecord: false, State: "yes"},
		Record{ID: "123452", Hostname: "@", Type: "MX", Priority: "10", Destination: "mail.example.com", DeleteRecord: false, State: "yes"},
		Record{ID: "123453", Hostname: "mail", Type: "AAAA", Priority: "0", Destination: "1234:5678:90ab:cdef::1", DeleteRecord: false, State: "yes"},
		Record{ID: "123454", Hostname: "mail", Type: "A", Priority: "0", Destination: "127.0.0.1", DeleteRecord: false, State: "yes"},
		Record{ID: "123455", Hostname: "www", Type: "CNAME", Priority: "0", Destination: "@", DeleteRecord: false, State: "yes"},
	}

	if !reflect.DeepEqual(wantRecords, records) {
		t.Error("GetRecords did not return the expected records")
	}
}

func TestUpdateDNSRecords(t *testing.T) {
	setup()
	defer teardown()

	// fake login
	client.sessionID = "thisisasessionid"

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		want := `{"action":"updateDnsRecords","param":{"apikey":"key","apisessionid":"thisisasessionid","customernumber":1234,"dnsrecordset":{"dnsrecords":[{"deleterecord":false,"destination":"test","hostname":"@","id":"","priority":"0","state":"yes","type":"TXT"}]},"domainname":"example.com"}}`
		body, err := ioutil.ReadAll(r.Body)

		if err != nil || string(body) != want {
			t.Error("Client did not send correct updateDnsRecords request.")
		}

		fmt.Fprint(w, `{
			"serverrequestid": "",
			"clientrequestid": "",
			"action": "updateDnsRecords",
			"status": "success",
			"statuscode": 2000,
			"shortmessage": "DNS records successful updated",
			"longmessage": "The given DNS records for this zone were updated.",
			"responsedata": {
			  "dnsrecords": [
				{
				  "id": "123451",
				  "hostname": "@",
				  "type": "A",
				  "priority": "0",
				  "destination": "127.0.0.1",
				  "deleterecord": false,
				  "state": "yes"
				},
				{
				  "id": "123452",
				  "hostname": "@",
				  "type": "MX",
				  "priority": "10",
				  "destination": "mail.example.com",
				  "deleterecord": false,
				  "state": "yes"
				},
				{
				  "id": "123453",
				  "hostname": "mail",
				  "type": "AAAA",
				  "priority": "0",
				  "destination": "1234:5678:90ab:cdef::1",
				  "deleterecord": false,
				  "state": "yes"
				},
				{
				  "id": "123454",
				  "hostname": "mail",
				  "type": "A",
				  "priority": "0",
				  "destination": "127.0.0.1",
				  "deleterecord": false,
				  "state": "yes"
				},
				{
				  "id": "123455",
				  "hostname": "www",
				  "type": "CNAME",
				  "priority": "0",
				  "destination": "@",
				  "deleterecord": false,
				  "state": "yes"
				},
				{
				  "id": "123456",
				  "hostname": "@",
				  "type": "TXT",
				  "priority": "0",
				  "destination": "test",
				  "deleterecord": false,
				  "state": "yes"
				}
			  ]
			}
		  }`)
	})
	
	updateRecord := Record{ID: "", Hostname: "@", Type: "TXT", Priority: "0", Destination: "test", DeleteRecord: false, State: "yes"}
		
	records, err := client.UpdateRecords("example.com", []Record{updateRecord})
	if err != nil {
		t.Error(err)
	}

	wantRecords := []Record{
		Record{ID: "123451", Hostname: "@", Type: "A", Priority: "0", Destination: "127.0.0.1", DeleteRecord: false, State: "yes"},
		Record{ID: "123452", Hostname: "@", Type: "MX", Priority: "10", Destination: "mail.example.com", DeleteRecord: false, State: "yes"},
		Record{ID: "123453", Hostname: "mail", Type: "AAAA", Priority: "0", Destination: "1234:5678:90ab:cdef::1", DeleteRecord: false, State: "yes"},
		Record{ID: "123454", Hostname: "mail", Type: "A", Priority: "0", Destination: "127.0.0.1", DeleteRecord: false, State: "yes"},
		Record{ID: "123455", Hostname: "www", Type: "CNAME", Priority: "0", Destination: "@", DeleteRecord: false, State: "yes"},
		Record{ID: "123456", Hostname: "@", Type: "TXT", Priority: "0", Destination: "test", DeleteRecord: false, State: "yes"},
	}

	if !reflect.DeepEqual(wantRecords, records) {
		t.Error("UpdateRecords did not return the expected records")
	}
}