//go:build wireinject
// +build wireinject

package injector

import (
	"net/http"

	"github.com/assidik12/go-restfull-api/config"
	handler "github.com/assidik12/go-restfull-api/internal/delivery/http/handler"
	"github.com/assidik12/go-restfull-api/internal/delivery/http/middleware"
	"github.com/assidik12/go-restfull-api/internal/delivery/http/route"

	"github.com/assidik12/go-restfull-api/internal/event"
	"github.com/assidik12/go-restfull-api/internal/infrastructure"
	mysql "github.com/assidik12/go-restfull-api/internal/repository/mysql"
	redis "github.com/assidik12/go-restfull-api/internal/repository/redis"
	service "github.com/assidik12/go-restfull-api/internal/service"
	"github.com/go-playground/validator/v10"
	"github.com/google/wire"
	"github.com/julienschmidt/httprouter"
)

var validatorSet = wire.NewSet(
	validator.New,
	wire.Value([]validator.Option{}),
)

// Setup Kafka (Pastikan file infrastructure/kafka.go dan event/kafka_producer.go sudah dibuat)
var eventSet = wire.NewSet(
	infrastructure.NewKafkaWriter,
	event.NewKafkaProducer,
	wire.Bind(new(event.Producer), new(*event.KafkaProducer)),
)

var userSet = wire.NewSet(
	mysql.NewUserRepository,
	service.NewUserService,
	handler.NewUserHandler,
)

var productSet = wire.NewSet(
	mysql.NewProductRepository,
	service.NewProductService,
	handler.NewProductHandler,
)

var transactionSet = wire.NewSet(
	mysql.NewTransactionRepository,
	service.NewTransactionService,
	handler.NewTransactionHandler,
)

// Hapus parameter cache.Wrapper dari sini, biarkan Wire yang buat
func InitializedServer(cfg config.Config) (*http.Server, func(), error) {
	wire.Build(
		// 1. Infrastructure (Singletons)
		infrastructure.DatabaseConnection,
		infrastructure.RedisConnection, // Redis di-init sekali di sini

		// 2. Utils & Wrappers
		redis.NewWrapper,
		validatorSet,

		// 3. Feature Sets
		eventSet,
		userSet,
		productSet,
		transactionSet,

		// 4. HTTP & Routing
		route.NewRouter,
		wire.Bind(new(http.Handler), new(*httprouter.Router)),
		middleware.NewAuthMiddleware,
		config.NewServer,
	)
	return nil, nil, nil
}
