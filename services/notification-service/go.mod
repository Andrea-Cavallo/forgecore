module github.com/yourorg/golang-modules/services/notification-service

go 1.24

require (
	github.com/google/uuid v1.6.0
	github.com/jackc/pgx/v5 v5.7.0
	github.com/nats-io/nats.go v1.37.0
	github.com/yourorg/golang-modules/shared v0.0.0
)

require (
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20240606120523-5a60cdf6a761 // indirect
	github.com/jackc/puddle/v2 v2.2.1 // indirect
	github.com/klauspost/compress v1.17.2 // indirect
	github.com/nats-io/nkeys v0.4.7 // indirect
	github.com/nats-io/nuid v1.0.1 // indirect
	golang.org/x/crypto v0.18.0 // indirect
	golang.org/x/sync v0.1.0 // indirect
	golang.org/x/sys v0.16.0 // indirect
	golang.org/x/text v0.14.0 // indirect
)

replace github.com/yourorg/golang-modules/shared => ../../shared
