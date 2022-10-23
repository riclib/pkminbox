package main

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/widget"
	"github.com/riclib/pkminbox/readwise"
	"golang.design/x/clipboard"
	"log"
	"os"
	"strings"
)

const maxMenuTextSize = 60

var a fyne.App

func main() {

	err := clipboard.Init()
	if err != nil {
		panic(err)
	}

	a = app.NewWithID("pt.liberato.pkminbox")
	token := a.Preferences().String("readwise.token")

	w := a.NewWindow("SysTray")

	api := readwise.NewApi(token)
	api.GetBooks(false)

	desk, ok := a.(desktop.App)
	if !ok {
		fyne.LogError("Not running on desktop", nil)
		os.Exit(1)
	}

	m := buildMenu(desk, &api)
	desk.SetSystemTrayMenu(m)

	w.SetContent(widget.NewLabel("Fyne System Tray"))
	w.SetCloseIntercept(func() {
		w.Hide()
	})
	a.Run()
}

func menuClicked(api *readwise.ReadwiseAPI, h readwise.Highlight) {

}
func buildMenu(desk desktop.App, api *readwise.ReadwiseAPI) *fyne.Menu {
	var items []*fyne.MenuItem

	api.GetBooks(true)
	highlights := api.GetLatestHighlights(10)
	books := make(map[int]bool)

	if api.Token != "" {
		for _, highlight := range highlights {
			menuText := highlight.HighlightedAt.Format("15:04") + " - " + highlight.Text
			if len(menuText) > maxMenuTextSize {
				menuText = menuText[:maxMenuTextSize]
			}
			high := highlight
			item := fyne.NewMenuItem(menuText, func() {
				clipboard.Write(clipboard.FmtText, FormatHighlightAsTana(api, high))
			})
			items = append(items, item)
			books[highlight.BookId] = true
		}
		items = append(items, fyne.NewMenuItemSeparator())

		for id, _ := range books {
			bookId := id
			menuText := api.Books[bookId].Title
			if len(menuText) > maxMenuTextSize {
				menuText = menuText[:maxMenuTextSize]
			}

			item := fyne.NewMenuItem(menuText, func() {
				clipboard.Write(clipboard.FmtText, FormatBookAsTana(api, bookId))
			})
			items = append(items, item)
		}
	}

	items = append(items, fyne.NewMenuItemSeparator())
	enterTokenMenu := fyne.NewMenuItem("Enter Token", func() {
		enterToken(api)
	})
	refreshMenu := fyne.NewMenuItem("Refresh", func() {
		desk.SetSystemTrayMenu(buildMenu(desk, api))
	})
	items = append(items, enterTokenMenu)
	items = append(items, refreshMenu)

	m := fyne.NewMenu("Readwise", items...)
	return m
}

func FormatBookAsTana(api *readwise.ReadwiseAPI, bookId int) (b []byte) {
	var text string
	book, found := api.Books[bookId]
	if !found {
		log.Print("Didn't find book", "id", bookId)
	}
	highlights := api.GeHighlightsOfBook(bookId)

	text = text + "- [[" + book.Title + "]] #readwise-highlight\n"
	text = text + "  - type:: " + book.Category + "\n"
	if strings.Contains(book.SourceUrl, "http") {
		text = text + "  - source:: " + book.SourceUrl + "\n"
	}
	if len(book.Author) > 0 {
		text = text + "  - author:: " + book.SourceUrl + "\n"
	}
	for _, h := range highlights {
		text = text + "  - " + h.Text + "\n"
	}
	return TanaPaste(text)
}

func FormatHighlightAsTana(api *readwise.ReadwiseAPI, h readwise.Highlight) (b []byte) {
	book, found := api.Books[h.BookId]
	if !found {
		log.Print("Didn't find book", "id", h.BookId)
	}
	text := "- [[" + book.Title + "]]\n" +
		"  - " + h.Text + "\n"
	return TanaPaste(text)
}

func TanaPaste(text string) []byte {
	return []byte("%%tana%%\n" + text)
}

func enterToken(api *readwise.ReadwiseAPI) {
	api.Token = ""
	var formItems []*widget.FormItem
	window := a.NewWindow("hello")
	size := fyne.Size{
		Width:  500,
		Height: 500,
	}
	window.Resize(size)
	window.Show()
	tokenField := widget.NewEntry()
	tokenFormField := widget.NewFormItem("token", tokenField)
	formItems = append(formItems, tokenFormField)
	dialog.ShowForm("Enter Readwise Token", "Ok", "Cancel", formItems, func(confirm bool) {
		window.Hide()
		if confirm {
			api.Token = tokenField.Text
			a.Preferences().SetString("readwise.token", tokenField.Text)
			desk := a.(desktop.App)
			desk.SetSystemTrayMenu(buildMenu(desk, api))
		}
	}, window)

}
