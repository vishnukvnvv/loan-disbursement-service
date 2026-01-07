package services

import (
	"mock-payment-gateway/config"
	"mock-payment-gateway/failures"
	"mock-payment-gateway/models"
	"mock-payment-gateway/payment"
	"mock-payment-gateway/types"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Mock PaymentChannel
type MockPaymentChannel struct {
	mock.Mock
}

func (m *MockPaymentChannel) ValidateLimit(amount float64) error {
	args := m.Called(amount)
	return args.Error(0)
}

func (m *MockPaymentChannel) Transfer(request models.PaymentRequest) (*models.Payment, error) {
	args := m.Called(request)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Payment), args.Error(1)
}

func (m *MockPaymentChannel) GetTransaction(transactionID string) (*models.Payment, error) {
	args := m.Called(transactionID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Payment), args.Error(1)
}

func TestPaymentService_Process(t *testing.T) {
	t.Run("successful payment processing", func(t *testing.T) {
		cfg := &config.Configuration{
			InvalidIFSC:          []string{},
			InvalidAccountNumber: []string{},
			BeneficiaryBankDown:  []string{},
		}

		// Initialize payment methods
		payment.Initialize(payment.PaymentMethods{
			UPI: payment.PaymentConstraints{
				Limit:       100000,
				Timeout:     3,
				FailureRate: 0.09,
			},
			IMPS: payment.PaymentConstraints{
				Limit:       500000,
				Timeout:     4.5,
				FailureRate: 0.06,
			},
			NEFT: payment.PaymentConstraints{
				Timeout:     6,
				FailureRate: 0.005,
			},
		})

		service := NewPaymentService(cfg)

		req := models.PaymentRequest{
			ReferenceID: "REF123",
			Amount:      5000.0,
			Mode:        types.UPI,
			Beneficiary: models.Beneficiary{
				Name:    "John Doe",
				Account: "1234567890",
				IFSC:    "IFSC0001234",
				Bank:    "Test Bank",
			},
		}

		result, err := service.Process(req)

		// Since payment processing involves randomness (failure rate), we check for either success or failure
		assert.NotNil(t, result)
		// If no error, should have a transaction
		if err == nil {
			assert.NotEmpty(t, result.TransactionID)
		}
	})

	t.Run("returns error for duplicate reference ID", func(t *testing.T) {
		cfg := &config.Configuration{
			InvalidIFSC:          []string{},
			InvalidAccountNumber: []string{},
			BeneficiaryBankDown:  []string{},
		}

		payment.Initialize(payment.PaymentMethods{
			UPI: payment.PaymentConstraints{
				Limit:       100000,
				Timeout:     3,
				FailureRate: 0.0, // No failures for testing
			},
		})

		service := NewPaymentService(cfg)

		req := models.PaymentRequest{
			ReferenceID: "REF123",
			Amount:      5000.0,
			Mode:        types.UPI,
			Beneficiary: models.Beneficiary{
				Name:    "John Doe",
				Account: "1234567890",
				IFSC:    "IFSC0001234",
				Bank:    "Test Bank",
			},
		}

		// First request should succeed
		_, err1 := service.Process(req)
		assert.NoError(t, err1)

		// Second request with same reference ID should fail
		_, err2 := service.Process(req)
		assert.Error(t, err2)
		assert.Contains(t, err2.Error(), "Reference ID already processed")
	})

	t.Run("returns error for invalid IFSC", func(t *testing.T) {
		cfg := &config.Configuration{
			InvalidIFSC:          []string{"INVALID_IFSC"},
			InvalidAccountNumber: []string{},
			BeneficiaryBankDown:  []string{},
		}

		payment.Initialize(payment.PaymentMethods{
			UPI: payment.PaymentConstraints{
				Limit:       100000,
				Timeout:     3,
				FailureRate: 0.0,
			},
		})

		service := NewPaymentService(cfg)

		req := models.PaymentRequest{
			ReferenceID: "REF123",
			Amount:      5000.0,
			Mode:        types.UPI,
			Beneficiary: models.Beneficiary{
				Name:    "John Doe",
				Account: "1234567890",
				IFSC:    "INVALID_IFSC",
				Bank:    "Test Bank",
			},
		}

		_, err := service.Process(req)
		assert.Error(t, err)
		assert.Equal(t, failures.INVALID_IFSC, err)
	})

	t.Run("returns error for inactive account", func(t *testing.T) {
		cfg := &config.Configuration{
			InvalidIFSC:          []string{},
			InvalidAccountNumber: []string{"INACTIVE_ACCOUNT"},
			BeneficiaryBankDown:  []string{},
		}

		payment.Initialize(payment.PaymentMethods{
			UPI: payment.PaymentConstraints{
				Limit:       100000,
				Timeout:     3,
				FailureRate: 0.0,
			},
		})

		service := NewPaymentService(cfg)

		req := models.PaymentRequest{
			ReferenceID: "REF123",
			Amount:      5000.0,
			Mode:        types.UPI,
			Beneficiary: models.Beneficiary{
				Name:    "John Doe",
				Account: "INACTIVE_ACCOUNT",
				IFSC:    "IFSC0001234",
				Bank:    "Test Bank",
			},
		}

		_, err := service.Process(req)
		assert.Error(t, err)
		assert.Equal(t, failures.INACTIVE_ACCOUNT, err)
	})

	t.Run("returns error for bank down", func(t *testing.T) {
		cfg := &config.Configuration{
			InvalidIFSC:          []string{},
			InvalidAccountNumber: []string{},
			BeneficiaryBankDown:  []string{"DOWN_BANK"},
		}

		payment.Initialize(payment.PaymentMethods{
			UPI: payment.PaymentConstraints{
				Limit:       100000,
				Timeout:     3,
				FailureRate: 0.0,
			},
		})

		service := NewPaymentService(cfg)

		req := models.PaymentRequest{
			ReferenceID: "REF123",
			Amount:      5000.0,
			Mode:        types.UPI,
			Beneficiary: models.Beneficiary{
				Name:    "John Doe",
				Account: "1234567890",
				IFSC:    "IFSC0001234",
				Bank:    "DOWN_BANK",
			},
		}

		_, err := service.Process(req)
		assert.Error(t, err)
		assert.Equal(t, failures.BENEFICIARY_BANK_DOWN, err)
	})

	t.Run("returns error for invalid payment mode", func(t *testing.T) {
		cfg := &config.Configuration{
			InvalidIFSC:          []string{},
			InvalidAccountNumber: []string{},
			BeneficiaryBankDown:  []string{},
		}

		payment.Initialize(payment.PaymentMethods{
			UPI: payment.PaymentConstraints{
				Limit:       100000,
				Timeout:     3,
				FailureRate: 0.0,
			},
		})

		service := NewPaymentService(cfg)

		req := models.PaymentRequest{
			ReferenceID: "REF123",
			Amount:      5000.0,
			Mode:        types.PaymentMode("INVALID"),
			Beneficiary: models.Beneficiary{
				Name:    "John Doe",
				Account: "1234567890",
				IFSC:    "IFSC0001234",
				Bank:    "Test Bank",
			},
		}

		_, err := service.Process(req)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid payment mode")
	})
}

