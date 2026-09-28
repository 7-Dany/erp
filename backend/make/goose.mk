##  Migrations (goose)

.PHONY: migrate-up migrate-down migrate-status migrate-create

migrate-up:
	goose -dir sql/schema postgres "$(DB_URL)" up

migrate-down:
	goose -dir sql/schema postgres "$(DB_URL)" down

migrate-status:
	goose -dir sql/schema postgres "$(DB_URL)" status

migrate-create:
	@read -p "Migration name: " name; \
	goose -dir sql/schema create $$name sql
