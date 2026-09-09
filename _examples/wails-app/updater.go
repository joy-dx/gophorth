package main

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/joy-dx/gophorth/pkg/releaser/releaserdto"
	"github.com/joy-dx/gophorth/pkg/updater/updaterdto"
	"github.com/joy-dx/relay/dto"
)

// UpdaterInterface struct
type UpdaterInterface struct {
	ctx        context.Context
	relay      dto.RelayInterface
	updaterSvc updaterdto.UpdaterInterface
}

// SetContext Allows for attaching Wails context for proper shutdown
func (i *UpdaterInterface) SetContext(ctx context.Context) {
	i.ctx = ctx
}

func (i *UpdaterInterface) UpdaterStateGet() *updaterdto.UpdaterState {
	return i.updaterSvc.State()
}

func (a *UpdaterInterface) CheckForUpdate() (releaserdto.ReleaseAsset, error) {
	return a.updaterSvc.CheckLatest(a.ctx)
}

func (a *UpdaterInterface) StartUpdate() error {
	if a.updaterSvc.Status() != updaterdto.UPDATE_AVAILABLE {
		return errors.New("no update available")
	}
	if downloadErr := a.updaterSvc.DownloadUpdate(a.ctx, nil); downloadErr != nil {
		return fmt.Errorf("problem downloading latest version: %w", downloadErr)
	}
	if updateErr := a.updaterSvc.PerformUpdate(a.ctx); updateErr != nil {
		return fmt.Errorf("problem performing update: %w", updateErr)
	}

	// Exit the app for update
	os.Exit(0)

	return nil
}
