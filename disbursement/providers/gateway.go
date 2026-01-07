package providers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	httpclient "loan-disbursement-service/http"
	"loan-disbursement-service/models"
)

type GatewayProvider struct {
	baseURL string
	client  httpclient.HTTPClient
}

func NewGatewayProvider(baseURL string, client httpclient.HTTPClient) *GatewayProvider {
	return &GatewayProvider{
		baseURL: baseURL,
		client:  client,
	}
}

func (g GatewayProvider) Transfer(
	ctx context.Context,
	req models.PaymentRequest,
) (models.PaymentResponse, error) {
	resp, err := g.client.POST(
		ctx,
		fmt.Sprintf("%s/api/v1/payment/init", g.baseURL),
		req,
		map[string]string{
			"Content-Type": "application/json",
		},
	)
	if err != nil {
		return models.PaymentResponse{}, errors.New("gateway error: " + err.Error())
	}

	if resp.StatusCode != http.StatusOK {
		var errBody map[string]any
		_ = json.NewDecoder(resp.Body).Decode(&errBody)
		if errorMessage, ok := errBody["error"].(string); ok {
			return models.PaymentResponse{}, errors.New(errorMessage)
		}
		return models.PaymentResponse{}, fmt.Errorf(
			"gateway error: status=%d body=%v",
			resp.StatusCode,
			errBody,
		)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errBody map[string]any
		_ = json.NewDecoder(resp.Body).Decode(&errBody)
		if errorMessage, ok := errBody["error"].(string); ok {
			return models.PaymentResponse{}, errors.New(errorMessage)
		}
		return models.PaymentResponse{}, fmt.Errorf(
			"gateway error: status=%d body=%v",
			resp.StatusCode,
			errBody,
		)
	}

	var result models.PaymentResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return models.PaymentResponse{}, err
	}

	return result, nil
}

func (g GatewayProvider) Fetch(
	ctx context.Context,
	transactionId string,
) (models.PaymentResponse, error) {
	if transactionId == "" {
		return models.PaymentResponse{}, errors.New("transactionId is required")
	}

	resp, err := g.client.GET(
		context.Background(),
		fmt.Sprintf("%s/api/v1/payment/status/%s", g.baseURL, transactionId),
		nil,
	)
	if err != nil {
		return models.PaymentResponse{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return models.PaymentResponse{}, fmt.Errorf("transaction not found")
	}
	if resp.StatusCode != http.StatusOK {
		var errBody map[string]any
		_ = json.NewDecoder(resp.Body).Decode(&errBody)
		if errorMessage, ok := errBody["error"].(string); ok {
			return models.PaymentResponse{}, errors.New(errorMessage)
		}
		return models.PaymentResponse{}, fmt.Errorf(
			"gateway error: status=%d body=%v",
			resp.StatusCode,
			errBody,
		)
	}

	var result models.PaymentResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return models.PaymentResponse{}, err
	}

	return result, nil
}
