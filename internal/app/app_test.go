package app_test

import (
	"context"
	"errors"
	"testing"

	"github.com/JuanCarlosAcostaPeraba/sparks-cli/internal/app"
	"github.com/JuanCarlosAcostaPeraba/sparks-cli/internal/model"
)

func TestParseID(t *testing.T) {
	id, err := app.ParseID("42")
	if err != nil {
		t.Fatalf("ParseID returned error: %v", err)
	}
	if id != 42 {
		t.Fatalf("expected 42, got %d", id)
	}
}

func TestParseIDRejectsInvalidValues(t *testing.T) {
	for _, raw := range []string{"", "abc", "0", "-1"} {
		if _, err := app.ParseID(raw); err == nil {
			t.Fatalf("expected error for %q", raw)
		}
	}
}

func TestAddParsesParentID(t *testing.T) {
	repo := &fakeRepo{}
	application := app.New(repo)

	if _, err := application.Add(context.Background(), "child", app.AddOptions{Parent: "42"}); err != nil {
		t.Fatalf("Add returned error: %v", err)
	}
	if repo.addOpts.ParentID == nil || *repo.addOpts.ParentID != 42 {
		t.Fatalf("expected parent id 42, got %#v", repo.addOpts.ParentID)
	}
}

func TestAddRejectsInvalidParentID(t *testing.T) {
	repo := &fakeRepo{}
	application := app.New(repo)

	if _, err := application.Add(context.Background(), "child", app.AddOptions{Parent: "abc"}); err == nil {
		t.Fatal("expected invalid parent id error")
	}
	if repo.addCalled {
		t.Fatal("expected repository not to be called")
	}
}

type fakeRepo struct {
	addCalled bool
	addOpts   model.AddOptions
}

func (f *fakeRepo) Add(_ context.Context, title string, opts model.AddOptions) (model.Spark, error) {
	f.addCalled = true
	f.addOpts = opts
	return model.Spark{ID: 1, Title: title, ParentID: opts.ParentID}, nil
}

func (f *fakeRepo) List(context.Context, model.ListOptions) ([]model.Spark, error) {
	return nil, errors.New("not implemented")
}

func (f *fakeRepo) Search(context.Context, string, model.ListOptions) ([]model.Spark, error) {
	return nil, errors.New("not implemented")
}

func (f *fakeRepo) MarkDone(context.Context, int64) (model.Spark, error) {
	return model.Spark{}, errors.New("not implemented")
}

func (f *fakeRepo) ToggleImportant(context.Context, int64) (model.Spark, error) {
	return model.Spark{}, errors.New("not implemented")
}

func (f *fakeRepo) Remove(context.Context, int64) error {
	return errors.New("not implemented")
}

func (f *fakeRepo) ClearCompleted(context.Context) (int64, error) {
	return 0, errors.New("not implemented")
}

func (f *fakeRepo) ClearAll(context.Context) (int64, error) {
	return 0, errors.New("not implemented")
}
