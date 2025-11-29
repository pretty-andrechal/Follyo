package storage

import "github.com/pretty-andrechal/follyo/internal/models"

// getItems is a generic helper that returns a copy of a slice.
// This prevents external modification of the underlying data.
func getItems[T any](items []T) []T {
	result := make([]T, len(items))
	copy(result, items)
	return result
}

// removeByID is a generic helper that removes an item by ID from a slice.
// Returns the filtered slice and whether an item was removed.
func removeByID[T models.Entity](items []T, id string) ([]T, bool) {
	originalLen := len(items)
	filtered := make([]T, 0, len(items))
	for _, item := range items {
		if item.GetID() != id {
			filtered = append(filtered, item)
		}
	}
	return filtered, len(filtered) < originalLen
}
