package ovh

import (
	"bytes"
	"crypto/sha1"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/user"
	"strings"
	"time"

	"gopkg.in/ini.v1"
)

// DefaultTimeout api requests after 180s
const DefaultTimeout = 180

// Endpoint reprensents an API endpoint
type Endpoint string

// Endpoints conveniently maps endpoints names to their real URI
var Endpoints = map[string]Endpoint{
	"ovh-eu":        Endpoint("https://eu.api.ovh.com/1.0"),
	"ovh-ca":        Endpoint("https://ca.api.ovh.com/1.0"),
	"kimsufi-eu":    Endpoint("https://eu.api.kimsufi.com/1.0"),
	"kimsufi-ca":    Endpoint("https://ca.api.kimsufi.com/1.0"),
	"soyoustart-eu": Endpoint("https://eu.api.soyoustart.com/1.0"),
	"soyoustart-ca": Endpoint("https://ca.api.soyoustart.com/1.0"),
	"runabove-ca":   Endpoint("https://api.runabove.com/1.0"),
}

// Client represents an an OVH API client
type Client struct {
	endpoint          Endpoint
	applicationKey    string
	applicationSecret string
	consumerKey       string
	Timeout           time.Duration
	timeDelta         int64
	client            *http.Client

	// sync.Once would consider init done, even in case of error
	// running it multiple times/races are not issue. Hence a good
	// old flag
	timeDeltaDone bool
}

// APIResponse represents a response from OVH API
type APIResponse struct {
	StatusCode int
	Status     string
	Body       []byte
}

// APIError represents an unmarshalled reponse from OVH in case of error
type APIError struct {
	ErrorCode string `json:"errorCode"`
	HTTPCode  string `json:"httpCode"`
	Message   string `json:"message"`
}

// Util: get user home
func currentUserHome() (string, error) {
	usr, err := user.Current()
	if err != nil {
		return "", err
	}
	return usr.HomeDir, nil
}

// NewDefaultClient returns an OVH API Client from external configuration
func NewDefaultClient() (*Client, error) {
	return NewClient("", "", "", "")
}

// NewEndpointClient returns an OVH API Client from external configuration, for a specific endpoint
func NewEndpointClient(endpoint string) (*Client, error) {
	return NewClient(endpoint, "", "", "")
}

// NewClient returns an OVH API Client.
func NewClient(endpointName, applicationKey, applicationSecret, consumerKey string) (*Client, error) {
	// Load configuration files by order of increasing priority. All configuration
	// files are optional. Only load file from user home if home could be resolve
	cfg := ini.Empty()
	cfg.Append("/etc/ovh.conf")
	if home, err := currentUserHome(); err == nil {
		cfg.Append(home + "/.ovh.conf")
	}
	cfg.Append("./ovh.conf")

	// Canonicalize configuration
	if endpointName == "" {
		endpointName = getConfigValue(cfg, "default", "endpoint")
	}

	if applicationKey == "" {
		applicationKey = getConfigValue(cfg, endpointName, "application_key")
	}

	if applicationSecret == "" {
		applicationSecret = getConfigValue(cfg, endpointName, "application_secret")
	}

	if consumerKey == "" {
		consumerKey = getConfigValue(cfg, endpointName, "consumer_key")
	}

	// Load real endpoint URL by name. If endpoint contains a '/', consider it as a URL
	var endpoint Endpoint
	if strings.Contains(endpointName, "/") {
		endpoint = Endpoint(endpointName)
	} else {
		endpoint = Endpoints[endpointName]
	}

	// Timeout
	timeout := time.Duration(DefaultTimeout * time.Second)

	// Create client
	client := &Client{
		endpoint:          endpoint,
		applicationKey:    applicationKey,
		applicationSecret: applicationSecret,
		consumerKey:       consumerKey,
		Timeout:           timeout,
		client:            &http.Client{},
	}

	return client, nil
}

// getConfigValue returns the value of OVH_<NAME> or ``name`` value from ``section``
func getConfigValue(cfg *ini.File, section, name string) string {
	// Attempt to load from environment
	fromEnv := os.Getenv("OVH_" + strings.ToUpper(name))
	if len(fromEnv) > 0 {
		return fromEnv
	}

	// Attempt to load from configuration
	fromSection := cfg.Section(section)
	if fromSection == nil {
		return ""
	}

	fromSectionKey := fromSection.Key(name)
	if fromSectionKey == nil {
		return ""
	}
	return fromSectionKey.String()
}

