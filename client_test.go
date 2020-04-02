package cloudhealth

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestBadApiKey(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		if r.Method != "GET" {
			t.Errorf("Expected ‘GET’ request, got ‘%s’", r.Method)
		}
		expectedURL := fmt.Sprintf("/aws_accounts/:id/generate_external_id")
		if r.URL.EscapedPath() != expectedURL {
			t.Errorf("Expected request to ‘%s’, got ‘%s’", expectedURL, r.URL.EscapedPath())
		}
	}))
	defer ts.Close()

	c, err := NewClient("apiKey", ts.URL)
	if err != nil {
		t.Errorf("NewClient() returned an error: %s", err)
		return
	}

	_, err = c.GetAwsExternalID()
	if err != ErrClientAuthenticationError {
		t.Errorf("GetAwsExternalID() returned the wrong error: %s", err)
		return
	}
}

func TestTimeoutArg(t *testing.T) {
	testTimeout := 42
	c, err := NewClient("apiKey", "https://api.foo.bar", testTimeout)
	if err != nil {
		t.Errorf("NewClient() returned an error: %s", err)
		return
	}
	if c.Timeout != testTimeout {
		t.Errorf("Unexpected NewClient() Timeout value: %d != %d", c.Timeout, testTimeout)
		return
	}
}

func TestDefaultTimeout(t *testing.T) {
	defaultTimeout := 15
	c, err := NewClient("apiKey", "https://api.foo.bar")
	if err != nil {
		t.Errorf("NewClient() returned an error: %s", err)
		return
	}
	if c.Timeout != defaultTimeout {
		t.Errorf("Unexpected NewClient() Timeout value: %d != %d", c.Timeout, defaultTimeout)
		return
	}
}
