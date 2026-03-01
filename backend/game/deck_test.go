package game

import (
	"testing"
)

func TestNewDeckSandCardCounts(t *testing.T) {
	d := newDeck(SuitSand, 0)

	if len(d.Cards) != 22 {
		t.Fatalf("expected 22 cards in a single-suit deck, got %d", len(d.Cards))
	}

	// Count by kind
	valueCount := 0
	impostorCount := 0
	sylopCount := 0
	for _, c := range d.Cards {
		switch c.Kind {
		case KindValue:
			valueCount++
		case KindImpostor:
			impostorCount++
		case KindSylop:
			sylopCount++
		default:
			t.Errorf("unexpected card kind: %s", c.Kind)
		}
	}

	if valueCount != 18 {
		t.Errorf("expected 18 value cards (6 values * 3 copies), got %d", valueCount)
	}
	if impostorCount != 2 {
		t.Errorf("expected 2 impostor cards, got %d", impostorCount)
	}
	if sylopCount != 2 {
		t.Errorf("expected 2 sylop cards, got %d", sylopCount)
	}
}

func TestNewDeckValueDistribution(t *testing.T) {
	d := newDeck(SuitBlood, 0)

	valueCounts := map[int]int{}
	for _, c := range d.Cards {
		if c.Kind == KindValue {
			valueCounts[c.Value]++
		}
	}

	for v := 1; v <= 6; v++ {
		if valueCounts[v] != 3 {
			t.Errorf("expected 3 copies of value %d, got %d", v, valueCounts[v])
		}
	}
}

func TestNewDeckSuitAssignment(t *testing.T) {
	d := newDeck(SuitSand, 0)
	for _, c := range d.Cards {
		if c.Suit != SuitSand {
			t.Errorf("expected all cards to be SuitSand, got %s", c.Suit)
		}
	}

	d2 := newDeck(SuitBlood, 22)
	for _, c := range d2.Cards {
		if c.Suit != SuitBlood {
			t.Errorf("expected all cards to be SuitBlood, got %s", c.Suit)
		}
	}
}

func TestNewDeckUniqueIDs(t *testing.T) {
	d := newDeck(SuitSand, 0)
	seen := map[int]bool{}
	for _, c := range d.Cards {
		if seen[c.ID] {
			t.Errorf("duplicate card ID: %d", c.ID)
		}
		seen[c.ID] = true
	}
}

func TestNewDeckStartIDOffset(t *testing.T) {
	d := newDeck(SuitBlood, 100)
	for _, c := range d.Cards {
		if c.ID < 100 {
			t.Errorf("expected card ID >= 100 with startID=100, got %d", c.ID)
		}
	}
}

func TestNewDecksProducesTwoDecks(t *testing.T) {
	sand, blood := NewDecks()

	if sand.Remaining() != 22 {
		t.Errorf("expected 22 sand cards, got %d", sand.Remaining())
	}
	if blood.Remaining() != 22 {
		t.Errorf("expected 22 blood cards, got %d", blood.Remaining())
	}

	// Verify suits
	for _, c := range sand.Cards {
		if c.Suit != SuitSand {
			t.Errorf("sand deck has card with suit %s", c.Suit)
		}
	}
	for _, c := range blood.Cards {
		if c.Suit != SuitBlood {
			t.Errorf("blood deck has card with suit %s", c.Suit)
		}
	}
}

func TestNewDecksIDsDoNotOverlap(t *testing.T) {
	sand, blood := NewDecks()
	ids := map[int]bool{}
	for _, c := range sand.Cards {
		ids[c.ID] = true
	}
	for _, c := range blood.Cards {
		if ids[c.ID] {
			t.Errorf("blood deck card ID %d overlaps with sand deck", c.ID)
		}
	}
}

func TestDeckShuffle(t *testing.T) {
	// Create two decks with the same content and shuffle one.
	// There's a vanishingly small chance they stay in order, but
	// running multiple shuffles should demonstrate randomness.
	d1 := newDeck(SuitSand, 0)
	d2 := newDeck(SuitSand, 0)

	d2.Shuffle()

	sameOrder := true
	for i := range d1.Cards {
		if d1.Cards[i].ID != d2.Cards[i].ID {
			sameOrder = false
			break
		}
	}

	// Run again if the first shuffle happened to match (extremely unlikely)
	if sameOrder {
		d2.Shuffle()
		sameOrder = true
		for i := range d1.Cards {
			if d1.Cards[i].ID != d2.Cards[i].ID {
				sameOrder = false
				break
			}
		}
		if sameOrder {
			t.Error("deck shuffle did not change card order after two attempts")
		}
	}
}

func TestDeckDraw(t *testing.T) {
	d := newDeck(SuitSand, 0)
	firstCard := d.Cards[0]

	card, ok := d.Draw()
	if !ok {
		t.Fatal("expected draw to succeed")
	}
	if card.ID != firstCard.ID {
		t.Errorf("expected drawn card ID %d, got %d", firstCard.ID, card.ID)
	}
	if d.Remaining() != 21 {
		t.Errorf("expected 21 remaining after draw, got %d", d.Remaining())
	}
}

func TestDeckDrawMultiple(t *testing.T) {
	d := newDeck(SuitSand, 0)
	total := d.Remaining()

	for i := 0; i < total; i++ {
		_, ok := d.Draw()
		if !ok {
			t.Fatalf("draw %d failed unexpectedly", i+1)
		}
	}

	if d.Remaining() != 0 {
		t.Errorf("expected 0 remaining after drawing all cards, got %d", d.Remaining())
	}
}

func TestDeckDrawFromEmpty(t *testing.T) {
	d := Deck{Cards: []Card{}}

	card, ok := d.Draw()
	if ok {
		t.Error("expected draw from empty deck to return false")
	}
	if card != (Card{}) {
		t.Error("expected zero-value Card from empty deck draw")
	}
}

func TestDeckRemaining(t *testing.T) {
	d := newDeck(SuitSand, 0)
	if d.Remaining() != 22 {
		t.Errorf("expected 22 remaining, got %d", d.Remaining())
	}

	d.Draw()
	if d.Remaining() != 21 {
		t.Errorf("expected 21 remaining after one draw, got %d", d.Remaining())
	}
}

func TestDeckImpostorAndSylopHaveZeroValue(t *testing.T) {
	d := newDeck(SuitSand, 0)
	for _, c := range d.Cards {
		if c.Kind == KindImpostor && c.Value != 0 {
			t.Errorf("impostor card should have value 0, got %d", c.Value)
		}
		if c.Kind == KindSylop && c.Value != 0 {
			t.Errorf("sylop card should have value 0, got %d", c.Value)
		}
	}
}
