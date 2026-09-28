## sqlc

.PHONY: sqlc-vet sqlc-generate

sqlc-vet:
	sqlc vet

sqlc-generate:
	sqlc generate
