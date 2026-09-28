include .env
export

DB_URL := postgres://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=$(DB_SSLMODE)

include make/goose.mk
include make/sqlc.mk
include make/dev.mk
include make/go.mk
