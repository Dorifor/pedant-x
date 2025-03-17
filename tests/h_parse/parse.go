package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"regexp"
	"strings"
)

type AppState struct {
	PageId        int
	PageTokens    []WordToken
	PageFinalHTML string
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

var title = "Albert Camus"
var baseHTML = "<p class=\"mw-empty-elt\">\n</p>\n\n\n<p><b>Albert Camus</b>, né le <time class=\"nowrap bday\" datetime=\"1913-11-07\" data-sort-value=\"1913-11-07\">7 novembre 1913</time> à Mondovi dans le département de Constantine (aujourd'hui Dréan dans la wilaya d'El Tarf), en Algérie pendant la période coloniale française, et mort par accident le <time class=\"nowrap dday\" datetime=\"1960-01-04\" data-sort-value=\"1960-01-04\">4 janvier 1960</time> à Villeblevin en France, est un philosophe, écrivain, journaliste militant, romancier, dramaturge, essayiste et nouvelliste français, lauréat du prix Nobel de littérature en 1957.\n</p><p>Né sur la côte orientale de l'Algérie, à proximité de Bône (aujourd'hui Annaba), de parents pieds-noirs, Camus passe son enfance dans les quartiers pauvres et populaires. Grâce à son instituteur Louis Germain, il est reçu au Grand Lycée d’Alger et entre par la suite en hypokhâgne à l'université, où Jean Grenier est son professeur de philosophie. Mais sa santé — dégradée par la tuberculose — ne lui permet pas d'accéder à une carrière universitaire. Après des débuts journalistiques et littéraires et la publication de deux de ses plus grandes œuvres : <i>L'Étranger</i> et <i>Le Mythe de Sisyphe,</i> il s'engage dans la Résistance française lors de l'Occupation, où il devient, fin 1943, rédacteur en chef du journal <i>Combat.</i>\n</p><p>Son œuvre comprend des pièces de théâtre, des romans, des nouvelles, des films, des poèmes et des essais dans lesquels il développe un humanisme sceptique et lucide fondé sur la prise de conscience de l'<i>absurde</i>, de la condition humaine et de la <i>révolte</i>, qui conduit à l'action, à la justice, et qui donne un sens au monde et à l'existence ; l'œuvre de Camus a par conséquent contribué à la montée de la philosophie de l'absurde. Rattaché à l'existentialisme, dans le sens où « l'absurde camusien » est aussi une réponse au nihilisme, l'écrivain a toujours refusé d'être étiqueté à ce courant.\n</p><p>Internationaliste réformiste, moraliste, abolitionniste et proche des courants libertaires, il prend notamment position sur la question de l'indépendance de l'Algérie et ses rapports avec le Parti communiste algérien, qu'il quitte après un court passage de deux ans. Il proteste également contre les inégalités et la misère qui frappent les indigènes d'Afrique du Nord tout comme il dénonce la caricature du pied-noir exploiteur, tout en prenant la défense des Espagnols exilés antifascistes, des victimes du stalinisme ou encore des objecteurs de conscience. En marge des courants philosophiques, Camus est d'abord <span>« témoin de son temps et ne cesse de lutter contre les idéologies et les abstractions qui détournent de l'humain »</span>. Il est ainsi amené à s'opposer aussi bien au libéralisme qu’à l'existentialisme et au marxisme. Lors de la sortie de <i>L'Homme révolté</i> en 1951, sa critique de la légitimation de la violence et son anti-soviétisme lui valent les anathèmes des intellectuels communistes, ainsi que sa rupture avec Jean-Paul Sartre.\n</p><p>En janvier 1960, victime d'un accident de voiture brutal alors qu'il se rendait à Paris avec Janine, Anne et Michel Gallimard, il meurt sur le coup, à 46 ans, et laisse derrière lui une partie inachevée de son œuvre.\n</p>"

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
	for _, token := range state.PageTokens {
		if SanitizeWord(token.Word) == SanitizeWord(word) {
			similarWord := WordSimilarity{TokenId: token.Id, Similarity: 100, SimilarWord: token.Word}
			foundSimilarities = append(foundSimilarities, similarWord)
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

func main() {
	baseHTML = strings.Replace(baseHTML, "<p class=\"mw-empty-elt\">\n</p>\n\n\n", "", 1)
	baseHTML = RemoveTagProperties(baseHTML)
	baseHTML = "<h2>" + title + "</h2>" + baseHTML
	ignored := GetIgnoredIndexes(baseHTML)

	r := regexp.MustCompile("[\\p{L}]+|[[:digit:]]+")
	tokenId := 0
	lastEndIndex := 0

	for _, match := range r.FindAllStringIndex(baseHTML, -1) {
		if !IsIndexIgnored(ignored, match[0]) {
			tokenId++
			newToken := WordToken{Id: tokenId, StartIndex: match[0], Word: baseHTML[match[0]:match[1]]}
			state.PageTokens = append(state.PageTokens, newToken)

			var spanHTML = fmt.Sprintf(`<span id="t%d" data-len=%d>%s</span>`, tokenId, match[1]-match[0], strings.Repeat(" ", match[1]-match[0]))

			state.PageFinalHTML += baseHTML[lastEndIndex:match[0]] + spanHTML
			lastEndIndex = match[1]
		}
	}

	state.PageFinalHTML += baseHTML[lastEndIndex:]

	http.HandleFunc("/", MainHandler)
	http.HandleFunc("/word", CheckUserWordHandler)
	http.ListenAndServe(":3333", nil)
	// fmt.Println(r.FindAllString(camus, -1))
}
