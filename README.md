# PoC of using AWS Cognito with IDP

Repository used in the following video blog: https://youtu.be/s5VTyRYXMJI

## How to update the Lambda

The Lambda function should be created with the runtime `Amazon Provided Linux 2`. The function will be updated by uploading a zip file with the binary. For this purpose we need to build and zip the code.

```
GOOS=linux GOARCH=amd64 go build -o bootstrap pre-sign-up-lambda/main.go \
&& zip pre-sign-up-lambda.zip bootstrap \
&& rm bootstrap
```

Upload `pre-sign-up-lambda.zip` on the Lambda function UI.

## Test the event on the Lambda UI

Test event:

```json
{
  "triggerSource": "PreSignUp_ExternalProvider",
  "region": "eu-north-1",
  "userPoolId": "eu-north-1_IsWS7SXW4",
  "userName": "Facebook_12324325436",
  "request": {
    "userAttributes": {
      "email": "non@user.com"
    }
  }
}
```

## Policy needed to run the code

```json
{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Sid": "VisualEditor0",
            "Effect": "Allow",
            "Action": [
                "cognito-idp:AdminLinkProviderForUser",
                "cognito-idp:ListUsers"
            ],
            "Resource": "arn-cognito-user-pool"
        }
    ]
}
```