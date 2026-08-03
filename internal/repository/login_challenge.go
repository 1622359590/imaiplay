package repository

import (
	"context"
	"time"

	"github.com/1622359590/imaiplay/internal/domain"
)

type LoginChallengeRepository interface {
	Create(context.Context, *domain.LoginChallenge) error
	Consume(
		context.Context,
		string,
		time.Time,
	) (*domain.LoginChallenge, error)
}
