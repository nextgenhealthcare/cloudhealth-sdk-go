package cloudhealth

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
)

var defaultPerspectiveID = "1234567839263"

var defaultPerspective = Perspective{
	Schema: Schema{
		Name:             "test",
		IncludeInReports: "true",
	},
}

var defaultPerspectiveMap = PerspectiveMap{
	defaultPerspectiveID: PerspectiveStatus{
		Name: "test",
	},
}

func TestGetPerspectiveOK(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		if r.Method != "GET" {
			t.Errorf("Expected ‘GET’ request, got ‘%s’", r.Method)
		}
		expectedURL := fmt.Sprintf("/perspective_schemas/%s", defaultPerspectiveID)
		if r.URL.EscapedPath() != expectedURL {
			t.Errorf("Expected request to ‘%s’, got ‘%s’", expectedURL, r.URL.EscapedPath())
		}
		body, _ := json.Marshal(defaultPerspective)
		w.Write(body)
	}))
	defer ts.Close()

	c, err := NewClient("apiKey", ts.URL)
	if err != nil {
		t.Errorf("NewClient() returned an error: %s", err)
		return
	}

	returnedPerspective, err := c.GetPerspective(defaultPerspectiveID)
	if err != nil {
		t.Errorf("GetPerspective() returned an error: %s", err)
		return
	}
	if reflect.DeepEqual(&returnedPerspective, defaultPerspective) {
		t.Errorf("GetPerspective() returned something unexpected")
		return
	}
}

func TestGetPerspectiveDoesntExist(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		if r.Method != "GET" {
			t.Errorf("Expected ‘GET’ request, got ‘%s’", r.Method)
		}
		expectedURL := fmt.Sprintf("/perspective_schemas/%s", defaultPerspectiveID)
		if r.URL.EscapedPath() != expectedURL {
			t.Errorf("Expected request to ‘%s’, got ‘%s’", expectedURL, r.URL.EscapedPath())
		}
		body, _ := json.Marshal(defaultPerspective)
		w.Write(body)
	}))
	defer ts.Close()

	c, err := NewClient("apiKey", ts.URL)
	if err != nil {
		t.Errorf("NewClient() returned an error: %s", err)
		return
	}

	_, err = c.GetPerspective(defaultPerspectiveID)
	if err != ErrPerspectiveNotFound {
		t.Errorf("GetPerspective() returned the wrong error: %s", err)
		return
	}
}

func TestGetAllPerspectivesOK(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		if r.Method != "GET" {
			t.Errorf("Expected ‘GET’ request, got ‘%s’", r.Method)
		}
		expectedURL := "/perspective_schemas"
		if r.URL.EscapedPath() != expectedURL {
			t.Errorf("Expected request to ‘%s’, got ‘%s’", expectedURL, r.URL.EscapedPath())
		}
		body, _ := json.Marshal(defaultPerspectiveMap)
		w.Write(body)
	}))
	defer ts.Close()

	c, err := NewClient("apiKey", ts.URL)
	if err != nil {
		t.Errorf("NewClient() returned an error: %s", err)
		return
	}

	perspectives, err := c.GetAllPerspectives()
	if err != nil {
		t.Errorf("GetAllPerspectives() returned the error: %s", err)
		return
	}

	if !reflect.DeepEqual(defaultPerspectiveMap, *perspectives) {
		t.Errorf("GetAllPerspectives() result:\n%#v\n not equal to expected value:\n%#v", *perspectives, defaultPerspectiveMap)
	}
}

func TestCreatePerspectiveOK(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
		if r.Method != "POST" {
			t.Errorf("Expected ‘POST’ request, got ‘%s’", r.Method)
		}
		if r.URL.EscapedPath() != "/perspective_schemas/" {
			t.Errorf("Expected request to ‘/perspective_schemas/, got ‘%s’", r.URL.EscapedPath())
		}
		if ctype := r.Header.Get("Content-Type"); ctype != "application/json" {
			t.Errorf("Expected response to be content-type ‘application/json’, got ‘%s’", ctype)
		}
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			t.Error("Unable to read request body")
		}

		perspective := new(Perspective)
		err = json.Unmarshal(body, &perspective)
		if err != nil {
			t.Errorf("Unable to unmarshal Perspective, got `%s`, error:\n%s", body, err)
		}
		if perspective.Schema.Name != "test" {
			t.Errorf("Expected request to include Perspective Schema name ‘test’, got ‘%s’", perspective.Schema.Name)
		}

		resp := fmt.Sprintf("Perspective %s created\n", defaultPerspectiveID)
		w.Write([]byte(resp))
	}))
	defer ts.Close()

	c, err := NewClient("apiKey", ts.URL)
	if err != nil {
		t.Errorf("NewClient() returned an error: %s", err)
		return
	}

	returnedID, err := c.CreatePerspective(&defaultPerspective)
	if err != nil {
		t.Errorf("CreatePerspective() returned an error: %v", err)
		return
	}
	if returnedID != defaultPerspectiveID {
		t.Errorf("CreatePerspective() expected ID `%s`, got `%s`", defaultPerspectiveID, returnedID)
		return
	}
}

func TestUpdatePerspectiveOK(t *testing.T) {
	updatedPerspective := defaultPerspective
	updatedPerspective.Schema.IncludeInReports = "false"
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		if r.Method != "PUT" {
			t.Errorf("Expected ‘PUT’ request, got ‘%s’", r.Method)
		}
		expectedURL := fmt.Sprintf("/perspective_schemas/%s", defaultPerspectiveID)
		if r.URL.EscapedPath() != expectedURL {
			t.Errorf("Expected request to ‘%s’, got ‘%s’", expectedURL, r.URL.EscapedPath())
		}
		body, _ := json.Marshal(updatedPerspective)
		w.Write(body)
	}))
	defer ts.Close()

	c, err := NewClient("apiKey", ts.URL)
	if err != nil {
		t.Errorf("NewClient() returned an error: %s", err)
		return
	}

	returnedPerspective, err := c.UpdatePerspective(defaultPerspectiveID, updatedPerspective)
	if err != nil {
		t.Errorf("UpdatePerspective() returned an error: %s", err)
		return
	}
	if returnedPerspective.Schema.Name != updatedPerspective.Schema.Name {
		t.Errorf("UpdatePerspective() expected Schema.Name `%s`, got `%s`", updatedPerspective.Schema.Name, returnedPerspective.Schema.Name)
		return
	}
	if returnedPerspective.Schema.IncludeInReports == defaultPerspective.Schema.IncludeInReports {
		t.Errorf("UpdatePerspective() did not update include_in_reports")
		return
	}
}

func TestDeletePerspectiveOK(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		if r.Method != "DELETE" {
			t.Errorf("Expected ‘DELETE’ request, got ‘%s’", r.Method)
		}
		expectedURL := fmt.Sprintf("/perspective_schemas/%s", defaultPerspectiveID)
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

	err = c.DeletePerspective(defaultPerspectiveID)
	if err != nil {
		t.Errorf("DeletePerspective() returned an error: %s", err)
		return
	}
}

func TestDeletePerspectiveDoesntExist(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		if r.Method != "DELETE" {
			t.Errorf("Expected ‘DELETE’ request, got ‘%s’", r.Method)
		}
		expectedURL := fmt.Sprintf("/perspective_schemas/%s", defaultPerspectiveID)
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

	err = c.DeletePerspective(defaultPerspectiveID)
	if err != ErrPerspectiveNotFound {
		t.Errorf("DeletePerspective() returned the wrong error: %s", err)
		return
	}
}
