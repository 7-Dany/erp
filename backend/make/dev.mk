##  Convenience / combined workflows

.PHONY: db-sync psql check dev

db-sync: migrate-up sqlc-vet sqlc-generate

psql: psql -U $(DB_USER) -h $(DB_HOST) -p $(DB_PORT) -d $(DB_NAME)

check: fmt vet lint test

dev: check build run
