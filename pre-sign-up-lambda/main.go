package main

import (
	"context"
	"fmt"
	"strings"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	cognito "github.com/aws/aws-sdk-go/service/cognitoidentityprovider"
)

type Request struct {
	UserAttributes map[string]string `json:"userAttributes"`
}

type MyEvent struct {
	Version       string   `json:"version"`
	TriggerSource string   `json:"triggerSource"`
	Region        string   `json:"region"`
	UserPoolID    string   `json:"userPoolId"`
	UserName      string   `json:"userName"`
	Request       Request  `json:"request"`
	Response      struct{} `json:"response"`
}

func HandleRequest(ctx context.Context, evt MyEvent) (MyEvent, error) {
	if evt.TriggerSource == "PreSignUp_ExternalProvider" {

		// Start Cognito
		conf := &aws.Config{Region: &evt.Region}
		sess, err := session.NewSession(conf)
		if err != nil {
			return evt, fmt.Errorf("could not start session: %w", err)
		}
		cognitoClient := cognito.New(sess)

		// Check if user email exists
		users, err := cognitoClient.ListUsersWithContext(ctx, &cognito.ListUsersInput{
			Filter:     aws.String(fmt.Sprintf(`email = "%s"`, evt.Request.UserAttributes["email"])),
			UserPoolId: &evt.UserPoolID,
		})
		if err != nil {
			return evt, fmt.Errorf("could not list users: %w", err)
		}

		// If no user exists with the email the trigger just returns
		if len(users.Users) == 0 {
			return evt, nil
		}

		// Get the federated sign up parts
		providerName := strings.Split(evt.UserName, "_")[0]
		providerName = cases.Title(language.AmericanEnglish).String(providerName)
		providerUserID := strings.Split(evt.UserName, "_")[1]

		// Link the provider to the existing user
		_, err = cognitoClient.AdminLinkProviderForUserWithContext(
			ctx,
			&cognito.AdminLinkProviderForUserInput{
				DestinationUser: &cognito.ProviderUserIdentifierType{
					ProviderName:           aws.String("Cognito"),
					ProviderAttributeValue: users.Users[0].Username,
				},
				SourceUser: &cognito.ProviderUserIdentifierType{
					ProviderAttributeName:  aws.String("Cognito_Subject"),
					ProviderAttributeValue: &providerUserID,
					ProviderName:           &providerName,
				},
				UserPoolId: &evt.UserPoolID,
			})
		if err != nil {
			return evt, fmt.Errorf("could not link provider for user: %w", err)
		}
	}

	return evt, nil
}

func main() {
	lambda.Start(HandleRequest)
}
