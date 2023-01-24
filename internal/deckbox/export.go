package deckbox

import (
	"encoding/csv"
	"fmt"
	"mtg-bulk-input/internal/data"
	"os"
	"strings"
)

type SelectedCard struct {
	Set      string
	Quantity int
	Number   string
	Foil     bool
}

func Export(path string, cards []SelectedCard, store data.Store) error {
	out := make([][]string, 1, len(cards)+1)

	// Set the header
	out[0] = make([]string, 5)

	out[0][0] = "Count"
	out[0][1] = "Name"
	out[0][2] = "Edition"
	out[0][3] = "Card Number"
	out[0][4] = "Foil"

	for _, sCard := range cards {
		row := make([]string, 5)

		card, ok := store.SetCards[sCard.Set][sCard.Number]
		if !ok {
			return fmt.Errorf("could not find card '%s' in set '%s'", sCard.Number, sCard.Set)
		}

		row[0] = fmt.Sprint(sCard.Quantity)
		row[1] = card.Name    // TODO: Handle two sided-cards?
		row[2] = card.SetName // TODO: Handle SetName mapping?
		row[3] = card.CollectorNumber

		foilS := ""
		if sCard.Foil {
			foilS = "foil"
		}

		row[4] = foilS

		out = append(out, row)
	}

	// Create file name
	csvPath := fmt.Sprintf("%s.csv", strings.TrimSuffix(path, ".json"))

	// Open file
	f, err := os.OpenFile(csvPath, os.O_CREATE|os.O_RDWR, 0755)
	if err != nil {
		return fmt.Errorf("failed to open output file: %w", err)
	}

	w := csv.NewWriter(f)

	err = w.WriteAll(out)
	if err != nil {
		return fmt.Errorf("failed to write CSV file: %w", err)
	}

	return nil
}
