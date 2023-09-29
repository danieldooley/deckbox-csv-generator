package ui

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"mtg-bulk-input/internal/data"
	"mtg-bulk-input/internal/deckbox"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
)

const (
	cardInputRegex = `^(?:([0-9a-z]{3})\.)?(\d+)(f?)$`
)

var (
	cardInputRegexEval = regexp.MustCompile(cardInputRegex)
)

func Start(filepath string, store data.Store) error {
	var selCards []deckbox.SelectedCard

	b, err := os.ReadFile(filepath)
	if errors.Is(err, os.ErrNotExist) {
		selCards = make([]deckbox.SelectedCard, 0)
	} else if err != nil {
		return fmt.Errorf("could not open file '%s': %w", filepath, err)
	} else {
		err = json.Unmarshal(b, &selCards)
		if err != nil {
			return fmt.Errorf("failed to parse file '%s' as json: %w", filepath, err)
		}
	}

	a := &app{
		filepath: filepath,

		store: store,

		autosaveTimer: time.NewTicker(time.Second * 10),

		selectedCards: selCards,
	}

	go func() {
		for {
			_ = <-a.autosaveTimer.C

			b, err := json.Marshal(a.selectedCards)
			if err != nil {
				panic(err) // TODO: Do I need to handle this?
			}

			err = os.WriteFile(filepath, b, 0755)
			if err != nil {
				panic(err)
			}
		}
	}()

	return a.start()
}

type app struct {
	store data.Store

	selectedSet string

	selectedCards []deckbox.SelectedCard

	autosaveTimer *time.Ticker

	filepath string
}

func (a *app) start() error {
	tviewApp := tview.NewApplication()

	/*
		Card Table
	*/
	cardsTable := tview.NewTable()
	cardsTable.SetSelectable(true, false)
	cardsTable.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		row, _ := cardsTable.GetSelection()

		switch event.Rune() {
		case '+':
			// Increment Quantity
			sCard := a.selectedCards[row-1]
			sCard.Quantity += 1
			a.selectedCards[row-1] = sCard
			return nil
		case '-':
			// Decrement Quantity

			sCard := a.selectedCards[row-1]
			sCard.Quantity -= 1

			if sCard.Quantity == 0 {
				cardsTable.RemoveRow(row)
			} else {
				a.selectedCards[row-1] = sCard
			}
			return nil

		case 'D':
			cardsTable.RemoveRow(row)
			return nil
		}

		return event
	})

	cardsTable.SetContent(a)

	tableFrame := tview.NewFrame(cardsTable)
	tableFrame.SetBorders(0, 0, 0, 1, 0, 0)
	tableFrame.SetBorder(true).SetTitle("Selected Cards")
	tableFrame.AddText("+: Increment Quantity - -: Decrement Quantity - D: Delete Row", false, tview.AlignCenter, tcell.ColorYellow)

	/*
		Card Input
	*/
	cardInput := tview.NewFlex()
	cardInput.SetBorder(true).SetTitle("Add Cards")

	cardField := tview.NewInputField()
	cardField.SetLabel("Add Card: ")

	cardField.SetDoneFunc(func(key tcell.Key) {
		if key == tcell.KeyEnter {
			s := strings.ToLower(strings.TrimSpace(cardField.GetText()))

			matches := cardInputRegexEval.FindStringSubmatch(s)

			if len(matches) == 4 {
				cardSet := matches[1]
				cardNum := matches[2]
				foil := matches[3] == "f"

				// Trim leading zeroes from the card number
				cardNum = strings.TrimLeft(cardNum, "0")

				// Prefer the set code from the card input
				selectedSet := a.selectedSet
				if cardSet != "" {
					selectedSet = cardSet
				}

				card, ok := a.store.SetCards[selectedSet][cardNum]
				if !ok {
					a.beep(1)
					return
				}

				if foil && !card.Foil || !foil && !card.Nonfoil {
					a.beep(1)
					return
				}

				var price string
				if foil {
					price = card.Prices.UsdFoil
				} else {
					price = card.Prices.Usd
				}

				// Notify if price above threshold
				priceF, err := strconv.ParseFloat(price, 64)
				if err == nil {
					if priceF > 10 {
						a.beep(3)
					} else if priceF > 2.5 {
						a.beep(2)
					}
				}

				// Find if card has already been added
				index := -1
				sCard := deckbox.SelectedCard{
					Set:      selectedSet,
					Quantity: 0,
					Number:   cardNum,
					Foil:     foil,
				}

				for i, c := range a.selectedCards {
					if c.Set == selectedSet && c.Number == cardNum && c.Foil == foil {
						sCard = c
						index = i
						break
					}
				}

				sCard.Quantity += 1

				if index == -1 {
					a.selectedCards = append(a.selectedCards, sCard)
					index = len(a.selectedCards) - 1
				} else {
					a.selectedCards[index] = sCard
				}

				cardsTable.Select(index+1, 0)

				cardField.SetText("")
			} else {
				a.beep(1)
			}
		}
	})

	cardInput.AddItem(cardField, 0, 1, true)

	/*
		Set Selector
	*/
	setSelector := tview.NewFlex()
	setSelector.SetBorder(true).SetTitle("Select Set")

	setField := tview.NewDropDown().
		SetLabel("Set Name: ")

	names := make([]string, 0, len(a.store.Sets))

	for name, _ := range a.store.Sets {
		names = append(names, name)
	}

	sort.Slice(names, func(i, j int) bool {
		return names[i] < names[j]
	})

	for _, name := range names {
		setField.AddOption(name, nil)
	}

	setField.SetSelectedFunc(func(text string, index int) {
		a.selectedSet = a.store.Sets[text]
		tviewApp.SetFocus(cardInput)
	})

	setSelector.AddItem(setField, 0, 1, true)

	/*
		Root
	*/

	root := tview.NewFlex()
	root.SetBorder(true).SetTitle("deckbox-csv-generator")

	root.SetDirection(tview.FlexRow)

	root.AddItem(tview.NewFlex().
		AddItem(setSelector, 0, 1, true).
		AddItem(cardInput, 0, 1, false),
		0, 1, true)

	root.AddItem(tableFrame, 0, 10, false)

	frame := tview.NewFrame(root).
		SetBorders(0, 0, 0, 1, 0, 0).
		AddText("S: Select Set - A: Add Cards - T: Selected Cards - X: Export", false, tview.AlignCenter, tcell.ColorYellow)

	root.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Rune() {
		case 'S':
			tviewApp.SetFocus(setSelector)
			return nil
		case 'A':
			tviewApp.SetFocus(cardInput)
			return nil
		case 'T':
			tviewApp.SetFocus(cardsTable)
			return nil
		case 'X':
			err := deckbox.Export(a.filepath, a.selectedCards, a.store)
			if err != nil {
				panic(err)
			}
			a.beep(2)
			return nil
		default:
			return event
		}
	})

	return tviewApp.SetRoot(frame, true).Run()
}

