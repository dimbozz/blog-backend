// service_test/post_publish_test.go
package service_test

import (
	"context"
	"testing"
	"time"

	"blog-backend/internal/config"
	"blog-backend/internal/model"
	"blog-backend/service"
)

func TestPostService_PublishLogic(t *testing.T) {
	memoryRepo := NewMemoryPostStorage()
	testConfig := &config.Config{}

	svc := service.NewPostService(memoryRepo, nil, testConfig)

	tests := []struct {
		name           string
		publishAtDelta time.Duration
		expectedStatus string
	}{
		{
			name:           "past_publish_at_auto_publishes",
			publishAtDelta: -1 * time.Hour,
			expectedStatus: "published",
		},
		{
			name:           "future_publish_at_stays_draft",
			publishAtDelta: 1 * time.Hour,
			expectedStatus: "draft",
		},
		{
			name:           "exactly_now_publishes",
			publishAtDelta: 0,
			expectedStatus: "published",
		},
	}

	ctx := context.Background()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			publishTime := time.Now().Add(tt.publishAtDelta)
			post := &model.Post{
				Title:     tt.name,
				Content:   "test content",
				Status:    "draft",
				PublishAt: &publishTime,
			}

			created, err := svc.CreatePost(ctx, 1, post)
			if err != nil {
				t.Fatalf("CreatePost failed: %v", err)
			}

			if created.Status != tt.expectedStatus {
				t.Errorf("expected status %q, got %q", tt.expectedStatus, created.Status)
			}
		})
	}
}

func TestPostService_ListOnlyPublished(t *testing.T) {
	memoryRepo := NewMemoryPostStorage()
	testConfig := &config.Config{}
	svc := service.NewPostService(memoryRepo, nil, testConfig)

	ctx := context.Background()

	// Создаем посты разных статусов
	pastTime := time.Now().Add(-1 * time.Hour)
	futureTime := time.Now().Add(1 * time.Hour)
	publishedPost := &model.Post{
		Title:     "Published Post",
		Status:    "published",
		PublishAt: &pastTime,
	}
	draftPost := &model.Post{
		Title:     "Draft Post",
		Status:    "draft",
		PublishAt: &futureTime,
	}

	svc.CreatePost(ctx, 1, publishedPost)
	svc.CreatePost(ctx, 1, draftPost)

	// ListPosts возвращает только published
	listPosts, total, err := svc.GetAllPosts(ctx, 10, 0)
	if err != nil {
		t.Fatalf("GetAllPosts failed: %v", err)
	}

	if total != 1 {
		t.Errorf("expected 1 published post, got total=%d", total)
	}
	if len(listPosts) != 1 || listPosts[0].Title != "Published Post" {
		t.Errorf("expected only published post, got: %v", listPosts)
	}
}
