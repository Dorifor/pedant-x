let lastSentWord = "";
let lastRequestDate = Date.now();
let searchForm = document.querySelector("#word_search");
let foundMatches = document.querySelector(".found-matches");
let similarMatches = document.querySelector(".similar-matches");
let wordHistoryList = document.querySelector(".word-history");
let searchInput = searchForm.querySelector("input");
let wordNotFoundLabel = document.querySelector('#not-found');
let titleFoundLabel = document.querySelector('#title-found');
let revealButton = document.querySelector('#reveal-button');
let wikiArticle = document.querySelector('article');

const wordHistory = [];
let lastFoundTokens = [];

document.addEventListener("keydown", (e) => {
    searchInput.focus();
});

revealButton.addEventListener("click", async e => {
    const res = await fetch("reveal");

    if (res.status == 403) {
        return;
    }

    wikiArticle.innerHTML = await res.text();
});

searchForm.addEventListener("submit", async (e) => {
    e.preventDefault();
    const elapsedTime = Date.now() - lastRequestDate;
    const word = searchInput.value;

    foundMatches.textContent = null;
    similarMatches.textContent = null;

    if (word.length <= 0 || elapsedTime < 250) return;
    wordNotFoundLabel.classList.add('hidden');

    searchInput.value = null;
    searchInput.placeholder = word;

    lastFoundTokens.forEach((token) => {
        token.classList.remove("just-found");
        token.classList.remove("just-similar");
    });
    lastFoundTokens = [];

    if (word == lastSentWord) return;

    const payload = {
        session_id: "",
        word: word,
    };

    const res = await fetch("word", {
        method: "POST",
        body: JSON.stringify(payload),
    });

    if (res.status == 404) {
        wordNotFoundLabel.classList.remove('hidden');
    }

    /**
     * @type { { TitleFound: boolean, SimilarTokens: { TokenId: number, Similarity: number, SimilarWord: string } }[] }
     */
    const jsonResponse = await res.json();

    if (jsonResponse.TitleFound) {
        titleFoundLabel.classList.remove('hidden');
        revealButton.classList.remove('hidden');
    }

    applySimilarTokens(jsonResponse.SimilarTokens);

    if (!wordHistory.includes(word)) {
        wordHistory.push(word);
        const historyListItem = document.createElement('li');
        historyListItem.textContent = word;
        wordHistoryList.appendChild(historyListItem);
        historyListItem.scrollIntoView();
    }

    lastSentWord = word;
    lastRequestDate = Date.now();
});

function applySimilarTokens(tokens) {
    tokens.forEach((sim) => {
        const matchedToken = document.querySelector(`#t${sim.TokenId}`);
        if (sim.Similarity >= .99) {
            matchedToken.classList.add("found");
            matchedToken.classList.add("just-found");
            matchedToken.textContent = sim.SimilarWord;
            foundMatches.textContent += "_";
        } else {
            similarMatches.textContent += "_";
            matchedToken.classList.add("just-similar");
            if (matchedToken.textContent.length <= sim.SimilarWord.length) {
                matchedToken.textContent = " ".repeat(
                    sim.SimilarWord.length,
                );
            }
            matchedToken.setAttribute("data-similar", sim.SimilarWord);
            matchedToken.style.setProperty("--opacity", sim.Similarity + .2);
        }

        lastFoundTokens.push(matchedToken);
    });
}