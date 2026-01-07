package providers

import (
	"context"
	httpclient "loan-disbursement-service/http"
	"loan-disbursement-service/models"
)

type PaymentProvider interface {
	Transfer(ctx context.Context, req models.PaymentRequest) (models.PaymentResponse, error)
	Fetch(ctx context.Context, transactionId string) (models.PaymentResponse, error)
}

func NewPaymentProvider(baseURL string, client httpclient.HTTPClient) (PaymentProvider, error) {
	return NewGatewayProvider(baseURL, client), nil
}
