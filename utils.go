package main

import (
	"fmt"
	"regexp"
	"strings"
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
	for i := range len(state.PageTokens) {
		token := state.PageTokens[i+1]
		if !token.IsTitle {
			break
		}

		if state.TokensState[token.Id].Similarity < 1 {
			return false
		}
	}

	return true
}

func GetFinalHtmlFromPage(page PageContent) (tokens map[int]WordToken, final_html string) {
	tokens = make(map[int]WordToken)
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
			tokens[token_id] = newToken

			var spanHTML = fmt.Sprintf(`<span id="t%d" data-len=%d>%s</span>`, token_id, match[1]-match[0], strings.Repeat(" ", match[1]-match[0]))

			final_html += base_html[last_end_index:match[0]] + spanHTML
			last_end_index = match[1]
		}
	}

	final_html += base_html[last_end_index:]

	return
}
