package usecase

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"time"

	"github.com/google/uuid"

	"gopr/internal/domain"
	"gopr/internal/repo"
)

var (
	ErrPRMerged    = errors.New("PR_MERGED")
	ErrNotAssigned = errors.New("NOT_ASSIGNED")
	ErrNoCandidate = errors.New("NO_CANDIDATE")
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

	if err := p.prRepo.Create(ctx, pr); err != nil {
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
		if err := p.prRepo.AddReviewer(ctx, pr.Id, r.Id); err != nil {
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

	if err := p.prRepo.UpdateStatusMerged(ctx, pr.Id); err != nil {
		return nil, fmt.Errorf("failed to merge: %w", err)
	}

	pr.Status = string(domain.PullRequestStatusClosed)
	now := time.Now()
	pr.MergedAt = &now
	pr.UpdatedAt = now

	revs, _ := p.prRepo.ListReviewers(ctx, pr.Id)

	return &domain.PullRequestWithReviewers{
		PR:        pr,
		Reviewers: revs,
	}, nil
}

func (p *PullRequest) Reassign(ctx context.Context, input *domain.ReassignPullRequest) (*domain.PullRequestWithReviewers, string, error) {
	pr, err := p.prRepo.GetByID(ctx, input.Id)
	if err != nil {
		return nil, "", fmt.Errorf("failed to load PR: %w", err)
	}

	if pr.Status == string(domain.PullRequestStatusClosed) {
		return nil, "", ErrPRMerged
	}

	current, err := p.prRepo.ListReviewers(ctx, pr.Id)
	if err != nil {
		return nil, "", fmt.Errorf("failed to load reviewers: %w", err)
	}

	found := false
	for _, r := range current {
		if r == input.OldReviewerId {
			found = true
			break
		}
	}
	if !found {
		return nil, "", ErrNotAssigned
	}

	oldUser, err := p.userRepo.GetByID(ctx, input.OldReviewerId)
	if err != nil {
		return nil, "", fmt.Errorf("failed to load old reviewer: %w", err)
	}

	candidates, err := p.userRepo.ListByTeam(ctx, oldUser.TeamId, true)
	if err != nil {
		return nil, "", fmt.Errorf("failed to load candidates: %w", err)
	}

	// build set of current reviewers
	currentSet := make(map[string]struct{})
	for _, r := range current {
		currentSet[r] = struct{}{}
	}

	filtered := make([]*domain.User, 0)
	for _, u := range candidates {
		if u.Id == input.OldReviewerId {
			continue
		}
		if u.Id == pr.AuthorId {
			continue
		}

		if _, exists := currentSet[u.Id]; !exists {
			filtered = append(filtered, u)
		}
	}

	// Если некого поставить вместо старого — просто удаляем
	if len(filtered) == 0 {
		if err := p.prRepo.RemoveReviewer(ctx, pr.Id, input.OldReviewerId); err != nil {
			return nil, "", fmt.Errorf("failed to remove reviewer: %w", err)
		}
		revs, _ := p.prRepo.ListReviewers(ctx, pr.Id)
		return &domain.PullRequestWithReviewers{
			PR:        pr,
			Reviewers: revs,
		}, "", ErrNoCandidate
	}

	newReviewer := filtered[rand.Intn(len(filtered))]
	newReviewerID := newReviewer.Id

	if err := p.prRepo.RemoveReviewer(ctx, pr.Id, input.OldReviewerId); err != nil {
		return nil, "", fmt.Errorf("failed to remove old reviewer: %w", err)
	}

	if err := p.prRepo.AddReviewer(ctx, pr.Id, newReviewerID); err != nil {
		return nil, "", fmt.Errorf("failed to add new reviewer: %w", err)
	}

	revs, _ := p.prRepo.ListReviewers(ctx, pr.Id)

	return &domain.PullRequestWithReviewers{
		PR:        pr,
		Reviewers: revs,
	}, newReviewerID, nil
}
