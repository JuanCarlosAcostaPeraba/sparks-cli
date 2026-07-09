package app

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/JuanCarlosAcostaPeraba/sparks-cli/internal/model"
	"github.com/JuanCarlosAcostaPeraba/sparks-cli/internal/storage"
)

type Repository interface {
	Add(context.Context, string, model.AddOptions) (model.Spark, error)
	List(context.Context, model.ListOptions) ([]model.Spark, error)
	Search(context.Context, string, model.ListOptions) ([]model.Spark, error)
	MarkDone(context.Context, int64) (model.Spark, error)
	ToggleImportant(context.Context, int64) (model.Spark, error)
	Remove(context.Context, int64) error
	ClearCompleted(context.Context) (int64, error)
	ClearAll(context.Context) (int64, error)
}

type App struct {
	repo Repository
}

type AddOptions struct {
	Parent string
}

func New(repo Repository) *App {
	return &App{repo: repo}
}

func (a *App) Add(ctx context.Context, title string, opts AddOptions) (model.Spark, error) {
	if strings.TrimSpace(title) == "" {
		return model.Spark{}, errors.New("add a spark title, for example: sparks add \"ship v0.1.0\"")
	}

	addOpts := model.AddOptions{}
	if strings.TrimSpace(opts.Parent) != "" {
		parentID, err := ParseID(opts.Parent)
		if err != nil {
			return model.Spark{}, fmt.Errorf("invalid parent id: %w", err)
		}
		addOpts.ParentID = &parentID
	}

	return a.repo.Add(ctx, title, addOpts)
}

func (a *App) List(ctx context.Context, opts model.ListOptions) ([]model.Spark, error) {
	return a.repo.List(ctx, opts)
}

func (a *App) Search(ctx context.Context, query string, opts model.ListOptions) ([]model.Spark, error) {
	if strings.TrimSpace(query) == "" {
		return nil, errors.New("add a search query, for example: sparks search \"release\"")
	}
	return a.repo.Search(ctx, query, opts)
}

func (a *App) Done(ctx context.Context, rawID string) (model.Spark, error) {
	id, err := ParseID(rawID)
	if err != nil {
		return model.Spark{}, err
	}
	return a.repo.MarkDone(ctx, id)
}

func (a *App) Important(ctx context.Context, rawID string) (model.Spark, error) {
	id, err := ParseID(rawID)
	if err != nil {
		return model.Spark{}, err
	}
	return a.repo.ToggleImportant(ctx, id)
}

func (a *App) Remove(ctx context.Context, rawID string) error {
	id, err := ParseID(rawID)
	if err != nil {
		return err
	}
	return a.repo.Remove(ctx, id)
}

func (a *App) Clear(ctx context.Context, all bool) (int64, error) {
	if all {
		return a.repo.ClearAll(ctx)
	}
	return a.repo.ClearCompleted(ctx)
}

func ParseID(raw string) (int64, error) {
	id, err := strconv.ParseInt(strings.TrimSpace(raw), 10, 64)
	if err != nil || id <= 0 {
		return 0, fmt.Errorf("invalid spark id %q", raw)
	}
	return id, nil
}

func FriendlyError(err error) error {
	if errors.Is(err, storage.ErrNotFound) {
		return errors.New("spark not found")
	}
	return err
}
