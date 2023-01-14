package main

import (
	"encoding/json"
	"fmt"
	"github.com/slack-go/slack"
	"log"
	"net/http"
	"strings"
)

var viewJSON string = `
{
	"title": {
		"type": "plain_text",
		"text": "Modal Title"
	},
	"submit": {
		"type": "plain_text",
		"text": "Submit"
	},
	"blocks": [
		{	
			"block_id": "4007890",
			"type": "input",
			"element": {
				"type": "plain_text_input",
				"action_id": "sl_input",
				"placeholder": {
					"type": "plain_text",
					"text": "Placeholder text for single-line input"
				}
			},
			"label": {
				"type": "plain_text",
				"text": "Label"
			},
			"hint": {
				"type": "plain_text",
				"text": "Hint text"
			}
		},
		{
			"block_id": "4007891",
			"type": "input",
			"element": {
				"type": "plain_text_input",
				"action_id": "ml_input",
				"multiline": true,
				"placeholder": {
					"type": "plain_text",
					"text": "Placeholder text for multi-line input"
				}
			},
			"label": {
				"type": "plain_text",
				"text": "Label"
			},
			"hint": {
				"type": "plain_text",
				"text": "Hint text"
			}
		}
	],
	"type": "modal"
}
`

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
	triggerID := r.Form.Get("trigger_id")

	// creating view
	var viewRequest slack.ModalViewRequest
	if err := json.Unmarshal([]byte(viewJSON), &viewRequest); err != nil {
		log.Println(err)
		return
	}

	client := slack.New("xoxb-3067711698450-4645480503873-qoIqRE703fst4OIlcI1KFn6K")

	args := strings.Fields(commandText)
	if len(args) != 1 {
		log.Printf("Error: Argument requirements were not fulfilled! Slack user: %s", commandUser)
		fmt.Fprint(w, "Error: Argument requirements were not fulfilled!")
		//TODO: Help display function here
		return
	}

	switch args[0] {

	case "create":
		openView, err := client.OpenView(triggerID, viewRequest)
		if err != nil {
			log.Printf("Error opening view: %s\n", err)
			return
		}
		log.Printf("View opened successfully: %s\n", openView)

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