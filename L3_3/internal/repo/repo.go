package repo

import (
	"fmt"
	"github.com/1lostsun/L3/tree/main/L3_3/cmd/server/internal/entity"
	"sync"
)

type Repository struct {
	storage map[string]entity.Comment
	lock    *sync.RWMutex
}

func New() *Repository {
	return &Repository{
		storage: make(map[string]entity.Comment),
		lock:    &sync.RWMutex{},
	}
}

func (repo *Repository) Add(comment entity.Comment) (string, error) {
	repo.lock.Lock()
	defer repo.lock.Unlock()

	if _, exists := repo.storage[comment.ID]; exists {
		return "", fmt.Errorf("comment already exists")
	}

	repo.storage[comment.ID] = comment
	return fmt.Sprintf("Comment successfully created: %v", comment.ID), nil
}

func (repo *Repository) GetAll() map[string]entity.Comment {
	repo.lock.RLock()
	defer repo.lock.RUnlock()

	c := make(map[string]entity.Comment, len(repo.storage))
	for k, v := range repo.storage {
		c[k] = v
	}
	return c
}

func (repo *Repository) GetByParent(parent *string) []entity.Comment {
	repo.lock.RLock()
	defer repo.lock.RUnlock()

	var result []entity.Comment
	for _, v := range repo.storage {
		if sameParent(parent, v.Parent) {
			result = append(result, v)
		}
	}

	return result
}

func (repo *Repository) GetByID(id string) (*entity.Comment, bool) {
	repo.lock.RLock()
	defer repo.lock.RUnlock()

	comment, ok := repo.storage[id]
	if !ok {
		return nil, false
	}
	return &comment, true
}

func sameParent(a, b *string) bool {
	if a == nil && b == nil {
		return true
	}
	if a != nil && b != nil && *a == *b {
		return true
	}
	return false
}

func (repo *Repository) Remove(comments []*entity.Comment) error {
	for _, comment := range comments {
		if _, exists := repo.storage[comment.ID]; !exists {
			return fmt.Errorf("comment does not exist")
		}

		delete(repo.storage, comment.ID)
		if err := repo.Remove(comment.CommentsTree); err != nil {
			return fmt.Errorf("failed remove comments tree: %v", err)
		}
	}

	return nil
}
