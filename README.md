# Pédant ∙ x

An attempt to remake the game [pédantix / pedantle](https://pedantix.certitudes.org/) from [@enigmathix](https://github.com/enigmathix)

# 🤨 Concept
Random wiki page, blank words, user types word, if word in text, reveal, if not show semantic similarity.

If title revealed 🏆

Multiplayer (create group and invite friends / colleagues)

# 📷 Screenshots
[pedantox.webm](https://github.com/user-attachments/assets/a6c04209-f288-4699-ab0e-08b939ac59dc)


# 🤓 Usage

Download the [latest release](https://github.com/Dorifor/pedant-x/releases/latest) (or build it yourself)

> You'll **need** a word2vec pre-trained binary

Launch the server:  
```
./pedantox -b <mandatory_word2vec_binary> [-p <port>] [-l <lang>] [-v <min_page_views>] [-d]
```

`-b` to specify the **word2vec** pre-trained binary the semantic analyzer will use  
`-l` to indicate language of the app and wikipedia content (default: en)  
`-v` to specify the minimum views a page has to accumulate in the last month to be picked at random (default: 3500)  
`-d` to toggle debug mode  

## Debug Mode
The debug mode exposes two endpoints :
- `/debug/state` returns the AppState as JSON
- `/debug/fetch` updates manually the random wiki page 

## Translations
If you need a translation for another language, you just have to add yours inside `translations.json` with the lang code and the app translation strings
If the lang code provided with the `-l` flag does not have a translation, it will fallback to english by default.

# 🪛 Build

1. Clone the repo
2. Install the go project dependencies
3. `go build .`

# 🙏 Credits
- [pedantix](https://pedantix.certitudes.org/) - Original game idea
- [wikipedia](https://wikipedia.org) - Random page content
- [sajari/word2vec](github.com/sajari/word2vec) - Go library for performing computations in word2vec binary models 
- [melody](https://github.com/olahol/melody/) - 🎶 Minimalist websocket framework for Go
- [lucide](https://lucide.dev) - Beautiful & consistent icons made by the community.
- [Jean-Philippe Fauconnier](https://fauconnier.github.io/#data) - French word2vec binaries
