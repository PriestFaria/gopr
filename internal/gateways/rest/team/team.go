package team

import (
	"net/http"

	"gopr/internal/domain"
	"gopr/internal/dto"
	"gopr/internal/usecase"

	"github.com/gin-gonic/gin"
)

func Setup(v1 *gin.RouterGroup, cases usecase.Cases) {
	g := v1.Group("/team")

	g.POST("/add", addTeam(cases.Team))
	g.GET("/get", getTeam(cases.Team))
}

// @Summary Создать команду с участниками (создаёт/обновляет пользователей)
// @Tags Teams
// @Accept json
// @Produce json
// @Param team body domain.TeamAddInput true "Team object"
// @Success 201 {object} map[string]dto.Team
// @Failure 400 {object} dto.ErrorResponse
// @Router /team/add [post]
func addTeam(teamCase *usecase.Team) gin.HandlerFunc {
	return func(c *gin.Context) {
		input := &domain.TeamAddInput{}
		if err := c.ShouldBindJSON(input); err != nil {
			c.JSON(http.StatusBadRequest, dto.ErrorResponse{
				Error: dto.ErrorObject{
					Code:    "BAD_REQUEST",
					Message: "invalid json",
				},
			})
			return
		}

		domainTeam, err := teamCase.AddTeam(c, input)
		if err != nil {
			c.JSON(http.StatusBadRequest, dto.ErrorResponse{
				Error: dto.ErrorObject{
					Code:    "TEAM_EXISTS",
					Message: err.Error(),
				},
			})
			return
		}

		c.JSON(http.StatusCreated, gin.H{
			"team": convertTeam(domainTeam),
		})
	}
}

// @Summary Получить команду с участниками
// @Tags Teams
// @Produce json
// @Param team_name query string true "Team name"
// @Success 200 {object} dto.Team
// @Failure 404 {object} dto.ErrorResponse
// @Router /team/get [get]
func getTeam(teamCase *usecase.Team) gin.HandlerFunc {
	return func(c *gin.Context) {
		teamName := c.Query("team_name")
		if teamName == "" {
			c.JSON(http.StatusBadRequest, dto.ErrorResponse{
				Error: dto.ErrorObject{
					Code:    "BAD_REQUEST",
					Message: "team_name required",
				},
			})
			return
		}

		domainTeam, err := teamCase.GetTeam(c, teamName)
		if err != nil {
			c.JSON(http.StatusNotFound, dto.ErrorResponse{
				Error: dto.ErrorObject{
					Code:    "NOT_FOUND",
					Message: err.Error(),
				},
			})
			return
		}

		c.JSON(http.StatusOK, convertTeam(domainTeam))
	}
}

func convertTeam(t *domain.TeamWithMembers) dto.Team {
	members := make([]dto.TeamMember, 0, len(t.Members))
	for _, m := range t.Members {
		members = append(members, dto.TeamMember{
			UserID:   m.Id,
			Username: m.Username,
			IsActive: m.IsActive,
		})
	}

	return dto.Team{
		TeamName: t.Team.Name,
		Members:  members,
	}
}
