package service

import (
	"sync"
)

type HistoryService interface {
	Add(url string)
	GetRecent() []string
}

type InMemoryHistory struct {
	urls  []string
	limit int
	mu    sync.RWMutex
}

func NewInMemoryHistory(limit int) *InMemoryHistory {
	return &InMemoryHistory{
		urls:  make([]string, 0, limit),
		limit: limit,
	}
}

// Add добавляет URL в начало списка и удаляет лишние
func (h *InMemoryHistory) Add(url string) {
	h.mu.Lock()         // Блокируем запись
	defer h.mu.Unlock() // Разблокируем при выходе

	// Добавляем новый элемент в начало (самый простой способ для "стека")
	// Или добавляем в конец, зависит от того, как хотите отображать.
	// Вариант: Новый в начало списка
	h.urls = append([]string{url}, h.urls...)

	// Если превысили лимит, обрезаем хвост
	if len(h.urls) > h.limit {
		h.urls = h.urls[:h.limit]
	}
}

// GetRecent возвращает копию списка
func (h *InMemoryHistory) GetRecent() []string {
	h.mu.RLock() // Блокируем только для чтения (разрешает параллельные чтения)
	defer h.mu.RUnlock()

	// Важно вернуть копию, чтобы избежать гонки, если вызывающий код захочет изменить слайс
	result := make([]string, len(h.urls))
	copy(result, h.urls)

	return result
}
