package cloudhealth

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
)

var defaultAWSAccount = AwsAccount{
	ID:   1234567890,
	Name: "test",
}

var defaultAWSAccounts = AwsAccounts{
	AwsAccounts: []AwsAccount{
		AwsAccount{
			ID:   1234567890,
			Name: "test",
		},
		AwsAccount{
			ID:   9876543210,
			Name: "tset",
		},
	},
}

func TestGetAwsAccountOK(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		if r.Method != "GET" {
			t.Errorf("Expected ‘GET’ request, got ‘%s’", r.Method)
		}
		expectedURL := fmt.Sprintf("/aws_accounts/%d", defaultAWSAccount.ID)
		if r.URL.EscapedPath() != expectedURL {
			t.Errorf("Expected request to ‘%s’, got ‘%s’", expectedURL, r.URL.EscapedPath())
		}
		body, _ := json.Marshal(defaultAWSAccount)
		w.Write(body)
	}))
	defer ts.Close()

	c, err := NewClient("apiKey", ts.URL)
	if err != nil {
		t.Errorf("NewClient() returned an error: %s", err)
		return
	}

	returnedAwsAccount, err := c.GetAwsAccount(defaultAWSAccount.ID)
	if err != nil {
		t.Errorf("GetAwsAccount() returned an error: %s", err)
		return
	}
	if returnedAwsAccount.ID != defaultAWSAccount.ID {
		t.Errorf("GetAwsAccount() expected ID `%d`, got `%d`", defaultAWSAccount.ID, returnedAwsAccount.ID)
		return
	}
}

func TestGetAwsAccountsOK(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		if r.Method != "GET" {
			t.Errorf("Expected ‘GET’ request, got ‘%s’", r.Method)
		}
		expectedURL := "/aws_accounts/"
		if r.URL.EscapedPath() != expectedURL {
			t.Errorf("Expected request to ‘%s’, got ‘%s’", expectedURL, r.URL.EscapedPath())
		}
		body, _ := json.Marshal(defaultAWSAccounts)
		w.Write(body)
	}))
	defer ts.Close()

	c, err := NewClient("apiKey", ts.URL)
	if err != nil {
		t.Errorf("NewClient() returned an error: %s", err)
		return
	}

	returnedAwsAccounts, err := c.GetAwsAccounts()
	if err != nil {
		t.Errorf("NewClient() returned an error: %s", err)
		return
	}
	fmt.Println(returnedAwsAccounts.AwsAccounts)
	if len(returnedAwsAccounts.AwsAccounts) != 2 {
		t.Errorf("All accounts have not be retrieved")
		return
	}
	if returnedAwsAccounts.AwsAccounts[0].ID != defaultAWSAccounts.AwsAccounts[0].ID && returnedAwsAccounts.AwsAccounts[0].ID != defaultAWSAccounts.AwsAccounts[1].ID {
		t.Errorf("Retrieved accounts don't match")
		return
	}
	return
}

func TestGetAwsAccountDoesntExist(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		if r.Method != "GET" {
			t.Errorf("Expected ‘GET’ request, got ‘%s’", r.Method)
		}
		expectedURL := fmt.Sprintf("/aws_accounts/%d", defaultAWSAccount.ID)
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

	_, err = c.GetAwsAccount(defaultAWSAccount.ID)
	if err != ErrAwsAccountNotFound {
		t.Errorf("GetAwsAccount() returned the wrong error: %s", err)
		return
	}
}

func TestCreateAwsAccountOk(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
		if r.Method != "POST" {
			t.Errorf("Expected ‘POST’ request, got ‘%s’", r.Method)
		}
		if r.URL.EscapedPath() != "/aws_accounts" {
			t.Errorf("Expected request to ‘/aws_accounts, got ‘%s’", r.URL.EscapedPath())
		}
		if ctype := r.Header.Get("Content-Type"); ctype != "application/json" {
			t.Errorf("Expected response to be content-type ‘application/json’, got ‘%s’", ctype)
		}
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			t.Error("Unable to read response body")
		}

		account := new(AwsAccount)
		err = json.Unmarshal(body, &account)
		if err != nil {
			t.Errorf("Unable to unmarshal AwsAccount, got `%s`", body)
		}
		if account.Name != "test" {
			t.Errorf("Expected request to include AWS Account name ‘test’, got ‘%s’", account.Name)
		}
		account.ID = 1234567890
		js, err := json.Marshal(account)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Write(js)
	}))
	defer ts.Close()

	c, err := NewClient("apiKey", ts.URL)
	if err != nil {
		t.Errorf("NewClient() returned an error: %s", err)
		return
	}

	returnedAccount, err := c.CreateAwsAccount(AwsAccount{
		Name: "test",
	})
	if err != nil {
		t.Errorf("CreateAwsAccount() returned an error: %s", err)
		return
	}
	if returnedAccount.ID != 1234567890 {
		t.Errorf("CreateAwsAccount() expected ID 1234567890, got `%d`", returnedAccount.ID)
		return
	}
}

