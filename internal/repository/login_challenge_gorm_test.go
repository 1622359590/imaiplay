package repository

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/1622359590/imaiplay/internal/domain"
	"github.com/1622359590/imaiplay/internal/migration"
	"gorm.io/gorm"
)

func TestLoginChallengeConsumeIsSingleUse(t *testing.T) {
	database := openTestDatabase(t)
	if err := migration.AutoMigrate(database); err != nil {
		t.Fatal(err)
	}
	repo := NewLoginChallengeRepository(database)
	now := time.Now().UTC()
	challenge := &domain.LoginChallenge{
		TokenHash:        "single-use-hash",
		CandidateUserIDs: `["user-a","user-b"]`,
		ExpiresAt:        now.Add(time.Minute),
	}
	if err := repo.Create(context.Background(), challenge); err != nil {
		t.Fatal(err)
	}
	if challenge.ID == "" {
		t.Fatal("Create() did not generate a challenge ID")
	}

	consumed, err := repo.Consume(
		context.Background(),
		challenge.TokenHash,
		now,
	)
	if err != nil {
		t.Fatal(err)
	}
	if consumed.ID != challenge.ID || consumed.ConsumedAt == nil {
		t.Fatalf("consumed=%#v", consumed)
	}
	if _, err := repo.Consume(
		context.Background(),
		challenge.TokenHash,
		now,
	); !errors.Is(err, gorm.ErrRecordNotFound) {
		t.Fatalf("second consume error=%v", err)
	}
}

func TestLoginChallengeConsumeRejectsExpired(t *testing.T) {
	database := openTestDatabase(t)
	if err := migration.AutoMigrate(database); err != nil {
		t.Fatal(err)
	}
	repo := NewLoginChallengeRepository(database)
	now := time.Now().UTC()
	challenge := &domain.LoginChallenge{
		TokenHash:        "expired-hash",
		CandidateUserIDs: `["user-a"]`,
		ExpiresAt:        now.Add(-time.Second),
	}
	if err := repo.Create(context.Background(), challenge); err != nil {
		t.Fatal(err)
	}

	if _, err := repo.Consume(
		context.Background(),
		challenge.TokenHash,
		now,
	); !errors.Is(err, gorm.ErrRecordNotFound) {
		t.Fatalf("expired consume error=%v", err)
	}
}

func TestLoginChallengeConcurrentConsumeHasOneWinner(t *testing.T) {
	database := openTestDatabase(t)
	if err := migration.AutoMigrate(database); err != nil {
		t.Fatal(err)
	}
	repo := NewLoginChallengeRepository(database)
	now := time.Now().UTC()
	challenge := &domain.LoginChallenge{
		TokenHash:        "concurrent-hash",
		CandidateUserIDs: `["user-a"]`,
		ExpiresAt:        now.Add(time.Minute),
	}
	if err := repo.Create(context.Background(), challenge); err != nil {
		t.Fatal(err)
	}

	var successes atomic.Int32
	var wait sync.WaitGroup
	wait.Add(2)
	for range 2 {
		go func() {
			defer wait.Done()
			if _, err := repo.Consume(
				context.Background(),
				challenge.TokenHash,
				now,
			); err == nil {
				successes.Add(1)
			} else if !errors.Is(err, gorm.ErrRecordNotFound) {
				t.Errorf("consume error=%v", err)
			}
		}()
	}
	wait.Wait()
	if successes.Load() != 1 {
		t.Fatalf("successful consumes=%d want=1", successes.Load())
	}
}
