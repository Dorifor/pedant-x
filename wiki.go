package main

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strconv"
)

// regex for words / numbers : /[\p{L}]+|[[:digit:]]+/gm

type PageContent struct {
	Title   string
	Extract string
}

type WikiPageQuery struct {
	Query struct {
		Pages map[string]PageContent
	}
}

type WikiRandomQuery struct {
	Query struct {
		Random []struct {
			Id int
		}
	}
}

func get_article_content(id int) PageContent {
	str_id := strconv.Itoa(id)
	content_url := "https://fr.wikipedia.org/w/api.php?action=query&prop=extracts&exintro=&format=json&pageids=" + str_id

	req, _ := http.NewRequest("GET", content_url, nil)
	res, err := http.DefaultClient.Do(req)

	if err != nil {
		panic(err)
	}

	defer res.Body.Close()
	body, _ := io.ReadAll(res.Body)

	var q WikiPageQuery

	err = json.Unmarshal(body, &q)
	if err != nil {
		log.Println(err)
	}

	return q.Query.Pages[str_id]
}

func get_random_article() int {
	random_url := "https://fr.wikipedia.org/w/api.php?action=query&format=json&list=random&rnnamespace=0"

	req, _ := http.NewRequest("GET", random_url, nil)
	res, err := http.DefaultClient.Do(req)

	if err != nil {
		panic(err)
	}

	defer res.Body.Close()
	body, _ := io.ReadAll(res.Body)

	var q WikiRandomQuery

	err = json.Unmarshal(body, &q)
	if err != nil {
		log.Println(err)
	}

	return q.Query.Random[0].Id
}
