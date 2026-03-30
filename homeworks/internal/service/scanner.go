package service

import (
	"context"
	"mini-asm/internal/model"
)

type Scanner interface {
	Scan(ctx context.Context, asset *model.Asset, job *model.ScanJob) (int, error)
}