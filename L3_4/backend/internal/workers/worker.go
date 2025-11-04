package workers

import (
	"backend/internal/entity"
	"backend/internal/infra/cache/redis"
	"backend/internal/infra/queue/kafka/consumer"
	"backend/internal/storage/MinIO"
	"backend/internal/usecase"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/disintegration/imaging"
	"github.com/fogleman/gg"
	"golang.org/x/image/font/inconsolata"
	"image"
	"image/jpeg"
	"image/png"
	"net/http"
)

type Worker struct {
	consumer *consumer.Consumer
	storage  *MinIO.Storage
	r        *redis.Cache
}

func New(consumer *consumer.Consumer, storage *MinIO.Storage, redis *redis.Cache) *Worker {
	return &Worker{
		consumer: consumer,
		storage:  storage,
		r:        redis,
	}
}

func (w *Worker) Run(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			return nil
		default:
			message, err := w.consumer.Fetch(ctx)
			if ctx.Err() != nil {
				return nil
			}

			if err != nil {
				fmt.Println("error fetching message:", err)
				continue
			}
			var task entity.ImageTask
			if err := json.Unmarshal(message.Value, &task); err != nil {
				fmt.Println("error unmarshalling message:", err)
				continue
			}

			processing := entity.ImageStatus{Status: usecase.Processing, OutputPath: task.OutputPath}
			if b, err := json.Marshal(processing); err == nil {
				_ = w.r.Client.Set(ctx, task.ImageID, b)
			}

			fail := func(reason string, err error) error {
				failed := entity.ImageStatus{Status: usecase.Failed, OutputPath: task.OutputPath}
				if b, mErr := json.Marshal(failed); mErr == nil {
					_ = w.r.Client.Set(ctx, task.ImageID, b)
				}
				return fmt.Errorf("%s: %w", reason, err)
			}

			data, err := w.storage.Download(ctx, task.InputPath)
			if err != nil {
				fmt.Println(fail("error downloading image", err))
				continue
			}

			if task.Operations.Resize != nil {
				data, err = w.resize(data, task.Operations.Resize)
				if err != nil {
					fmt.Println(fail(fmt.Sprintf("error resizing image"), err))
					continue
				}
			}

			if task.Operations.WaterMark != nil {
				data, err = w.addWatermark(data, task.Operations.WaterMark.Text)
				if err != nil {
					fmt.Println(fail(fmt.Sprintf("error adding watermark"), err))
					continue
				}

			}

			if err := w.storage.Upload(ctx, data, task.OutputPath, http.DetectContentType(data)); err != nil {
				fmt.Println(fail("error uploading image", err))
				continue
			}

			done := entity.ImageStatus{Status: usecase.Done, OutputPath: task.OutputPath}
			if b, err := json.Marshal(done); err == nil {
				if err := w.r.Client.Set(ctx, task.ImageID, b); err != nil {
					fmt.Println("error updating redis state:", err)
				}
			}
		}
	}

}

func (w *Worker) resize(data []byte, sizes *entity.Resize) ([]byte, error) {
	img, format, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("error decoding image: %w", err)
	}

	resized := imaging.Resize(img, sizes.Width, sizes.Height, imaging.Lanczos)

	var buf bytes.Buffer
	switch format {
	case "png":
		err = png.Encode(&buf, resized)
	default:
		err = jpeg.Encode(&buf, resized, &jpeg.Options{Quality: 90})
	}

	if err != nil {
		return nil, fmt.Errorf("failed to encode resized image: %w", err)
	}

	return buf.Bytes(), nil
}

func (w *Worker) addWatermark(data []byte, text string) ([]byte, error) {
	img, format, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("error decoding image: %w", err)
	}

	dc := gg.NewContextForImage(img)
	dc.SetFontFace(inconsolata.Bold8x16)

	width := float64(dc.Width())
	height := float64(dc.Height())

	dc.SetRGBA(1, 1, 1, 0.25)

	stepX := 350.0
	stepY := 220.0

	for x := -width; x < 2*width; x += stepX {
		for y := -height; y < 2*height; y += stepY {
			dc.Push()
			dc.DrawStringAnchored(text, x, y, 0.5, 0.5)
			dc.Pop()
		}
	}

	finalImg := dc.Image()

	var buf bytes.Buffer
	switch format {
	case "png":
		err = png.Encode(&buf, finalImg)
	default:
		err = jpeg.Encode(&buf, finalImg, &jpeg.Options{Quality: 90})
	}

	if err != nil {
		return nil, fmt.Errorf("failed to encode resized image: %w", err)
	}

	return buf.Bytes(), nil
}
