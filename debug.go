package main

import (
	"encoding/json"
	"fmt"
	"net/http"
)

func DebugPrintAppStateHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Debug: app state requested")
	res, err := json.Marshal(state)
	if err != nil {
		fmt.Println(err.Error())
	}
	w.Write(res)
}
