package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
)

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

type WikiPageViewsQuery struct {
	Query struct {
		Pages map[string]struct {
			Pageid    int
			Pageviews map[string]int
		}
	}
}

func GetArticleContent(id int) PageContent {
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

func GetMostViewedArticle(query *WikiRandomQuery) (page_id, sum_view_count int) {
	page_ids := make([]string, 0)

	for _, page := range query.Query.Random {
		page_ids = append(page_ids, fmt.Sprint(page.Id))
	}

	page_views_url := "https://fr.wikipedia.org/w/api.php?action=query&minsize=60000&prop=pageviews&pvipdays=30&format=json&pageids=" + strings.Join(page_ids, "|")

	req, _ := http.NewRequest("GET", page_views_url, nil)
	res, err := http.DefaultClient.Do(req)

	if err != nil {
		panic(err)
	}

	defer res.Body.Close()

	body, _ := io.ReadAll(res.Body)

	var q WikiPageViewsQuery

	err = json.Unmarshal(body, &q)
	if err != nil {
		panic(err)
	}

	sum_page_views := make(map[int]int, 0)

	for _, page := range q.Query.Pages {
		sum := 0
		for _, views := range page.Pageviews {
			sum += views
		}
		sum_page_views[page.Pageid] = sum
	}

	for id, sum := range sum_page_views {
		if sum_view_count > sum {
			continue
		}

		page_id = id
		sum_view_count = sum
	}

	fmt.Printf("most viewed: %d with %d views this month.\n", page_id, sum_view_count)
	return
}

func GetRandomArticles(count int) WikiRandomQuery {
	random_url := "https://fr.wikipedia.org/w/api.php?action=query&format=json&list=random&rnnamespace=0&rnlimit=" + strconv.Itoa(count)

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
		panic(err)
	}

	return q
}

func GetRandomArticle(minPageViews int) int {
	var page_view_count int
	var page_id int

	for page_view_count <= minPageViews {
		random_list := GetRandomArticles(50)
		page_id, page_view_count = GetMostViewedArticle(&random_list)
	}

	return page_id
}
