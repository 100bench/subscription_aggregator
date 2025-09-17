package entities

import "errors"

var (
	ErrNilDependency        = errors.New("nil dependency: ")
	ErrSubscriptionNotFound = errors.New("subscription not found")
)
