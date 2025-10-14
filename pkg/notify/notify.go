package notify

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type (
	NotifyClient struct {
		bearerToken string
		endpoint    string
		httpClient  *http.Client
	}

	NotifyPayload struct {
		SenderAddress string `json:"senderAddress"`
		Code          string `json:"code"`
		Size          string `json:"size"`
	}

	NotifyResponse struct {
		Success bool   `json:"success"`
		Message string `json:"message"`
	}
)

func New(bearerToken string, endpoint string) *NotifyClient {
	nClient := &NotifyClient{
		bearerToken: bearerToken,
		endpoint:    endpoint,
		httpClient: &http.Client{
			Timeout: time.Second * 10,
		},
	}

	return nClient
}

func (n *NotifyClient) setDefaultHeaders(req *http.Request) *http.Request {
	req.Header.Set("Authorization", "Bearer "+n.bearerToken)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")

	return req
}

func (n *NotifyClient) postRequestWithCtx(ctx context.Context, url string, body io.Reader) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, body)
	if err != nil {
		return nil, err
	}

	return n.do(req)
}

func (n *NotifyClient) do(req *http.Request) (*http.Response, error) {
	return n.httpClient.Do(n.setDefaultHeaders(req))
}

func parseResponse(resp *http.Response, target interface{}) error {
	defer resp.Body.Close()

	if resp.StatusCode >= http.StatusBadRequest {
		b, err := io.ReadAll(resp.Body)
		if err != nil {
			return err
		}

		return fmt.Errorf("notify server error: code=%s: response_body=%s", resp.Status, string(b))
	}

	return json.NewDecoder(resp.Body).Decode(target)
}

func (n *NotifyClient) SendNotification(ctx context.Context, input NotifyPayload) (NotifyResponse, error) {
	var (
		buf            bytes.Buffer
		notifyResponse NotifyResponse
	)

	if err := json.NewEncoder(&buf).Encode(input); err != nil {
		return notifyResponse, err
	}

	resp, err := n.postRequestWithCtx(ctx, n.endpoint+"/api/v1/external/inethi", &buf)
	if err != nil {
		return notifyResponse, err
	}

	if err := parseResponse(resp, &notifyResponse); err != nil {
		return notifyResponse, err
	}

	return notifyResponse, nil
}
