.PHONY: setup dev-up dev-down dev-logs test tidy up down logs restart-db release

COMPOSE      = docker compose
COMPOSE_DEV  = docker compose -f docker-compose.yml -f docker-compose.dev.yml
GO_IMAGE     = golang:1.26-alpine

GO_CACHE = -v gonami-gomod:/go/pkg/mod -v gonami-gocache:/root/.cache/go-build

setup:
	@test -f .env || cp .env.example .env
	@echo "Edit .env lalu isi GEMINI_API_KEY, jalankan: make dev-up"

dev-up:
	$(COMPOSE_DEV) up -d --build
	@echo ""
	@echo "Log & QR pairing : make dev-logs"
	@echo "Console MinIO    : http://localhost:9001"

dev-down:
	$(COMPOSE_DEV) down

dev-logs:
	$(COMPOSE_DEV) logs -f bot

test:
	docker run --rm \
		$(GO_CACHE) \
		-v "$(CURDIR)":/src -w /src \
		$(GO_IMAGE) go test ./... -count=1

tidy:
	docker run --rm \
		$(GO_CACHE) \
		-v "$(CURDIR)":/src -w /src \
		$(GO_IMAGE) sh -c "go mod tidy && chown -R $(shell id -u):$(shell id -g) /src/go.mod /src/go.sum"

up:
	$(COMPOSE) pull bot
	$(COMPOSE) up -d

down:
	$(COMPOSE) down

logs:
	$(COMPOSE) logs -f bot

restart-db:
	$(COMPOSE) restart db bot

release:
	@test -n "$(version)" || { echo "pakai: make release version=v1.0.0"; exit 1; }
	git tag $(version)
	git push origin $(version)
	@echo ""
	@echo "Tag $(version) ter-push. CI akan:"
	@echo "  1. test -> build multi-arch -> push ghcr.io:$(version)"
	@echo "  2. membuat GitHub Release dengan changelog otomatis"
