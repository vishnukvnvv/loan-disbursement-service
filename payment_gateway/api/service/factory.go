package service

import (
	"payment-gateway/db"
	"payment-gateway/models"
	"payment-gateway/utils"
)

type ServiceFactory interface {
	GetAccountService() AccountService
	GetPaymentChannelService() PaymentChannelService
	GetAvailabilitySchedule() AvailabilitySchedule
	GetPaymentService() PaymentService
}

type ServiceFactoryImpl struct {
	accountService        AccountService
	paymentChannelService PaymentChannelService
	availabilitySchedule  AvailabilitySchedule
	paymentService        PaymentService
}

func NewServiceFactory(
	db *db.Database,
	processor chan models.ProcessorMessage,
	idGenerator utils.IdGenerator,
) ServiceFactory {
	return &ServiceFactoryImpl{
		accountService: NewAccountService(db.GetAccountRepository(), idGenerator),
		paymentChannelService: NewPaymentChannelService(
			db.GetPaymentChannelRepository(),
			idGenerator,
		),
		availabilitySchedule: NewAvailabilitySchedule(),
		paymentService: NewPaymentService(
			processor,
			db.GetPaymentChannelRepository(),
			db.GetTransactionRepository(),
			idGenerator,
		),
	}
}

func (f *ServiceFactoryImpl) GetAccountService() AccountService {
	return f.accountService
}

func (f *ServiceFactoryImpl) GetPaymentChannelService() PaymentChannelService {
	return f.paymentChannelService
}

func (f *ServiceFactoryImpl) GetAvailabilitySchedule() AvailabilitySchedule {
	return f.availabilitySchedule
}

func (f *ServiceFactoryImpl) GetPaymentService() PaymentService {
	return f.paymentService
}
