package main

import (
	"encoding/json"
	"html/template"
	"log"
	"maps"
	"math"
	"net/http"
	"slices"
	"strconv"

	"fmt"

	"github.com/olahol/melody"
	"github.com/sajari/word2vec"
)

type WSLobbyResponse struct {
	Type string
	Data struct {
		PlayerCount int
	}
}

type WSWordResponse struct {
	Type   string
	Status int
	Data   UserWordResponse
}

type WSInitResponse struct {
	Type string
	Data struct {
		WordsHistory      []string
		CurrentTokenState []WordSimilarity
		TitleFound        bool
		WikiPageUrl       string
	}
}

type WSBaseResponse struct {
	Type string
}

type WSHintResponse struct {
	Type string
	Data struct {
		TokenId   int
		Hint      string
		Remaining int
	}
}

type WSRequest struct {
	Type string
	Data any
}

func HandleInit(s *melody.Session) {
	id := ids.Add(1)
	s.Set("id", id)
	sessions[id] = s

	var init_response WSInitResponse
	init_response.Type = "init"
	init_response.Data.TitleFound = state.FoundTitle
	init_response.Data.WordsHistory = state.WordsHistory
	init_response.Data.CurrentTokenState = slices.Collect(maps.Values(state.TokensState))
	if state.FoundTitle {
		init_response.Data.WikiPageUrl = "https://" + settings.Lang + ".wikipedia.org/wiki?curid=" + fmt.Sprint(state.PageId)
	}

	init_response_json, err := json.Marshal(init_response)
	if err != nil {
		log.Println(err)
	}

	s.Write(init_response_json)

	SendLobbyUpdate()
}

func HandleMessage(s *melody.Session, msg []byte) {
	var data WSRequest

	err := json.Unmarshal(msg, &data)
	if err != nil {
		log.Println(err)
	}

	switch data.Type {
	case "word":
		word := data.Data.(string)
		wordResponse := HandleWord(word)
		response := WSWordResponse{
			Type: "word",
			Data: wordResponse,
		}

		if wordResponse.IsUnknown {
			response.Status = 404
			response_json, err := json.Marshal(response)
			if err != nil {
				log.Println(err)
				return
			}
			s.Write(response_json)
		} else {
			response_json, err := json.Marshal(response)
			if err != nil {
				log.Println(err)
				return
			}
			m.Broadcast(response_json)
			state.WordsHistory = append(state.WordsHistory, word)
		}
	case "hint":
		tokenId, err := strconv.Atoi(data.Data.(string))
		if err != nil {
			log.Println(err)
			break
		}
		HandleHint(tokenId)
	case "refresh":
		var loadingResponse WSBaseResponse
		loadingResponse.Type = "loading"
		loading_response_json, err := json.Marshal(loadingResponse)
		if err != nil {
			log.Println(err)
		}
		m.Broadcast(loading_response_json)

		FetchRandomPage()

		var refresh_response WSBaseResponse
		refresh_response.Type = "refresh"

		refresh_response_json, err := json.Marshal(refresh_response)
		if err != nil {
			log.Println(err)
		}
		m.Broadcast(refresh_response_json)
	}
}