//
// High level API
//

// DecodeError return error on unexpected HTTP code
func (r *APIResponse) DecodeError(expectedHTTPCode []int) (*APIError, error) {
	for _, code := range expectedHTTPCode {
		if r.StatusCode == code {
			return nil, nil
		}
	}

	// Decode OVH error informations from response
	if r.Body != nil {
		var ovhResponse *APIError
		err := json.Unmarshal(r.Body, ovhResponse)
		if err == nil {
			return ovhResponse, errors.New(ovhResponse.Message)
		}
	}
	return nil, fmt.Errorf("%d - %s", r.StatusCode, r.Status)
}

// Get Issues an authenticated get request on /path
func (c *Client) Get(path string) (*APIResponse, error) {
	return c.Call("GET", path, nil, true)
}

// GetUnAuth Issues an un-authenticated get request on /path
func (c *Client) GetUnAuth(path string) (*APIResponse, error) {
	return c.Call("GET", path, nil, false)
}

// Post Issues an authenticated get request on /path
func (c *Client) Post(path string, data interface{}) (*APIResponse, error) {
	return c.Call("POST", path, data, true)
}

// PostUnAuth Issues an un-authenticated get request on /path
func (c *Client) PostUnAuth(path string, data interface{}) (*APIResponse, error) {
	return c.Call("POST", path, data, false)
}

// Put Issues an authenticated get request on /path
func (c *Client) Put(path string, data interface{}) (*APIResponse, error) {
	return c.Call("PUT", path, data, true)
}

// PutUnAuth Issues an un-authenticated get request on /path
func (c *Client) PutUnAuth(path string, data interface{}) (*APIResponse, error) {
	return c.Call("PUT", path, data, false)
}

// Delete Issues an authenticated get request on /path
func (c *Client) Delete(path string) (*APIResponse, error) {
	return c.Call("DELETE", path, nil, true)
}

// DeleteUnAuth Issues an un-authenticated get request on /path
func (c *Client) DeleteUnAuth(path string) (*APIResponse, error) {
	return c.Call("DELETE", path, nil, false)
}

//
// Low Level Helpers
//

// Account for clock delay in API in signatures
func (c *Client) getTimeDelta() int64 {
	if c.timeDeltaDone != true {
		// Attempt to get timeDelta or fallback on 0
		timeDelta, err := c.GetUnAuth("/auth/time")
		if err != nil {
			return 0
		}

		// Attempt to load timeDelta or fallback on 0
		var serverTime int64
		err = json.Unmarshal(timeDelta.Body, &serverTime)
		if err != nil {
			return 0
		}
		c.timeDelta = time.Now().Unix() - serverTime
		c.timeDeltaDone = true
	}
	return c.timeDelta
}

// Call calls OVH's API and signs the request if ``needAuth`` is ``true``
func (c *Client) Call(method, path string, data interface{}, needAuth bool) (*APIResponse, error) {
	var body []byte
	var err error

	if data != nil {
		body, err = json.Marshal(data)
		if err != nil {
			return nil, err
		}
	}

	target := fmt.Sprintf("%s%s", c.endpoint, path)
	req, err := http.NewRequest(method, target, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}

	if body != nil {
		req.Header.Add("Content-Type", "application/json;charset=utf-8")
	}
	req.Header.Add("X-Ovh-Application", c.applicationKey)

	// Some methods do not need authentication, especially /time, /auth and some
	// /order methods are actually broken if authenticated.
	if needAuth {
		timestamp := time.Now().Unix() - c.getTimeDelta()

		req.Header.Add("X-Ovh-Timestamp", fmt.Sprintf("%d", timestamp))
		req.Header.Add("X-Ovh-Consumer", c.consumerKey)
		req.Header.Add("Accept", "application/json")

		h := sha1.New()
		h.Write([]byte(fmt.Sprintf("%s+%s+%s+%s+%s+%d",
			c.applicationSecret,
			c.consumerKey,
			method,
			target,
			body,
			timestamp,
		)))
		req.Header.Add("X-Ovh-Signature", fmt.Sprintf("$1$%x", h.Sum(nil)))
	}

	c.client.Timeout = c.Timeout
	r, err := c.client.Do(req)

	if err != nil {
		return nil, err
	}
	defer r.Body.Close()

	response, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}

	return &APIResponse{
		StatusCode: r.StatusCode,
		Status:     r.Status,
		Body:       response,
	}, nil
}
