package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/slack-go/slack"
	"log"
	"net/http"
	"strings"
)

var responseUrl string

var client *slack.Client = slack.New("xoxb-3067711698450-4645480503873-qoIqRE703fst4OIlcI1KFn6K")

type ViewSubmission struct {
	Type string `json:"type"`
	Team struct {
		ID     string `json:"id"`
		Domain string `json:"domain"`
	} `json:"team"`
	User struct {
		ID       string `json:"id"`
		Username string `json:"username"`
		Name     string `json:"name"`
		TeamID   string `json:"team_id"`
	} `json:"user"`
	APIAppID  string `json:"api_app_id"`
	Token     string `json:"token"`
	TriggerID string `json:"trigger_id"`
	View      struct {
		ID     string `json:"id"`
		TeamID string `json:"team_id"`
		Type   string `json:"type"`
		Blocks []struct {
			Type    string `json:"type"`
			BlockID string `json:"block_id"`
			Label   struct {
				Type  string `json:"type"`
				Text  string `json:"text"`
				Emoji bool   `json:"emoji"`
			} `json:"label"`
			Hint struct {
				Type  string `json:"type"`
				Text  string `json:"text"`
				Emoji bool   `json:"emoji"`
			} `json:"hint"`
			Optional       bool `json:"optional"`
			DispatchAction bool `json:"dispatch_action"`
			Element        struct {
				Type        string `json:"type"`
				ActionID    string `json:"action_id"`
				Placeholder struct {
					Type  string `json:"type"`
					Text  string `json:"text"`
					Emoji bool   `json:"emoji"`
				} `json:"placeholder"`
				DispatchActionConfig struct {
					TriggerActionsOn []string `json:"trigger_actions_on"`
				} `json:"dispatch_action_config"`
			} `json:"element,omitempty"`
			Element0 struct {
				Type        string `json:"type"`
				ActionID    string `json:"action_id"`
				Placeholder struct {
					Type  string `json:"type"`
					Text  string `json:"text"`
					Emoji bool   `json:"emoji"`
				} `json:"placeholder"`
				Multiline            bool `json:"multiline"`
				DispatchActionConfig struct {
					TriggerActionsOn []string `json:"trigger_actions_on"`
				} `json:"dispatch_action_config"`
			} `json:"element,omitempty"`
		} `json:"blocks"`
		PrivateMetadata string `json:"private_metadata"`
		CallbackID      string `json:"callback_id"`
		State           struct {
			Values struct {
				Num4007890 struct {
					SlInput struct {
						Type  string `json:"type"`
						Value string `json:"value"`
					} `json:"sl_input"`
				} `json:"4007890"`
				Num4007891 struct {
					MlInput struct {
						Type  string `json:"type"`
						Value string `json:"value"`
					} `json:"ml_input"`
				} `json:"4007891"`
			} `json:"values"`
		} `json:"state"`
		Hash  string `json:"hash"`
		Title struct {
			Type  string `json:"type"`
			Text  string `json:"text"`
			Emoji bool   `json:"emoji"`
		} `json:"title"`
		ClearOnClose  bool        `json:"clear_on_close"`
		NotifyOnClose bool        `json:"notify_on_close"`
		Close         interface{} `json:"close"`
		Submit        struct {
			Type  string `json:"type"`
			Text  string `json:"text"`
			Emoji bool   `json:"emoji"`
		} `json:"submit"`
		PreviousViewID     interface{} `json:"previous_view_id"`
		RootViewID         string      `json:"root_view_id"`
		AppID              string      `json:"app_id"`
		ExternalID         string      `json:"external_id"`
		AppInstalledTeamID string      `json:"app_installed_team_id"`
		BotID              string      `json:"bot_id"`
	} `json:"view"`
	ResponseUrls        []interface{} `json:"response_urls"`
	IsEnterpriseInstall bool          `json:"is_enterprise_install"`
	Enterprise          interface{}   `json:"enterprise"`
}

