package route

import (
	"net/http"

	"github.com/assidik12/catalyst/internal/delivery/http/handler"
	"github.com/assidik12/catalyst/internal/delivery/http/middleware"
	"github.com/julienschmidt/httprouter"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func NewRouter(
	userHandler *handler.UserHandler,
	productHandler *handler.ProductHandler,
	transactionHandler *handler.TransactionHandler,
	jwtSecret string,
) *httprouter.Router {
	router := httprouter.New()
	authMiddleware := middleware.NewAuthMiddleware(router)

	docsDir := "./docs/swagger"
	fileServer := http.FileServer(http.Dir(docsDir))

	// Helper to wrap handlers with tracing and metrics
	wrap := func(path string, h httprouter.Handle) httprouter.Handle {
		return middleware.TracingMiddleware(middleware.MetricsMiddleware(h), path)
	}

	// Expose Prometheus metrics endpoint
	router.Handler("GET", "/metrics", promhttp.Handler())

	// Grup rute untuk API v1
	router.POST("/api/v1/users/register", wrap("/api/v1/users/register", userHandler.Register))
	router.POST("/api/v1/users/login", wrap("/api/v1/users/login", userHandler.Login))

	router.GET("/api/v1/products", wrap("/api/v1/products", productHandler.GetAllProducts))
	router.GET("/api/v1/products/:id", wrap("/api/v1/products/:id", productHandler.GetProductById))
	router.POST("/api/v1/products", wrap("/api/v1/products", authMiddleware.Middleware("admin", productHandler.CreateProduct, jwtSecret)))
	router.PUT("/api/v1/products/:id", wrap("/api/v1/products/:id", authMiddleware.Middleware("admin", productHandler.UpdateProduct, jwtSecret)))
	router.DELETE("/api/v1/products/:id", wrap("/api/v1/products/:id", authMiddleware.Middleware("admin", productHandler.DeleteProduct, jwtSecret)))

	router.GET("/api/v1/transactions", wrap("/api/v1/transactions", authMiddleware.Middleware("user", transactionHandler.GetAllTransaction, jwtSecret)))
	router.GET("/api/v1/transactions/:id", wrap("/api/v1/transactions/:id", authMiddleware.Middleware("user", transactionHandler.GetTransactionById, jwtSecret)))
	router.POST("/api/v1/transactions", wrap("/api/v1/transactions", authMiddleware.Middleware("user", transactionHandler.CreateTransaction, jwtSecret)))

	router.Handler("GET", "/api/v1/docs/*filepath", http.StripPrefix("/api/v1/docs/", fileServer))

	return router
}
