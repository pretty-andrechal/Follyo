package storage

import (
	"testing"

	"github.com/pretty-andrechal/follyo/internal/models"
)

func TestGetItems(t *testing.T) {
	t.Run("returns copy of slice", func(t *testing.T) {
		original := []models.Holding{
			{ID: "1", Coin: "BTC", Amount: 1.0},
			{ID: "2", Coin: "ETH", Amount: 10.0},
		}

		result := getItems(original)

		// Verify same length and content
		if len(result) != len(original) {
			t.Errorf("got length %d, want %d", len(result), len(original))
		}

		// Modify the result
		result[0].Amount = 999.0

		// Original should be unchanged
		if original[0].Amount != 1.0 {
			t.Errorf("original was modified, got amount %f, want 1.0", original[0].Amount)
		}
	})

	t.Run("empty slice returns empty slice", func(t *testing.T) {
		var original []models.Loan
		result := getItems(original)

		if result == nil {
			t.Error("expected non-nil slice")
		}
		if len(result) != 0 {
			t.Errorf("got length %d, want 0", len(result))
		}
	})

	t.Run("works with different entity types", func(t *testing.T) {
		// Test with Sales
		sales := []models.Sale{
			{ID: "s1", Coin: "BTC", Amount: 0.5},
		}
		salesResult := getItems(sales)
		if len(salesResult) != 1 || salesResult[0].ID != "s1" {
			t.Error("getItems failed for Sales")
		}

		// Test with Stakes
		stakes := []models.Stake{
			{ID: "st1", Coin: "ETH", Amount: 10.0},
		}
		stakesResult := getItems(stakes)
		if len(stakesResult) != 1 || stakesResult[0].ID != "st1" {
			t.Error("getItems failed for Stakes")
		}
	})
}

func TestRemoveByID(t *testing.T) {
	t.Run("removes existing item", func(t *testing.T) {
		items := []models.Holding{
			{ID: "1", Coin: "BTC", Amount: 1.0},
			{ID: "2", Coin: "ETH", Amount: 10.0},
			{ID: "3", Coin: "SOL", Amount: 100.0},
		}

		result, removed := removeByID(items, "2")

		if !removed {
			t.Error("expected removed to be true")
		}
		if len(result) != 2 {
			t.Errorf("got length %d, want 2", len(result))
		}

		// Verify correct item was removed
		for _, item := range result {
			if item.ID == "2" {
				t.Error("item with ID '2' should have been removed")
			}
		}
	})

	t.Run("returns false for non-existent ID", func(t *testing.T) {
		items := []models.Loan{
			{ID: "1", Coin: "BTC"},
			{ID: "2", Coin: "ETH"},
		}

		result, removed := removeByID(items, "999")

		if removed {
			t.Error("expected removed to be false for non-existent ID")
		}
		if len(result) != 2 {
			t.Errorf("got length %d, want 2", len(result))
		}
	})

	t.Run("handles empty slice", func(t *testing.T) {
		var items []models.Sale

		result, removed := removeByID(items, "1")

		if removed {
			t.Error("expected removed to be false for empty slice")
		}
		if len(result) != 0 {
			t.Errorf("got length %d, want 0", len(result))
		}
	})

	t.Run("removes first item", func(t *testing.T) {
		items := []models.Stake{
			{ID: "first", Coin: "BTC"},
			{ID: "second", Coin: "ETH"},
		}

		result, removed := removeByID(items, "first")

		if !removed {
			t.Error("expected removed to be true")
		}
		if len(result) != 1 {
			t.Errorf("got length %d, want 1", len(result))
		}
		if result[0].ID != "second" {
			t.Error("wrong item remained")
		}
	})

	t.Run("removes last item", func(t *testing.T) {
		items := []models.Holding{
			{ID: "first", Coin: "BTC"},
			{ID: "last", Coin: "ETH"},
		}

		result, removed := removeByID(items, "last")

		if !removed {
			t.Error("expected removed to be true")
		}
		if len(result) != 1 {
			t.Errorf("got length %d, want 1", len(result))
		}
		if result[0].ID != "first" {
			t.Error("wrong item remained")
		}
	})

	t.Run("removes only item", func(t *testing.T) {
		items := []models.Loan{
			{ID: "only", Coin: "BTC"},
		}

		result, removed := removeByID(items, "only")

		if !removed {
			t.Error("expected removed to be true")
		}
		if len(result) != 0 {
			t.Errorf("got length %d, want 0", len(result))
		}
	})

	t.Run("does not modify original slice", func(t *testing.T) {
		items := []models.Sale{
			{ID: "1", Coin: "BTC"},
			{ID: "2", Coin: "ETH"},
		}
		originalLen := len(items)

		removeByID(items, "1")

		if len(items) != originalLen {
			t.Error("original slice was modified")
		}
	})
}

func TestEntityInterface(t *testing.T) {
	// Verify all model types implement Entity interface
	var _ models.Entity = models.Holding{}
	var _ models.Entity = models.Loan{}
	var _ models.Entity = models.Sale{}
	var _ models.Entity = models.Stake{}

	t.Run("Holding implements Entity", func(t *testing.T) {
		h := models.Holding{ID: "test-id"}
		if h.GetID() != "test-id" {
			t.Errorf("got %q, want %q", h.GetID(), "test-id")
		}
	})

	t.Run("Loan implements Entity", func(t *testing.T) {
		l := models.Loan{ID: "loan-id"}
		if l.GetID() != "loan-id" {
			t.Errorf("got %q, want %q", l.GetID(), "loan-id")
		}
	})

	t.Run("Sale implements Entity", func(t *testing.T) {
		s := models.Sale{ID: "sale-id"}
		if s.GetID() != "sale-id" {
			t.Errorf("got %q, want %q", s.GetID(), "sale-id")
		}
	})

	t.Run("Stake implements Entity", func(t *testing.T) {
		st := models.Stake{ID: "stake-id"}
		if st.GetID() != "stake-id" {
			t.Errorf("got %q, want %q", st.GetID(), "stake-id")
		}
	})
}
