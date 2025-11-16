package usecase

import (
	context "context"
	"errors"
	"fmt"
	"math/rand"
	"time"

	"github.com/google/uuid"

	"gopr/internal/domain"
	"gopr/internal/repo"
)

type PullRequest struct {
	prRepo   repo.PullRequest
	userRepo repo.User
}

func NewPullRequest(prRepo repo.PullRequest, userRepo repo.User) *PullRequest {
	return &PullRequest{
		prRepo:   prRepo,
		userRepo: userRepo,
	}
}

func (p *PullRequest) Create(ctx context.Context, input *domain.CreatePullRequest) (*domain.PullRequestWithReviewers, error) {
	pr := &domain.PullRequest{
		Id:        uuid.NewString(),
		AuthorId:  input.AuthorId,
		Name:      input.Name,
		Status:    string(domain.PullRequestStatusOpen),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	author, err := p.userRepo.GetByID(ctx, pr.AuthorId)
	if err != nil {
		return nil, fmt.Errorf("failed to load author: %w", err)
	}

	err = p.prRepo.Create(ctx, pr)
	if err != nil {
		return nil, fmt.Errorf("failed to create PR: %w", err)
	}

	members, err := p.userRepo.ListByTeam(ctx, author.TeamId, true)
	if err != nil {
		return nil, fmt.Errorf("failed to list team members: %w", err)
	}

	candidates := make([]*domain.User, 0, len(members))
	for _, m := range members {
		if m.Id != author.Id {
			candidates = append(candidates, m)
		}
	}

	rand.Shuffle(len(candidates), func(i, j int) {
		candidates[i], candidates[j] = candidates[j], candidates[i]
	})

	limit := 2
	if len(candidates) < 2 {
		limit = len(candidates)
	}

	reviewers := make([]string, 0, limit)

	for i := 0; i < limit; i++ {
		r := candidates[i]
		err = p.prRepo.AddReviewer(ctx, pr.Id, r.Id)
		if err != nil {
			return nil, fmt.Errorf("failed to add reviewer: %w", err)
		}
		reviewers = append(reviewers, r.Id)
	}

	return &domain.PullRequestWithReviewers{
		PR:        pr,
		Reviewers: reviewers,
	}, nil
}

func (p *PullRequest) Merge(ctx context.Context, input *domain.MergePullRequest) (*domain.PullRequestWithReviewers, error) {
	pr, err := p.prRepo.GetByID(ctx, input.Id)
	if err != nil {
		return nil, fmt.Errorf("failed to get PR: %w", err)
	}

	if pr.Status == string(domain.PullRequestStatusClosed) {
		revs, _ := p.prRepo.ListReviewers(ctx, pr.Id)
		return &domain.PullRequestWithReviewers{
			PR:        pr,
			Reviewers: revs,
		}, nil
	}

	err = p.prRepo.UpdateStatusMerged(ctx, pr.Id)
	if err != nil {
		return nil, fmt.Errorf("failed to merge: %w", err)
	}

	pr.Status = string(domain.PullRequestStatusClosed)
	pr.MergedAt = time.Now()
	pr.UpdatedAt = time.Now()

	revs, _ := p.prRepo.ListReviewers(ctx, pr.Id)

	return &domain.PullRequestWithReviewers{
		PR:        pr,
		Reviewers: revs,
	}, nil
}

func (p *PullRequest) Reassign(ctx context.Context, input *domain.ReassignPullRequest) (*domain.PullRequestWithReviewers, error) {
	pr, err := p.prRepo.GetByID(ctx, input.Id)
	if err != nil {
		return nil, fmt.Errorf("failed to load PR: %w", err)
	}

	if pr.Status == string(domain.PullRequestStatusClosed) {
		return nil, errors.New("cannot reassign reviewer for merged PR")
	}

	current, err := p.prRepo.ListReviewers(ctx, pr.Id)
	if err != nil {
		return nil, fmt.Errorf("failed to load reviewers: %w", err)
	}

	found := false
	for _, r := range current {
		if r == input.OldReviewerId {
			found = true
			break
		}
	}
	if !found {
		return nil, errors.New("old reviewer is not assigned to this PR")
	}

	oldUser, err := p.userRepo.GetByID(ctx, input.OldReviewerId)
	if err != nil {
		return nil, fmt.Errorf("failed to load old reviewer: %w", err)
	}

	candidates, err := p.userRepo.ListByTeam(ctx, oldUser.TeamId, true)
	if err != nil {
		return nil, fmt.Errorf("failed to load candidates: %w", err)
	}

	filtered := make([]*domain.User, 0)
	for _, u := range candidates {
		if u.Id != input.OldReviewerId && u.Id != pr.AuthorId {
			filtered = append(filtered, u)
		}
	}

	if len(filtered) == 0 {
		err = p.prRepo.RemoveReviewer(ctx, pr.Id, input.OldReviewerId)
		if err != nil {
			return nil, fmt.Errorf("failed to remove reviewer: %w", err)
		}
	} else {
		newReviewer := filtered[rand.Intn(len(filtered))]

		err = p.prRepo.RemoveReviewer(ctx, pr.Id, input.OldReviewerId)
		if err != nil {
			return nil, fmt.Errorf("failed to remove reviewer: %w", err)
		}

		err = p.prRepo.AddReviewer(ctx, pr.Id, newReviewer.Id)
		if err != nil {
			return nil, fmt.Errorf("failed to add new reviewer: %w", err)
		}
	}

	revs, _ := p.prRepo.ListReviewers(ctx, pr.Id)

	return &domain.PullRequestWithReviewers{
		PR:        pr,
		Reviewers: revs,
	}, nil
}

func (p *PullRequest) ListByReviewer(ctx context.Context, userID string) (*domain.UserReviews, error) {
	prs, err := p.prRepo.ListByReviewer(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to list PRs by reviewer: %w", err)
	}

	return &domain.UserReviews{
		UserID: userID,
		PRs:    prs,
	}, nil
}
