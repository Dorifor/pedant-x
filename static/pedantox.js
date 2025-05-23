let lastSentWord = "";
let lastRequestDate = Date.now();
const searchForm = document.querySelector("#word_search");
const foundMatches = document.querySelector(".found-matches");
const similarMatches = document.querySelector(".similar-matches");
const wordHistoryList = document.querySelector(".word-history");
const searchInput = searchForm.querySelector("input");
const wordNotFoundLabel = document.querySelector('#not-found');
const titleFoundLabel = document.querySelector('#title-found');
const revealButton = document.querySelector('#reveal-button');
const wikiArticle = document.querySelector('article');
const lobbyLabel = document.querySelector('p.lobby-label > span');
const hintsLabel = document.querySelector('p.hint-label > span');
const refreshButton = document.querySelector('p.refresh');
const seePageButton = document.querySelector('#see-page-button');

const wordHistory = [];
let lastFoundTokens = [];
const tokens = document.querySelectorAll('article span');
tokens.forEach(token => {
    token.addEventListener('click', e => {
        if (e.target.classList.contains('found'))
            return;

        var tokenId = e.target.id.replace('t', '');

        const payload = {
            type: "hint",
            session_id: "",
            data: tokenId
        };

        webSocket.send(JSON.stringify(payload))
    });
})

const webSocket = new WebSocket(`ws://${window.location.host}/ws`);

webSocket.onmessage = event => {
    /**
     * @type { { Type: string, Status: number, Data: any } }
     */
    const message = JSON.parse(event.data);

    switch (message.Type) {
        case "lobby":
            lobbyLabel.textContent = message.Data.PlayerCount;
            break;

        case "init":
            if (message.Data.TitleFound) {
                titleFoundLabel.classList.remove('hidden');
                revealButton.classList.remove('hidden');
                seePageButton.classList.remove('hidden');
                seePageButton.setAttribute('href', message.Data.WikiPageURL)
            }

            applySimilarTokens(message.Data.CurrentTokenState)

            if (message.Data.WordsHistory)
                message.Data.WordsHistory.forEach(word => addWordToHistory(word))
            break;

        case "word":
            if (message.Status == 404) {
                wordNotFoundLabel.classList.remove('hidden');
            }
            if (message.Data.TitleFound) {
                titleFoundLabel.classList.remove('hidden');
                revealButton.classList.remove('hidden');
                seePageButton.classList.remove('hidden');
                seePageButton.setAttribute('href', message.Data.WikiPageURL)
            }

            foundMatches.textContent = null;
            similarMatches.textContent = null;

            if (message.Status != 404) {
                addWordToHistory(message.Data.Word);
            }
            applySimilarTokens(message.Data.SimilarTokens);
            break;

        case "loading":
            refreshButton.classList.add('loading');
            break;

        case "refresh":
            location.reload();
            break;

        case "hint":
            hintsLabel.textContent = message.Data.Remaining;
            applyHint(message.Data.TokenId, message.Data.Hint);
            break;

        default:
            break;
    }
}

refreshButton.addEventListener('click', e => {
    if (refreshButton.classList.contains('loading'))
        return;
    
    const payload = {
        type: "refresh",
        session_id: "",
        data: null
    };

    webSocket.send(JSON.stringify(payload))
});


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

    if (word == lastSentWord) return;

    const payload = {
        type: "word",
        session_id: "",
        data: word
    };

    webSocket.send(JSON.stringify(payload))

    lastSentWord = word;
    lastRequestDate = Date.now();
});



function addWordToHistory(word) {
    if (!wordHistory.includes(word)) {
        wordHistory.push(word);
        const historyListItem = document.createElement('li');
        historyListItem.textContent = word;
        wordHistoryList.appendChild(historyListItem);
        historyListItem.scrollIntoView();
    }
}

function applySimilarTokens(tokens) {
    if (!tokens) return;

    lastFoundTokens.forEach((token) => {
        token.classList.remove("just-found");
        token.classList.remove("just-similar");
        token.classList.remove("just-hinted");
    });
    lastFoundTokens = [];

    tokens.forEach((sim) => {
        const matchedToken = document.querySelector(`#t${sim.TokenId}`);
        if (sim.Similarity >= .95) {
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

function applyHint(tokenId, hintWord) {
    lastFoundTokens.forEach((token) => {
        token.classList.remove("just-found");
        token.classList.remove("just-similar");
        token.classList.remove("just-hinted");
    });
    lastFoundTokens = [];

    const matchedToken = document.querySelector(`#t${tokenId}`);
    matchedToken.classList.add("just-hinted");
    if (matchedToken.textContent.length <= hintWord.length) {
        matchedToken.textContent = " ".repeat(
            hintWord.length,
        );
    }
    matchedToken.setAttribute("data-similar", hintWord);
    matchedToken.style.setProperty("--opacity", .7);
    lastFoundTokens.push(matchedToken);
}