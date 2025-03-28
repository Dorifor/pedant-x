package main

import (
	"context"
	"regexp"
	"slices"
	"strings"

	"github.com/coder/websocket"
)

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

func CheckIfTitleFound() bool {
	for _, token := range state.PageTokens {
		if !token.IsTitle {
			break
		}

		if state.TokensState[token.Id].Similarity < 1 {
			return false
		}
	}

	return true
}

func RemoveAllClosedClients() {
	Clients = slices.DeleteFunc(Clients, func(client *websocket.Conn) bool {
		_, _, err := client.Reader(context.Background())
		return err == nil
	})
}
