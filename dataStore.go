package main

var responseBegin string = `
{
	"blocks": [
		{
			"type": "section",
			"text": {
				"type": "mrkdwn",
				"text": "*Project: Service Desk*\n*Company: Allwhere*\n*List of created issues*\n*In historical order*"
			},
			"accessory": {
				"type": "image",
				"image_url": "https://media.licdn.com/dms/image/C560BAQFSEYHt0DOivw/company-logo_200_200/0/1656514866210?e=2147483647&v=beta&t=aWpk6b-Eh783Hyx8CKjJSCQz7tqMXLX0RM4XizcW6H4",
				"alt_text": "allwhere"
			}
		},
`

var responseList string = `
		{
			"type": "divider"
		},
		{
			"type": "section",
			"text": {
				"type": "mrkdwn",
				"text": "*Summary*: %s\n*Reported description*: %s...\n*Ticket number*: %s\n*Status*: %s\n*Reported date*: %s"
			}
		},
`

var responseEnd string = `
	]
}
`

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
		"text": "Report Issue"
	},
	"submit": {
		"type": "plain_text",
		"text": "Submit"
	},
	"blocks": [
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
				"text": "Please describe your issue with as much detail as possible."
			}
		}
	],
	"type": "modal"
}
`
var createTaskResponse string = `
{
	"blocks": [
		{
			"type": "section",
			"text": {
				"type": "mrkdwn",
				"text": "New issue *%s* reported by @%s\n\n*Reported description*: %s"
			}
		}
	]
}
`

var helpText = "`/service_desk report` to report a new issue\n`/service_desk status` to list active issues\n`/service_desk comment <IssueKey> <Comment>` to add a new comment\n`/service_desk help` to open help menu"
