package services

import (
	"loan-disbursement-service/db"
	"loan-disbursement-service/providers"
	"loan-disbursement-service/utils"
)

type ServiceFactory struct {
	database       *db.Database
	paymentService PaymentService
	disbursement   DisbursementService
	loanService    LoanService
	retryPolicy    RetryPolicy
	reconciliation ReconciliationService
}

func New(
	database *db.Database,
	idGenerator utils.IdGenerator,
	paymentProvider providers.PaymentProvider,
	notificationURL string,
) *ServiceFactory {
	retryPolicy := NewRetryPolicy()
	return &ServiceFactory{
		database:    database,
		retryPolicy: retryPolicy,
		disbursement: NewDisbursementService(
			idGenerator,
			database.GetLoanRepository(),
			database.GetDisbursementRepository(),
			database.GetTransactionRepository(),
			database.GetBeneficiaryRepository(),
		),
		loanService: NewLoanService(database.GetLoanRepository(), idGenerator),
		paymentService: NewPaymentService(
			database,
			database.GetDisbursementRepository(),
			database.GetTransactionRepository(),
			database.GetLoanRepository(),
			database.GetBeneficiaryRepository(),
			retryPolicy,
			paymentProvider,
			idGenerator,
			notificationURL,
		),
		reconciliation: NewReconciliationService(
			idGenerator,
			database.GetTransactionRepository(),
		),
	}
}

func (f *ServiceFactory) GetDisbursementService() DisbursementService {
	return f.disbursement
}

func (f *ServiceFactory) GetLoanService() LoanService {
	return f.loanService
}

func (f *ServiceFactory) GetPaymentService() PaymentService {
	return f.paymentService
}

func (f *ServiceFactory) GetRetryPolicy() RetryPolicy {
	return f.retryPolicy
}

func (f *ServiceFactory) GetReconciliationService() ReconciliationService {
	return f.reconciliation
}
