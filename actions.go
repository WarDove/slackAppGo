package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

type Payload struct {
	Blocks []struct {
		Type    string `json:"type"`
		Element struct {
			Type     string `json:"type"`
			ActionID string `json:"action_id"`
		} `json:"element"`
	} `json:"blocks"`
}

func handleSlackRequest(w http.ResponseWriter, r *http.Request) {
	// Read the request body
	body, _ := ioutil.ReadAll(r.Body)
	//TODO: use r.ParseForm() and r.PostForm["payload"][0] instead
	// Get Json decoded to pointer:
	//func handler(w http.ResponseWriter, r *http.Request) {
	//	var req UserRequest
	//	err := json.NewDecoder(r.Body).Decode(&req)
	//	if err != nil {
	//		http.Error(w, err.Error(), 400)
	//		return
	//	}
	//	// process with the request parameters
	//}

	// Unmarshal the JSON payload
	var payload Payload
	json.Unmarshal(body, &payload)

	// Iterate through the blocks and check the action_id
	for _, block := range payload.Blocks {
		switch block.Element.ActionID {
		case "plain_text_input-action":
			fmt.Println("User entered text in the plain text input element")
		case "static_select-action":
			fmt.Println("User selected an option from the static select element")
		case "button-action":
			fmt.Println("User clicked on the button")
		}
	}
}

// Sample with slack sdk
//func main() {
//	api := slack.New("YOUR_ACCESS_TOKEN_HERE")
//
//	// Build the modal view
//	view := slack.ModalViewRequest{
//		Type: "modal",
//		CallbackID: "modal-identifier",
//		Title: &slack.TextBlockObject{
//			Type: "plain_text",
//			Text: "Just a modal",
//		},
//		Blocks: slack.Blocks{
//			slack.BlockSet{
//				Type: "section",
//				BlockID: "section-identifier",
//				Text: &slack.TextBlockObject{
//					Type: "mrkdwn",
//					Text: "*Welcome* to ~my~ Block Kit _modal_!",
//				},
//				Accessory: &slack.ButtonBlockElement{
//					Type: "button",
//					Text: &slack.TextBlockObject{
//						Type: "plain_text",
//						Text: "Just a button",
//					},
//					ActionID: "button-identifier",
//				},
//			},
//		},
//	}
//
//	triggerID := "156772938.1827394"
//	openView, err := api.OpenView(context.Background(), triggerID, view)
//	if err != nil {
//		fmt.Printf("Error opening view: %s\n", err)
//		return
//	}
//	fmt.Printf("View opened successfully: %s\n", openView)
//}

//view := slack.View{
//Type: "modal",
//Blocks: slack.Blocks{
//BlockSet: viewJSON,
//},
//PrivateMetadata: "",
//CallbackID: "",
//ExternalID: "",
//AppID: "",
//}
