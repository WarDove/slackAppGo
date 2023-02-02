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
	"github.com/trivago/tgo/tcontainer"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
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

	customFields := tcontainer.NewMarshalMap()
	customFields["customfield_10038"] = issueDescription // Reported Description
	customFields["customfield_10041"] = slackUsername    // Reported By
	currentDate := time.Now().Format("01-02-2006")

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
			Summary:  fmt.Sprintf("%s%s / %s", issueSummary, currentDate, slackUsername),
			Unknowns: customFields,
		},
	}

	createdIssue, resp, err := jiraClient.Issue.Create(&issue)
	if err != nil {
		log.Println("Error creating Jira task:", resp.Status, err)
	} else {
		log.Println("Jira task created successfully!")
	}

	issueUrl := fmt.Sprintf("%s/browse/%s", jiraBaseUrl, createdIssue.Key)

	return createdIssue.Key, issueUrl
}

func lambdaHandler(event events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {

	var responseBody string
	var responseCode int = 200

	// Handle slash command "/service_desk"
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
			slackUsername := getSlackUserName(commandUserID)

			args := strings.Fields(commandText)

			switch args[0] {

			case "report":

				if len(args) != 1 {
					log.Printf("Error: Argument requirements were not fulfilled! Slack user: %s", commandUserName)
					responseBody = "Error: `report` command doesn't need any arguments, enter `/service_desk help` for more info"
					return
				}

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

			case "status":

				if len(args) != 1 {
					log.Printf("Error: Argument requirements were not fulfilled! Slack user: %s", commandUserName)
					responseBody = "Error: `status` command doesn't need any arguments, enter `/service_desk help` for more info"
					return
				}

				JQLQuery := fmt.Sprintf("project = '%s' AND summary ~ '%s' AND status not in ('DONE', 'NO ACTION NEEDED') ORDER BY created DESC", os.Getenv("JIRA_PROJECT_KEY"), slackUsername)

				issues, _, err := jiraClient.Issue.Search(JQLQuery, nil)
				if err != nil {
					log.Fatal(err)
				}

				if len(issues) > 0 {
					listOutput := ""
					for _, issue := range issues {
						issueUrl := jiraBaseUrl + "/browse/" + issue.Key

						descriptionSummary := ""
						if len(issue.Fields.Description) > 50 {
							descriptionSummary = issue.Fields.Description[:50]
						} else {
							descriptionSummary = issue.Fields.Description
						}

						splitSummary := strings.Split(issue.Fields.Summary, " / ")
						reportedDate := splitSummary[0]
						listOutput += fmt.Sprintf(responseList, issue.Fields.Summary, descriptionSummary, issue.Key, issue.Fields.Status.Name, reportedDate, issueUrl)
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

			case "comment":

				if len(args) < 3 {
					log.Printf("Error: Argument requirements were not fulfilled! Slack user: %s", commandUserName)
					responseBody = "Error: `comment` command needs 2 arguments, enter `/service_desk help` for more info"
					return
				}

				issue, response, err := jiraClient.Issue.Get(args[1], nil)
				if err != nil {
					Error := fmt.Sprintf("Error: Issue with ID %s created by %s does not exist", args[1], slackUsername)
					log.Println(Error)
					responseBody = Error
					return
				}

				if response.StatusCode == http.StatusOK {

					// Define the comment to be added
					comment := fmt.Sprintf("%s \n [Author: %s]", strings.Join(args[2:], " "), slackUsername)

					newComment := jira.Comment{
						Body: comment,
					}

					// Add the comment to the issue
					_, _, err := jiraClient.Issue.AddComment(issue.Key, &newComment)

					issueUrl := jiraBaseUrl + "/browse/" + issue.Key

					if err != nil {
						log.Fatalf("Error adding comment to JIRA issue: %s\n", err)
						return
					}
					log.Printf("comment created successfully")
					responseBody = fmt.Sprintf("Comment added to issue: *<%s| %s>*\n\nAdded comment: %s", issueUrl, issue.Key, comment)

				} else if response.StatusCode == http.StatusNotFound {
					Error := fmt.Sprintf("Error: Issue with ID %s created by %s does not exist", args[1], slackUsername)
					log.Println(Error)
					responseBody = Error
					return
				}

			case "help":

				if len(args) != 1 {
					log.Printf("Error: Argument requirements were not fulfilled! Slack user: %s", commandUserName)
					responseBody = "Error: `help` command doesn't need any arguments, enter `/service_desk help` for more info"
					return
				}
				responseBody = helpText

			default:
				responseBody = "Invalid argument\n`/service_desk report` to report a new issue\n`/service_desk status` to list active issues\n`/service_desk help` to open help menu"
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

			descriptionSummary := ""
			if len(issueDescription) > 50 {
				descriptionSummary = issueDescription[:50]
			} else {
				descriptionSummary = issueDescription
			}

			responseJson := fmt.Sprintf(createTaskResponse, issueUrl, issueKey, viewSubmission.User.Username, descriptionSummary)

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
