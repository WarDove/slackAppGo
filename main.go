package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/andygrunwald/go-jira"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/slack-go/slack"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
)

var slackClient *slack.Client = slack.New(os.Getenv("SLACK_TOKEN"))
var jiraClient *jira.Client = createJiraClient()
var jiraBaseUrl string = os.Getenv("JIRA_URL")

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
		Username: os.Getenv("JIRA_USERNAME"),
		Password: os.Getenv("JIRA_PASSWORD"),
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
				AccountID: os.Getenv("JIRA_ASSIGNEE_ACCOUNT_ID"),
			},
			Type: jira.IssueType{
				Name: os.Getenv("JIRA_ISSUE_TYPE"),
			},
			Description: issueDescription,
			Project: jira.Project{
				Key: os.Getenv("JIRA_PROJECT_KEY"),
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

func lambdaHandler(event events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {

	var responseBody string
	var responseCode int = 200

	// Handle slash command "/task"
	if event.RawPath == "/task" {
		func() {

			base64EncodedBody := event.Body
			urlEncodedBody, err := base64.StdEncoding.DecodeString(base64EncodedBody)
			if err != nil {
				log.Println("Base64 Decode error:", err)
			}

			body, err := url.QueryUnescape(string(urlEncodedBody))
			if err != nil {
				log.Println(err)
			}

			log.Printf("Request URI: %s", event.RawPath)
			log.Printf("Request body: %s", body)
			log.Printf("Request headers: %v", event.Headers)

			urlValues := make(map[string]string)
			for _, pair := range strings.Split(body, "&") {
				kv := strings.Split(pair, "=")
				urlValues[kv[0]] = kv[1]
			}

			commandText := urlValues["text"]
			commandUserName := urlValues["user_name"]
			commandUserID := urlValues["user_id"]
			responseUrl := urlValues["response_url"]

			args := strings.Fields(commandText)
			if len(args) != 1 {
				log.Printf("Error: Argument requirements were not fulfilled! Slack user: %s", commandUserName)
				responseBody = "Error: Argument requirements were not fulfilled!\n`/task create` to create a new issue\n`/task list` to list created issues"
				return
			}

			switch args[0] {

			case "create":

				// creating view
				triggerID := urlValues["trigger_id"]
				var viewCreateRequest slack.ModalViewRequest
				if err := json.Unmarshal([]byte(viewCreateJSON), &viewCreateRequest); err != nil {
					log.Println(err)
					return
				}

				// opening view
				openView, err := slackClient.OpenView(triggerID, viewCreateRequest)
				if err != nil {
					log.Printf("Error opening view: %s\n", err)
					responseCode = 500
					return
				}
				log.Printf("View opened successfully, ID %s\n", openView.View.ID)

			case "list":
				slackUsername := getSlackUserName(commandUserID)
				JQLQuery := fmt.Sprintf("project = '%s' AND summary ~ 'Author: %s' ORDER BY created DESC", os.Getenv("JIRA_PROJECT_KEY"), slackUsername)

				issues, _, err := jiraClient.Issue.Search(JQLQuery, nil)
				if err != nil {
					log.Fatal(err)
				}

				if len(issues) > 0 {
					listOutput := ""
					for _, issue := range issues {
						issueUrl := jiraBaseUrl + "/browse/" + issue.Key
						listOutput += fmt.Sprintf(responseList, issue.Key, issue.Fields.Summary, issue.Fields.Status.Name, issueUrl)
					}

					responseJson := responseBegin + listOutput + responseEnd

					resp, err := http.Post(responseUrl, "application/json", bytes.NewBuffer([]byte(responseJson)))
					if err != nil {
						log.Println(err)
					} else {
						log.Println(resp.Status)
					}

				} else {
					responseBody = "There are no active tasks :catshake:"
					return
				}

			default:
				responseBody = "Invalid argument\n`/task create` to create a new issue\n`/task list` to list created issues"
				return
			}
		}()

	} else if event.RawPath == "/action" {
		func() {
			base64EncodedBody := event.Body
			urlEncodedBody, err := base64.StdEncoding.DecodeString(base64EncodedBody)
			if err != nil {
				log.Println("Base64 Decode error:", err)
			}

			body, err := url.QueryUnescape(string(urlEncodedBody))
			if err != nil {
				log.Println("Unescape error:", err)
			}

			JSONBody := strings.Replace(body, "payload=", "", -1)

			// log incoming requests
			log.Printf("Request URI: %s", event.RawPath)
			log.Printf("Request body: %s", body)
			log.Printf("Request headers: %v", event.Headers)

			// Decode JSON body
			bodyReader := strings.NewReader(JSONBody)
			var viewSubmission ViewSubmission
			err = json.NewDecoder(bodyReader).Decode(&viewSubmission)
			if err != nil {
				responseCode = 500
				return
			}
			// Create issue
			issueSummary := viewSubmission.View.State.Values.Summary.SlInput.Value
			issueDescription := viewSubmission.View.State.Values.Description.MlInput.Value
			slackUsername := getSlackUserName(viewSubmission.User.ID)
			issueKey, issueUrl := createJiraIssue(issueSummary, issueDescription, slackUsername)
			log.Printf("Issue %v successfully created", issueKey)

			responseJson := fmt.Sprintf(createTaskResponse, issueKey, viewSubmission.User.Username, issueUrl)

			resp, err := http.Post(os.Getenv("SLACK_WEBHOOK"), "application/json", bytes.NewBuffer([]byte(responseJson)))
			if err != nil {
				log.Println(err)
			} else {
				log.Println(resp.Status)
			}
		}()
	}

	return events.APIGatewayV2HTTPResponse{
		StatusCode: responseCode,
		Body:       responseBody,
	}, nil
}

func main() {
	lambda.Start(lambdaHandler)
}
