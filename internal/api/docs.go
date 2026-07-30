// @title ImaiPlay API
// @version 0.1.0
// @description Multi-tenant learning platform API.
// @host localhost:8080
// @BasePath /
// @schemes http
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
package api

// APIResponse is the common envelope returned by API handlers.
type APIResponse struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}
