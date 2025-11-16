package rest

import (
	"context"
	"gopr/docs"
	"gopr/internal/gateways/rest/middlewares"
	"gopr/internal/gateways/rest/pullrequest"
	"gopr/internal/gateways/rest/team"
	"gopr/internal/gateways/rest/user"
	"gopr/internal/usecase"

	"github.com/gin-gonic/gin"
	swaggerfiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func setupRouter(ctx context.Context, r *gin.Engine, useCases usecase.Cases) {
	r.HandleMethodNotAllowed = true
	r.Use(middlewares.AllowOrigin())
	r.Use(middlewares.Logger(ctx))

	// Swagger
	docs.SwaggerInfo.BasePath = "/api/v1"
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerfiles.Handler))

	v1 := r.Group("/api/v1")

	user.Setup(v1, useCases)
	team.Setup(v1, useCases)
	pullrequest.Setup(v1, useCases)
}
