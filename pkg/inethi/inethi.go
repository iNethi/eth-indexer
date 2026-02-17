package inethi

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
	InethiClient struct {
		apiKey     string
		endpoint   string
		httpClient *http.Client
	}

	VoucherPayload struct {
		SenderAddress    string
		RecipientAddress string
		Amount           string
		TokenSymbol      string
		CouponSize       int
	}

	VoucherResponse struct {
		Voucher string `json:"voucher"`
	}
)

func New(apiKey string, endpoint string) *InethiClient {
	iClient := &InethiClient{
		apiKey:   apiKey,
		endpoint: endpoint,
		httpClient: &http.Client{
			Timeout: time.Second * 10,
		},
	}

	return iClient
}

func (i *InethiClient) setDefaultHeaders(req *http.Request) *http.Request {
	req.Header.Set("Authorization", "ApiKey "+i.apiKey)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")

	return req
}

func (i *InethiClient) postRequestWithCtx(ctx context.Context, url string, body io.Reader) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, body)
	if err != nil {
		return nil, err
	}

	return i.do(req)
}

func (i *InethiClient) do(req *http.Request) (*http.Response, error) {
	return i.httpClient.Do(i.setDefaultHeaders(req))
}

func parseResponse(resp *http.Response, target interface{}) error {
	defer resp.Body.Close()

	if resp.StatusCode >= http.StatusBadRequest {
		b, err := io.ReadAll(resp.Body)
		if err != nil {
			return err
		}

		return fmt.Errorf("iNethi server error: code=%s: response_body=%s", resp.Status, string(b))
	}

	return json.NewDecoder(resp.Body).Decode(target)
}

func (i *InethiClient) GenerateVoucher(ctx context.Context, input VoucherPayload) (VoucherResponse, error) {
	var (
		buf             bytes.Buffer
		voucherResponse VoucherResponse
	)

	if err := json.NewEncoder(&buf).Encode(struct {
		RadiusDeskInstancePK int    `json:"radius_desk_instance_pk"`
		RadiusDeskProfilePK  int    `json:"radius_desk_profile_pk"`
		RadiusDeskCloudPK    int    `json:"radius_desk_cloud_pk"`
		RadiusDeskRealmPK    int    `json:"radius_desk_realm_pk"`
		SenderAddress        string `json:"sender_address"`
		RecipientAddress     string `json:"recipient_address"`
		Amount               string `json:"amount"`
		Category             string `json:"category"`
		Token                string `json:"token"`
	}{
		RadiusDeskInstancePK: 3,
		RadiusDeskProfilePK:  input.CouponSize,
		RadiusDeskCloudPK:    3,
		RadiusDeskRealmPK:    3,
		SenderAddress:        input.SenderAddress,
		RecipientAddress:     input.RecipientAddress,
		Amount:               input.Amount,
		Token:                input.TokenSymbol,
		Category:             "INTERNET_COUPON",
	}); err != nil {
		return voucherResponse, err
	}

	resp, err := i.postRequestWithCtx(ctx, i.endpoint+"/api/v1/vouchers/add_voucher/", &buf)
	if err != nil {
		return voucherResponse, err
	}

	if err := parseResponse(resp, &voucherResponse); err != nil {
		return voucherResponse, err
	}

	return voucherResponse, nil
}
