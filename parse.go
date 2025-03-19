package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"html/template"
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
	PageFinalHTML string
	PageTitle     string
	FoundTitle    bool
}

type WordToken struct {
	Id         int
	StartIndex int
	Word       string
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

var state AppState
var model *word2vec.Model

func RemoveTagProperties(input string) string {
	r := regexp.MustCompile(`<(\/?\w+)[^>]*>`)
	return r.ReplaceAllString(input, "<$1>")
}

func GetIgnoredIndexes(input string) [][]int {
	r := regexp.MustCompile(`(</?\w+>)|\\n`)
	return r.FindAllStringIndex(input, -1)
}

func IsIndexIgnored(ignoredList [][]int, index int) bool {
	for _, ignored := range ignoredList {
		if index >= ignored[0] && index < ignored[1] {
			return true
		}
	}
	return false
}

func SanitizeWord(word string) string {
	return strings.ToLower(strings.TrimSpace(word))
}

func CheckUserWordHandler(w http.ResponseWriter, r *http.Request) {
	var payload UserWordRequestPayload

	err := json.NewDecoder(r.Body).Decode(&payload)

	if err != nil {
		fmt.Println(err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	word := payload.Word
	var foundSimilarities []WordSimilarity = make([]WordSimilarity, 0)

	e1 := word2vec.Expr{word: 1}

	for _, token := range state.PageTokens {
		if SanitizeWord(token.Word) == SanitizeWord(word) {
			similarWord := WordSimilarity{TokenId: token.Id, Similarity: 1, SimilarWord: token.Word}
			state.TokensState[token.Id] = similarWord
			foundSimilarities = append(foundSimilarities, similarWord)
		} else {
			e2 := word2vec.Expr{token.Word: 1}
			similarity, _ := model.Cos(e1, e2)
			lastSim, stateExists := state.TokensState[token.Id]
			if stateExists && similarity < lastSim.Similarity || similarity < 0.1 {
				continue
			}
			similarWord := WordSimilarity{TokenId: token.Id, Similarity: similarity, SimilarWord: word}
			foundSimilarities = append(foundSimilarities, similarWord)
			state.TokensState[token.Id] = similarWord
		}
	}

	jsonBytes, _ := json.Marshal(foundSimilarities)
	w.Write(jsonBytes)
}

func MainHandler(w http.ResponseWriter, r *http.Request) {
	params := struct{ PedHTML template.HTML }{
		PedHTML: template.HTML(state.PageFinalHTML),
	}
	t, _ := template.ParseFiles("ped.html")
	t.Execute(w, params)
}

func FetchRandomPage() {
	randomArticleId := get_random_article()
	content := get_article_content(randomArticleId)

	state.FoundTitle = false
	state.TokensState = make(map[int]WordSimilarity, 0)
	state.PageTitle = content.Title

	state.PageTokens, state.PageFinalHTML = GetFinalHtmlFromPage(content)
}

func GetFinalHtmlFromPage(page PageContent) (tokens []WordToken, finalHTML string) {
	baseHTML := page.Extract
	baseHTML = strings.Replace(baseHTML, "<p class=\"mw-empty-elt\">\n</p>\n\n\n", "", 1)
	baseHTML = RemoveTagProperties(baseHTML)
	baseHTML = "<h2>" + page.Title + "</h2>" + baseHTML
	ignored := GetIgnoredIndexes(baseHTML)

	r := regexp.MustCompile("[\\p{L}]+|[[:digit:]]+")
	tokenId := 0
	lastEndIndex := 0

	for _, match := range r.FindAllStringIndex(baseHTML, -1) {
		if !IsIndexIgnored(ignored, match[0]) {
			tokenId++
			newToken := WordToken{Id: tokenId, StartIndex: match[0], Word: baseHTML[match[0]:match[1]]}
			tokens = append(tokens, newToken)

			var spanHTML = fmt.Sprintf(`<span id="t%d" data-len=%d>%s</span>`, tokenId, match[1]-match[0], strings.Repeat(" ", match[1]-match[0]))

			finalHTML += baseHTML[lastEndIndex:match[0]] + spanHTML
			lastEndIndex = match[1]
		}
	}

	finalHTML += baseHTML[lastEndIndex:]

	return
}

func main() {
	binary := flag.String("b", "", "The word embedding binary (word2vec format)")

	flag.Parse()

	if *binary == "" {
		panic("No binary submitted, use '-b' to specify word embeding path.")
	}

	fmt.Println("Loading the word2vec binary...")

	content, err := os.Open(*binary)
	if err != nil {
		panic(err)
	}

	defer content.Close()

	model, err = word2vec.FromReader(content)
	if err != nil {
		panic(err)
	}

	fmt.Println("Loaded")

	FetchRandomPage()
	fmt.Printf("Fetched the page \"%s\"\n", state.PageTitle)
	http.HandleFunc("/", MainHandler)
	http.HandleFunc("/word", CheckUserWordHandler)
	http.ListenAndServe(":3333", nil)
	// fmt.Println(r.FindAllString(camus, -1))
}
