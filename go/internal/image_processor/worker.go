package image_processor

import "context"

type Worker struct{}

func (r *Worker) Work(ctx context.Context, t Task) error {
	return ctx.Err()
}
