package main

import (
	"fmt"
	"log"
	"net/http"
	"strings"
)

func slashCmdHandle(w http.ResponseWriter, r *http.Request) {
	// Parse the request body to get the command text
	r.ParseForm()

	// log incoming requests
	log.Printf("Request body: %s", r.PostForm)
	log.Printf("Request header: %v", r.Header)

	// get values from request
	r.ParseForm()
	commandText := r.Form.Get("text")
	commandUser := r.Form.Get("user_name")

	args := strings.Fields(commandText)
	if len(args) != 1 {
		log.Printf("Error: Argument requirements were not fulfilled! Slack user: %s", commandUser)
		fmt.Fprint(w, "Error: Argument requirements were not fulfilled!")
		//TODO: Help display function here
		return
	}

	switch args[0] {

	case "create":
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, createJson)
	default:
		fmt.Fprint(w, "Invalid argument")
		//TODO: Help display function here
		return
	}
}

func main() {

	http.HandleFunc("/test", slashCmdHandle)
	http.ListenAndServe(":80", nil)

}
