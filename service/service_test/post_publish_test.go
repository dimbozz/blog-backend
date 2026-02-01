// service_test/post_publish_test.go
package service_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"blog-backend/internal/config"
	"blog-backend/internal/model"
	"blog-backend/service"
)

func TestPostService_PublishLogic(t *testing.T) {
	memoryRepo := NewMemoryPostStorage()
	userRepo := NewMockUserRepo()
	testConfig := &config.Config{
		PostTickerDuration: 30 * time.Second,
	}

	svc := service.NewPostService(memoryRepo, userRepo, testConfig)

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
			// log.Printf("TEST: %s, PublishAt=%v, Now=%v", tt.name, publishTime, time.Now())
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

			// log.Printf("AFTER CreatePost: Status=%q, PublishAt=%v", created.Status, created.PublishAt)

			if created.Status != tt.expectedStatus {
				t.Errorf("expected status %q, got %q", tt.expectedStatus, created.Status)
			}
		})
	}
}

func TestGetAllPosts(t *testing.T) {
	memoryRepo := NewMemoryPostStorage()
	userRepo := NewMockUserRepo()
	testConfig := &config.Config{
		PostTickerDuration: 30 * time.Second,
	}
	svc := service.NewPostService(memoryRepo, userRepo, testConfig)

	ctx := context.Background()

	// без пагинации
	tests := []struct {
		name           string
		postsToCreate  int
		limit          int
		offset         int
		wantPostsCount int
		wantTotal      int
	}{
		{
			name:           "no_posts",
			postsToCreate:  0,
			limit:          10,
			offset:         0,
			wantPostsCount: 0,
			wantTotal:      0,
		},
		{
			name:           "one_post",
			postsToCreate:  1,
			limit:          10,
			offset:         0,
			wantPostsCount: 1,
			wantTotal:      1,
		},
		{
			name:           "three_posts_all",
			postsToCreate:  3,
			limit:          10,
			offset:         0,
			wantPostsCount: 3,
			wantTotal:      3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// ✅ Очищаем репозиторий между тестами
			memoryRepo = NewMemoryPostStorage()
			svc = service.NewPostService(memoryRepo, NewMockUserRepo(), testConfig)

			// Создаем нужное количество постов
			for i := 0; i < tt.postsToCreate; i++ {
				post := &model.Post{
					Title:   fmt.Sprintf("post-%d", i),
					Content: fmt.Sprintf("content-%d", i),
					Status:  "published",
				}
				_, err := svc.CreatePost(ctx, 1, post)
				if err != nil {
					t.Fatalf("CreatePost failed: %v", err)
				}
			}

			// Тестируем GetAllPosts
			gotPosts, gotTotal, err := svc.GetAllPosts(ctx, tt.limit, tt.offset)
			if err != nil {
				t.Fatalf("GetAllPosts() error = %v", err)
			}

			if len(gotPosts) != tt.wantPostsCount {
				t.Errorf("expected %d posts, got %d", tt.wantPostsCount, len(gotPosts))
			}

			if gotTotal != tt.wantTotal {
				t.Errorf("expected total %d, got %d", tt.wantTotal, gotTotal)
			}
		})
	}
}
