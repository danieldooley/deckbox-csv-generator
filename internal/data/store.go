package data

import (
	"fmt"
	"mtg-bulk-input/internal/scryfall"
	"os"
	"path"
)

type Store struct {
	// Sets is keyed: Card.SetName and the value is Card.Set
	Sets map[string]string

	// SetCards is keyed: Card.Set -> Card.CollectorNumber
	SetCards map[string]map[string]scryfall.Card
}

func BuildDataStore() (Store, error) {
	out := Store{
		Sets:     make(map[string]string),
		SetCards: make(map[string]map[string]scryfall.Card),
	}

	p := path.Join(WorkingDirectory, BulkDataDirectory, "default_cards.json")

	f, err := os.Open(p)
	if err != nil {
		return out, fmt.Errorf("failed to open file '%s': %w", p, err)
	}

	chanCards, chanErr := StreamJsonList(f)

	moreCards := true

	for moreCards {
		select {
		case c, ok := <-chanCards:
			if ok {
				// Populate the mapping of set codes to names
				out.Sets[c.SetName] = c.Set

				// Put the card in the SetCards map
				set, ok := out.SetCards[c.Set]
				if !ok {
					set = make(map[string]scryfall.Card)
				}

				set[c.CollectorNumber] = c

				out.SetCards[c.Set] = set
			} else {
				moreCards = false
			}
		case err := <-chanErr:
			return out, fmt.Errorf("json stream failed: %w", err)
		}
	}

	return out, nil
}
