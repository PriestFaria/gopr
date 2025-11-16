package user

import (
	"net/http"

	"gopr/internal/dto"
	"gopr/internal/usecase"

	"github.com/gin-gonic/gin"
)

type setActiveInput struct {
	UserID   string `json:"user_id"`
	IsActive bool   `json:"is_active"`
}

func Setup(v1 *gin.RouterGroup, cases usecase.Cases) {
	g := v1.Group("/users")

	g.POST("/setIsActive", setActive(cases.User))
}

// @Summary Установить флаг активности пользователя
// @Tags Users
// @Accept json
// @Produce json
// @Param body body setActiveInput true "User ID and active flag"
// @Success 200 {object} map[string]dto.User
// @Failure 404 {object} dto.ErrorResponse
// @Router /users/setIsActive [post]
func setActive(userCase *usecase.User) gin.HandlerFunc {
	return func(c *gin.Context) {
		var input setActiveInput
		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, dto.ErrorResponse{
				Error: dto.ErrorObject{
					Code:    "BAD_REQUEST",
					Message: "invalid json",
				},
			})
			return
		}

		user, teamName, err := userCase.SetActive(c, input.UserID, input.IsActive)
		if err != nil {
			c.JSON(http.StatusNotFound, dto.ErrorResponse{
				Error: dto.ErrorObject{
					Code:    "NOT_FOUND",
					Message: err.Error(),
				},
			})
			return
		}

		resp := dto.User{
			UserID:   user.Id,
			Username: user.Username,
			TeamName: teamName,
			IsActive: user.IsActive,
		}

		c.JSON(http.StatusOK, gin.H{"user": resp})
	}
}
