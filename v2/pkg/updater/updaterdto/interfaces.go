package updaterdto

import (
	"context"
	"time"

	"github.com/Masterminds/semver"
	"github.com/joy-dx/gophorth/pkg/releaser/releaserdto"
)

type UpdaterInterface interface {
	CheckLatest(ctx context.Context) (releaserdto.ReleaseAsset, error)
	DownloadUpdate(ctx context.Context, link *releaserdto.ReleaseAsset) error
	Hydrate(ctx context.Context) error
	PerformUpdate(ctx context.Context) error
	PostInstallCleanup() error
	State() *UpdaterState
	Status() UpdateStatus
	UpdateLog() string
}

type CheckClientInterface interface {
	CheckUpdate(ctx context.Context, cfg *UpdaterConfig) (releaserdto.ReleaseAsset, error)
	GetRef() string
	GetVersionLink() (releaserdto.ReleaseAsset, error)
}

// UpdateClientInterface Common Methods used to
type UpdateClientInterface interface {
	ArtefactPath() string
	Changelog() string
	DownloadVersion(releaserdto.ReleaseAsset) (artefactPath string, fetchErr error)
	PublishedAt() *time.Time
	ReleaseURL() string
	Size() int64
	Version() *semver.Version
}

type VerificationMethodInterface interface {
	GetRef() string
	Verify(artefactPath string) error
}

type GenericConfig interface {
	GetRef() string
}
