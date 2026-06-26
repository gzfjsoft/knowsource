package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type Client struct {
	BaseURL    string
	Token      string
	HTTPClient *http.Client
}

func NewClient(baseURL string) *Client {
	return &Client{
		BaseURL: baseURL,
		HTTPClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (c *Client) SetToken(token string) {
	c.Token = token
}

func (c *Client) request(method, path string, reqBody interface{}, respDest interface{}) error {
	var bodyReader io.Reader
	if reqBody != nil {
		reqBytes, err := json.Marshal(reqBody)
		if err != nil {
			return fmt.Errorf("failed to marshal request: %w", err)
		}
		bodyReader = bytes.NewReader(reqBytes)
	}

	url := c.BaseURL + path
	req, err := http.NewRequest(method, url, bodyReader)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	if c.Token != "" {
		req.Header.Set("Authorization", "Bearer "+c.Token)
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("http request failed: %w", err)
	}
	defer resp.Body.Close()

	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("http status error: %d, body: %s", resp.StatusCode, string(respBytes))
	}

	if err := json.Unmarshal(respBytes, respDest); err != nil {
		return fmt.Errorf("failed to unmarshal response: %w, body: %s", err, string(respBytes))
	}

	return nil
}

type LoginRequest struct {
	ClientID  string `json:"clientId"`
	EmpCode   string `json:"empCode"`
	Password  string `json:"password"`
	CaptchaId string `json:"captchaId"`
	Captcha   string `json:"captcha"`
	IsDebug   int64  `json:"isDebug"`
}

type LoginResponse struct {
	Code    int64      `json:"code"`
	Message string     `json:"message"`
	Data    *LoginData `json:"data,omitempty"`
}

type LoginData struct {
	Token    string  `json:"token"`
	UserInfo EmpInfo `json:"userInfo"`
}

type EmpInfo struct {
	EmpCode    string   `json:"empCode"`
	EmpName    string   `json:"empName"`
	ClientID   string   `json:"clientId,optional"`
	ClientName string   `json:"clientName,optional"`
	DeptCode   string   `json:"deptCode"`
	DeptName   string   `json:"deptName"`
	Status     int64    `json:"status"`
	Position   string   `json:"position"`
	Mobile     string   `json:"mobile,optional"`
	Email      string   `json:"email,optional"`
	Roles      []string `json:"roles,optional"`
}

func (c *Client) Login(req LoginRequest) (*LoginResponse, error) {
	var resp LoginResponse
	err := c.request("POST", "/api/knowsource/login", req, &resp)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

type DocumentsType struct {
	Code        string   `json:"code,optional"`
	Name        string   `json:"name"`
	IsDisabled  int64    `json:"isDisabled,optional"`
	Description string   `json:"description"`
	Tags        []string `json:"tags,optional"`
	CreatedAt   int64    `json:"createdAt,optional"`
	UpdatedAt   int64    `json:"updatedAt,optional"`
}

type DocumentsTypeData struct {
	List  []DocumentsType `json:"list"`
	Total int64           `json:"total"`
}

type ListDocumentsTypeResponse struct {
	Code    int64              `json:"code"`
	Message string             `json:"message"`
	Data    *DocumentsTypeData `json:"data"`
}

func (c *Client) ListDocumentsType() (*ListDocumentsTypeResponse, error) {
	var resp ListDocumentsTypeResponse
	err := c.request("POST", "/api/documents/type/list", nil, &resp)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}
