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
		t.Error("That's not good!")
	}
}

func TestParseServerStatus(t *testing.T) {
	ts := httptest.NewServer(http.FileServer(http.Dir("../test/http")))
	url := fmt.Sprintf("%s/index.html", ts.URL)
	serverStatus := getServerStatus(url)
	if serverStatus == nil {
		t.Error("That's not good!")
	}
	stringResults := parseServerStatus(serverStatus)
	if stringResults == nil {
		t.Error("There are no results.")
	}
}

func TestParseServerStats(t *testing.T) {
	ts := httptest.NewServer(http.FileServer(http.Dir("../test/http")))
	url := fmt.Sprintf("%s/index.html", ts.URL)
	serverStatus := getServerStatus(url)
	stringResults := parseServerStatus(serverStatus)
	ApacheProcesses := parseProcessStats(stringResults)
	if ApacheProcesses == nil {
		t.Error("No process data - that's bad.")
	}
}
