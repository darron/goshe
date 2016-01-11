package cmd

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGetServerStatus(t *testing.T) {
	ts := httptest.NewServer(http.FileServer(http.Dir("../test/http")))
	url := fmt.Sprintf("%s/index.html", ts.URL)
	serverStatus := getServerStatus(url)
	if serverStatus == nil {
		t.Error("That's not good! Not getting html.Nodes back.")
	}
}

func TestParseServerStatus(t *testing.T) {
	ts := httptest.NewServer(http.FileServer(http.Dir("../test/http")))
	url := fmt.Sprintf("%s/index.html", ts.URL)
	serverStatus := getServerStatus(url)
	stringResults := parseServerStatus(serverStatus)
	if len(stringResults) != 64 {
		t.Error("We're not getting the right number of strings back. Should be 64.")
	}
}

func TestParseServerStats(t *testing.T) {
	ts := httptest.NewServer(http.FileServer(http.Dir("../test/http")))
	url := fmt.Sprintf("%s/index.html", ts.URL)
	serverStatus := getServerStatus(url)
	stringResults := parseServerStatus(serverStatus)
	ApacheProcesses := parseProcessStats(stringResults)
	if len(ApacheProcesses) != 9 {
		t.Error("That's bad - we should see 9 Apache structs.")
	}
}
