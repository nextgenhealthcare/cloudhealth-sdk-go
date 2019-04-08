package cloudhealth

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"regexp"
	"time"
)

type Clause struct {
	Field    []string `json:"field,omitempty"`
	TagField []string `json:"tag_field,omitempty"`
	Op       string   `json:"op,omitempty"`
	Val      string   `json:"val,omitempty"`
}

type Condition struct {
	CombineWith string   `json:"combine_with,omitempty"`
	Clauses     []Clause `json:"clauses,omitempty"`
}

type Rule struct {
	Type      string     `json:"type,omitempty"`
	Asset     string     `json:"asset,omitempty"`
	To        string     `json:"to,omitempty"`
	RefID     string     `json:"ref_id,omitempty"`    // for type='categorize'
	Name      string     `json:"name,omitempty"`      // for type='categorize'
	Field     []string   `json:"field,omitempty"`     // for type='categorize'
	TagField  []string   `json:"tag_field,omitempty"` // for type='categorize'
	Condition *Condition `json:"condition,omitempty"`
}

type ConstantItem struct {
	RefID   string  `json:"ref_id,omitempty"`
	BlkID   *string `json:"blk_id,omitempty"` // for Dynamic Groups
	Name    string  `json:"name,omitempty"`
	Val     string  `json:"val,omitempty"`      // for Dynamic Groups
	IsOther string  `json:"is_other,omitempty"` // the "Other" for Static Groups
}

type Constant struct {
	Type string         `json:"type,omitempty"`
	List []ConstantItem `json:"list,omitempty"`
}

type Perspective struct {
	Schema Schema `json:"schema"`
}

type Schema struct {
	Name             string        `json:"name"`
	IncludeInReports string        `json:"include_in_reports"`
	Rules            []Rule        `json:"rules"`
	Constants        []Constant    `json:"constants"`
	Merges           []interface{} `json:"merges"` // Not supported
}

type PerspectiveMap map[string]PerspectiveStatus

type PerspectiveStatus struct {
	Name   string `json:"name"`
	Active bool   `json:"active"`
}

type Group map[string]interface{}

var emptyPerspective = Perspective{
	Schema: Schema{
		Name: "Empty",
	},
}

var ErrPerspectiveNotFound = errors.New("Perspective not found")

func (s *Client) GetAllPerspectives() (*PerspectiveMap, error) {
	relativeURL, _ := url.Parse(fmt.Sprintf("perspective_schemas?api_key=%s", s.ApiKey))
	url := s.EndpointURL.ResolveReference(relativeURL)

	req, err := http.NewRequest("GET", url.String(), nil)

	client := &http.Client{
		Timeout: time.Second * 15,
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	responseBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	switch resp.StatusCode {
	case http.StatusOK:
		var perspectives = new(PerspectiveMap)
		err = json.Unmarshal(responseBody, &perspectives)
		if err != nil {
			return nil, err
		}
		return perspectives, nil
	case http.StatusUnauthorized:
		return nil, ErrClientAuthenticationError
	default:
		return nil, fmt.Errorf("Unknown Response with CloudHealth: `%d`", resp.StatusCode)
	}
}

func (s *Client) GetPerspective(id string) (*Perspective, error) {
	relativeURL, _ := url.Parse(fmt.Sprintf("perspective_schemas/%s?api_key=%s", id, s.ApiKey))
	url := s.EndpointURL.ResolveReference(relativeURL)

	req, err := http.NewRequest("GET", url.String(), nil)

	client := &http.Client{
		Timeout: time.Second * 15,
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	responseBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	switch resp.StatusCode {
	case http.StatusOK:
		var perspective = new(Perspective)
		err = json.Unmarshal(responseBody, &perspective)
		if err != nil {
			return nil, err
		}
		return perspective, nil
	case http.StatusUnauthorized:
		return nil, ErrClientAuthenticationError
	case http.StatusNotFound:
		return nil, ErrPerspectiveNotFound
	default:
		return nil, fmt.Errorf("Unknown Response with CloudHealth: `%d`", resp.StatusCode)
	}
}

func (s *Client) CreatePerspective(perspective *Perspective) (string, error) {

	body, _ := json.Marshal(perspective)

	relativeURL, _ := url.Parse(fmt.Sprintf("perspective_schemas/?api_key=%s", s.ApiKey))
	url := s.EndpointURL.ResolveReference(relativeURL)

	req, err := http.NewRequest("POST", url.String(), bytes.NewBuffer(body))

	req.Header.Add("Content-Type", "application/json")

	client := &http.Client{
		Timeout: time.Second * 15,
	}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	responseBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	switch resp.StatusCode {
	case http.StatusOK, http.StatusCreated:
		re := regexp.MustCompile(`Perspective (\d*) created`)
		match := re.FindStringSubmatch(string(responseBody))
		if match == nil || len(match) != 2 {
			return "", fmt.Errorf("Created perspective but didn't understand response to extract ID: %s", responseBody)
		}
		return match[1], nil
	case http.StatusUnauthorized:
		return "", ErrClientAuthenticationError
	case http.StatusNotFound:
		return "", ErrPerspectiveNotFound
	default:
		return "", fmt.Errorf("Unknown Response with CloudHealth: `%d`", resp.StatusCode)
	}
}

func (s *Client) UpdatePerspective(perspectiveID string, perspective Perspective) (*Perspective, error) {

	relativeURL, _ := url.Parse(fmt.Sprintf("perspective_schemas/%s?api_key=%s", perspectiveID, s.ApiKey))
	url := s.EndpointURL.ResolveReference(relativeURL)

	body, _ := json.Marshal(perspective)

	req, err := http.NewRequest("PUT", url.String(), bytes.NewBuffer((body)))
	req.Header.Add("Content-Type", "application/json")

	client := &http.Client{
		Timeout: time.Second * 15,
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	responseBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	switch resp.StatusCode {
	case http.StatusOK:
		var updatedPerspective = new(Perspective)
		err = json.Unmarshal(responseBody, &updatedPerspective)
		if err != nil {
			return nil, err
		}

		return updatedPerspective, nil
	case http.StatusUnauthorized:
		return nil, ErrClientAuthenticationError
	case http.StatusNotFound:
		return nil, ErrPerspectiveNotFound
	case http.StatusUnprocessableEntity:
		return nil, fmt.Errorf("Bad Request. Please check if a Perspective with this name `%s` already exists", perspective.Schema.Name)
	default:
		return nil, fmt.Errorf("Unknown Response with CloudHealth: `%d`", resp.StatusCode)
	}
}

func (s *Client) DeletePerspective(id string) error {

	relativeURL, _ := url.Parse(fmt.Sprintf("perspective_schemas/%s?api_key=%s", id, s.ApiKey))
	url := s.EndpointURL.ResolveReference(relativeURL)

	req, err := http.NewRequest("DELETE", url.String(), nil)

	client := &http.Client{
		Timeout: time.Second * 15,
	}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusOK:
		return nil
	case http.StatusNoContent:
		return nil
	case http.StatusNotFound:
		return ErrPerspectiveNotFound
	case http.StatusUnauthorized:
		return ErrClientAuthenticationError
	default:
		return fmt.Errorf("Unknown Response with CloudHealth: `%d`", resp.StatusCode)
	}
}
