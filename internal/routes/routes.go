package routes

import (
	"net/http"

	"github.com/kosttiik/pvz-service/internal/handlers"
	"github.com/kosttiik/pvz-service/internal/middleware"
)

func SetupRoutes() {
	http.HandleFunc("/ping", handlers.PingHandler)

	http.HandleFunc("/dummyLogin", handlers.DummyLoginHandler)
	http.HandleFunc("/register", handlers.RegisterHandler)
	http.HandleFunc("/login", handlers.LoginHandler)

	http.HandleFunc("/pvz", middleware.AuthMiddleware(
		middleware.RoleMiddleware("moderator")(handlers.CreatePVZHandler)),
	)

	// http.HandleFunc("/receptions", middleware.AuthMiddleware(
	// 	middleware.RoleMiddleware("employee")(handlers.CreateReceptionHandler)),
	// )
	// http.HandleFunc("/pvz/{pvzId}/close_last_reception", middleware.AuthMiddleware(
	// 	middleware.RoleMiddleware("employee")(handlers.CloseReceptionHandler)),
	// )

	// http.HandleFunc("/products", middleware.AuthMiddleware(
	// 	middleware.RoleMiddleware("employee")(handlers.AddProductHandler)),
	// )
	// http.HandleFunc("/pvz/{pvzId}/delete_last_product", middleware.AuthMiddleware(
	// 	middleware.RoleMiddleware("employee")(handlers.DeleteLastProductHandler)),
	// )
}
