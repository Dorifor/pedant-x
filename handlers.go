package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"

	"github.com/sajari/word2vec"
)

func CheckUserWordHandler(w http.ResponseWriter, r *http.Request) {
	var payload UserWordRequestPayload

	err := json.NewDecoder(r.Body).Decode(&payload)

	if err != nil {
		fmt.Println(err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	word := payload.Word
	var found_similarities []WordSimilarity = make([]WordSimilarity, 0)

	e1 := word2vec.Expr{word: 1}

	for _, token := range state.PageTokens {
		if SanitizeWord(token.Word) == SanitizeWord(word) {
			similar_word := WordSimilarity{TokenId: token.Id, Similarity: 1, SimilarWord: token.Word}
			state.TokensState[token.Id] = similar_word
			found_similarities = append(found_similarities, similar_word)
		} else {
			if token.IsTitle {
				continue
			}
			e2 := word2vec.Expr{token.Word: 1}
			similarity, _ := model.Cos(e1, e2)
			last_sim, state_exists := state.TokensState[token.Id]
			if state_exists && similarity < last_sim.Similarity || similarity < 0.4 {
				continue
			}
			similar_word := WordSimilarity{TokenId: token.Id, Similarity: similarity, SimilarWord: word}
			found_similarities = append(found_similarities, similar_word)
			state.TokensState[token.Id] = similar_word
		}
	}

	jsonBytes, _ := json.Marshal(found_similarities)
	w.Write(jsonBytes)
}

func MainHandler(w http.ResponseWriter, r *http.Request) {
	params := struct{ ArticleHTML template.HTML }{
		ArticleHTML: template.HTML(state.PageFinalHTML),
	}
	t, _ := template.ParseFiles("app.html")
	t.Execute(w, params)
}
