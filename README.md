# Go, RabbitMQ and gRPC [Clean Architecture](https://blog.cleancoder.com/uncle-bob/2012/08/13/the-clean-architecture.html) microservice 👋
Here is article about create the example of similar to real production mail microservice using RabbitMQ, gRPC, Prometheus, Grafana monitoring and Jaeger opentracing ⚡️


#### 👨‍💻 Full list what has been used:
* [GRPC](https://grpc.io/) - gRPC
* [RabbitMQ](https://github.com/streadway/amqp) - RabbitMQ
* [sqlx](https://github.com/jmoiron/sqlx) - Extensions to database/sql.
* [pgx](https://github.com/jackc/pgx) - PostgreSQL driver and toolkit for Go
* [viper](https://github.com/spf13/viper) - Go configuration with fangs
* [go-redis](https://github.com/go-redis/redis) - Redis client for Golang
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

First run all necessary docker containers:

```go
run make local
```

UI interfaces will be available on ports:
### RabbitMQ UI: http://localhost:15672
### Jaeger UI: http://localhost:16686
### Prometheus UI: http://localhost:9090
### Grafana UI: http://localhost:3000

The RabbitMQ management plugin UI available on http://localhost:15672 default login/password is **guest/guest**:
<img src="https://i.postimg.cc/xjPYGbLW/Rabbit-MQ-Management-2020-12-30-11-46-13.png" />


As soon as we run containers, Prometheus and Grafana UI is available too, and after sending any requests, you are able to monitoring of metrics at the dashboard:<br/>
Grafana default login/password is **admin** and password **admin**<br/>
<img src="https://i.postimg.cc/JnxXrCky/g-RPC-Sever-Grafana-2020-12-24-13-42-29.png" />
<img src="https://i.postimg.cc/2jJf2zT2/Prometheus-Time-Series-Collection-and-Processing-Server-2020-12-24-13-42-50.png" />


RabbitMQ have Go client library and good documentation with nice tutorial<br/>
Dial connection:
```go
// Initialize new RabbitMQ connection
func NewRabbitMQConn(cfg *config.Config) (*amqp.Connection, error) {
	connAddr := fmt.Sprintf(
		"amqp://%s:%s@%s:%s/",
		cfg.RabbitMQ.User,
		cfg.RabbitMQ.Password,
		cfg.RabbitMQ.Host,
		cfg.RabbitMQ.Port,
	)
	return amqp.Dial(connAddr)
}
```
Let's create our producer and consumer, pass amqpConn to the constructor<br/>

```go
// Images Rabbitmq consumer
type EmailsConsumer struct {
	amqpConn *amqp.Connection
	logger   logger.Logger
	emailUC  email.EmailsUseCase
}
```
Create new channel method, amqpConn.Channel opens a unique, concurrent server channel.
Next we need declare Exchange and Queue, then bind them using bindingKey Routing key,
The ch.Qos method allows us prefetch messages. With a prefetch count greater than zero, the server will deliver that many
messages to consumers before acknowledgments are received.

```go
// Consume messages
func (c *EmailsConsumer) CreateChannel(exchangeName, queueName, bindingKey, consumerTag string) (*amqp.Channel, error) {
	ch, err := c.amqpConn.Channel()
	if err != nil {
		return nil, errors.Wrap(err, "Error amqpConn.Channel")
	}

	c.logger.Infof("Declaring exchange: %s", exchangeName)
	err = ch.ExchangeDeclare(
		exchangeName,
		exchangeKind,
		exchangeDurable,
		exchangeAutoDelete,
		exchangeInternal,
		exchangeNoWait,
		nil,
	)
	if err != nil {
		return nil, errors.Wrap(err, "Error ch.ExchangeDeclare")
	}

	queue, err := ch.QueueDeclare(
		queueName,
		queueDurable,
		queueAutoDelete,
		queueExclusive,
		queueNoWait,
		nil,
	)
	if err != nil {
		return nil, errors.Wrap(err, "Error ch.QueueDeclare")
	}

	c.logger.Infof("Declared queue, binding it to exchange: Queue: %v, messagesCount: %v, "+
		"consumerCount: %v, exchange: %v, bindingKey: %v",
		queue.Name,
		queue.Messages,
		queue.Consumers,
		exchangeName,
		bindingKey,
	)

	err = ch.QueueBind(
		queue.Name,
		bindingKey,
		exchangeName,
		queueNoWait,
		nil,
	)
	if err != nil {
		return nil, errors.Wrap(err, "Error ch.QueueBind")
	}

	c.logger.Infof("Queue bound to exchange, starting to consume from queue, consumerTag: %v", consumerTag)

	err = ch.Qos(
		prefetchCount,  // prefetch count
		prefetchSize,   // prefetch size
		prefetchGlobal, // global
	)
	if err != nil {
		return nil, errors.Wrap(err, "Error  ch.Qos")
	}

	return ch, nil
}

```
StartConsumer method accept exchange, queue params and [Worker Pools](https://gobyexample.com/worker-pools) size<br/>

```
// Start new rabbitmq consumer
func (c *EmailsConsumer) StartConsumer(workerPoolSize int, exchange, queueName, bindingKey, consumerTag string) error {
	ch, err := c.CreateChannel(exchange, queueName, bindingKey, consumerTag)
	if err != nil {
		return errors.Wrap(err, "CreateChannel")
	}
	defer ch.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	deliveries, err := ch.Consume(
		queueName,
		consumerTag,
		consumeAutoAck,
		consumeExclusive,
		consumeNoLocal,
		consumeNoWait,
		nil,
	)
	if err != nil {
		return errors.Wrap(err, "Consume")
	}

	wg := &sync.WaitGroup{}
	wg.Add(workerPoolSize)
	for i := 0; i < workerPoolSize; i++ {
		go c.worker(ctx, deliveries, wg)
	}

	wg.Wait()
	return nil
}
```
Very popular phrase in Go: “Do not communicate by sharing memory; instead, share memory by communicating.”<br/>
Run workers for the number of concurrency what we need:
```go
	for i := 0; i < workerPoolSize; i++ {
		go c.worker(ctx, deliveries, wg)
	}
```

Each worker reads from jobs channel and produces a job:<br/>
All deliveries in AMQP must be acknowledged. <br/>
If you called Channel.Consumewith autoAck true then the server will be automatically ack each message and
this method should not be called. Otherwise, you must call Delivery.Ack after
you have successfully processed this delivery.<br/>
Either Delivery.Ack, Delivery.Reject or Delivery.Nack must be called for every
delivery that is not automatically acknowledged.<br/>
Reject delegates a negatively acknowledgement through the Acknowledger interface.

```go
func (c *EmailsConsumer) worker(ctx context.Context, messages <-chan amqp.Delivery, wg *sync.WaitGroup) {
	defer wg.Done()

	for delivery := range messages {
		span, ctx := opentracing.StartSpanFromContext(ctx, "EmailsConsumer.worker")

		c.logger.Infof("processDeliveries deliveryTag% v", delivery.DeliveryTag)

		incomingMessages.Inc()

		err := c.emailUC.SendEmail(ctx, delivery.Body)
		if err != nil {
			if err := delivery.Reject(false); err != nil {
				c.logger.Errorf("Err delivery.Reject: %v", err)
			}
			c.logger.Errorf("Failed to process delivery: %v", err)
			errorMessages.Inc()
			span.Finish()
		} else {
			err = delivery.Ack(false)
			if err != nil {
				c.logger.Errorf("Failed to acknowledge delivery: %v", err)
			}
			successMessages.Inc()
			span.Finish()
		}
	}

	c.logger.Info("Deliveries channel closed")
}
```


I like to use [evans](https://github.com/ktr0731/evans) for simple testing gRPC.
<img src="https://i.postimg.cc/gc8n9RPQ/evans-evans-2020-12-25-16-45-21.png" />


In **cmd** folder let's init all dependencies and start the app.<br/>
[Viper](https://github.com/spf13/viper) is very good and common choice as complete configuration solution for Go applications.<br/>
We use here config-local.yml file approach.

```go
configPath := utils.GetConfigPath(os.Getenv("config"))
	cfg, err := config.GetConfig(configPath)
	if err != nil {
		log.Fatalf("Loading config: %v", err)
	}
```

Next let's create logger, here i used Uber's [Zap](https://github.com/uber-go/zap) under the hood, important here is to create Logger interface for be able to replace logger in the future if it's need.

```go
type Logger interface {
	InitLogger()
	Debug(args ...interface{})
	Debugf(template string, args ...interface{})
	Info(args ...interface{})
	Infof(template string, args ...interface{})
	Warn(args ...interface{})
	Warnf(template string, args ...interface{})
	Error(args ...interface{})
	Errorf(template string, args ...interface{})
	DPanic(args ...interface{})
	DPanicf(template string, args ...interface{})
	Fatal(args ...interface{})
	Fatalf(template string, args ...interface{})
}
```

Setup postgres and redis<br/>
Usually production SQL db standard solution for these days is combination of [sqlx](https://github.com/jmoiron/sqlx) and [pgx](https://github.com/jackc/pgx).<br/>
Good Redis Go clients is [go-redis](https://github.com/go-redis/redis) and [redigo](https://github.com/gomodule/redigo), i used first.

```go
func NewPsqlDB(c *config.Config) (*sqlx.DB, error) {
	dataSourceName := fmt.Sprintf("host=%s port=%s user=%s dbname=%s sslmode=disable password=%s",
		c.Postgres.PostgresqlHost,
		c.Postgres.PostgresqlPort,
		c.Postgres.PostgresqlUser,
		c.Postgres.PostgresqlDbname,
		c.Postgres.PostgresqlPassword,
	)

	db, err := sqlx.Connect(c.Postgres.PgDriver, dataSourceName)
	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(maxOpenConns)
	db.SetConnMaxLifetime(connMaxLifetime * time.Second)
	db.SetMaxIdleConns(maxIdleConns)
	db.SetConnMaxIdleTime(connMaxIdleTime * time.Second)
	if err = db.Ping(); err != nil {
		return nil, err
	}

	return db, nil
}

func NewRedisClient(cfg *config.Config) *redis.Client {
	redisHost := cfg.Redis.RedisAddr

	if redisHost == "" {
		redisHost = ":6379"
	}

	client := redis.NewClient(&redis.Options{
		Addr:         redisHost,
		MinIdleConns: cfg.Redis.MinIdleConns,
		PoolSize:     cfg.Redis.PoolSize,
		PoolTimeout:  time.Duration(cfg.Redis.PoolTimeout) * time.Second,
		Password:     cfg.Redis.Password, // no password set
		DB:           cfg.Redis.DB,       // use default DB
	})

	return client
}
```

And let's set up [Jaeger](https://www.jaegertracing.io/):

```go
func InitJaeger(cfg *config.Config) (opentracing.Tracer, io.Closer, error) {
	jaegerCfgInstance := jaegercfg.Configuration{
		ServiceName: cfg.Jaeger.ServiceName,
		Sampler: &jaegercfg.SamplerConfig{
			Type:  jaeger.SamplerTypeConst,
			Param: 1,
		},
		Reporter: &jaegercfg.ReporterConfig{
			LogSpans:           cfg.Jaeger.LogSpans,
			LocalAgentHostPort: cfg.Jaeger.Host,
		},
	}

	return jaegerCfgInstance.NewTracer(
		jaegercfg.Logger(jaegerlog.StdLogger),
		jaegercfg.Metrics(metrics.NullFactory),
	)
}
```

And add the global tracer to our application:

```go
	tracer, closer, err := jaegerTracer.InitJaeger(cfg)
	if err != nil {
		appLogger.Fatal("cannot create tracer", err)
	}
	appLogger.Info("Jaeger connected")

	opentracing.SetGlobalTracer(tracer)
	defer closer.Close()
```

Prometheus has have [4 types of metrics](https://prometheus.io/docs/concepts/metric_types/): Counter, Gauge, Histogram, Summary <br/>
To expose Prometheus metrics in a Go application, you need to provide a /metrics HTTP endpoint. <br/>
You can use the prometheus/promhttp library's HTTP Handler as the handler function.<br/>

```go
func CreateMetrics(address string, name string) (Metrics, error) {
	var metr PrometheusMetrics
	metr.HitsTotal = prometheus.NewCounter(prometheus.CounterOpts{
		Name: name + "_hits_total",
	})
	if err := prometheus.Register(metr.HitsTotal); err != nil {
		return nil, err
	}
	metr.Hits = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: name + "_hits",
		},
		[]string{"status", "method", "path"},
	)
	if err := prometheus.Register(metr.Hits); err != nil {
		return nil, err
	}
	metr.Times = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name: name + "_times",
		},
		[]string{"status", "method", "path"},
	)
	if err := prometheus.Register(metr.Times); err != nil {
		return nil, err
	}
	if err := prometheus.Register(prometheus.NewBuildInfoCollector()); err != nil {
		return nil, err
	}
	go func() {
		router := echo.New()
		router.GET("/metrics", echo.WrapHandler(promhttp.Handler()))
		if err := router.Start(address); err != nil {
			log.Fatal(err)
		}
	}()
	return &metr, nil
}
```

user.proto file<br/>
In gRPC documentation we can find good practice recommendations and naming conventions for writing proto files.<br/>
As you can see, each field in the message definition has a unique number. These field numbers are used to identify your fields in the message binary format, and should not be changed once your message type is in use. Note that field numbers in the range 1 through 15 take one byte to encode, including the field number and the field's type (you can find out more about this in Protocol Buffer Encoding).

```
message LoginRequest {
  string email = 1;
  string password = 2;
}

message LoginResponse {
  User user = 1;
  string session_id = 2;
}

service UserService{
  rpc Register(RegisterRequest) returns (RegisterResponse);
  rpc FindByEmail(FindByEmailRequest) returns (FindByEmailResponse);
  rpc FindByID(FindByIDRequest) returns (FindByIDResponse);
  rpc Login(LoginRequest) returns (LoginResponse);
  rpc GetMe(GetMeRequest) returns(GetMeResponse);
  rpc Logout(LogoutRequest) returns(LogoutResponse);
}
```

generate your user.[proto](https://developers.google.com/protocol-buffers/docs/gotutorial) file 🤓

```
protoc --go_out=plugins=grpc:. *.proto
```

it creates user.pb.go file with server and client interfaces what need to implement in our microservice:

```go
// UserServiceServer is the service API for UserService service.
type UserServiceServer interface {
	Register(context.Context, *RegisterRequest) (*RegisterResponse, error)
	FindByEmail(context.Context, *FindByEmailRequest) (*FindByEmailResponse, error)
	FindByID(context.Context, *FindByIDRequest) (*FindByIDResponse, error)
	Login(context.Context, *LoginRequest) (*LoginResponse, error)
	GetMe(context.Context, *GetMeRequest) (*GetMeResponse, error)
	Logout(context.Context, *LogoutRequest) (*LogoutResponse, error)
}
```

Then in server.go initialize the repository, use cases, metrics and so on then start gRPC server:

```go
func (s *Server) Run() error {
	metrics, err := metric.CreateMetrics(s.cfg.Metrics.URL, s.cfg.Metrics.ServiceName)
	if err != nil {
		s.logger.Errorf("CreateMetrics Error: %s", err)
	}
	s.logger.Info(
		"Metrics available URL: %s, ServiceName: %s",
		s.cfg.Metrics.URL,
		s.cfg.Metrics.ServiceName,
	)

	im := interceptors.NewInterceptorManager(s.logger, s.cfg, metrics)
	userRepo := userRepository.NewUserPGRepository(s.db)
	sessRepo := sessRepository.NewSessionRepository(s.redisClient, s.cfg)
	userRedisRepo := userRepository.NewUserRedisRepo(s.redisClient, s.logger)
	userUC := userUseCase.NewUserUseCase(s.logger, userRepo, userRedisRepo)
	sessUC := sessUseCase.NewSessionUseCase(sessRepo, s.cfg)

	l, err := net.Listen("tcp", s.cfg.Server.Port)
	if err != nil {
		return err
	}
	defer l.Close()

	server := grpc.NewServer(grpc.KeepaliveParams(keepalive.ServerParameters{
		MaxConnectionIdle: s.cfg.Server.MaxConnectionIdle * time.Minute,
		Timeout:           s.cfg.Server.Timeout * time.Second,
		MaxConnectionAge:  s.cfg.Server.MaxConnectionAge * time.Minute,
		Time:              s.cfg.Server.Timeout * time.Minute,
	}),
		grpc.UnaryInterceptor(im.Logger),
		grpc.ChainUnaryInterceptor(
			grpc_ctxtags.UnaryServerInterceptor(),
			grpc_prometheus.UnaryServerInterceptor,
			grpcrecovery.UnaryServerInterceptor(),
		),
	)

	if s.cfg.Server.Mode != "Production" {
		reflection.Register(server)
	}

	authGRPCServer := authServerGRPC.NewAuthServerGRPC(s.logger, s.cfg, userUC, sessUC)
	userService.RegisterUserServiceServer(server, authGRPCServer)

	grpc_prometheus.Register(server)
	http.Handle("/metrics", promhttp.Handler())

	go func() {
		s.logger.Infof("Server is listening on port: %v", s.cfg.Server.Port)
		if err := server.Serve(l); err != nil {
			s.logger.Fatal(err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	<-quit
	server.GracefulStop()
	s.logger.Info("Server Exited Properly")

	return nil
}
```

I found this is very good [gRPC Middleware repository](https://github.com/grpc-ecosystem/go-grpc-middleware), but we easy can create our own, for example logger interceptor:

```go
func (im *InterceptorManager) Logger(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
	start := time.Now()
	md, _ := metadata.FromIncomingContext(ctx)
	reply, err := handler(ctx, req)
	im.logger.Infof("Method: %s, Time: %v, Metadata: %v, Err: %v", info.FullMethod, time.Since(start), md, err)

	return reply, err
}
```

We can access grpc metadata in service handlers too, for example here we extract and validate ***session_id*** which client must send in the request context:

```go
md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return "", status.Errorf(codes.Unauthenticated, "metadata.FromIncomingContext: %v", grpc_errors.ErrNoCtxMetaData)
	}
sessionID := md.Get("session_id")
```

So let's create unary service handler for creating the new user:

```go
func (u *usersService) Register(ctx context.Context, r *userService.RegisterRequest) (*userService.RegisterResponse, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "user.Create")
	defer span.Finish()

	user, err := u.registerReqToUserModel(r)
	if err != nil {
		u.logger.Errorf("registerReqToUserModel: %v", err)
		return nil, status.Errorf(grpc_errors.ParseGRPCErrStatusCode(err), "registerReqToUserModel: %v", err)
	}

	if err := utils.ValidateStruct(ctx, user); err != nil {
		u.logger.Errorf("ValidateStruct: %v", err)
		return nil, status.Errorf(grpc_errors.ParseGRPCErrStatusCode(err), "ValidateStruct: %v", err)
	}

	createdUser, err := u.userUC.Register(ctx, user)
	if err != nil {
		u.logger.Errorf("userUC.Register: %v", err)
		return nil, status.Errorf(grpc_errors.ParseGRPCErrStatusCode(err), "Register: %v", err)
	}

	return &userService.RegisterResponse{User: u.userModelToProto(createdUser)}, nil
}
```

On the first lines start tracing span. The “span” is the primary building block of a distributed trace.<br/>
Each component of the distributed system contributes a span - a named, timed operation representing a piece of the workflow.<br/>
[opentracing](https://opentracing.io/docs/overview/spans/#:~:text=The%20%E2%80%9Cspan%E2%80%9D%20is%20the%20primary,a%20piece%20of%20the%20workflow)

```go
span, ctx := opentracing.StartSpanFromContext(ctx, "user.Create")
defer span.Finish()
```

let's check how it's look in Jaeger:<br/>
open http://localhost:16686/<br/>
<img src="https://i.postimg.cc/8zyh2n3k/Jaeger-UI-2020-12-24-16-24-33.png" />

Then we usually have to validate request input, for errors gRPC has packages [status](https://godoc.org/google.golang.org/grpc/status) and [codes](https://godoc.org/google.golang.org/grpc/codes)<br/>
I found good practice to parse and log errors in handler layer, here i use **ParseGRPCErrStatusCode** method, which parse err and returns matched gRPC code.<br/>
[Validator](https://github.com/go-playground/validator) is good solution for validation.<br/>

```go
user, err := u.registerReqToUserModel(r)
	if err != nil {
		u.logger.Errorf("registerReqToUserModel: %v", err)
		return nil, status.Errorf(grpc_errors.ParseGRPCErrStatusCode(err), "registerReqToUserModel: %v", err)
	}

	if err := utils.ValidateStruct(ctx, user); err != nil {
		u.logger.Errorf("ValidateStruct: %v", err)
		return nil, status.Errorf(grpc_errors.ParseGRPCErrStatusCode(err), "ValidateStruct: %v", err)
	}
```

After request input validation call use case method which contains business logic and works with users repository:
```go
createdUser, err := u.userUC.Register(ctx, user)
	if err != nil {
		u.logger.Errorf("userUC.Register: %v", err)
		return nil, status.Errorf(grpc_errors.ParseGRPCErrStatusCode(err), "Register: %v", err)
	}
```

Inside **a.user.Register(ctx, user)** method we start new tracing span and call user repository methods:<br/>

```go
func (u *userUseCase) Register(ctx context.Context, user *models.User) (*models.User, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "UserUseCase.Register")
	defer span.Finish()

	existsUser, err := u.userPgRepo.FindByEmail(ctx, user.Email)
	if existsUser != nil || err == nil {
		return nil, grpc_errors.ErrEmailExists
	}

	return u.userPgRepo.Create(ctx, user)
}
```

In user repository ***Create*** method we again start new tracing span and run our query</br>
Important note here:</br>
Good practice is always wrap err with some additional information, it's will make debugging much easier in the future 👍</br>
On Repository and UseCase levels usually we don't log errors, only wrap with the message and returns, because we logging errors on top layer in handlers.👨‍💻</br>
So we don't need log the one error multiple time, already warped it with debug message whats went wrong.

```go
func (r *UserRepository) Create(ctx context.Context, user *models.User) (*models.User, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "UserRepository.Create")
	defer span.Finish()

	createdUser := &models.User{}
	if err := r.db.QueryRowxContext(
		ctx,
		createUserQuery,
		user.FirstName,
		user.LastName,
		user.Email,
		user.Password,
		user.Role,
		user.Avatar,
	).StructScan(createdUser); err != nil {
		return nil, errors.Wrap(err, "Create.QueryRowxContext")
	}

	return createdUser, nil
}
```


Finally, service handler must return response object generated by proto, here usually we need to create helpers for map our internal business logic models to response object for return it.<br/>

```go
return &userService.RegisterResponse{User: u.userModelToProto(createdUser)}, nil
```

Every app must be covered by tests, I didn't completely cover all code this one, but wrote some test of course.
For testing and mocking [testify](https://github.com/stretchr/testify) and [gomock](https://github.com/golang/mock) is very good tools.

Source code and list of all used tools u can find [here](https://github.com/AleksK1NG/Go-GRPC-Auth-Microservice) 👨‍💻 :)
I hope this article is usefully and helpfully, I'll be happy to receive any feedbacks or questions :)
