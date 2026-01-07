package services

import (
	"loan-disbursement-service/db"
	"loan-disbursement-service/providers"
)

type ServiceFactory struct {
	database            *db.Database
	paymentProvider     providers.PaymentProvider
	paymentService      *PaymentService
	disbursementService *DisbursementService
	loanService         *LoanService
}

func New(
	database *db.Database,
	paymentProvider providers.PaymentProvider,
) *ServiceFactory {
	return &ServiceFactory{
		database:            database,
		paymentProvider:     paymentProvider,
		paymentService:      nil,
		disbursementService: nil,
		loanService:         nil,
	}
}

func (f *ServiceFactory) GetPaymentService() *PaymentService {
	if f.paymentService == nil {
		f.paymentService = NewPaymentService(
			f.database.GetLoanDAO(),
			NewRetryPolicy(),
			f.database.GetBeneficiaryDAO(),
			f.database.GetDisbursementDAO(),
			f.database.GetTransactionDAO(),
			f.paymentProvider,
		)
	}
	return f.paymentService
}

func (f *ServiceFactory) GetDisbursementService() *DisbursementService {
	if f.disbursementService == nil {
		f.disbursementService = NewDisbursementService(
			f.database.GetLoanDAO(),
			f.database.GetDisbursementDAO(),
			f.database.GetTransactionDAO(),
			f.database.GetBeneficiaryDAO(),
		)
	}
	return f.disbursementService
}

func (f *ServiceFactory) GetLoanService() *LoanService {
	if f.loanService == nil {
		f.loanService = NewLoanService(f.database.GetLoanDAO())
	}
	return f.loanService
}
