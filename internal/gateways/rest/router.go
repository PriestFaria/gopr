package rest

import (
	"context"
	"gopr/internal/gateways/rest/middlewares"
	"gopr/internal/gateways/rest/pullrequest"
	"gopr/internal/gateways/rest/team"
	"gopr/internal/gateways/rest/user"
	"gopr/internal/usecase"

	"github.com/gin-gonic/gin"
	swaggerfiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"github.com/swaggo/swag/example/basic/docs"
)

func setupRouter(ctx context.Context, r *gin.Engine, useCases usecase.Cases) {
	r.HandleMethodNotAllowed = true
	r.Use(middlewares.AllowOrigin())
	r.Use(middlewares.Logger(ctx))

	v1 := r.Group("/api/v1")
	docs.SwaggerInfo.BasePath = "/api/v1"
	v1.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerfiles.Handler))

	user.Setup(v1, useCases)
	team.Setup(v1, useCases)
	pullrequest.Setup(v1, useCases)
}