var viewCreateJSON string = `
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

func actionHandle(w http.ResponseWriter, r *http.Request) {
	// Parse the request body to get the command text
	r.ParseForm()

	// log incoming requests
	JSONBody := r.PostForm["payload"][0]
	log.Printf("Request URI: %s", r.RequestURI)
	log.Printf("Request body: %s", JSONBody)
	log.Printf("Request header: %v", r.Header)

	// Decode JSON body
	bodyReader := strings.NewReader(JSONBody)
	var viewSubmission ViewSubmission
	err := json.NewDecoder(bodyReader).Decode(&viewSubmission)
	if err != nil {
		http.Error(w, err.Error(), 400)
		return
	}

	// Respond Successfully
	w.WriteHeader(200)

	responseJson := fmt.Sprintf(`{
	"blocks": [
		{
			"type": "section",
			"text": {
				"type": "mrkdwn",
				"text": ":catjam: Issue %s created!\nThanks for adding another one :catshake:\n<%s| Click here>:point_left: to view the task"
			}
		}
	]
}`, "test", "test")

	resp, err := http.Post(responseUrl, "application/json", bytes.NewBuffer([]byte(responseJson)))
	if err != nil {
		log.Println(err)
	} else {
		log.Println(resp.Status)
	}
}

func slashCmdHandle(w http.ResponseWriter, r *http.Request) {
	// Parse the request body to get the command text
	r.ParseForm()

	// log incoming requests
	log.Printf("Request URI: %s", r.RequestURI)
	log.Printf("Request body: %s", r.PostForm)
	log.Printf("Request header: %v", r.Header)

	// get values from request
	r.ParseForm()
	commandText := r.Form.Get("text")
	commandUser := r.Form.Get("user_name")
	triggerID := r.Form.Get("trigger_id")
	responseUrl = r.Form.Get("response_url")

	// creating initial view
	var viewCreateRequest slack.ModalViewRequest
	if err := json.Unmarshal([]byte(viewCreateJSON), &viewCreateRequest); err != nil {
		log.Println(err)
		return
	}

	args := strings.Fields(commandText)
	if len(args) != 1 {
		log.Printf("Error: Argument requirements were not fulfilled! Slack user: %s", commandUser)
		fmt.Fprint(w, "Error: Argument requirements were not fulfilled!")
		//TODO: Help display function here
		return
	}

	switch args[0] {

	case "create":
		openView, err := client.OpenView(triggerID, viewCreateRequest)
		if err != nil {
			log.Printf("Error opening view: %s\n", err)
			return
		}
		log.Printf("View opened successfully: %s\n", openView.ExternalID)

	default:
		fmt.Fprint(w, "Invalid argument")
		//TODO: Help display function here
		return
	}
}

func main() {

	http.HandleFunc("/action", actionHandle)
	http.HandleFunc("/test", slashCmdHandle)
	http.ListenAndServe(":80", nil)

}

//Request body: map[payload:[{"type":"view_submission","team":{"id":"T031ZLXLJD8","domain":"azintelecomgroup"},"user":{"id":"U033TG7F0QG","username":"tarlan.huseynov512","name":"tarlan.huseynov512","team_id":"T031ZLXLJD8"},"api_app_id":"A04JLQ9K1QA","token":"QRCbbgaxGSXfHzPJ3h0sSakG","trigger_id":"4668448250768.3067711698450.61b0e8481d80414ec3a442a5e0b20b93","view":{"id":"V04JYN337V0","team_id":"T031ZLXLJD8","type":"modal","blocks":[{"type":"input","block_id":"4007890","label":{"type":"plain_text","text":"Label","emoji":true},"hint":{"type":"plain_text","text":"Hint text","emoji":true},"optional":false,"dispatch_action":false,"element":{"type":"plain_text_input","action_id":"sl_input","placeholder":{"type":"plain_text","text":"Placeholder text for single-line input","emoji":true},"dispatch_action_config":{"trigger_actions_on":["on_enter_pressed"]}}},{"type":"input","block_id":"4007891","label":{"type":"plain_text","text":"Label","emoji":true},"hint":{"type":"plain_text","text":"Hint text","emoji":true},"optional":false,"dispatch_action":false,"element":{"type":"plain_text_input","action_id":"ml_input","placeholder":{"type":"plain_text","text":"Placeholder text for multi-line input","emoji":true},"multiline":true,"dispatch_action_config":{"trigger_actions_on":["on_enter_pressed"]}}}],"private_metadata":"","callback_id":"","state":{"values":{"4007890":{"sl_input":{"type":"plain_text_input","value":"qwe"}},"4007891":{"ml_input":{"type":"plain_text_input","value":"qwe"}}}},"hash":"1673733919.dxcfXsW2","title":{"type":"plain_text","text":"Modal Title","emoji":true},"clear_on_close":false,"notify_on_close":false,"close":null,"submit":{"type":"plain_text","text":"Submit","emoji":true},"previous_view_id":null,"root_view_id":"V04JYN337V0","app_id":"A04JLQ9K1QA","external_id":"","app_installed_team_id":"T031ZLXLJD8","bot_id":"B04JE77FAAJ"},"response_urls":[],"is_enterprise_install":false,"enterprise":null}]]
// we will use func (*Client) UpdateViewContext
// we may need to add  w.WriteHeader(200) before updating
