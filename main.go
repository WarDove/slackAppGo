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

//Request body: map[payload:[{"type":"view_submission","team":{"id":"T031ZLXLJD8","domain":"azintelecomgroup"},"user":{"id":"U033TG7F0QG","username":"tarlan.huseynov512","name":"tarlan.huseynov512","team_id":"T031ZLXLJD8"},"api_app_id":"A04JLQ9K1QA","token":"QRCbbgaxGSXfHzPJ3h0sSakG","trigger_id":"4668448250768.3067711698450.61b0e8481d80414ec3a442a5e0b20b93","view":{"id":"V04JYN337V0","team_id":"T031ZLXLJD8","type":"modal","blocks":[{"type":"input","block_id":"4007890","label":{"type":"plain_text","text":"Label","emoji":true},"hint":{"type":"plain_text","text":"Hint text","emoji":true},"optional":false,"dispatch_action":false,"element":{"type":"plain_text_input","action_id":"sl_input","placeholder":{"type":"plain_text","text":"Placeholder text for single-line input","emoji":true},"dispatch_action_config":{"trigger_actions_on":["on_enter_pressed"]}}},{"type":"input","block_id":"4007891","label":{"type":"plain_text","text":"Label","emoji":true},"hint":{"type":"plain_text","text":"Hint text","emoji":true},"optional":false,"dispatch_action":false,"element":{"type":"plain_text_input","action_id":"ml_input","placeholder":{"type":"plain_text","text":"Placeholder text for multi-line input","emoji":true},"multiline":true,"dispatch_action_config":{"trigger_actions_on":["on_enter_pressed"]}}}],"private_metadata":"","callback_id":"","state":{"values":{"4007890":{"sl_input":{"type":"plain_text_input","value":"qwe"}},"4007891":{"ml_input":{"type":"plain_text_input","value":"qwe"}}}},"hash":"1673733919.dxcfXsW2","title":{"type":"plain_text","text":"Modal Title","emoji":true},"clear_on_close":false,"notify_on_close":false,"close":null,"submit":{"type":"plain_text","text":"Submit","emoji":true},"previous_view_id":null,"root_view_id":"V04JYN337V0","app_id":"A04JLQ9K1QA","external_id":"","app_installed_team_id":"T031ZLXLJD8","bot_id":"B04JE77FAAJ"},"response_urls":[],"is_enterprise_install":false,"enterprise":null}]]
// we will use func (*Client) UpdateViewContext
// we may need to add  w.WriteHeader(200) before updating
