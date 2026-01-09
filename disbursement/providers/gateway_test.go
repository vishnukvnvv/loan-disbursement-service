package providers

import (
	"context"
	"encoding/json"
	"errors"
	"loan-disbursement-service/models"
	http_test "loan-disbursement-service/test/http"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGatewayProvider_Transfer(t *testing.T) {
	ctx := context.Background()
	baseURL := "http://localhost:8080"

	t.Run("successfully transfers payment", func(t *testing.T) {
		mockClient := new(http_test.MockHTTPClient)
		provider := NewGatewayProvider(baseURL, mockClient)

		request := models.PaymentRequest{
			ReferenceID: "REF-123",
			Amount:      5000.0,
			Channel:     models.PaymentChannelUPI,
			Beneficiary: models.Beneficiary{
				Name:    "John Doe",
				Account: "1234567890",
				IFSC:    "IFSC0001234",
				Bank:    "Test Bank",
			},
			Metadata: models.PaymentMetadata{
				NotificationURL: "https://example.com/webhook",
				LoanID:          "LOAN-001",
				DisbursementID:  "DISB-001",
			},
		}

		expectedResponse := models.PaymentResponse{
			TransactionID: "TXN-123456789012",
			ReferenceID:   request.ReferenceID,
			Amount:        request.Amount,
			Status:        "initiated",
			Channel:       request.Channel,
			Beneficiary:   request.Beneficiary,
			Metadata:      request.Metadata,
			AcceptedAT:    time.Now(),
			ProcessedAT:   time.Time{},
		}

		responseBody, _ := json.Marshal(expectedResponse)
		response := http_test.NewJSONResponse(http.StatusOK, string(responseBody))

		mockClient.On("POST", ctx, "http://localhost:8080/api/v1/payment", request, mock.MatchedBy(func(headers map[string]string) bool {
			return headers["Content-Type"] == "application/json"
		})).
			Return(response, nil).
			Once()

		result, err := provider.Transfer(ctx, request)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, expectedResponse.TransactionID, result.TransactionID)
		assert.Equal(t, expectedResponse.ReferenceID, result.ReferenceID)
		assert.Equal(t, expectedResponse.Amount, result.Amount)
		assert.Equal(t, expectedResponse.Status, result.Status)

		mockClient.AssertExpectations(t)
		response.Body.Close()
	})

	t.Run("returns network error when POST request fails", func(t *testing.T) {
		mockClient := new(http_test.MockHTTPClient)
		provider := NewGatewayProvider(baseURL, mockClient)

		request := models.PaymentRequest{
			ReferenceID: "REF-123",
			Amount:      5000.0,
			Channel:     models.PaymentChannelUPI,
			Beneficiary: models.Beneficiary{
				Name:    "John Doe",
				Account: "1234567890",
				IFSC:    "IFSC0001234",
				Bank:    "Test Bank",
			},
		}

		networkError := errors.New("connection refused")

		mockClient.On("POST", ctx, "http://localhost:8080/api/v1/payment", request, mock.Anything).
			Return(nil, networkError).Once()

		result, err := provider.Transfer(ctx, request)

		assert.Error(t, err)
		assert.Equal(t, models.NETWORK_ERROR, err)
		assert.Equal(t, models.PaymentResponse{}, result)

		mockClient.AssertExpectations(t)
	})

	t.Run("returns error when status code is not OK", func(t *testing.T) {
		mockClient := new(http_test.MockHTTPClient)
		provider := NewGatewayProvider(baseURL, mockClient)

		request := models.PaymentRequest{
			ReferenceID: "REF-123",
			Amount:      5000.0,
			Channel:     models.PaymentChannelUPI,
			Beneficiary: models.Beneficiary{
				Name:    "John Doe",
				Account: "1234567890",
				IFSC:    "IFSC0001234",
				Bank:    "Test Bank",
			},
		}

		errorBody := `{"error": "Invalid beneficiary details"}`
		response := http_test.NewJSONResponse(http.StatusBadRequest, errorBody)

		mockClient.On("POST", ctx, "http://localhost:8080/api/v1/payment", request, mock.Anything).
			Return(response, nil).Once()

		result, err := provider.Transfer(ctx, request)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "Invalid beneficiary details")
		assert.Equal(t, models.PaymentResponse{}, result)

		mockClient.AssertExpectations(t)
		response.Body.Close()
	})

	t.Run(
		"returns formatted error when status code is not OK and error message not in body",
		func(t *testing.T) {
			mockClient := new(http_test.MockHTTPClient)
			provider := NewGatewayProvider(baseURL, mockClient)

			request := models.PaymentRequest{
				ReferenceID: "REF-123",
				Amount:      5000.0,
				Channel:     models.PaymentChannelUPI,
				Beneficiary: models.Beneficiary{
					Name:    "John Doe",
					Account: "1234567890",
					IFSC:    "IFSC0001234",
					Bank:    "Test Bank",
				},
			}

			errorBody := `{"message": "Something went wrong"}`
			response := http_test.NewJSONResponse(http.StatusInternalServerError, errorBody)

			mockClient.On("POST", ctx, "http://localhost:8080/api/v1/payment", request, mock.Anything).
				Return(response, nil).
				Once()

			result, err := provider.Transfer(ctx, request)

			assert.Error(t, err)
			assert.Contains(t, err.Error(), "gateway error: status=500")
			assert.Equal(t, models.PaymentResponse{}, result)

			mockClient.AssertExpectations(t)
			response.Body.Close()
		},
	)

	t.Run("returns error when response body cannot be decoded", func(t *testing.T) {
		mockClient := new(http_test.MockHTTPClient)
		provider := NewGatewayProvider(baseURL, mockClient)

		request := models.PaymentRequest{
			ReferenceID: "REF-123",
			Amount:      5000.0,
			Channel:     models.PaymentChannelUPI,
			Beneficiary: models.Beneficiary{
				Name:    "John Doe",
				Account: "1234567890",
				IFSC:    "IFSC0001234",
				Bank:    "Test Bank",
			},
		}

		invalidJSON := `{"invalid": json}`
		response := http_test.NewJSONResponse(http.StatusOK, invalidJSON)

		mockClient.On("POST", ctx, "http://localhost:8080/api/v1/payment", request, mock.Anything).
			Return(response, nil).Once()

		result, err := provider.Transfer(ctx, request)

		assert.Error(t, err)
		assert.Equal(t, models.PaymentResponse{}, result)

		mockClient.AssertExpectations(t)
		response.Body.Close()
	})

	t.Run("handles different payment modes", func(t *testing.T) {
		testCases := []struct {
			name    string
			channel models.PaymentChannel
		}{
			{"UPI", models.PaymentChannelUPI},
			{"NEFT", models.PaymentChannelNEFT},
			{"IMPS", models.PaymentChannelIMPS},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				mockClient := new(http_test.MockHTTPClient)
				provider := NewGatewayProvider(baseURL, mockClient)

				request := models.PaymentRequest{
					ReferenceID: "REF-123",
					Amount:      5000.0,
					Channel:     tc.channel,
					Beneficiary: models.Beneficiary{
						Name:    "John Doe",
						Account: "1234567890",
						IFSC:    "IFSC0001234",
						Bank:    "Test Bank",
					},
				}

				expectedResponse := models.PaymentResponse{
					TransactionID: "TXN-123456789012",
					ReferenceID:   request.ReferenceID,
					Amount:        request.Amount,
					Status:        "initiated",
					Channel:       tc.channel,
				}

				responseBody, _ := json.Marshal(expectedResponse)
				response := http_test.NewJSONResponse(http.StatusOK, string(responseBody))

				mockClient.On("POST", ctx, "http://localhost:8080/api/v1/payment", request, mock.Anything).
					Return(response, nil).
					Once()

				result, err := provider.Transfer(ctx, request)

				assert.NoError(t, err)
				assert.Equal(t, tc.channel, result.Channel)

				mockClient.AssertExpectations(t)
				response.Body.Close()
			})
		}
	})
}