func TestUpdateAwsAccountAlreadyExists(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnprocessableEntity)
		if r.Method != "POST" {
			t.Errorf("Expected ‘POST’ request, got ‘%s’", r.Method)
		}
		if r.URL.EscapedPath() != "/aws_accounts" {
			t.Errorf("Expected request to ‘/aws_accounts, got ‘%s’", r.URL.EscapedPath())
		}
		if ctype := r.Header.Get("Content-Type"); ctype != "application/json" {
			t.Errorf("Expected response to be content-type ‘application/json’, got ‘%s’", ctype)
		}
	}))
	defer ts.Close()

	c, err := NewClient("apiKey", ts.URL)
	if err != nil {
		t.Errorf("NewClient() returned an error: %s", err)
		return
	}

	_, err = c.CreateAwsAccount(AwsAccount{
		Name: "test",
	})
	if err == nil {
		t.Errorf("CreateAwsAccount() did not return an error: %s", err)
		return
	}
}

func TestUpdateAwsAccountOK(t *testing.T) {
	updatedAwsAccount := defaultAWSAccount
	updatedAwsAccount.Name = "Updated"
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		if r.Method != "PUT" {
			t.Errorf("Expected ‘PUT’ request, got ‘%s’", r.Method)
		}
		expectedURL := fmt.Sprintf("/aws_accounts/%d", defaultAWSAccount.ID)
		if r.URL.EscapedPath() != expectedURL {
			t.Errorf("Expected request to ‘%s’, got ‘%s’", expectedURL, r.URL.EscapedPath())
		}
		body, _ := json.Marshal(updatedAwsAccount)
		w.Write(body)
	}))
	defer ts.Close()

	c, err := NewClient("apiKey", ts.URL)
	if err != nil {
		t.Errorf("NewClient() returned an error: %s", err)
		return
	}

	returnedAwsAccount, err := c.UpdateAwsAccount(updatedAwsAccount)
	if err != nil {
		t.Errorf("UpdateAwsAccount() returned an error: %s", err)
		return
	}
	if returnedAwsAccount.ID != updatedAwsAccount.ID {
		t.Errorf("UpdateAwsAccount() expected ID `%d`, got `%d`", defaultAWSAccount.ID, returnedAwsAccount.ID)
		return
	}
	if returnedAwsAccount.Name == defaultAWSAccount.Name {
		t.Errorf("UpdateAwsAccount() did not update the name")
		return
	}
}

func TestUpdateAwsAccountNameConflict(t *testing.T) {
	updatedAwsAccount := defaultAWSAccount
	updatedAwsAccount.Name = "Updated"
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnprocessableEntity)
		if r.Method != "PUT" {
			t.Errorf("Expected ‘PUT’ request, got ‘%s’", r.Method)
		}
		expectedURL := fmt.Sprintf("/aws_accounts/%d", defaultAWSAccount.ID)
		if r.URL.EscapedPath() != expectedURL {
			t.Errorf("Expected request to ‘%s’, got ‘%s’", expectedURL, r.URL.EscapedPath())
		}
		body, _ := json.Marshal(updatedAwsAccount)
		w.Write(body)
	}))
	defer ts.Close()

	c, err := NewClient("apiKey", ts.URL)
	if err != nil {
		t.Errorf("NewClient() returned an error: %s", err)
		return
	}

	_, err = c.UpdateAwsAccount(updatedAwsAccount)
	if err == nil {
		t.Errorf("UpdateAwsAccount() did not return an error: %s", err)
		return
	}
}

func TestDeleteAwsAccountOK(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		if r.Method != "DELETE" {
			t.Errorf("Expected ‘DELETE’ request, got ‘%s’", r.Method)
		}
		expectedURL := fmt.Sprintf("/aws_accounts/%d", defaultAWSAccount.ID)
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

	err = c.DeleteAwsAccount(defaultAWSAccount.ID)
	if err != nil {
		t.Errorf("DeleteAwsAccount() returned an error: %s", err)
		return
	}
}

func TestDeleteAwsAccountDoesntExist(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		if r.Method != "DELETE" {
			t.Errorf("Expected ‘DELETE’ request, got ‘%s’", r.Method)
		}
		expectedURL := fmt.Sprintf("/aws_accounts/%d", defaultAWSAccount.ID)
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

	err = c.DeleteAwsAccount(defaultAWSAccount.ID)
	if err != ErrAwsAccountNotFound {
		t.Errorf("DeleteAwsAccount() returned the wrong error: %s", err)
		return
	}
}