func (a *app) GetCell(row, column int) *tview.TableCell {
	if row == 0 { // Header row
		switch column {
		case 0:
			return tview.NewTableCell("Quantity").SetTextColor(tcell.ColorYellow)
		case 1:
			return tview.NewTableCell("Set").SetTextColor(tcell.ColorYellow)
		case 2:
			return tview.NewTableCell("Number").SetTextColor(tcell.ColorYellow)
		case 3:
			return tview.NewTableCell("Name").SetTextColor(tcell.ColorYellow)
		case 4:
			return tview.NewTableCell("Foil").SetTextColor(tcell.ColorYellow)
		case 5:
			return tview.NewTableCell("Price").SetTextColor(tcell.ColorYellow)
		default:
			return nil
		}
	} else {
		sCard := a.selectedCards[row-1]
		card := a.store.SetCards[sCard.Set][sCard.Number]

		price := ""

		if card.Foil {
			price = card.Prices.UsdFoil
		} else {
			price = card.Prices.Usd
		} // TODO: Etched price?

		color := tcell.ColorWhite

		priceF, err := strconv.ParseFloat(price, 64)
		if err == nil {
			if priceF > 10 {
				color = tcell.ColorRed
			} else if priceF > 2.5 {
				color = tcell.ColorOrange
			}
		}

		switch column {
		case 0:
			return tview.NewTableCell(fmt.Sprint(sCard.Quantity)).SetTextColor(color)
		case 1:
			return tview.NewTableCell(card.Set).SetTextColor(color)
		case 2:
			return tview.NewTableCell(card.CollectorNumber).SetTextColor(color)
		case 3:
			return tview.NewTableCell(card.Name).SetTextColor(color)
		case 4:
			return tview.NewTableCell(fmt.Sprint(sCard.Foil)).SetTextColor(color)
		case 5:

			return tview.NewTableCell(price).SetTextColor(color)
		default:
			return nil
		}
	}
}

func (a *app) GetRowCount() int {
	return len(a.selectedCards) + 1 // 1 row per card plus a header row
}

func (a *app) GetColumnCount() int {
	/*
		Columns are:
		* Quantity - int
		* Set Code - string
		* Card Number - string
		* Card Name - string
		* Foil - bool
		* Price - int
	*/
	return 6
}

func (a *app) SetCell(row, column int, cell *tview.TableCell) {
	// Not a function
}

func (a *app) RemoveRow(row int) {
	if row > 0 { // Can't remove header
		a.selectedCards = append(a.selectedCards[:row-1], a.selectedCards[row:]...)
	}
}

func (a *app) RemoveColumn(column int) {
	// Not a function
}

func (a *app) InsertRow(row int) {
	// Not a function
}

func (a *app) InsertColumn(column int) {
	// Not a function
}

func (a *app) Clear() {
	// Not a function // TODO: Or is it?
}

func (a *app) beep(n int) {
	go func() {
		for i := 0; i < n; i++ {
			fmt.Print("\a")
			time.Sleep(time.Millisecond * 200)
		}
	}()
}