func HandleWord(word string) (response UserWordResponse) {
	response.SimilarTokens = make([]WordSimilarity, 0)
	response.TitleFound = false
	response.IsUnknown = false
	response.Word = word

	e1 := word2vec.Expr{word: 1}
	word_int, err := strconv.Atoi(word)
	is_word_number := err == nil

	_, err = model.Eval(e1)
	word_is_unknown := err != nil

	for _, token := range state.PageTokens {
		if SanitizeWord(token.Word) == SanitizeWord(word) || CheckPlural(word, token.Word) {
			similar_word := WordSimilarity{TokenId: token.Id, Similarity: 1, SimilarWord: token.Word}
			state.TokensState[token.Id] = similar_word
			response.SimilarTokens = append(response.SimilarTokens, similar_word)
		} else if is_word_number {
			token_int, err := strconv.Atoi(token.Word)
			is_token_number := err == nil

			if !is_token_number {
				continue
			}
			current_diff := math.Abs(float64(word_int - token_int))

			last_sim, state_exists := state.TokensState[token.Id]
			last_int, _ := strconv.Atoi(last_sim.SimilarWord)
			last_diff := math.Abs(float64(last_int - token_int))

			if state_exists && last_diff < current_diff || current_diff > float64(50)/100*float64(token_int) {
				continue
			}

			similar_int := WordSimilarity{TokenId: token.Id, Similarity: .5, SimilarWord: word}
			response.SimilarTokens = append(response.SimilarTokens, similar_int)
			state.TokensState[token.Id] = similar_int
		} else if !word_is_unknown || len(word) == 1 {
			if token.IsTitle {
				continue
			}

			e2 := word2vec.Expr{token.Word: 1}
			similarity, _ := model.Cos(e1, e2)

			last_sim, state_exists := state.TokensState[token.Id]
			if len(word) > 4 && similarity < 0.3 {
				continue
			}
			if state_exists && similarity < last_sim.Similarity || similarity < 0.2 {
				continue
			}

			similar_word := WordSimilarity{TokenId: token.Id, Similarity: similarity, SimilarWord: word}
			response.SimilarTokens = append(response.SimilarTokens, similar_word)
			state.TokensState[token.Id] = similar_word
		}
	}

	if len(response.SimilarTokens) <= 0 && word_is_unknown && !is_word_number {
		response.IsUnknown = true
	}

	if !response.IsUnknown && !state.FoundTitle && CheckIfTitleFound() {
		response.TitleFound = true
		response.WikiPageURL = "https://" + settings.Lang + ".wikipedia.org/wiki?curid=" + fmt.Sprint(state.PageId)
		state.FoundTitle = true
	}

	return
}

func HandleHint(tokenId int) {
	if state.HintsRemaining <= 0 {
		return
	}

	hint_count, ok_hint := state.TokenHints[tokenId]
	token, ok_token := state.PageTokens[tokenId]
	if !ok_token || hint_count >= 3 {
		return
	}

	if !ok_hint {
		state.TokenHints[tokenId] = 0
	}

	var hint_response WSHintResponse

	matches, err := model.CosN(word2vec.Expr{token.Word: 1}, 5)
	if err != nil {
		// log.Println(err)
		if state.TokenHints[tokenId]+1 >= len(token.Word) {
			return
		}
		hint_response.Data.Hint = token.Word[:state.TokenHints[tokenId]+1] + "~"
	} else {
		hint_response.Data.Hint = matches[3-hint_count].Word
	}

	hint_count++
	state.HintsRemaining--

	hint_response.Type = "hint"
	hint_response.Data.TokenId = tokenId
	hint_response.Data.Remaining = state.HintsRemaining

	hint_response_json, err := json.Marshal(hint_response)
	if err != nil {
		log.Println(err)
	}

	state.TokenHints[tokenId]++
	m.Broadcast(hint_response_json)
}

func RevealPageHandler(w http.ResponseWriter, r *http.Request) {
	if !state.FoundTitle {
		http.Error(w, "Forbidden.", http.StatusForbidden)
		return
	}
	w.Write([]byte(state.PageBaseHTML))
}

func MainHandler(w http.ResponseWriter, r *http.Request) {
	params := struct {
		ArticleHTML    template.HTML
		T              AppTranslationStrings
		PageViews      int
		HintsRemaining int
	}{
		ArticleHTML:    template.HTML(state.PageFinalHTML),
		T:              translations,
		PageViews:      state.PageViews,
		HintsRemaining: state.HintsRemaining,
	}
	t, _ := template.ParseFiles("app.html")
	t.Execute(w, params)
}

func SendLobbyUpdate() {
	var lobby_update WSLobbyResponse
	lobby_update.Type = "lobby"
	lobby_update.Data.PlayerCount = len(sessions)

	lobby_update_json, err := json.Marshal(lobby_update)
	if err != nil {
		log.Println(err)
	}

	m.BroadcastMultiple(lobby_update_json, slices.Collect(maps.Values(sessions)))
}
