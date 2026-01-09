package providers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	httpclient "loan-disbursement-service/http"
	"loan-disbursement-service/models"

	"github.com/rs/zerolog/log"
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
		fmt.Sprintf("%s/api/v1/payment", g.baseURL),
		req,
		map[string]string{
			"Content-Type": "application/json",
		},
	)
	if err != nil {
		log.Error().Err(err).Msg("gateway error")
		return models.PaymentResponse{}, models.NETWORK_ERROR
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errBody map[string]any
		if err := json.NewDecoder(resp.Body).Decode(&errBody); err != nil {
			log.Error().Err(err).Msg("failed to decode error response")
			return models.PaymentResponse{}, fmt.Errorf("gateway error: status=%d", resp.StatusCode)
		}
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
	channel models.PaymentChannel,
	transactionID string,
) (models.PaymentResponse, error) {
	if transactionID == "" {
		return models.PaymentResponse{}, models.TRANSACTION_ID_REQUIRED
	}

	resp, err := g.client.GET(
		ctx,
		fmt.Sprintf("%s/api/v1/payment/%s/txn/%s", g.baseURL, channel, transactionID),
		nil,
	)
	if err != nil {
		return models.PaymentResponse{}, models.NETWORK_ERROR
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return models.PaymentResponse{}, models.TRANSACTION_NOT_FOUND
	}
	if resp.StatusCode != http.StatusOK {
		var errBody map[string]any
		if err := json.NewDecoder(resp.Body).Decode(&errBody); err != nil {
			log.Error().Err(err).Msg("failed to decode error response")
			return models.PaymentResponse{}, fmt.Errorf("gateway error: status=%d", resp.StatusCode)
		}
		if errorMessage, ok := errBody["error"].(string); ok {
			return models.PaymentResponse{}, errors.New(errorMessage)
		}
		log.Error().Int("status_code", resp.StatusCode).
			RawJSON("body", []byte(fmt.Sprintf("%v", errBody))).
			Msg("gateway error")
		return models.PaymentResponse{}, models.UNKNOWN_ERROR
	}

	var result models.PaymentResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return models.PaymentResponse{}, err
	}

	return result, nil
}

func (g GatewayProvider) IsActive(
	ctx context.Context,
	channel models.PaymentChannel,
) (bool, error) {
	resp, err := g.client.GET(
		ctx,
		fmt.Sprintf("%s/api/v1/channel/%s/status", g.baseURL, channel),
		nil,
	)
	if err != nil {
		return false, models.NETWORK_ERROR
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return false, models.UNKNOWN_ERROR
	}

	var result models.PaymentChannelResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return false, err
	}

	return result.Active, nil
}
