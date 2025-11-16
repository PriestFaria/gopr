package pullrequest

import (
	"errors"
	"net/http"
	"time"

	"gopr/internal/domain"
	"gopr/internal/dto"
	"gopr/internal/usecase"

	"github.com/gin-gonic/gin"
)

func Setup(v1 *gin.RouterGroup, cases usecase.Cases) {
	g := v1.Group("/pullRequest")

	g.POST("/create", addPR(cases.PullRequest))
	g.POST("/merge", mergePR(cases.PullRequest))
	g.POST("/reassign", reassignPR(cases.PullRequest))
}

// @Summary Создать PR и автоматически назначить до 2 ревьюверов из команды автора
// @Tags PullRequests
// @Accept json
// @Produce json
// @Param pr body domain.CreatePullRequest true "PR create payload"
// @Success 201 {object} map[string]dto.PullRequest
// @Failure 404 {object} dto.ErrorResponse
// @Failure 409 {object} dto.ErrorResponse
// @Router /pullRequest/create [post]
func addPR(prCase *usecase.PullRequest) gin.HandlerFunc {
	return func(c *gin.Context) {
		input := &domain.CreatePullRequest{}
		if err := c.ShouldBindJSON(input); err != nil {
			c.JSON(http.StatusBadRequest, dto.ErrorResponse{
				Error: dto.ErrorObject{
					Code:    "BAD_REQUEST",
					Message: "invalid json",
				},
			})
			return
		}

		res, err := prCase.Create(c, input)
		if err != nil {
			// тут можно разрулить код ошибки, если будешь прокидывать из usecase
			c.JSON(http.StatusConflict, dto.ErrorResponse{
				Error: dto.ErrorObject{
					Code:    "PR_EXISTS",
					Message: err.Error(),
				},
			})
			return
		}

		c.JSON(http.StatusCreated, gin.H{"pr": convertPR(res)})
	}
}

// @Summary Пометить PR как MERGED (идемпотентная операция)
// @Tags PullRequests
// @Accept json
// @Produce json
// @Param pr body domain.MergePullRequest true "Merge request"
// @Success 200 {object} map[string]dto.PullRequest
// @Failure 404 {object} dto.ErrorResponse
// @Router /pullRequest/merge [post]
func mergePR(prCase *usecase.PullRequest) gin.HandlerFunc {
	return func(c *gin.Context) {
		input := &domain.MergePullRequest{}
		if err := c.ShouldBindJSON(input); err != nil {
			c.JSON(http.StatusBadRequest, dto.ErrorResponse{
				Error: dto.ErrorObject{
					Code:    "BAD_REQUEST",
					Message: "invalid json",
				},
			})
			return
		}

		res, err := prCase.Merge(c, input)
		if err != nil {
			c.JSON(http.StatusNotFound, dto.ErrorResponse{
				Error: dto.ErrorObject{
					Code:    "NOT_FOUND",
					Message: err.Error(),
				},
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{"pr": convertPR(res)})
	}
}

// @Summary Переназначить конкретного ревьювера на другого из его команды
// @Tags PullRequests
// @Accept json
// @Produce json
// @Param reassign body domain.ReassignPullRequest true "Reassign payload"
// @Success 200 {object} dto.PullRequestReassignResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 409 {object} dto.ErrorResponse
// @Router /pullRequest/reassign [post]
func reassignPR(prCase *usecase.PullRequest) gin.HandlerFunc {
	return func(c *gin.Context) {
		input := &domain.ReassignPullRequest{}
		if err := c.ShouldBindJSON(input); err != nil {
			c.JSON(http.StatusBadRequest, dto.ErrorResponse{
				Error: dto.ErrorObject{
					Code:    "BAD_REQUEST",
					Message: "invalid json",
				},
			})
			return
		}

		res, newReviewer, err := prCase.Reassign(c, input)
		if err != nil {
			code := "PR_ERROR"

			switch {
			case errors.Is(err, usecase.ErrPRMerged):
				code = "PR_MERGED"
			case errors.Is(err, usecase.ErrNotAssigned):
				code = "NOT_ASSIGNED"
			case errors.Is(err, usecase.ErrNoCandidate):
				code = "NO_CANDIDATE"
			default:
				code = "REASSIGN_ERROR"
			}

			c.JSON(http.StatusConflict, dto.ErrorResponse{
				Error: dto.ErrorObject{
					Code:    code,
					Message: err.Error(),
				},
			})
			return
		}

		resp := dto.PullRequestReassignResponse{
			PR:         convertPR(res),
			ReplacedBy: newReviewer,
		}

		c.JSON(http.StatusOK, resp)
	}
}

func convertPR(p *domain.PullRequestWithReviewers) dto.PullRequest {
	var createdAt, mergedAt *string

	if !p.PR.CreatedAt.IsZero() {
		s := p.PR.CreatedAt.Format(time.RFC3339)
		createdAt = &s
	}

	if p.PR.MergedAt != nil {
		s := p.PR.MergedAt.Format(time.RFC3339)
		mergedAt = &s
	}

	return dto.PullRequest{
		PullRequestID:     p.PR.Id,
		PullRequestName:   p.PR.Name,
		AuthorID:          p.PR.AuthorId,
		Status:            p.PR.Status,
		AssignedReviewers: p.Reviewers,
		CreatedAt:         createdAt,
		MergedAt:          mergedAt,
	}
}