func TestGatewayProvider_Fetch(t *testing.T) {
	ctx := context.Background()
	baseURL := "http://localhost:8080"

	t.Run("successfully fetches transaction", func(t *testing.T) {
		mockClient := new(http_test.MockHTTPClient)
		provider := NewGatewayProvider(baseURL, mockClient)

		channel := models.PaymentChannelUPI
		transactionID := "TXN-123456789012"

		expectedResponse := models.PaymentResponse{
			TransactionID: transactionID,
			ReferenceID:   "REF-123",
			Amount:        5000.0,
			Status:        "success",
			Channel:       channel,
			Beneficiary: models.Beneficiary{
				Name:    "John Doe",
				Account: "1234567890",
				IFSC:    "IFSC0001234",
				Bank:    "Test Bank",
			},
			AcceptedAT:  time.Now().Add(-1 * time.Hour),
			ProcessedAT: time.Now(),
		}

		responseBody, _ := json.Marshal(expectedResponse)
		response := http_test.NewJSONResponse(http.StatusOK, string(responseBody))

		expectedURL := "http://localhost:8080/api/v1/payment/UPI/txn/TXN-123456789012"
		mockClient.On("GET", ctx, expectedURL, mock.Anything).
			Return(response, nil).Once()

		result, err := provider.Fetch(ctx, channel, transactionID)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, transactionID, result.TransactionID)
		assert.Equal(t, expectedResponse.Status, result.Status)
		assert.Equal(t, expectedResponse.Amount, result.Amount)

		mockClient.AssertExpectations(t)
		response.Body.Close()
	})

	t.Run("returns error when transaction ID is empty", func(t *testing.T) {
		mockClient := new(http_test.MockHTTPClient)
		provider := NewGatewayProvider(baseURL, mockClient)

		channel := models.PaymentChannelUPI
		transactionID := ""

		result, err := provider.Fetch(ctx, channel, transactionID)

		assert.Error(t, err)
		assert.Equal(t, models.TRANSACTION_ID_REQUIRED, err)
		assert.Equal(t, models.PaymentResponse{}, result)

		mockClient.AssertNotCalled(t, "GET")
	})

	t.Run("returns network error when GET request fails", func(t *testing.T) {
		mockClient := new(http_test.MockHTTPClient)
		provider := NewGatewayProvider(baseURL, mockClient)

		channel := models.PaymentChannelUPI
		transactionID := "TXN-123456789012"

		networkError := errors.New("connection refused")

		expectedURL := "http://localhost:8080/api/v1/payment/UPI/txn/TXN-123456789012"
		mockClient.On("GET", ctx, expectedURL, mock.Anything).
			Return(nil, networkError).Once()

		result, err := provider.Fetch(ctx, channel, transactionID)

		assert.Error(t, err)
		assert.Equal(t, models.NETWORK_ERROR, err)
		assert.Equal(t, models.PaymentResponse{}, result)

		mockClient.AssertExpectations(t)
	})

	t.Run("returns transaction not found when status is 404", func(t *testing.T) {
		mockClient := new(http_test.MockHTTPClient)
		provider := NewGatewayProvider(baseURL, mockClient)

		channel := models.PaymentChannelUPI
		transactionID := "TXN-NONEXISTENT"

		response := http_test.NewJSONResponse(
			http.StatusNotFound,
			`{"error": "Transaction not found"}`,
		)

		expectedURL := "http://localhost:8080/api/v1/payment/UPI/txn/TXN-NONEXISTENT"
		mockClient.On("GET", ctx, expectedURL, mock.Anything).
			Return(response, nil).Once()

		result, err := provider.Fetch(ctx, channel, transactionID)

		assert.Error(t, err)
		assert.Equal(t, models.TRANSACTION_NOT_FOUND, err)
		assert.Equal(t, models.PaymentResponse{}, result)

		mockClient.AssertExpectations(t)
		response.Body.Close()
	})

	t.Run("returns error when status code is not OK and not 404", func(t *testing.T) {
		mockClient := new(http_test.MockHTTPClient)
		provider := NewGatewayProvider(baseURL, mockClient)

		channel := models.PaymentChannelUPI
		transactionID := "TXN-123456789012"

		errorBody := `{"error": "Internal server error"}`
		response := http_test.NewJSONResponse(http.StatusInternalServerError, errorBody)

		expectedURL := "http://localhost:8080/api/v1/payment/UPI/txn/TXN-123456789012"
		mockClient.On("GET", ctx, expectedURL, mock.Anything).
			Return(response, nil).Once()

		result, err := provider.Fetch(ctx, channel, transactionID)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "Internal server error")
		assert.Equal(t, models.PaymentResponse{}, result)

		mockClient.AssertExpectations(t)
		response.Body.Close()
	})

	t.Run(
		"returns unknown error when status code is not OK and error message not in body",
		func(t *testing.T) {
			mockClient := new(http_test.MockHTTPClient)
			provider := NewGatewayProvider(baseURL, mockClient)

			channel := models.PaymentChannelUPI
			transactionID := "TXN-123456789012"

			errorBody := `{"message": "Something went wrong"}`
			response := http_test.NewJSONResponse(http.StatusInternalServerError, errorBody)

			expectedURL := "http://localhost:8080/api/v1/payment/UPI/txn/TXN-123456789012"
			mockClient.On("GET", ctx, expectedURL, mock.Anything).
				Return(response, nil).Once()

			result, err := provider.Fetch(ctx, channel, transactionID)

			assert.Error(t, err)
			assert.Equal(t, models.UNKNOWN_ERROR, err)
			assert.Equal(t, models.PaymentResponse{}, result)

			mockClient.AssertExpectations(t)
			response.Body.Close()
		},
	)

	t.Run("returns error when response body cannot be decoded", func(t *testing.T) {
		mockClient := new(http_test.MockHTTPClient)
		provider := NewGatewayProvider(baseURL, mockClient)

		channel := models.PaymentChannelUPI
		transactionID := "TXN-123456789012"

		invalidJSON := `{"invalid": json}`
		response := http_test.NewJSONResponse(http.StatusOK, invalidJSON)

		expectedURL := "http://localhost:8080/api/v1/payment/UPI/txn/TXN-123456789012"
		mockClient.On("GET", ctx, expectedURL, mock.Anything).
			Return(response, nil).Once()

		result, err := provider.Fetch(ctx, channel, transactionID)

		assert.Error(t, err)
		assert.Equal(t, models.PaymentResponse{}, result)

		mockClient.AssertExpectations(t)
		response.Body.Close()
	})

	t.Run("handles different payment channels", func(t *testing.T) {
		testCases := []struct {
			name    string
			channel models.PaymentChannel
		}{
			{"UPI", models.PaymentChannelUPI},
			{"NEFT", models.PaymentChannelNEFT},
			{"IMPS", models.PaymentChannelIMPS},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				mockClient := new(http_test.MockHTTPClient)
				provider := NewGatewayProvider(baseURL, mockClient)

				transactionID := "TXN-123456789012"

				expectedResponse := models.PaymentResponse{
					TransactionID: transactionID,
					ReferenceID:   "REF-123",
					Amount:        5000.0,
					Status:        "success",
					Channel:       tc.channel,
				}

				responseBody, _ := json.Marshal(expectedResponse)
				response := http_test.NewJSONResponse(http.StatusOK, string(responseBody))

				expectedURL := strings.ToLower(
					"http://localhost:8080/api/v1/payment/" + string(
						tc.channel,
					) + "/txn/" + transactionID,
				)
				// Note: The actual URL construction uses the channel as-is, so we need to match it properly
				expectedURL = "http://localhost:8080/api/v1/payment/" + string(
					tc.channel,
				) + "/txn/" + transactionID
				mockClient.On("GET", ctx, expectedURL, mock.Anything).
					Return(response, nil).Once()

				result, err := provider.Fetch(ctx, tc.channel, transactionID)

				assert.NoError(t, err)
				assert.Equal(t, tc.channel, result.Channel)

				mockClient.AssertExpectations(t)
				response.Body.Close()
			})
		}
	})

	t.Run(
		"uses context.Background() instead of provided context for GET request",
		func(t *testing.T) {
			mockClient := new(http_test.MockHTTPClient)
			provider := NewGatewayProvider(baseURL, mockClient)

			channel := models.PaymentChannelUPI
			transactionID := "TXN-123456789012"

			expectedResponse := models.PaymentResponse{
				TransactionID: transactionID,
				Status:        "success",
			}

			responseBody, _ := json.Marshal(expectedResponse)
			response := http_test.NewJSONResponse(http.StatusOK, string(responseBody))

			expectedURL := "http://localhost:8080/api/v1/payment/UPI/txn/TXN-123456789012"
			mockClient.On("GET", ctx, expectedURL, mock.Anything).
				Return(response, nil).Once()

			result, err := provider.Fetch(ctx, channel, transactionID)

			assert.NoError(t, err)
			assert.NotNil(t, result)

			mockClient.AssertExpectations(t)
			response.Body.Close()
		},
	)
}

