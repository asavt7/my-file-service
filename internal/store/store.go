package store

import (
	"github.com/asavt7/my-file-service/internal/model"
	"golang.org/x/net/context"
)

type Store interface {
	SaveFile(ctx context.Context, f model.FileToStore) (model.StoredFile, error)
	LoadFile(ctx context.Context, f model.FileToDownload) (model.LoadedFile, error)
}
