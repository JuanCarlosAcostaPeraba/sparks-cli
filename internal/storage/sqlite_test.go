package storage_test

import (
	"context"
	"errors"
	"testing"

	"github.com/JuanCarlosAcostaPeraba/sparks-cli/internal/model"
	"github.com/JuanCarlosAcostaPeraba/sparks-cli/internal/storage"
)

func TestStoreInitializesAndAddsSpark(t *testing.T) {
	store := newStore(t)

	spark, err := store.Add(context.Background(), "Create GoReleaser config", model.AddOptions{})
	if err != nil {
		t.Fatalf("Add returned error: %v", err)
	}
	if spark.ID == 0 {
		t.Fatal("expected spark ID to be set")
	}
	if spark.Title != "Create GoReleaser config" {
		t.Fatalf("unexpected title: %q", spark.Title)
	}
}

func TestStoreUpdatesSparkTitle(t *testing.T) {
	store := newStore(t)
	ctx := context.Background()
	spark, err := store.Add(ctx, "old title", model.AddOptions{})
	if err != nil {
		t.Fatal(err)
	}

	updated, err := store.UpdateTitle(ctx, spark.ID, "  new title  ")
	if err != nil {
		t.Fatal(err)
	}
	if updated.Title != "new title" {
		t.Fatalf("expected trimmed updated title, got %q", updated.Title)
	}
	if updated.ID != spark.ID || updated.CreatedAt != spark.CreatedAt {
		t.Fatalf("expected identity and creation date to be preserved, got %#v", updated)
	}
}

func TestStoreUpdateTitleRejectsMissingSpark(t *testing.T) {
	store := newStore(t)
	_, err := store.UpdateTitle(context.Background(), 999, "new title")
	if !errors.Is(err, storage.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestStoreListsActiveSparks(t *testing.T) {
	store := newStore(t)
	ctx := context.Background()
	first, err := store.Add(ctx, "active", model.AddOptions{})
	if err != nil {
		t.Fatal(err)
	}
	second, err := store.Add(ctx, "done", model.AddOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := store.MarkDone(ctx, second.ID); err != nil {
		t.Fatal(err)
	}

	active, err := store.List(ctx, model.ListOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if len(active) != 1 || active[0].ID != first.ID {
		t.Fatalf("expected only active spark, got %#v", active)
	}

	all, err := store.List(ctx, model.ListOptions{IncludeDone: true})
	if err != nil {
		t.Fatal(err)
	}
	if len(all) != 2 {
		t.Fatalf("expected active and done sparks, got %d", len(all))
	}
}

func TestStoreMarksDone(t *testing.T) {
	store := newStore(t)
	ctx := context.Background()
	spark, err := store.Add(ctx, "finish tests", model.AddOptions{})
	if err != nil {
		t.Fatal(err)
	}

	done, err := store.MarkDone(ctx, spark.ID)
	if err != nil {
		t.Fatal(err)
	}
	if !done.Done || done.CompletedAt == nil {
		t.Fatalf("expected done spark with completed_at, got %#v", done)
	}
}

func TestStoreTogglesImportant(t *testing.T) {
	store := newStore(t)
	ctx := context.Background()
	spark, err := store.Add(ctx, "publish tap", model.AddOptions{})
	if err != nil {
		t.Fatal(err)
	}

	important, err := store.ToggleImportant(ctx, spark.ID)
	if err != nil {
		t.Fatal(err)
	}
	if !important.Important {
		t.Fatal("expected spark to be important")
	}

	normal, err := store.ToggleImportant(ctx, spark.ID)
	if err != nil {
		t.Fatal(err)
	}
	if normal.Important {
		t.Fatal("expected spark to be unmarked")
	}
}

func TestStoreSoftRemovesSpark(t *testing.T) {
	store := newStore(t)
	ctx := context.Background()
	spark, err := store.Add(ctx, "remove me", model.AddOptions{})
	if err != nil {
		t.Fatal(err)
	}

	if err := store.Remove(ctx, spark.ID); err != nil {
		t.Fatal(err)
	}
	if _, err := store.Get(ctx, spark.ID); !errors.Is(err, storage.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestStoreSearchesSparks(t *testing.T) {
	store := newStore(t)
	ctx := context.Background()
	if _, err := store.Add(ctx, "Prepare Codex prompt", model.AddOptions{}); err != nil {
		t.Fatal(err)
	}
	if _, err := store.Add(ctx, "Publish Homebrew tap", model.AddOptions{}); err != nil {
		t.Fatal(err)
	}

	results, err := store.Search(ctx, "codex", model.ListOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 1 || results[0].Title != "Prepare Codex prompt" {
		t.Fatalf("unexpected search results: %#v", results)
	}
}

func TestStoreAddsChildSpark(t *testing.T) {
	store := newStore(t)
	ctx := context.Background()
	parent, err := store.Add(ctx, "parent", model.AddOptions{})
	if err != nil {
		t.Fatal(err)
	}

	child, err := store.Add(ctx, "child", model.AddOptions{ParentID: &parent.ID})
	if err != nil {
		t.Fatal(err)
	}
	if child.ParentID == nil || *child.ParentID != parent.ID {
		t.Fatalf("expected parent id %d, got %#v", parent.ID, child.ParentID)
	}
}

func TestStoreRejectsMissingParent(t *testing.T) {
	store := newStore(t)
	ctx := context.Background()
	parentID := int64(999)

	_, err := store.Add(ctx, "child", model.AddOptions{ParentID: &parentID})
	if !errors.Is(err, storage.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func newStore(t *testing.T) *storage.Store {
	t.Helper()
	store, err := storage.Open(t.TempDir() + "/sparks.db")
	if err != nil {
		t.Fatalf("Open returned error: %v", err)
	}
	t.Cleanup(func() {
		if err := store.Close(); err != nil {
			t.Fatalf("Close returned error: %v", err)
		}
	})
	return store
}
