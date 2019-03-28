package cloudhealth

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

// AwsAccounts represents all AWS Accounts enabled in CloudHealth with their configurations.
type AwsAccounts struct {
	AwsAccounts []AwsAccount `json:"aws_accounts"`
}

// AwsAccount represents the configuration of an AWS Account enabled in CloudHealth.
type AwsAccount struct {
	ID               int                      `json:"id"`
	Name             string                   `json:"name"`
	OwnerId          string                   `json:"owner_id,omitempty"`
	HidePublicFields bool                     `json:"hide_public_fields,omitempty"`
	Region           string                   `json:"region,omitempty"`
	CreatedAt        time.Time                `json:"created_at,omitempty"`
	UpdatedAt        time.Time                `json:"updated_at,omitempty"`
	AccountType      string                   `json:"account_type,omitempty"`
	VpcOnly          bool                     `json:"vpc_only,omitempty"`
	ClusterName      string                   `json:"cluster_name,omitempty"`
	Status           AwsAccountStatus         `json:"status,omitempty"`
	Authentication   AwsAccountAuthentication `json:"authentication"`
}

// AwsAccountStatus represents the status details for AWS integration.
type AwsAccountStatus struct {
	Level      string    `json:"level"`
	LastUpdate time.Time `json:"last_update,omitempty"`
}

// AwsAccountAuthentication represents the authentication details for AWS integration.
type AwsAccountAuthentication struct {
	Protocol             string `json:"protocol"`
	AccessKey            string `json:"access_key,omitempty"`
	SecreyKey            string `json:"secret_key,omitempty"`
	AssumeRoleArn        string `json:"assume_role_arn,omitempty"`
	AssumeRoleExternalID string `json:"assume_role_external_id,omitempty"`
}

// ErrAwsAccountNotFound is returned when an AWS Account doesn't exist on a Read or Delete.
// It's useful for ignoring errors (e.g. delete if exists).
var ErrAwsAccountNotFound = errors.New("AWS Account not found")

// GetAwsAccount gets the AWS Account with the specified CloudHealth Account ID.
func (s *Client) GetAwsAccount(id int) (*AwsAccount, error) {

	relativeURL, _ := url.Parse(fmt.Sprintf("aws_accounts/%d?api_key=%s", id, s.ApiKey))
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
		var account = new(AwsAccount)
		err = json.Unmarshal(responseBody, &account)
		if err != nil {
			return nil, err
		}

		return account, nil
	case http.StatusUnauthorized:
		return nil, ErrClientAuthenticationError
	case http.StatusNotFound:
		return nil, ErrAwsAccountNotFound
	default:
		return nil, fmt.Errorf("Unknown Response with CloudHealth: `%d`", resp.StatusCode)
	}
}

// GetAwsAccounts gets all AWS Accounts enabled in CloudHealth.
func (s *Client) GetAwsAccounts() (*AwsAccounts, error) {
	awsaccounts := new(AwsAccounts)
	err := iterateOverAwsAccountsPages(s, awsaccounts, 1)
	if err != nil {
		return nil, err
	}
	return awsaccounts, nil
}

// iterateOverAwsAccountsPages iterates over all pages returned by CloudHealth for listing enabled AWS accounts.
func iterateOverAwsAccountsPages(s *Client, awsaccounts *AwsAccounts, page int) error {
	params := url.Values{"page": {strconv.Itoa(page)}, "per_page": {"100"}, "api_key": {s.ApiKey}}
	relativeURL, _ := url.Parse(fmt.Sprintf("aws_accounts/?%s", params.Encode()))
	url := s.EndpointURL.ResolveReference(relativeURL)

	req, err := http.NewRequest("GET", url.String(), nil)

	client := &http.Client{
		Timeout: time.Second * 15,
	}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	responseBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	var pageofaccounts = new(AwsAccounts)
	switch resp.StatusCode {
	case http.StatusOK:
		err = json.Unmarshal(responseBody, &pageofaccounts)
		if err != nil {
			return err
		}
	case http.StatusUnauthorized:
		return ErrClientAuthenticationError
	default:
		return fmt.Errorf("Unknown Response with CloudHealth: `%d`", resp.StatusCode)
	}

	for _, a := range pageofaccounts.AwsAccounts {
		awsaccounts.AwsAccounts = append(awsaccounts.AwsAccounts, a)
	}

	if len(pageofaccounts.AwsAccounts) < 100 {
		return nil
	}

	return iterateOverAwsAccountsPages(s, awsaccounts, page+1)
}

// CreateAwsAccount enables a new AWS Account in CloudHealth.
func (s *Client) CreateAwsAccount(account AwsAccount) (*AwsAccount, error) {

	body, _ := json.Marshal(account)

	relativeURL, _ := url.Parse(fmt.Sprintf("aws_accounts?api_key=%s", s.ApiKey))
	url := s.EndpointURL.ResolveReference(relativeURL)

	req, err := http.NewRequest("POST", url.String(), bytes.NewBuffer(body))
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
	case http.StatusCreated:
		var account = new(AwsAccount)
		err = json.Unmarshal(responseBody, &account)
		if err != nil {
			return nil, err
		}

		return account, nil
	case http.StatusUnauthorized:
		return nil, ErrClientAuthenticationError
	case http.StatusUnprocessableEntity:
		return nil, fmt.Errorf("Bad Request. Please check if a AWS Account with this name `%s` already exists", account.Name)
	default:
		return nil, fmt.Errorf("Unknown Response with CloudHealth: `%d`", resp.StatusCode)
	}
}

// UpdateAwsAccount updates an existing AWS Account in CloudHealth.
func (s *Client) UpdateAwsAccount(account AwsAccount) (*AwsAccount, error) {

	relativeURL, _ := url.Parse(fmt.Sprintf("aws_accounts/%d?api_key=%s", account.ID, s.ApiKey))
	url := s.EndpointURL.ResolveReference(relativeURL)

	body, _ := json.Marshal(account)

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
		var account = new(AwsAccount)
		err = json.Unmarshal(responseBody, &account)
		if err != nil {
			return nil, err
		}

		return account, nil
	case http.StatusUnauthorized:
		return nil, ErrClientAuthenticationError
	case http.StatusUnprocessableEntity:
		return nil, fmt.Errorf("Bad Request. Please check if a AWS Account with this name `%s` already exists", account.Name)
	default:
		return nil, fmt.Errorf("Unknown Response with CloudHealth: `%d`", resp.StatusCode)
	}
}

// DeleteAwsAccount removes the AWS Account with the specified CloudHealth ID.
func (s *Client) DeleteAwsAccount(id int) error {

	relativeURL, _ := url.Parse(fmt.Sprintf("aws_accounts/%d?api_key=%s", id, s.ApiKey))
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
		return ErrAwsAccountNotFound
	case http.StatusUnauthorized:
		return ErrClientAuthenticationError
	default:
		return fmt.Errorf("Unknown Response with CloudHealth: `%d`", resp.StatusCode)
	}
}
