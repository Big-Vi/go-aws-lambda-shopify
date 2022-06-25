# Deploy lambda function using AWS CLI

## Prerequisites - AWS CLI & AWS Shell

**AWS Command Line Interface(CLI)** is used in this post to deploy the Lambda function. CLI needs to be [installed](https://docs.aws.amazon.com/cli/latest/userguide/getting-started-install.html) and [configured](https://docs.aws.amazon.com/cli/latest/userguide/cli-chap-configure.html) locally in order to use it.

**[AWS Shell](https://github.com/awslabs/aws-shell)** helps with auto-completion, fuzzy searching, and more.

## How to deploy it on AWS

We're uploading the Go binary file into AWS Lambda as a .zip file. In order to [build for Linux](https://github.com/aws/aws-lambda-go/blob/main/README.md#building-your-function), we're using the following command.

```bash
GOOS=linux GOARCH=amd64 go build -o main main.go
zip main.zip main
```

### Creating IAM roles & policies.

Create the `assume-role-policy.json` file:

```json:assume-role-policy.json
{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Effect": "Allow",
            "Principal": {
                "Service": "lambda.amazonaws.com"
            },
            "Action": "sts:AssumeRole"
        }
    ]
}
```

Create an IAM role

```shell:shell
iam create-role --role-name role-lambda-shopify --assume-role-policy-document file://assume-role-policy.json
```

```json:response
{
    "Role": {
        "Path": "/",
        "RoleName": "role-lambda-shopify",
        "RoleId": "AROARKDOZJ3E6FZQIORPR",
        "Arn": "arn:aws:iam::090426658505:role/role-lambda-shopify",
        "CreateDate": "2022-06-25T21:30:57Z",
        "AssumeRolePolicyDocument": {
            "Version": "2012-10-17",
            "Statement": [
                {
                    "Effect": "Allow",
                    "Principal": {
                        "Service": "lambda.amazonaws.com"
                    },
                    "Action": "sts:AssumeRole"
                }
            ]
        }
    }
}
```

Let's attach permission to this role. **AWSLambdaBasicExecutionRole** gives permission to write logs to cloudWatch.

```shell
iam attach-role-policy --role-name role-lambda-shopify --policy-arn arn:aws:iam::aws:policy/service-role/AWSLambdaBasicExecutionRole
```

Create an SNS topic

````shell
sns create-topic --name shopify-stock-notification


```json:response
{
    "TopicArn": "arn:aws:sns:ap-southeast-2:090426658505:shopify-stock-notification"
}
````

To subscribe to the SNS topic:

```shell
sns subscribe --topic-arn arn:aws:sns:ap-southeast-2:090426658505:shopify-stock-notification --protocol email --notification-endpoint test@hotmail.com
```

we're going to attach an inline policy that grants access to the Lambda function to trigger SNS notification.

```json:sns-policy-for-lambda.json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": "sns:Publish",
      "Resource": "arn:aws:sns:ap-southeast-2:090426658505:shopify-stock-notification"
    }
  ]
}

```

```shell
iam put-role-policy --role-name role-lambda-shopify --policy-name publish-to-sns --policy-document file://sns-policy-for-lambda.json
```

Create/Update Lambda function

```shell
lambda create-function --function-name function-lambda-shopify-stock --zip-file fileb://main.zip --handler main --runtime go1.x --role arn:aws:iam::090426658505:role/role-lambda-shopify
```

```json:response
{
    "FunctionName": "function-lambda-shopify-stock",
    "FunctionArn": "arn:aws:lambda:ap-southeast-2:090426658505:function:function-lambda-shopify-stock",
    "Runtime": "go1.x",
    "Role": "arn:aws:iam::090426658505:role/role-lambda-shopify",
    "Handler": "main",
    "CodeSize": 6662556,
    "Description": "",
    "Timeout": 3,
    "MemorySize": 128,
    "LastModified": "2022-06-25T22:04:02.491+0000",
    "CodeSha256": "uLsOQEp4VCFi294HpbKyl1i60uLGUIBwmpI2cg0wdeU=",
    "Version": "$LATEST",
    "TracingConfig": {
        "Mode": "PassThrough"
    },
    "RevisionId": "1ff52d69-6429-49d8-ba7f-492886e9aee3",
    "State": "Pending",
    "StateReason": "The function is being created.",
    "StateReasonCode": "Creating",
    "PackageType": "Zip",
    "Architectures": [
        "x86_64"
    ],
    "EphemeralStorage": {
        "Size": 512
    }
}
```

To add environmental variable to Lambda function:

```shell
lambda update-function-configuration --function-name function-lambda-shopify-stock --environment "Variables={SHOPIFY_API_KEY=<your-shopify-key>, SHOPIPY_API_PASSWORD=<your-shopify-password>, SHOPIFY_SHOPIFY_DOMAIN=c<your-shopify-domain>}”
```

To update the Lambda function:

```shell
lambda update-function-code --function-name function-lambda-shopify-stock --zip-file fileb://main.zip
```

To invoke the Lambda function and store the output to a text file:

```shell
lambda invoke --function-name function-lambda-shopify-stock output.txt
```

## EventBridge scheduler

EventBridge rule let you create a cron that triggers the Lambda function.

![image](https://cdn.sanity.io/images/bz8z0oa1/production/4667de87e806512057e23e7f117a5c93dca4ffd7-2386x1362.png?w=650)

![image](https://cdn.sanity.io/images/bz8z0oa1/production/be4327fe97398e6457ba3db5b6c669e7ed458a47-2238x1360.png?w=650)

![image](https://cdn.sanity.io/images/bz8z0oa1/production/42366a61e629f40365366152206c6e103e194472-2234x1364.png?w=650)

The code is available in [Github repo](https://github.com/Big-Vi/go-aws-lambda-shopify)
