package usecase

import (
	"github.com/1lostsun/L3/internal/infra/messaging"
	"github.com/1lostsun/L3/internal/repo/pg"
)

type UseCase struct {
	r  *pg.Repo
	pb *messaging.Publisher
}

func New(r *pg.Repo, pb *messaging.Publisher) *UseCase {
	return &UseCase{
		r:  r,
		pb: pb,
	}
}
