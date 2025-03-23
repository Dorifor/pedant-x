let lastSentWord = "";
let lastRequestDate = Date.now();
let searchForm = document.querySelector("#word_search");
let foundMatches = document.querySelector(".found-matches");
let similarMatches = document.querySelector(".similar-matches");
let wordHistoryList = document.querySelector(".word-history");
let searchInput = searchForm.querySelector("input");

const wordHistory = [];
let lastFoundTokens = [];

document.addEventListener("keydown", (e) => {
    searchInput.focus();
});

searchForm.addEventListener("submit", async (e) => {
    e.preventDefault();
    const elapsedTime = Date.now() - lastRequestDate;
    const word = searchInput.value;

    foundMatches.textContent = null;
    similarMatches.textContent = null;

    if (word.length <= 0 || elapsedTime < 250) return;

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

    const jsonResponse = await res.json();

    jsonResponse.forEach((sim) => {
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