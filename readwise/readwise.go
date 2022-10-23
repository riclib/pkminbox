package readwise

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math"
	"net/http"
	"time"
)

const (
	ReadwiseTimeFormat = time.RFC3339
	ReadwisePageSize   = 100
)

type ReadwiseAPI struct {
	Token             string
	lastBookSync      time.Time
	lastHighlightSync time.Time
	Books             map[int]Book
}

type ReadwiseBooks struct {
	Count    int         `json:"count"`
	Next     string      `json:"next"`
	Previous interface{} `json:"previous"`
	Results  []Book      `json:"results"`
}

type Book struct {
	Id              int           `json:"id"`
	Title           string        `json:"title"`
	Author          string        `json:"author"`
	Category        string        `json:"category"`
	NumHighlights   int           `json:"num_highlights"`
	LastHighlightAt time.Time     `json:"last_highlight_at"`
	Updated         time.Time     `json:"updated"`
	CoverImageUrl   string        `json:"cover_image_url"`
	HighlightsUrl   string        `json:"highlights_url"`
	SourceUrl       string        `json:"source_url"`
	Asin            string        `json:"asin"`
	Tags            []interface{} `json:"tags"`
	MemURL          string        `json:"mem_url"`
	MemId           string        `json:"mem_id"`
}

type ReadwiseHighlights struct {
	Count    int         `json:"count"`
	Next     string      `json:"next"`
	Previous interface{} `json:"previous"`
	Results  []Highlight `json:"results"`
}

type Highlight struct {
	Id            int       `json:"id"`
	Text          string    `json:"text"`
	Note          string    `json:"note"`
	Location      int       `json:"location"`
	LocationType  string    `json:"location_type"`
	HighlightedAt time.Time `json:"highlighted_at"`
	Url           string    `json:"url"`
	Color         string    `json:"color"`
	Updated       time.Time `json:"updated"`
	BookId        int       `json:"book_id"`
	Tags          []struct {
		Id   int    `json:"id"`
		Name string `json:"name"`
	} `json:"tags"`
}

func NewApi(token string) ReadwiseAPI {
	var api ReadwiseAPI
	api.Token = token
	api.Books = make(map[int]Book)
	return api
}

func (api *ReadwiseAPI) GetLatestHighlights(count int) []Highlight {
	return api.getHighlights(false, count, -1)
}

func (api *ReadwiseAPI) GeHighlightsOfBook(bookId int) []Highlight {
	return api.getHighlights(false, math.MaxInt, bookId)
}

func (api *ReadwiseAPI) getHighlights(update bool, max int, bookId int) []Highlight {

	// Get highlights (GET https://readwise.io/api/v2/highlights/)

	// Create client
	client := &http.Client{}
	var apiResult ReadwiseHighlights
	var result []Highlight

	// Fetch Request
	done := false
	next := "https://readwise.io/api/v2/highlights/"

	var pageSize = ReadwisePageSize
	if max < ReadwisePageSize {
		pageSize = max
	}

	count := 0
	for !done {
		req, err := http.NewRequest("GET", next, nil)
		req.Header.Add("Authorization", "token "+api.Token)

		q := req.URL.Query()
		q.Add("page_size", fmt.Sprint(pageSize))
		if bookId != -1 {
			q.Add("book_id", fmt.Sprint(bookId))
			//			log.Println("get highlights URL: ", req.URL.String())
		}

		if update {
			q.Add("updated__gt", api.lastHighlightSync.Format(ReadwiseTimeFormat))
			//			log.Println("get highlights URL: ", req.URL.String())
		}
		req.URL.RawQuery = q.Encode()
		resp, err := client.Do(req)
		if err != nil {
			fmt.Println("Failure : ", err)
		}
		respBody, _ := ioutil.ReadAll(resp.Body)
		err = json.Unmarshal(respBody, &apiResult)
		result = append(result, apiResult.Results...)
		log.Print("Got ", len(result), " of ", apiResult.Count, " highlights")

		if apiResult.Next == "" || apiResult.Next == next {
			done = true
		} else {
			next = apiResult.Next
		}
		count += len(apiResult.Results)
		if count >= max {
			done = true
		}
	}
	return result
}

func (api *ReadwiseAPI) GetBooks(update bool) {
	// Get Books (GET https://readwise.io/api/v2/books/)

	// Create client
	client := &http.Client{}
	var apiResult ReadwiseBooks
	var result []Book

	// Fetch Request
	done := false
	next := "https://readwise.io/api/v2/books/"
	//	api.lastBookSync = time.Now()

	for !done {
		req, err := http.NewRequest("GET", next, nil)
		req.Header.Add("Authorization", "token "+api.Token)
		if update {
			q := req.URL.Query()
			q.Add("updated__gt", api.lastBookSync.Format(ReadwiseTimeFormat))
			req.URL.RawQuery = q.Encode()
			//			log.Println("get books URL: ", req.URL.String())
		}
		resp, err := client.Do(req)

		if err != nil {
			fmt.Println("Failure : ", err)
		}

		// Read Response Body
		respBody, _ := ioutil.ReadAll(resp.Body)
		err = json.Unmarshal(respBody, &apiResult)
		if err != nil {
			log.Print("Error unmarshalling highlights", err)
		}

		for _, book := range apiResult.Results {
			api.Books[book.Id] = book
		}

		log.Print("Got ", len(result), " of ", apiResult.Count, " books")

		if apiResult.Next == "" || apiResult.Next == next {
			done = true
		} else {
			next = apiResult.Next
		}
	}
	return
}