func TestPaymentService_GetTransaction(t *testing.T) {
	t.Run("successful transaction fetch for UPI", func(t *testing.T) {
		cfg := &config.Configuration{}

		payment.Initialize(payment.PaymentMethods{
			UPI: payment.PaymentConstraints{
				Limit:       100000,
				Timeout:     3,
				FailureRate: 0.0,
			},
		})

		service := NewPaymentService(cfg)

		// First create a transaction
		req := models.PaymentRequest{
			ReferenceID: "REF123",
			Amount:      5000.0,
			Mode:        types.UPI,
			Beneficiary: models.Beneficiary{
				Name:    "John Doe",
				Account: "1234567890",
				IFSC:    "IFSC0001234",
				Bank:    "Test Bank",
			},
		}

		paymentResult, err := service.Process(req)
		if err != nil {
			t.Skip("Payment processing failed, skipping transaction fetch test")
		}

		// Now fetch the transaction
		result, err := service.GetTransaction(paymentResult.TransactionID)
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, paymentResult.TransactionID, result.TransactionID)
	})

	t.Run("returns error for empty transaction ID", func(t *testing.T) {
		cfg := &config.Configuration{}
		service := NewPaymentService(cfg)

		_, err := service.GetTransaction("")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "transaction ID cannot be empty")
	})

	t.Run("returns error for invalid transaction ID format", func(t *testing.T) {
		cfg := &config.Configuration{}
		service := NewPaymentService(cfg)

		_, err := service.GetTransaction("INVALID123")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid transaction ID format")
	})

	t.Run("returns error for transaction not found", func(t *testing.T) {
		cfg := &config.Configuration{}

		payment.Initialize(payment.PaymentMethods{
			UPI: payment.PaymentConstraints{
				Limit:       100000,
				Timeout:     3,
				FailureRate: 0.0,
			},
		})

		service := NewPaymentService(cfg)

		// Use a valid format but non-existent transaction ID
		_, err := service.GetTransaction("UPI000000000000")
		assert.Error(t, err)
	})
}

func TestPaymentService_GetPaymentMode(t *testing.T) {
	cfg := &config.Configuration{}
	service := NewPaymentService(cfg)

	t.Run("returns UPI for UPI prefix", func(t *testing.T) {
		mode, err := service.getPaymentMode("UPI123456789")
		assert.NoError(t, err)
		assert.Equal(t, types.UPI, mode)
	})

	t.Run("returns IMPS for IMPS prefix", func(t *testing.T) {
		mode, err := service.getPaymentMode("IMPS123456789")
		assert.NoError(t, err)
		assert.Equal(t, types.IMPS, mode)
	})

	t.Run("returns NEFT for NEFT prefix", func(t *testing.T) {
		mode, err := service.getPaymentMode("NEFT123456789")
		assert.NoError(t, err)
		assert.Equal(t, types.NEFT, mode)
	})

	t.Run("returns error for empty transaction ID", func(t *testing.T) {
		mode, err := service.getPaymentMode("")
		assert.Error(t, err)
		assert.Equal(t, types.PaymentMode(""), mode)
		assert.Contains(t, err.Error(), "transaction ID cannot be empty")
	})

	t.Run("returns error for invalid prefix", func(t *testing.T) {
		mode, err := service.getPaymentMode("INVALID123")
		assert.Error(t, err)
		assert.Equal(t, types.PaymentMode(""), mode)
		assert.Contains(t, err.Error(), "invalid transaction ID format")
	})
}

