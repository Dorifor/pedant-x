package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"regexp"
	"strings"

	"github.com/sajari/word2vec"
)

type AppState struct {
	PageId        int
	PageTokens    []WordToken
	TokensState   map[int]WordSimilarity
	PageBaseHTML  string
	PageFinalHTML string
	PageTitle     string
	FoundTitle    bool
}

type WordToken struct {
	Id         int
	StartIndex int
	Word       string
	IsTitle    bool
}

type WordSimilarity struct {
	TokenId     int
	Similarity  float32
	SimilarWord string
}

type UserWordRequestPayload struct {
	SessionId string
	Word      string
}

type UserWordResponse struct {
	TitleFound    bool
	SimilarTokens []WordSimilarity
}

var state AppState
var model *word2vec.Model

func FetchRandomPage() {
	random_article_id := GetRandomArticle(1000)
	content := GetArticleContent(random_article_id)

	state.FoundTitle = false
	state.TokensState = make(map[int]WordSimilarity, 0)
	state.PageTitle = content.Title

	state.PageTokens, state.PageFinalHTML = GetFinalHtmlFromPage(content)
}

func GetFinalHtmlFromPage(page PageContent) (tokens []WordToken, final_html string) {
	base_html := page.Extract
	base_html = strings.Replace(base_html, "<p class=\"mw-empty-elt\">\n</p>\n\n\n", "", 1)
	base_html = RemoveTagProperties(base_html)
	base_html = "<h2>" + page.Title + "</h2>" + base_html

	state.PageBaseHTML = base_html

	ignored := GetIgnoredIndexes(base_html)

	title_end_pos := strings.Index(base_html, "</h2>")

	r := regexp.MustCompile(`[\p{L}]+|[[:digit:]]+`)
	token_id := 0
	last_end_index := 0

	for _, match := range r.FindAllStringIndex(base_html, -1) {
		if !IsIndexIgnored(ignored, match[0]) {
			token_id++
			newToken := WordToken{Id: token_id, StartIndex: match[0], Word: base_html[match[0]:match[1]], IsTitle: match[0] < title_end_pos}
			tokens = append(tokens, newToken)

			var spanHTML = fmt.Sprintf(`<span id="t%d" data-len=%d>%s</span>`, token_id, match[1]-match[0], strings.Repeat(" ", match[1]-match[0]))

			final_html += base_html[last_end_index:match[0]] + spanHTML
			last_end_index = match[1]
		}
	}

	final_html += base_html[last_end_index:]

	return
}

func main() {
	binary := flag.String("b", "", "The word embedding binary (word2vec format)")
	debug := flag.Bool("d", false, "Activates debug mode (see debug.go)")

	flag.Parse()

	if *binary == "" {
		panic("No binary submitted, use '-b' to specify word embeding path.")
	}

	fmt.Println("Loading the word2vec binary... ⏳")

	content, err := os.Open(*binary)
	if err != nil {
		panic(err)
	}

	defer content.Close()

	model, err = word2vec.FromReader(content)
	if err != nil {
		panic(err)
	}

	fmt.Println("Binary loaded ! ✅")

	FetchRandomPage()
	// fmt.Printf("Fetched the page \"%s\"\n", state.PageTitle)
	http.HandleFunc("/", MainHandler)
	http.HandleFunc("/word", CheckUserWordHandler)
	http.HandleFunc("/reveal", RevealPageHandler)
	if *debug {
		fmt.Println("Debug mode: ON 🤖")
		http.HandleFunc("/debug/state", DebugPrintAppStateHandler)
	}

	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("./static"))))
	http.ListenAndServe(":3333", nil)
	// fmt.Println(r.FindAllString(camus, -1))
}
