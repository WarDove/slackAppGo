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
	commandText := r.Form.Get("text")
	commandUser := r.Form.Get("user_name")

	createJson := fmt.Sprintf(`{
        "blocks": [
                {
                        "type": "section",
                        "text": {
                                "type": "mrkdwn",
                                "text": ":tada: Test %s is successful!!!\n<https://huseynov.net| Click here> to view"
                        }
                }
        ]
}`, "test")

	args := strings.Fields(commandText)
	if len(args) > 1 {
		log.Printf("Error: too many arguments, Slack user: %s", commandUser)
		fmt.Fprint(w, "Error: too many arguments")
		return
	}

	switch args[0] {

	case "create":
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, createJson)
	}
}

func main() {

	http.HandleFunc("/test", slashCmdHandle)
	http.ListenAndServe(":80", nil)

}
