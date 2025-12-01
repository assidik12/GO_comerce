package route

import (
	"net/http"

	"github.com/assidik12/go-restfull-api/config"
	"github.com/assidik12/go-restfull-api/internal/delivery/http/handler"
	"github.com/assidik12/go-restfull-api/internal/delivery/http/middleware"
	"github.com/julienschmidt/httprouter"
)

func NewRouter(userHandler *handler.UserHandler, productHandler *handler.ProductHandler, transactionHandler *handler.TransactionHandler) *httprouter.Router {
	router := httprouter.New()
	middleware := middleware.NewAuthMiddleware(router)
	c := config.GetConfig()

	docsDir := "./docs/swagger"
	fileServer := http.FileServer(http.Dir(docsDir))

	// Grup rute untuk API v1
	router.POST("/api/v1/users/register", userHandler.Register)
	router.POST("/api/v1/users/login", userHandler.Login)

	router.GET("/api/v1/products", productHandler.GetAllProducts)
	router.GET("/api/v1/products/:id", productHandler.GetProductById)
	router.POST("/api/v1/products", middleware.Middleware("admin", productHandler.CreateProduct, c.JWTSecret))
	router.PUT("/api/v1/products/:id", middleware.Middleware("admin", productHandler.UpdateProduct, c.JWTSecret))
	router.DELETE("/api/v1/products/:id", middleware.Middleware("admin", productHandler.DeleteProduct, c.JWTSecret))

	router.GET("/api/v1/transactions", middleware.Middleware("user", transactionHandler.GetAllTransaction, c.JWTSecret))
	router.GET("/api/v1/transactions/:id", middleware.Middleware("user", transactionHandler.GetTransactionById, c.JWTSecret))
	router.POST("/api/v1/transactions", middleware.Middleware("user", transactionHandler.CreateTransaction, c.JWTSecret))

	router.Handler("GET", "/api/v1/docs/*filepath", http.StripPrefix("/api/v1/docs/", fileServer))

	return router
}
