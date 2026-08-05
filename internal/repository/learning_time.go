package repository

import (
	"context"

	"github.com/1622359590/imaiplay/internal/domain"
)

type LearningTimeRepository interface {
	Record(
		ctx context.Context,
		report *domain.LearningTimeReport,
		studyDate string,
	) (recorded bool, err error)
}
