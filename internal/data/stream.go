package data

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mtg-bulk-input/internal/scryfall"
	"strings"
)

func StreamJsonList(r io.Reader) (<-chan scryfall.Card, <-chan error) {
	chanCard := make(chan scryfall.Card, 100)
	chanErr := make(chan error, 1)

	go func() {
		scan := bufio.NewScanner(r)

		line := scan.Scan()
		if !line {
			chanErr <- fmt.Errorf("unable to scan first line from reader: %w", scan.Err())
			return
		}

		lineS := scan.Text()

		if !strings.HasPrefix(lineS, "[") {
			chanErr <- fmt.Errorf("first line on reader was not an opening list brace ('['), got: %s", lineS)
			return
		}

		// While more lines
		for scan.Scan() {
			lineB := scan.Bytes()

			if bytes.HasPrefix(lineB, []byte("]")) {
				break
			}

			lineB = bytes.TrimSuffix(bytes.TrimSpace(lineB), []byte(","))

			var card scryfall.Card

			err := json.Unmarshal(lineB, &card)
			if err != nil {
				chanErr <- fmt.Errorf("failed to unmarshal json from stream: %w", err)
				return
			}

			chanCard <- card
		}

		close(chanCard)
	}()

	return chanCard, chanErr
}
