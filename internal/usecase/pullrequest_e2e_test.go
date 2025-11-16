package usecase_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	"gopr/internal/domain"
	"gopr/internal/repo/pg"
	"gopr/internal/repo/testhelpers"
	"gopr/internal/usecase"

	"github.com/jackc/pgx/v5/pgxpool"
	tc "github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

func setupPostgres(t *testing.T) (*pgxpool.Pool, func()) {
	ctx := context.Background()

	req := tc.ContainerRequest{
		Image:        "postgres:16.7",
		Env:          map[string]string{"POSTGRES_PASSWORD": "root", "POSTGRES_USER": "root", "POSTGRES_DB": "testdb"},
		ExposedPorts: []string{"5432/tcp"},
		WaitingFor:   wait.ForListeningPort("5432/tcp"),
	}

	container, err := tc.GenericContainer(ctx, tc.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	require.NoError(t, err)

	host, _ := container.Host(ctx)
	port, _ := container.MappedPort(ctx, "5432")

	dsn := "postgres://root:root@" + host + ":" + port.Port() + "/testdb?sslmode=disable"

	dbpool, err := pgxpool.New(ctx, dsn)
	require.NoError(t, err)

	err = testhelpers.RunMigrations(dsn)
	require.NoError(t, err)

	cleanup := func() {
		dbpool.Close()
		container.Terminate(ctx)
	}

	return dbpool, cleanup
}

func TestPullRequestFlow_E2E(t *testing.T) {
	db, cleanup := setupPostgres(t)
	defer cleanup()

	ctx := context.Background()

	userRepo := pg.NewUserRepo(db)
	teamRepo := pg.NewTeamRepo(db)
	prRepo := pg.NewPullRequestRepo(db)

	uc := usecase.NewPullRequest(prRepo, userRepo)

	team := &domain.Team{
		Id:   uuid.New().String(),
		Name: "devs",
	}
	require.NoError(t, teamRepo.Create(ctx, team))

	u1 := &domain.User{Id: "u1", Username: "alice", TeamId: team.Id, IsActive: true}
	u2 := &domain.User{Id: "u2", Username: "bob", TeamId: team.Id, IsActive: true}
	u3 := &domain.User{Id: "u3", Username: "charlie", TeamId: team.Id, IsActive: true}

	require.NoError(t, userRepo.Create(ctx, u1))
	require.NoError(t, userRepo.Create(ctx, u2))
	require.NoError(t, userRepo.Create(ctx, u3))

	input := &domain.CreatePullRequest{
		AuthorId: u1.Id,
		Name:     "Fix login bug",
	}

	pr, err := uc.Create(ctx, input)
	require.NoError(t, err)

	require.Equal(t, "Fix login bug", pr.PR.Name)
	require.Equal(t, domain.PullRequestStatusOpen, domain.PullRequestStatus(pr.PR.Status))
	require.Len(t, pr.Reviewers, 2)

	oldReviewer := pr.Reviewers[0]

	reassign := &domain.ReassignPullRequest{
		Id:            pr.PR.Id,
		OldReviewerId: oldReviewer,
	}

	reassignedPR, newReviewer, err := uc.Reassign(ctx, reassign)

	if errors.Is(err, usecase.ErrNoCandidate) {
		require.Len(t, reassignedPR.Reviewers, 1)
		require.Equal(t, pr.Reviewers[1], reassignedPR.Reviewers[0])
	} else {
		require.NoError(t, err)
		require.NotEqual(t, oldReviewer, newReviewer)
		require.Len(t, reassignedPR.Reviewers, 2)
	}

	merged, err := uc.Merge(ctx, &domain.MergePullRequest{Id: pr.PR.Id})
	require.NoError(t, err)
	require.Equal(t, domain.PullRequestStatusClosed, domain.PullRequestStatus(merged.PR.Status))
	require.NotNil(t, merged.PR.MergedAt)

	merged2, err := uc.Merge(ctx, &domain.MergePullRequest{Id: pr.PR.Id})
	require.NoError(t, err)
	require.Equal(t, domain.PullRequestStatusClosed, domain.PullRequestStatus(merged2.PR.Status))
}

func TestPullRequest_ReassignOnMergedPR(t *testing.T) {
	db, cleanup := setupPostgres(t)
	defer cleanup()

	ctx := context.Background()

	userRepo := pg.NewUserRepo(db)
	teamRepo := pg.NewTeamRepo(db)
	prRepo := pg.NewPullRequestRepo(db)

	uc := usecase.NewPullRequest(prRepo, userRepo)

	team := &domain.Team{
		Id:   uuid.New().String(),
		Name: "qa",
	}
	require.NoError(t, teamRepo.Create(ctx, team))

	u1 := &domain.User{Id: "a1", Username: "author", TeamId: team.Id, IsActive: true}
	u2 := &domain.User{Id: "r1", Username: "rev1", TeamId: team.Id, IsActive: true}
	u3 := &domain.User{Id: "r2", Username: "rev2", TeamId: team.Id, IsActive: true}

	require.NoError(t, userRepo.Create(ctx, u1))
	require.NoError(t, userRepo.Create(ctx, u2))
	require.NoError(t, userRepo.Create(ctx, u3))

	pr, err := uc.Create(ctx, &domain.CreatePullRequest{
		AuthorId: u1.Id,
		Name:     "Add logging",
	})
	require.NoError(t, err)
	require.Len(t, pr.Reviewers, 2)

	oldReviewer := pr.Reviewers[0]

	merged, err := uc.Merge(ctx, &domain.MergePullRequest{
		Id: pr.PR.Id,
	})
	require.NoError(t, err)
	require.Equal(t, domain.PullRequestStatusClosed, domain.PullRequestStatus(merged.PR.Status))

	_, _, err = uc.Reassign(ctx, &domain.ReassignPullRequest{
		Id:            pr.PR.Id,
		OldReviewerId: oldReviewer,
	})

	require.Error(t, err)
	require.True(t, errors.Is(err, usecase.ErrPRMerged), "expected PR_MERGED error")
}

func TestPullRequest_NoReviewers_E2E(t *testing.T) {
	db, cleanup := setupPostgres(t)
	defer cleanup()

	ctx := context.Background()

	userRepo := pg.NewUserRepo(db)
	teamRepo := pg.NewTeamRepo(db)
	prRepo := pg.NewPullRequestRepo(db)

	uc := usecase.NewPullRequest(prRepo, userRepo)

	team := &domain.Team{Id: uuid.NewString(), Name: "solo"}
	require.NoError(t, teamRepo.Create(ctx, team))

	u1 := &domain.User{
		Id:       "solo1",
		Username: "lonely",
		TeamId:   team.Id,
		IsActive: true,
	}
	require.NoError(t, userRepo.Create(ctx, u1))

	pr, err := uc.Create(ctx, &domain.CreatePullRequest{
		AuthorId: u1.Id,
		Name:     "Solo update",
	})

	require.NoError(t, err)
	require.Len(t, pr.Reviewers, 0, "должно быть 0 ревьюверов, т.к. автор один в команде")
	require.Equal(t, domain.PullRequestStatusOpen, domain.PullRequestStatus(pr.PR.Status))
}

func TestPullRequest_ReassignInactiveReviewer_E2E(t *testing.T) {
	db, cleanup := setupPostgres(t)
	defer cleanup()

	ctx := context.Background()

	userRepo := pg.NewUserRepo(db)
	teamRepo := pg.NewTeamRepo(db)
	prRepo := pg.NewPullRequestRepo(db)

	uc := usecase.NewPullRequest(prRepo, userRepo)

	// Команда
	team := &domain.Team{Id: uuid.NewString(), Name: "backend"}
	require.NoError(t, teamRepo.Create(ctx, team))

	// Автор + 3 ревьювера = всего 4 человека
	u1 := &domain.User{Id: "auth1", Username: "author", TeamId: team.Id, IsActive: true}
	u2 := &domain.User{Id: "rev1", Username: "rev1", TeamId: team.Id, IsActive: true}
	u3 := &domain.User{Id: "rev2", Username: "rev2", TeamId: team.Id, IsActive: true}
	u4 := &domain.User{Id: "rev3", Username: "rev3", TeamId: team.Id, IsActive: true}

	require.NoError(t, userRepo.Create(ctx, u1))
	require.NoError(t, userRepo.Create(ctx, u2))
	require.NoError(t, userRepo.Create(ctx, u3))
	require.NoError(t, userRepo.Create(ctx, u4))

	// Создаём PR — назначится только 2 ревьювера из 3
	pr, err := uc.Create(ctx, &domain.CreatePullRequest{
		AuthorId: u1.Id,
		Name:     "Add cache",
	})
	require.NoError(t, err)
	require.Len(t, pr.Reviewers, 2)

	old := pr.Reviewers[0]

	// Делаем старого ревьювера неактивным
	require.NoError(t, userRepo.UpdateIsActive(ctx, old, false))

	// Теперь должен быть кандидат (u4)
	_, newRev, err := uc.Reassign(ctx, &domain.ReassignPullRequest{
		Id:            pr.PR.Id,
		OldReviewerId: old,
	})

	require.NoError(t, err)
	require.NotEqual(t, old, newRev)
	require.NotEmpty(t, newRev)
}

func TestPullRequest_Reassign_NoCandidates_E2E(t *testing.T) {
	db, cleanup := setupPostgres(t)
	defer cleanup()

	ctx := context.Background()

	userRepo := pg.NewUserRepo(db)
	teamRepo := pg.NewTeamRepo(db)
	prRepo := pg.NewPullRequestRepo(db)

	uc := usecase.NewPullRequest(prRepo, userRepo)

	team := &domain.Team{Id: uuid.NewString(), Name: "tiny-team"}
	require.NoError(t, teamRepo.Create(ctx, team))

	u1 := &domain.User{Id: "auth2", Username: "author", TeamId: team.Id, IsActive: true}
	u2 := &domain.User{Id: "revA", Username: "revA", TeamId: team.Id, IsActive: true}

	require.NoError(t, userRepo.Create(ctx, u1))
	require.NoError(t, userRepo.Create(ctx, u2))

	pr, err := uc.Create(ctx, &domain.CreatePullRequest{
		AuthorId: u1.Id,
		Name:     "Small fix",
	})
	require.NoError(t, err)

	require.Len(t, pr.Reviewers, 1)

	old := pr.Reviewers[0]

	_, _, err = uc.Reassign(ctx, &domain.ReassignPullRequest{
		Id:            pr.PR.Id,
		OldReviewerId: old,
	})

	require.Error(t, err)
	require.ErrorIs(t, err, usecase.ErrNoCandidate)

	revs, err2 := prRepo.ListReviewers(ctx, pr.PR.Id)
	require.NoError(t, err2)
	require.Empty(t, revs, "после удаления без кандидатов не должно быть ревьюверов")
}
