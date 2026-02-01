// service_test/in_memory_map_test.go
package service_test

import (
	"context"
	"testing"
	"time"

	"blog-backend/internal/config"
	"blog-backend/internal/model"
	"blog-backend/service"
)

func TestPublishLogic_WithMemoryStorage(t *testing.T) {
	// Тестируем полную цепочку: service → MemoryPostStorage
	memoryRepo := NewMemoryPostStorage()
	testConfig := &config.Config{}
	svc := service.NewPostService(memoryRepo, nil, testConfig)

	// Пост с publish_at в прошлом → должен стать published
	pastTime := time.Now().Add(-2 * time.Hour)
	post := &model.Post{
		Title:     "Memory Storage Test",
		Content:   "content",
		Status:    "draft",
		PublishAt: &pastTime, // прошлое время
	}

	ctx := context.Background()
	created, err := svc.CreatePost(ctx, 1, post)
	if err != nil {
		t.Fatalf("CreatePost failed: %v", err)
	}

	if created.Status != "published" {
		t.Errorf("expected published status from service, got %s", created.Status)
	}

	// Проверяем что в хранилище тоже published
	storedPost, err := svc.GetPost(ctx, created.ID)
	if err != nil {
		t.Fatalf("GetPost failed: %v", err)
	}
	if storedPost.Status != "published" {
		t.Errorf("expected stored post to be published, got %s", storedPost.Status)
	}
}

func TestMemoryStorage_Isolation(t *testing.T) {
	// Каждый тест имеет изолированное хранилище
	repo1 := NewMemoryPostStorage()
	repo2 := NewMemoryPostStorage()
	testConfig := &config.Config{}

	svc1 := service.NewPostService(repo1, nil, testConfig)
	svc2 := service.NewPostService(repo2, nil, testConfig)

	ctx := context.Background()

	// Создаем пост в первом хранилище
	svc1.CreatePost(ctx, 1, &model.Post{Title: "First Service"})

	// Проверяем изоляцию — во втором хранилище поста нет
	list1, _, _ := svc1.GetAllPosts(ctx, 10, 0)
	list2, _, _ := svc2.GetAllPosts(ctx, 10, 0)

	if len(list1) != 1 {
		t.Errorf("expected 1 post in first service, got %d", len(list1))
	}
	if len(list2) != 0 {
		t.Errorf("expected 0 posts in second service, got %d", len(list2))
	}
}