func TestPaymentService_ValidateBeneficiary(t *testing.T) {
	t.Run("validates beneficiary successfully", func(t *testing.T) {
		cfg := &config.Configuration{
			InvalidIFSC:          []string{},
			InvalidAccountNumber: []string{},
			BeneficiaryBankDown:  []string{},
		}

		service := NewPaymentService(cfg)

		beneficiary := models.Beneficiary{
			Name:    "John Doe",
			Account: "1234567890",
			IFSC:    "IFSC0001234",
			Bank:    "Test Bank",
		}

		err := service.validateBeneficiary(beneficiary)
		assert.NoError(t, err)
	})

	t.Run("returns error for invalid IFSC", func(t *testing.T) {
		cfg := &config.Configuration{
			InvalidIFSC:          []string{"INVALID_IFSC"},
			InvalidAccountNumber: []string{},
			BeneficiaryBankDown:  []string{},
		}

		service := NewPaymentService(cfg)

		beneficiary := models.Beneficiary{
			Name:    "John Doe",
			Account: "1234567890",
			IFSC:    "INVALID_IFSC",
			Bank:    "Test Bank",
		}

		err := service.validateBeneficiary(beneficiary)
		assert.Error(t, err)
		assert.Equal(t, failures.INVALID_IFSC, err)
	})

	t.Run("returns error for inactive account", func(t *testing.T) {
		cfg := &config.Configuration{
			InvalidIFSC:          []string{},
			InvalidAccountNumber: []string{"INACTIVE_ACCOUNT"},
			BeneficiaryBankDown:  []string{},
		}

		service := NewPaymentService(cfg)

		beneficiary := models.Beneficiary{
			Name:    "John Doe",
			Account: "INACTIVE_ACCOUNT",
			IFSC:    "IFSC0001234",
			Bank:    "Test Bank",
		}

		err := service.validateBeneficiary(beneficiary)
		assert.Error(t, err)
		assert.Equal(t, failures.INACTIVE_ACCOUNT, err)
	})

	t.Run("returns error for bank down", func(t *testing.T) {
		cfg := &config.Configuration{
			InvalidIFSC:          []string{},
			InvalidAccountNumber: []string{},
			BeneficiaryBankDown:  []string{"DOWN_BANK"},
		}

		service := NewPaymentService(cfg)

		beneficiary := models.Beneficiary{
			Name:    "John Doe",
			Account: "1234567890",
			IFSC:    "IFSC0001234",
			Bank:    "DOWN_BANK",
		}

		err := service.validateBeneficiary(beneficiary)
		assert.Error(t, err)
		assert.Equal(t, failures.BENEFICIARY_BANK_DOWN, err)
	})
}

func TestPaymentService_ReferenceIdTracking(t *testing.T) {
	t.Run("tracks reference IDs across multiple requests", func(t *testing.T) {
		cfg := &config.Configuration{
			InvalidIFSC:          []string{},
			InvalidAccountNumber: []string{},
			BeneficiaryBankDown:  []string{},
		}

		payment.Initialize(payment.PaymentMethods{
			UPI: payment.PaymentConstraints{
				Limit:       100000,
				Timeout:     3,
				FailureRate: 0.0,
			},
		})

		service := NewPaymentService(cfg)

		req1 := models.PaymentRequest{
			ReferenceID: "REF123",
			Amount:      5000.0,
			Mode:        types.UPI,
			Beneficiary: models.Beneficiary{
				Name:    "John Doe",
				Account: "1234567890",
				IFSC:    "IFSC0001234",
				Bank:    "Test Bank",
			},
		}

		req2 := models.PaymentRequest{
			ReferenceID: "REF456",
			Amount:      6000.0,
			Mode:        types.UPI,
			Beneficiary: models.Beneficiary{
				Name:    "Jane Doe",
				Account: "0987654321",
				IFSC:    "IFSC0005678",
				Bank:    "Test Bank",
			},
		}

		// Both should succeed with different reference IDs
		_, err1 := service.Process(req1)
		assert.NoError(t, err1)

		_, err2 := service.Process(req2)
		assert.NoError(t, err2)

		// But duplicate reference ID should fail
		_, err3 := service.Process(req1)
		assert.Error(t, err3)
		assert.Contains(t, err3.Error(), "Reference ID already processed")
	})
}
