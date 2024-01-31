package data

import (
	"mtg-bulk-input/internal/scryfall"
	"sort"
	"strings"
)

/*
	CardIndex is a fast data structure for getting cards by their names.
	It lower-cases the string values and strips any punctuation marks to make searching faster.
	It works by sorting together all names by the first three characters.
*/

type CardIndex struct {
	prefixIndex map[string][]scryfall.Card
}

func NewCardIndex() CardIndex {
	return CardIndex{
		prefixIndex: make(map[string][]scryfall.Card),
	}
}

func (ci CardIndex) Add(card scryfall.Card) {
	if !isLegal(card) {
		return
	}

	key := strings.ToLower(card.Name)

	var prefix string

	if len(key) > 3 {
		prefix = key[:3]
	} else {
		prefix = key
	}

	sharedPrefix, ok := ci.prefixIndex[prefix]
	if !ok {
		sharedPrefix = make([]scryfall.Card, 0)
	}

	sharedPrefix = append(sharedPrefix, card)
	ci.prefixIndex[prefix] = sharedPrefix
}

func (ci CardIndex) Sort() {
	for _, v := range ci.prefixIndex {
		v := v
		sort.Slice(v, func(i, j int) bool {
			return strings.Compare(v[i].Name, v[j].Name) == -1
		})
	}
}

func (ci CardIndex) Search(term string) []scryfall.Card {
	if len(term) < 3 {
		return []scryfall.Card{}
	}

	key := strings.ToLower(term)

	sharedPrefix, ok := ci.prefixIndex[key[:3]]
	if !ok {
		return []scryfall.Card{}
	}

	out := make([]scryfall.Card, 0)

	for _, c := range sharedPrefix {
		if strings.HasPrefix(strings.ToLower(c.Name), key) {
			out = append(out, c)
		}
	}

	return out
}

/*
isLegal checks if a 'card' is legal in at least one format
*/
func isLegal(card scryfall.Card) bool {
	return card.Legalities.Standard == "legal" ||
		card.Legalities.Future == "legal" ||
		card.Legalities.Historic == "legal" ||
		card.Legalities.Gladiator == "legal" ||
		card.Legalities.Pioneer == "legal" ||
		card.Legalities.Explorer == "legal" ||
		card.Legalities.Modern == "legal" ||
		card.Legalities.Legacy == "legal" ||
		card.Legalities.Pauper == "legal" ||
		card.Legalities.Vintage == "legal" ||
		card.Legalities.Penny == "legal" ||
		card.Legalities.Commander == "legal" ||
		card.Legalities.Brawl == "legal" ||
		card.Legalities.Historicbrawl == "legal" ||
		card.Legalities.Alchemy == "legal" ||
		card.Legalities.Paupercommander == "legal" ||
		card.Legalities.Duel == "legal" ||
		card.Legalities.Oldschool == "legal" ||
		card.Legalities.Premodern == "legal"
}
