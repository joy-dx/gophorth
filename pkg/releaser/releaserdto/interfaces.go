package releaserdto

import (
	"context"
)

type UpdatesInterface interface {
	GenerateArtefacts(ctx context.Context) (ReleaseAsset, error)
	Hydrate(ctx context.Context) error
	State() *ReleaserState
}
