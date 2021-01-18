# ==============================================================================
# Main

run:
	go run ./cmd/email_service/main.go

build:
	go build ./cmd/email_service/main.go

test:
	go test -cover ./...


# ==============================================================================
# Modules support

deps-reset:
	git checkout -- go.mod
	go mod tidy
	go mod vendor

tidy:
	go mod tidy
	go mod vendor

deps-upgrade:
	# go get $(go list -f '{{if not (or .Main .Indirect)}}{{.Path}}{{end}}' -m all)
	go get -u -t -d -v ./...
	go mod tidy
	go mod vendor

deps-cleancache:
	go clean -modcache

# ==============================================================================
# Tools commands

run-linter:
	echo "Starting linters"
	golangci-lint run ./...


# ==============================================================================
# Go migrate postgresql

DB_NAME = mails_db

force:
	migrate -database postgres://postgres:postgres@localhost:5432/$(DB_NAME)?sslmode=disable -path migrations force 1

version:
	migrate -database postgres://postgres:postgres@localhost:5432/$(DB_NAME)?sslmode=disable -path migrations version

migrate_up:
	migrate -database postgres://postgres:postgres@localhost:5432/$(DB_NAME)?sslmode=disable -path migrations up 1

migrate_down:
	migrate -database postgres://postgres:postgres@localhost:5432/$(DB_NAME)?sslmode=disable -path migrations down 1


# ==============================================================================
# Docker compose commands

dev_docker:
	echo "Starting docker environment"
	make jaeger
	docker-compose -f docker-compose.yml up --build


local:
	echo "Starting local environment"
	make jaeger
	docker-compose -f docker-compose.local.yml up --build


jaeger:
	echo "Starting jaeger containers"
	docker run -d --name jaeger \
      -e COLLECTOR_ZIPKIN_HTTP_PORT=9411 \
      -p 5775:5775/udp \
      -p 6831:6831/udp \
      -p 6832:6832/udp \
      -p 5778:5778 \
      -p 16686:16686 \
      -p 14268:14268 \
      -p 14250:14250 \
      -p 9411:9411 \
      jaegertracing/all-in-one:1.21


# ==============================================================================
# Docker support

FILES := $(shell docker ps -aq)

down-local:
	docker stop $(FILES)
	docker rm $(FILES)

clean:
	docker system prune -f

logs-local:
	docker logs -f $(FILES)