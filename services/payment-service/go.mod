module github.com/yourorg/golang-modules/services/payment-service

go 1.24

require (
	github.com/google/uuid v1.6.0
	github.com/jackc/pgx/v5 v5.7.0
	github.com/nats-io/nats.go v1.37.0
	github.com/yourorg/golang-modules/shared v0.0.0
	go.uber.org/zap v1.27.0
)

replace github.com/yourorg/golang-modules/shared => ../../shared
