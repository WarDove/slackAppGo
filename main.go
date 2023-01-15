package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/andygrunwald/go-jira"
	"github.com/joho/godotenv"
	"github.com/slack-go/slack"
	"log"
	"net/http"
	"os"
	"strings"
)

var responseUrl string
var slackUsername string

// will be used by multiple functions i.e getSlackUserName, slashCommandHandle
var slackClient *slack.Client = slack.New(GetDotEnv("SLACK_TOKEN"))
var jiraClient *jira.Client = createJiraClient()
var jiraBaseUrl string = GetDotEnv("JIRA_URL")

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
				Summary struct {
					SlInput struct {
						Type  string `json:"type"`
						Value string `json:"value"`
					} `json:"sl_input"`
				} `json:"summary"`
				Description struct {
					MlInput struct {
						Type  string `json:"type"`
						Value string `json:"value"`
					} `json:"ml_input"`
				} `json:"description"`
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
		"text": "Create Ops Issue"
	},
	"submit": {
		"type": "plain_text",
		"text": "Submit"
	},
	"blocks": [
		{
			"block_id": "summary",
			"type": "input",
			"element": {
				"type": "plain_text_input",
				"action_id": "sl_input",
				"placeholder": {
					"type": "plain_text",
					"text": "Brief summary text"
				}
			},
			"label": {
				"type": "plain_text",
				"text": "Summary"
			},
			"hint": {
				"type": "plain_text",
				"text": "A summary of what the issue is about"
			}
		},
		{
			"block_id": "description",
			"type": "input",
			"element": {
				"type": "plain_text_input",
				"action_id": "ml_input",
				"multiline": true,
				"placeholder": {
					"type": "plain_text",
					"text": "Issue description"
				}
			},
			"label": {
				"type": "plain_text",
				"text": "Description"
			},
			"hint": {
				"type": "plain_text",
				"text": "Describe issue and define acceptance criteria for this issue"
			}
		}
	],
	"type": "modal"
}
`

func GetDotEnv(key string) string {
	// load .env file
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatalf("Error loading .env file")
	}

	return os.Getenv(key)
}

func getSlackUserName(slackUserID string) string {
	user, err := slackClient.GetUserInfo(slackUserID)
	if err != nil {
		fmt.Println("Error getting Slack user information:", err)
		return ""
	}
	return user.Profile.RealName
}

func createJiraClient() *jira.Client {
	tp := jira.BasicAuthTransport{
		Username: GetDotEnv("JIRA_USERNAME"),
		Password: GetDotEnv("JIRA_PASSWORD"),
	}

	jiraClient, err := jira.NewClient(tp.Client(), jiraBaseUrl)
	if err != nil {
		log.Fatal(err)
	}
	return jiraClient
}

func createJiraIssue(issueSummary, issueDescription, slackUsername string) (string, string) {

	issue := jira.Issue{
		Fields: &jira.IssueFields{
			Assignee: &jira.User{
				AccountID: GetDotEnv("JIRA_ASSIGNEE_ACCOUNT_ID"),
			},
			Type: jira.IssueType{
				Name: GetDotEnv("JIRA_ISSUE_TYPE"),
			},
			Description: issueDescription,
			Project: jira.Project{
				Key: GetDotEnv("JIRA_PROJECT_KEY"),
			},
			Summary: fmt.Sprintf("%s [Author: %s]", issueSummary, slackUsername),
		},
	}

	createdIssue, resp, err := jiraClient.Issue.Create(&issue)
	if err != nil {
		fmt.Println("Error creating Jira task:", resp.Status, err)
	} else {
		fmt.Println("Jira task created successfully!")
	}

	issueUrl := fmt.Sprintf("%s/browse/%s", jiraBaseUrl, createdIssue.Key)

	return createdIssue.Key, issueUrl
}

func actionHandle(w http.ResponseWriter, r *http.Request) {
	// Parse the request body to get the command text
	r.ParseForm()

	// load .env file
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatalf("Error loading .env file")
	}

	// log incoming requests
	JSONBody := r.PostForm["payload"][0]
	log.Printf("Request URI: %s", r.RequestURI)
	log.Printf("Request body: %s", JSONBody)
	log.Printf("Request header: %v", r.Header)

	// Decode JSON body
	bodyReader := strings.NewReader(JSONBody)
	var viewSubmission ViewSubmission
	err = json.NewDecoder(bodyReader).Decode(&viewSubmission)
	if err != nil {
		http.Error(w, err.Error(), 400)
		return
	}
	// Create issue
	issueSummary := viewSubmission.View.State.Values.Summary.SlInput.Value
	issueDescription := viewSubmission.View.State.Values.Description.MlInput.Value
	slackUsername = getSlackUserName(viewSubmission.User.ID)
	issueKey, issueUrl := createJiraIssue(issueSummary, issueDescription, slackUsername)
	log.Printf("Issue %v successfully created", issueKey)
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
}`, issueKey, issueUrl)

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

	commandText := r.Form.Get("text")
	commandUser := r.Form.Get("user_name")
	responseUrl = r.Form.Get("response_url")

	args := strings.Fields(commandText)
	if len(args) != 1 {
		log.Printf("Error: Argument requirements were not fulfilled! Slack user: %s", commandUser)
		fmt.Fprint(w, "Error: Argument requirements were not fulfilled!")
		//TODO: Help display function here
		return
	}

	switch args[0] {

	case "create":

		// creating view
		triggerID := r.Form.Get("trigger_id")

		var viewCreateRequest slack.ModalViewRequest
		if err := json.Unmarshal([]byte(viewCreateJSON), &viewCreateRequest); err != nil {
			log.Println(err)
			return
		}

		// opening view
		openView, err := slackClient.OpenView(triggerID, viewCreateRequest)
		if err != nil {
			log.Printf("Error opening view: %s\n", err)
			return
		}
		log.Printf("View opened successfully, ID %s\n", openView.View.ID)

	case "list":

		JQLQuery := fmt.Sprintf("project = '%s' AND summary ~ 'Author: %s' ORDER BY created DESC", GetDotEnv("JIRA_PROJECT_KEY"), slackUsername)

		issues, _, err := jiraClient.Issue.Search(JQLQuery, nil)
		if err != nil {
			log.Fatal(err)
		}

		if len(issues) > 0 {
			listOutput := ""
			for _, issue := range issues {
				issueUrl := jiraBaseUrl + "/browse/" + issue.Key
				listOutput += fmt.Sprintf("*Issue [%s]*\nSummary: %s\nStatus: %s\n<%s|Clic here> to view\n_______________________________\n", issue.Key, issue.Fields.Summary, issue.Fields.Status.Name, issueUrl)
			}

			responseJson := fmt.Sprintf(`
{
	"blocks": [
		{
			"type": "section",
			"text": {
				"type": "mrkdwn",
				"text": "*List of created issues*"
			},
			"accessory": {
				"type": "image",
				"image_url": "https://media.licdn.com/dms/image/C560BAQFSEYHt0DOivw/company-logo_200_200/0/1656514866210?e=2147483647&v=beta&t=aWpk6b-Eh783Hyx8CKjJSCQz7tqMXLX0RM4XizcW6H4",
				"alt_text": "allwhere"
			}
		},
		{
			"type": "divider"
		},
		{
			"type": "section",
			"text": {
				"type": "mrkdwn",
				"text": "%s"
			}
		}
	]
}
`, listOutput)

			resp, err := http.Post(responseUrl, "application/json", bytes.NewBuffer([]byte(responseJson)))
			if err != nil {
				log.Println(err)
			} else {
				log.Println(resp.Status)
			}

		} else {
			fmt.Fprint(w, "There are no active tasks :catshake:")
		}

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
