package redis

import (
	"context"
	"os"
	"testing"
	"time"
)

func TestConnect(t *testing.T) {
	// переменные ОС для тестов
	os.Setenv("REDIS_HOST", "localhost")
	os.Setenv("REDIS_PORT", "6379")

	// Проверяем успешность соединения
	if err := Connect(); err != nil {
		t.Errorf("Connect() error = %v", err)
	}
	defer Close()

	if Client == nil {
		t.Error("Client should not be nil after successful connection")
	}

	// Тестируем операции редиса
	ctx := context.Background()
	key := "test_key"
	value := "test_value"

	if err := Set(ctx, key, value, time.Minute); err != nil {
		t.Errorf("Set() error = %v", err)
	}

	got, err := Get(ctx, key)
	if err != nil {
		t.Errorf("Get() error = %v", err)
	}
	if got != value {
		t.Errorf("Get() = %v, want %v", got, value)
	}

	if err := Delete(ctx, key); err != nil {
		t.Errorf("Delete() error = %v", err)
	}

	// Проверяем, что ключ был удалён (функция должна возвращать ошибку)
	_, err = Get(ctx, key)
	if err == nil {
		t.Error("Get() should return error for deleted key")
	}
}
