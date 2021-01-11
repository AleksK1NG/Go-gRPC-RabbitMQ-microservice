### Go, RabbitMQ and gRPC [Clean Architecture](https://blog.cleancoder.com/uncle-bob/2012/08/13/the-clean-architecture.html) microservice 👋


#### 👨‍💻 Full list what has been used:
* [GRPC](https://grpc.io/) - gRPC
* [RabbitMQ](https://github.com/streadway/amqp) - RabbitMQ
* [sqlx](https://github.com/jmoiron/sqlx) - Extensions to database/sql.
* [pgx](https://github.com/jackc/pgx) - PostgreSQL driver and toolkit for Go
* [viper](https://github.com/spf13/viper) - Go configuration with fangs
* [zap](https://github.com/uber-go/zap) - Logger
* [validator](https://github.com/go-playground/validator) - Go Struct and Field validation
* [migrate](https://github.com/golang-migrate/migrate) - Database migrations. CLI and Golang library.
* [testify](https://github.com/stretchr/testify) - Testing toolkit
* [gomock](https://github.com/golang/mock) - Mocking framework
* [CompileDaemon](https://github.com/githubnemo/CompileDaemon) - Compile daemon for Go
* [Docker](https://www.docker.com/) - Docker
* [Prometheus](https://prometheus.io/) - Prometheus
* [Grafana](https://grafana.com/) - Grafana
* [Jaeger](https://www.jaegertracing.io/) - Jaeger tracing
* [Bluemonday](https://github.com/microcosm-cc/bluemonday) - HTML sanitizer
* [Gomail](https://github.com/go-gomail/gomail/tree/v2) - Simple and efficient package to send emails
* [Go-sqlmock](https://github.com/DATA-DOG/go-sqlmock) - Sql mock driver for golang to test database interactions
* [Go-grpc-middleware](https://github.com/grpc-ecosystem/go-grpc-middleware) - interceptor chaining, auth, logging, retries and more
* [Opentracing-go](https://github.com/opentracing/opentracing-go) - OpenTracing API for Go
* [Prometheus-go-client](https://github.com/prometheus/client_golang) - Prometheus instrumentation library for Go applications

#### Recommendation for local development most comfortable usage:
    make local // run all containers
    make run // run the application

#### 🙌👨‍💻🚀 Docker-compose files:
    docker-compose.local.yml - run rabbitmq, postgresql, jaeger, prometheus, grafana containers
    docker-compose.yml - run all in docker

### Docker development usage:
    make docker

### Local development usage:
    make local
    make run

### Jaeger UI:

http://localhost:16686

### Prometheus UI:

http://localhost:9090

### Grafana UI:

http://localhost:3000

### RabbitMQ UI:

http://localhost:15672