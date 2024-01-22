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

	"golang.design/x/clipboard"
)

const (
	cardInputRegex      = `^(?:([0-9a-z]{3})\.)?(\d+)(f?)$`
	moxfieldImportRegex = `^(\d+)\s+(.+)\s+\(([0-9A-Z]{3})\)\s(\d+)(\s\*F\*)?$`

	mainPageName        = "main"
	importModalPageName = "importModal"
)

var (
	cardInputRegexEval      = regexp.MustCompile(cardInputRegex)
	moxfieldImportRegexEval = regexp.MustCompile(moxfieldImportRegex)
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

type cardMatch struct {
	count   int
	cardSet string
	cardNum string
	foil    bool
}

type app struct {
	store data.Store

	selectedSet string

	selectedCards []deckbox.SelectedCard

	autosaveTimer *time.Ticker

	filepath string
}

func (a *app) start() error {
	err := clipboard.Init()
	if err != nil {
		return fmt.Errorf("failed to init clipboard: %w", err)
	}

	tviewApp := tview.NewApplication()
	pages := tview.NewPages()

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

				index, err := a.AddCard(cardMatch{
					count:   1,
					cardSet: cardSet,
					cardNum: cardNum,
					foil:    foil,
				})

				if err != nil {
					a.beep(1)
					return
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
		Import Modal
	*/

	importField := tview.NewTextArea()

	importFlex := tview.NewFlex().AddItem(importField, 0, 1, false)

	importFrame := tview.NewFrame(importFlex).
		SetBorders(0, 0, 0, 1, 0, 0).
		AddText("P: Paste - I: Import Moxfield - X: Cancel", false, tview.AlignCenter, tcell.ColorYellow)

	importFrame.SetBorder(true).SetTitle("Import")

	importFlex.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Rune() {
		case 'P':
			importField.SetText(string(clipboard.Read(clipboard.FmtText)), true)
			return nil
		case 'I':
			lines := strings.Split(strings.ReplaceAll(importField.GetText(), "\r\n", "\n"), "\n")

			cardsToAdd := make([]cardMatch, 0, len(lines))

			for i, l := range lines {
				match := moxfieldImportRegexEval.FindStringSubmatch(l)

				if len(match) == 5 || len(match) == 6 {
					count, err := strconv.Atoi(match[1])
					if err != nil {
						a.beep(1)
						errIdx := mapTextAreaCoord(importField, i, 0)
						importField.Select(errIdx, errIdx)
						return nil
					}
					// match[2] is card name which we don't need/want
					cardSet := match[3]
					cardNumber := match[4]
					foil := false

					if len(match) == 6 {
						foil = len(match[5]) > 0
					}

					cardsToAdd = append(cardsToAdd, cardMatch{
						count:   count,
						cardSet: cardSet,
						cardNum: cardNumber,
						foil:    foil,
					})
				} else {
					a.beep(1)
					errIdx := mapTextAreaCoord(importField, i, 0)
					importField.Select(errIdx, errIdx)
					return nil
				}
			}

			for _, cm := range cardsToAdd {
				_, err := a.AddCard(cm)
				if err != nil {
					fmt.Print(err)
					a.beep(1)
					break
				}
			}

			importField.SetText("", true)
			pages.HidePage(importModalPageName)
			tviewApp.SetFocus(cardsTable)
			cardsTable.Select(len(a.selectedCards), 0)

			return nil
		case 'X':
			pages.HidePage(importModalPageName)
			return nil
		default:
			return event
		}
	})

	importModal := a.Modal(importFrame, 5, 5)

	/*
		Main page
	*/

	mainFlex := tview.NewFlex()
	mainFlex.SetBorder(true).SetTitle("deckbox-csv-generator")

	mainFlex.SetDirection(tview.FlexRow)

	mainFlex.AddItem(tview.NewFlex().
		AddItem(setSelector, 0, 1, true).
		AddItem(cardInput, 0, 1, false),
		0, 1, true)

	mainFlex.AddItem(tableFrame, 0, 10, false)

	mainFrame := tview.NewFrame(mainFlex).
		SetBorders(0, 0, 0, 1, 0, 0).
		AddText("S: Select Set - A: Add Cards - T: Selected Cards - X: Export - I: Import", false, tview.AlignCenter, tcell.ColorYellow)

	mainFlex.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
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
		case 'I':
			pages.ShowPage(importModalPageName)
			tviewApp.SetFocus(importFrame)
			return nil
		default:
			return event
		}
	})

	/*
		Pages
	*/

	pages.AddPage(mainPageName, mainFrame, true, true)
	pages.AddPage(importModalPageName, importModal, true, false)

	return tviewApp.SetRoot(pages, true).Run()
}

func (a *app) AddCard(cm cardMatch) (int, error) {
	// Trim leading zeroes from the card number
	cardNum := strings.TrimLeft(cm.cardNum, "0")

	// Prefer the set code from the card input
	selectedSet := a.selectedSet
	if cm.cardSet != "" {
		selectedSet = strings.ToLower(cm.cardSet)
	}

	card, ok := a.store.SetCards[selectedSet][cardNum]
	if !ok {
		a.beep(1)
		return 0, fmt.Errorf("SetCards did not contain %s - %s", selectedSet, cardNum)
	}

	if cm.foil && !card.Foil || !cm.foil && !card.Nonfoil {
		a.beep(1)
		return 0, fmt.Errorf("foil value '%v' is not valid for %s - %s", cm.foil, selectedSet, cardNum)
	}

	var price string
	if cm.foil {
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
		Foil:     cm.foil,
	}

	for i, c := range a.selectedCards {
		if c.Set == selectedSet && c.Number == cardNum && c.Foil == cm.foil {
			sCard = c
			index = i
			break
		}
	}

	sCard.Quantity += cm.count

	if index == -1 {
		a.selectedCards = append(a.selectedCards, sCard)
		index = len(a.selectedCards) - 1
	} else {
		a.selectedCards[index] = sCard
	}

	return index, nil
}

/*
Util method to put a UI component in a modal
width/height params are a proprotion, where the borders are proprotion 1
*/
func (a *app) Modal(p tview.Primitive, width, height int) tview.Primitive {
	return tview.NewFlex().
		AddItem(nil, 0, 1, false).
		AddItem(
			tview.NewFlex().SetDirection(tview.FlexRow).
				AddItem(nil, 0, 1, false).
				AddItem(p, 0, width, false).
				AddItem(nil, 0, 1, false),
			0, height, false).
		AddItem(nil, 0, 1, false)
}

/*
Util Method that converts a row/col to an index for a TextArea
*/
func mapTextAreaCoord(ta *tview.TextArea, row, col int) int {
	lines := strings.Split(strings.ReplaceAll(ta.GetText(), "\r\n", "\n"), "\n")

	if row >= len(lines) {
		return -1
	}

	count := 0
	i := 0

	for i = 0; i < row; i++ {
		count += len(lines[i])
	}

	if col > len(lines[i]) {
		return -1
	}

	return count + col
}

/*
Methods that make app implement a Table
*/
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

		if sCard.Foil {
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
