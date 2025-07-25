package biz

import "github.com/google/wire"

// ProviderSet is biz providers.
var ProviderSet = wire.NewSet(NewGreeterUsecase, NewUserUsecase, NewHouseUsecase, NewTransactionUsecase, NewPointsUsecase, NewCustomerUsecase)
