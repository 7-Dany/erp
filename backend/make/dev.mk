##  Convenience / combined workflows

.PHONY: db-sync psql check dev demo

db-sync: migrate-up sqlc-vet sqlc-generate

psql: psql -U $(DB_USER) -h $(DB_HOST) -p $(DB_PORT) -d $(DB_NAME)

check: fmt vet lint test

dev: check build run

demo: build
	set "ERP_DEMO=1" & bin\app.exe
