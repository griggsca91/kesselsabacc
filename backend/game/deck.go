package game

import "math/rand"

type CardSuit string
type CardKind string

const (
	SuitSand  CardSuit = "sand"
	SuitBlood CardSuit = "blood"

	KindValue    CardKind = "value"
	KindImpostor CardKind = "impostor"
	KindSylop    CardKind = "sylop"
)

type Card struct {
	Suit CardSuit `json:"suit"`
	Kind CardKind `json:"kind"`
	// Value is 1-6 for value cards, 0 for sylop/impostor (resolved later)
	Value int `json:"value"`
	// ID uniquely identifies this card instance in the deck
	ID int `json:"id"`
}

type Deck struct {
	Cards []Card
}

func newDeck(suit CardSuit, startID int) Deck {
	cards := []Card{}
	id := startID

	// Value cards: 1-6, 3 copies each
	for v := 1; v <= 6; v++ {
		for range 3 {
			cards = append(cards, Card{Suit: suit, Kind: KindValue, Value: v, ID: id})
			id++
		}
	}

	// Impostor cards: 2 copies
	for range 2 {
		cards = append(cards, Card{Suit: suit, Kind: KindImpostor, Value: 0, ID: id})
		id++
	}

	// Sylop cards: 2 copies
	for range 2 {
		cards = append(cards, Card{Suit: suit, Kind: KindSylop, Value: 0, ID: id})
		id++
	}

	return Deck{Cards: cards}
}

func (d *Deck) Shuffle() {
	rand.Shuffle(len(d.Cards), func(i, j int) {
		d.Cards[i], d.Cards[j] = d.Cards[j], d.Cards[i]
	})
}

func (d *Deck) Draw() (Card, bool) {
	if len(d.Cards) == 0 {
		return Card{}, false
	}
	card := d.Cards[0]
	d.Cards = d.Cards[1:]
	return card, true
}

func (d *Deck) Remaining() int {
	return len(d.Cards)
}

func NewDecks() (sand Deck, blood Deck) {
	sand = newDeck(SuitSand, 0)
	blood = newDeck(SuitBlood, 22)
	sand.Shuffle()
	blood.Shuffle()
	return
}
