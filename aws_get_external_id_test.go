package cloudhealth

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

var defaultAwsExternalID = AwsExternalID{
	ExternalID: "1234567890",
}

func TestGetAwsExternalIDOk(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		if r.Method != "GET" {
			t.Errorf("Expected ‘GET’ request, got ‘%s’", r.Method)
		}
		expectedURL := fmt.Sprintf("/aws_accounts/:id/generate_external_id")
		if r.URL.EscapedPath() != expectedURL {
			t.Errorf("Expected request to ‘%s’, got ‘%s’", expectedURL, r.URL.EscapedPath())
		}
		body, _ := json.Marshal(defaultAwsExternalID)
		w.Write(body)
	}))
	defer ts.Close()

	c, err := NewClient("apiKey", ts.URL)
	if err != nil {
		t.Errorf("NewClient() returned an error: %s", err)
		return
	}

	returnedAwsExternalID, err := c.GetAwsExternalID()
	if err != nil {
		t.Errorf("GetAwsExternalID() returned an error: %s", err)
		return
	}
	if returnedAwsExternalID != defaultAwsExternalID.ExternalID {
		t.Errorf("GetAwsExternalID() expected ID `%s`, got `%s`", defaultAwsExternalID.ExternalID, returnedAwsExternalID)
		return
	}
}
