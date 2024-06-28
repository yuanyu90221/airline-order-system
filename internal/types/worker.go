package types

import "context"

type Worker interface {
	Run(ctx context.Context) error
}