func TestGatewayProvider_IsActive(t *testing.T) {
	ctx := context.Background()
	baseURL := "http://localhost:8080"

	t.Run("successfully checks availability when channel is active", func(t *testing.T) {
		mockClient := new(http_test.MockHTTPClient)
		provider := NewGatewayProvider(baseURL, mockClient)

		channel := models.PaymentChannelUPI

		responseBody := `{"active": true}`
		response := http_test.NewJSONResponse(http.StatusOK, responseBody)

		expectedURL := "http://localhost:8080/api/v1/channel/UPI/status"
		mockClient.On("GET", ctx, expectedURL, mock.Anything).
			Return(response, nil).
			Once()

		result, err := provider.IsActive(ctx, channel)

		assert.NoError(t, err)
		assert.True(t, result)

		mockClient.AssertExpectations(t)
		response.Body.Close()
	})

	t.Run("successfully checks availability when channel is inactive", func(t *testing.T) {
		mockClient := new(http_test.MockHTTPClient)
		provider := NewGatewayProvider(baseURL, mockClient)

		channel := models.PaymentChannelNEFT

		responseBody := `{"active": false}`
		response := http_test.NewJSONResponse(http.StatusOK, responseBody)

		expectedURL := "http://localhost:8080/api/v1/channel/NEFT/status"
		mockClient.On("GET", ctx, expectedURL, mock.Anything).
			Return(response, nil).
			Once()

		result, err := provider.IsActive(ctx, channel)

		assert.NoError(t, err)
		assert.False(t, result)

		mockClient.AssertExpectations(t)
		response.Body.Close()
	})

	t.Run("returns network error when GET request fails", func(t *testing.T) {
		mockClient := new(http_test.MockHTTPClient)
		provider := NewGatewayProvider(baseURL, mockClient)

		channel := models.PaymentChannelUPI

		networkError := errors.New("connection refused")

		expectedURL := "http://localhost:8080/api/v1/channel/UPI/status"
		mockClient.On("GET", ctx, expectedURL, mock.Anything).
			Return(nil, networkError).
			Once()

		result, err := provider.IsActive(ctx, channel)

		assert.Error(t, err)
		assert.Equal(t, models.NETWORK_ERROR, err)
		assert.False(t, result)

		mockClient.AssertExpectations(t)
	})

	t.Run("returns unknown error when status code is not OK", func(t *testing.T) {
		mockClient := new(http_test.MockHTTPClient)
		provider := NewGatewayProvider(baseURL, mockClient)

		channel := models.PaymentChannelUPI

		response := http_test.NewJSONResponse(
			http.StatusInternalServerError,
			`{"error": "Internal server error"}`,
		)

		expectedURL := "http://localhost:8080/api/v1/channel/UPI/status"
		mockClient.On("GET", ctx, expectedURL, mock.Anything).
			Return(response, nil).
			Once()

		result, err := provider.IsActive(ctx, channel)

		assert.Error(t, err)
		assert.Equal(t, models.UNKNOWN_ERROR, err)
		assert.False(t, result)

		mockClient.AssertExpectations(t)
		response.Body.Close()
	})

	t.Run("returns error when response body cannot be decoded", func(t *testing.T) {
		mockClient := new(http_test.MockHTTPClient)
		provider := NewGatewayProvider(baseURL, mockClient)

		channel := models.PaymentChannelUPI

		invalidJSON := `{"invalid": json}`
		response := http_test.NewJSONResponse(http.StatusOK, invalidJSON)

		expectedURL := "http://localhost:8080/api/v1/channel/UPI/status"
		mockClient.On("GET", ctx, expectedURL, mock.Anything).
			Return(response, nil).
			Once()

		result, err := provider.IsActive(ctx, channel)

		assert.Error(t, err)
		assert.False(t, result)

		mockClient.AssertExpectations(t)
		response.Body.Close()
	})

	t.Run("handles different payment channels", func(t *testing.T) {
		testCases := []struct {
			name    string
			channel models.PaymentChannel
		}{
			{"UPI", models.PaymentChannelUPI},
			{"NEFT", models.PaymentChannelNEFT},
			{"IMPS", models.PaymentChannelIMPS},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				mockClient := new(http_test.MockHTTPClient)
				provider := NewGatewayProvider(baseURL, mockClient)

				responseBody := `{"active": true}`
				response := http_test.NewJSONResponse(http.StatusOK, responseBody)

				expectedURL := "http://localhost:8080/api/v1/channel/" + string(
					tc.channel,
				) + "/status"
				mockClient.On("GET", ctx, expectedURL, mock.Anything).
					Return(response, nil).
					Once()

				result, err := provider.IsActive(ctx, tc.channel)

				assert.NoError(t, err)
				assert.True(t, result)

				mockClient.AssertExpectations(t)
				response.Body.Close()
			})
		}
	})

	t.Run("returns false and unknown error for 404 status", func(t *testing.T) {
		mockClient := new(http_test.MockHTTPClient)
		provider := NewGatewayProvider(baseURL, mockClient)

		channel := models.PaymentChannelUPI

		response := http_test.NewJSONResponse(http.StatusNotFound, `{"error": "Channel not found"}`)

		expectedURL := "http://localhost:8080/api/v1/channel/UPI/status"
		mockClient.On("GET", ctx, expectedURL, mock.Anything).
			Return(response, nil).
			Once()

		result, err := provider.IsActive(ctx, channel)

		assert.Error(t, err)
		assert.Equal(t, models.UNKNOWN_ERROR, err)
		assert.False(t, result)

		mockClient.AssertExpectations(t)
		response.Body.Close()
	})

	t.Run("returns false and unknown error for 400 status", func(t *testing.T) {
		mockClient := new(http_test.MockHTTPClient)
		provider := NewGatewayProvider(baseURL, mockClient)

		channel := models.PaymentChannelIMPS

		response := http_test.NewJSONResponse(http.StatusBadRequest, `{"error": "Invalid channel"}`)

		expectedURL := "http://localhost:8080/api/v1/channel/IMPS/status"
		mockClient.On("GET", ctx, expectedURL, mock.Anything).
			Return(response, nil).
			Once()

		result, err := provider.IsActive(ctx, channel)

		assert.Error(t, err)
		assert.Equal(t, models.UNKNOWN_ERROR, err)
		assert.False(t, result)

		mockClient.AssertExpectations(t)
		response.Body.Close()
	})
}
