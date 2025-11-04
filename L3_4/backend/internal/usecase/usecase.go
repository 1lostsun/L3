package usecase

import (
	"backend/internal/entity"
	r "backend/internal/infra/cache/redis"
	"backend/internal/storage/MinIO"
	"context"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"github.com/wb-go/wbf/kafka"
	"mime"
	"net/http"
)

const (
	Pending    = "pending"
	Done       = "done"
	Processing = "processing"
	Failed     = "failed"
)

type UseCase struct {
	s *MinIO.Storage
	r *r.Cache
	k *kafka.Producer
}

func New(s *MinIO.Storage, r *r.Cache, k *kafka.Producer) *UseCase {
	return &UseCase{
		s: s,
		r: r,
		k: k,
	}
}

func (uc *UseCase) Upload(ctx context.Context, fileData []byte, operations *entity.Operations) (string, error) {
	imageID := uuid.New().String()
	ext := detectExt(fileData, ".jpg")
	inputPath := fmt.Sprintf("original/%s%s", imageID, ext)
	outputPath := fmt.Sprintf("processed/%s%s", imageID, ext)
	contentType := http.DetectContentType(fileData)

	task := entity.ImageTask{
		ImageID:    imageID,
		InputPath:  inputPath,
		OutputPath: outputPath,
		Operations: operations,
	}

	payload, err := json.Marshal(task)
	if err != nil {
		return "", fmt.Errorf("failed to marshal image task: %v", err)
	}

	imgStatus := entity.ImageStatus{
		Status:     Pending,
		OutputPath: outputPath,
	}
	status, err := json.Marshal(imgStatus)
	if err != nil {
		return "", fmt.Errorf("failed to marshal image status: %v", err)
	}

	if err := uc.r.Client.Set(ctx, imageID, status); err != nil {
		return "", fmt.Errorf("failed to set redis data: %v", err)
	}

	if err := uc.s.Upload(ctx, fileData, inputPath, contentType); err != nil {
		fmt.Println(uc.s)
		return "", fmt.Errorf("failed to upload image: %v", err)
	}

	if err := uc.k.Send(ctx, []byte(imageID), payload); err != nil {
		fmt.Println(uc.k.Writer)
		return "", fmt.Errorf("failed to send kafka message: %v", err)
	}

	return imageID, nil
}

func detectExt(fileData []byte, fallback string) string {
	contentType := http.DetectContentType(fileData)
	exts, _ := mime.ExtensionsByType(contentType)
	if len(exts) > 0 {
		return exts[0]
	}
	return fallback
}

func (uc *UseCase) Download(ctx context.Context, imageID string) ([]byte, string, error) {
	data, err := uc.r.Client.Get(ctx, imageID)
	if err != nil {
		return nil, "", fmt.Errorf("failed to fetch image: %v", err)
	}

	var imgStatus *entity.ImageStatus
	if err := json.Unmarshal([]byte(data), &imgStatus); err != nil {
		return nil, "", fmt.Errorf("failed to unmarshal image: %v", err)
	}

	switch imgStatus.Status {
	case Pending, Processing:
		return nil, imgStatus.Status, nil
	case Failed:
		return nil, imgStatus.Status, fmt.Errorf("image processing failed")
	case Done:
		if imgStatus.OutputPath == "" {
			return nil, "", fmt.Errorf("failed to fetch image: image %s has no output path", imageID)
		}
		img, err := uc.s.Download(ctx, imgStatus.OutputPath)
		if err != nil {
			return nil, imgStatus.Status, fmt.Errorf("failed to download image: %v", err)
		}
		return img, imgStatus.Status, nil
	default:
		return nil, imgStatus.Status, fmt.Errorf("unknown status: %v", imgStatus.Status)
	}
}

func (uc *UseCase) Delete(ctx context.Context, imageID string) error {
	data, err := uc.r.Client.Get(ctx, imageID)
	if err != nil {
		return fmt.Errorf("failed to get img status from redis: %v", err)
	}

	var imgStatus entity.ImageStatus
	if err := json.Unmarshal([]byte(data), &imgStatus); err != nil {
		return fmt.Errorf("failed to unmarshal image: %v", err)
	}

	if imgStatus.Status != Done {
		return fmt.Errorf("image %s is not done", imageID)
	}

	if err := uc.s.Delete(ctx, imgStatus.OutputPath); err != nil {
		return fmt.Errorf("failed to delete image: %v", err)
	}

	if err := uc.r.Client.Del(ctx, imageID); err != nil {
		return fmt.Errorf("failed to delete image from redis: %v", err)
	}

	return nil
}
