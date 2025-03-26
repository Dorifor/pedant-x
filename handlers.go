package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"math"
	"net/http"
	"strconv"

	"github.com/sajari/word2vec"
)

func CheckUserWordHandler(w http.ResponseWriter, r *http.Request) {
	var payload UserWordRequestPayload
	var response UserWordResponse
	response.SimilarTokens = make([]WordSimilarity, 0)
	response.TitleFound = false

	err := json.NewDecoder(r.Body).Decode(&payload)

	if err != nil {
		fmt.Println(err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	word := payload.Word

	e1 := word2vec.Expr{word: 1}
	word_int, err := strconv.Atoi(word)
	is_word_number := err == nil

	_, err = model.Eval(e1)
	word_is_unknown := err != nil

	for _, token := range state.PageTokens {
		if SanitizeWord(token.Word) == SanitizeWord(word) {
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
			if state_exists && similarity < last_sim.Similarity || similarity < 0.35 {
				continue
			}

			similar_word := WordSimilarity{TokenId: token.Id, Similarity: similarity, SimilarWord: word}
			response.SimilarTokens = append(response.SimilarTokens, similar_word)
			state.TokensState[token.Id] = similar_word
		}
	}

	if len(response.SimilarTokens) <= 0 && word_is_unknown && !is_word_number {
		http.Error(w, "Word unknown.", 404)
		return
	}

	if !state.FoundTitle && CheckIfTitleFound() {
		response.TitleFound = true
		state.FoundTitle = true
		fmt.Println("TITLE FOUND !!!!")
	}

	jsonBytes, _ := json.Marshal(response)
	w.Write(jsonBytes)
}

func RevealPageHandler(w http.ResponseWriter, r *http.Request) {
	if !state.FoundTitle {
		http.Error(w, "Forbidden.", http.StatusForbidden)
		return
	}
	w.Write([]byte(state.PageBaseHTML))
}

func MainHandler(w http.ResponseWriter, r *http.Request) {
	params := struct{ ArticleHTML template.HTML }{
		ArticleHTML: template.HTML(state.PageFinalHTML),
	}
	t, _ := template.ParseFiles("app.html")
	t.Execute(w, params)
}
