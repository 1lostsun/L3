package usecase

import (
	"fmt"
	"github.com/1lostsun/L3/tree/main/L3_3/cmd/server/internal/entity"
	"github.com/1lostsun/L3/tree/main/L3_3/cmd/server/internal/repo"
	"sort"
	"strconv"
	"strings"
	"time"
)

type SortOrder string

const (
	SortByDateAsc  SortOrder = "asc"
	SortByDateDesc SortOrder = "desc"
)

type UseCase struct {
	repo *repo.Repository
}

func New(repo *repo.Repository) *UseCase {
	return &UseCase{repo: repo}
}

func (uc *UseCase) AddComment(req entity.Request) (string, error) {
	id, err := uc.generateID(req.Parent)
	if err != nil {
		return "", fmt.Errorf("failed to generate id: %v", err)
	}

	comment := &entity.Comment{
		ID:           id,
		Text:         req.Text,
		Date:         time.Now(),
		Parent:       req.Parent,
		CommentsTree: []*entity.Comment{},
	}

	message, err := uc.repo.Add(*comment)
	if err != nil {
		return "", fmt.Errorf("failed to add comment: %v", err)
	}

	return message, nil
}

func (uc *UseCase) generateID(parent *string) (string, error) {
	storage := uc.repo.GetAll()

	if parent == nil {
		maxID := 0
		for id := range storage {
			if !strings.Contains(id, ".") {
				intID, err := strconv.Atoi(id)
				if err != nil {
					return "", fmt.Errorf("failed to convert string to int: %w", err)
				}

				if intID > maxID {
					maxID = intID
				}
			}
		}

		return fmt.Sprintf("%d", maxID+1), nil
	}

	prefix := *parent + "."
	m := 0
	for id := range storage {
		if strings.HasPrefix(id, prefix) {
			rest := strings.TrimPrefix(id, prefix)
			if !strings.Contains(rest, ".") {
				intID, err := strconv.Atoi(rest)
				if err != nil {
					return "", fmt.Errorf("invalid child in: %w", err)
				}

				if intID > m {
					m = intID
				}
			}
		}
	}

	return fmt.Sprintf("%s%d", prefix, m+1), nil
}

func (uc *UseCase) GetCommentsTree(parent *string) ([]*entity.Comment, error) {
	if parent == nil {
		storage := uc.repo.GetByParent(nil)
		result := make([]*entity.Comment, 0, len(storage))
		for i := range storage {
			comment := &storage[i]
			if err := uc.buildTree(comment); err != nil {
				return nil, err
			}
			result = append(result, comment)
		}

		fmt.Println(result)
		return result, nil
	}

	root, ok := uc.repo.GetByID(*parent)
	if !ok {
		return nil, fmt.Errorf("comment with id %s not found", *parent)
	}

	if err := uc.buildTree(root); err != nil {
		return nil, err
	}

	return []*entity.Comment{root}, nil
}

func (uc *UseCase) buildTree(comment *entity.Comment) error {
	childs := uc.repo.GetByParent(&comment.ID)

	for i := range childs {
		child := &childs[i]
		if child.ID == comment.ID {
			continue
		}

		if err := uc.buildTree(child); err != nil {
			return err
		}
		comment.CommentsTree = append(comment.CommentsTree, child)
	}
	return nil
}

func (uc *UseCase) GetPagedComments(page, limit int, order SortOrder) ([]*entity.Comment, error) {
	roots, err := uc.GetCommentsTree(nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get comments tree: %v", err)
	}

	sortComments(roots, order)

	start := (page - 1) * limit
	if start > len(roots) {
		return []*entity.Comment{}, nil
	}

	end := start + limit
	if end > len(roots) {
		end = len(roots)
	}

	return roots[start:end], nil
}

func sortComments(comments []*entity.Comment, order SortOrder) {
	sort.Slice(comments, func(i, j int) bool {
		if order == SortByDateAsc {
			return comments[i].Date.Before(comments[j].Date)
		}
		return comments[i].Date.After(comments[j].Date)
	})

	for _, comment := range comments {
		if len(comment.CommentsTree) > 0 {
			sortComments(comment.CommentsTree, order)
		}
	}
}

func (uc *UseCase) DeleteCommentsTree(commentID string) error {
	childs, err := uc.GetCommentsTree(&commentID)
	if err != nil {
		return fmt.Errorf("failed to get comments tree: %v", err)
	}

	if err := uc.repo.Remove(childs); err != nil {
		return fmt.Errorf("failed to remove comments tree: %v", err)
	}

	return nil
}

func (uc *UseCase) SearchComments(text string) ([]*entity.Comment, error) {
	allComments := uc.repo.GetAll()
	result := make([]*entity.Comment, 0, len(allComments)/2)

	for _, comment := range allComments {
		if strings.Contains(strings.ToLower(comment.Text), strings.ToLower(text)) {
			c := comment
			result = append(result, &c)
		}
	}

	if len(result) == 0 {
		return nil, fmt.Errorf("no comments found")
	}

	return result, nil
}
