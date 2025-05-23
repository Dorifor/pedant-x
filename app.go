package main

import (
	"encoding/json"
	"flag"
	"log"
	"net/http"
	"os"
	"sync/atomic"

	"github.com/olahol/melody"
	"github.com/sajari/word2vec"
)

type AppState struct {
	PageId         int
	PageViews      int
	PageTokens     map[int]WordToken
	TokensState    map[int]WordSimilarity
	WordsHistory   []string
	PageBaseHTML   string
	PageFinalHTML  string
	PageTitle      string
	FoundTitle     bool
	RevealVoteIDs  []int64
	TokenHints     map[int]int
	HintsRemaining int
}

type AppSettings struct {
	Lang         string
	WikiMinViews int
	ServerPort   string
}

type AppTranslationStrings struct {
	WordNotFound            string
	WinCongratulations      string
	RevealPageButton        string
	WordHistory             string
	FoundMatchesIndicator   string
	SimilarMatchesIndicator string
	SeePageButton           string
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
	IsUnknown     bool
	Word          string
	WikiPageURL   string
}

var (
	state        AppState
	model        *word2vec.Model
	sessions     map[int64]*melody.Session = make(map[int64]*melody.Session)
	m            *melody.Melody
	ids          atomic.Int64
	settings     AppSettings
	translations AppTranslationStrings
)

func FetchRandomPage() {
	random_article_id, random_article_page_views := GetRandomArticle(settings.WikiMinViews)
	content := GetArticleContent(random_article_id)

	state.FoundTitle = false
	state.TokensState = make(map[int]WordSimilarity, 0)
	state.PageTitle = content.Title
	state.PageId = random_article_id
	state.PageViews = random_article_page_views
	state.WordsHistory = make([]string, 0)
	state.TokenHints = make(map[int]int)
	state.HintsRemaining = 5

	state.PageTokens, state.PageFinalHTML = GetFinalHtmlFromPage(content)
}

func main() {
	binary := flag.String("b", "", "The word embedding binary (word2vec format)")
	debug := flag.Bool("d", false, "Activates debug mode (see debug.go)")
	server_port := flag.String("p", "3333", "Server port")
	wiki_min_views := flag.Int("v", 3500, "Minimum page views for a random wiki page to be picked")
	app_lang := flag.String("l", "en", "App & Wikipedia content language code (fr, en, de, ...)")

	flag.Parse()

	settings.ServerPort = *server_port
	settings.WikiMinViews = *wiki_min_views
	settings.Lang = *app_lang

	if *binary == "" {
		panic("No binary submitted, use '-b' to specify word embeding path.")
	}

	log.Println("Loading the word2vec binary... ⏳")

	content, err := os.Open(*binary)
	if err != nil {
		panic(err)
	}

	defer content.Close()

	model, err = word2vec.FromReader(content)
	if err != nil {
		panic(err)
	}

	log.Println("Binary loaded ! ✅")

	log.Println("Loading translations... ⏳")

	translations_data, err := os.ReadFile("./translations.json")
	if err != nil {
		log.Println("Error loading the translation file.")
		panic(err)
	}

	log.Println("Translations loaded ! ✅")

	var translations_map map[string]AppTranslationStrings

	err = json.Unmarshal(translations_data, &translations_map)
	if err != nil {
		log.Println("Error parsing the translation file.")
		panic(err)
	}

	if val, ok := translations_map[settings.Lang]; ok {
		translations = val
	} else {
		translations = translations_map["en"]
	}

	FetchRandomPage()

	m = melody.New()

	http.HandleFunc("/", MainHandler)
	http.HandleFunc("/reveal", RevealPageHandler)

	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		m.HandleRequest(w, r)
	})

	m.HandleConnect(HandleInit)

	m.HandleMessage(HandleMessage)

	m.HandleDisconnect(func(s *melody.Session) {
		id, _ := s.Get("id")
		delete(sessions, id.(int64))

		SendLobbyUpdate()
	})

	if *debug {
		log.Println("Debug mode: ON 🤖")
		http.HandleFunc("/debug/state", DebugPrintAppStateHandler)
		http.HandleFunc("/debug/fetch", DebugFetchRandomPage)
	}

	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("./static"))))
	http.ListenAndServe(":"+settings.ServerPort, nil)
}
