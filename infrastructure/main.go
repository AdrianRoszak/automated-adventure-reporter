package main

import (
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/dynamodb"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/lambda"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/sns"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {
		// Import existing DynamoDB table for Strava activities
		stravaTable, err := dynamodb.NewTable(ctx, "stravaActivities", &dynamodb.TableArgs{
			Name:        pulumi.String("StravaActivities"),
			BillingMode: pulumi.String("PAY_PER_REQUEST"),
			HashKey:     pulumi.String("pk"),
			RangeKey:    pulumi.String("sk"),
			Attributes: dynamodb.TableAttributeArray{
				&dynamodb.TableAttributeArgs{
					Name: pulumi.String("pk"),
					Type: pulumi.String("S"),
				},
				&dynamodb.TableAttributeArgs{
					Name: pulumi.String("sk"),
					Type: pulumi.String("S"),
				},
			},
			PointInTimeRecovery: &dynamodb.TablePointInTimeRecoveryArgs{
				Enabled: pulumi.Bool(false),
			},
			TableClass: pulumi.String("STANDARD"),
		}, pulumi.Protect(true))
		if err != nil {
			return err
		}

		// Import existing SNS topic for Strava notifications
		stravaNotifications, err := sns.NewTopic(ctx, "stravaNotifications", &sns.TopicArgs{
			Name: pulumi.String("StravaNotifications"),
		}, pulumi.Protect(true))
		if err != nil {
			return err
		}

		// Import existing Lambda function for Strava webhook handler
		_, err = lambda.NewFunction(ctx, "stravaWebhookHandler", &lambda.FunctionArgs{
			Name:        pulumi.String("arn:aws:lambda:eu-central-1:891377403759:function:stravaWebhookHandler"),
			Role:        pulumi.String("arn:aws:iam::891377403759:role/service-role/stravaWebhookHandler-role-4sow3cz9"),
			Runtime:     pulumi.String(lambda.RuntimePython3d14),
			Handler:     pulumi.String("lambda_function.lambda_handler"),
			PackageType: pulumi.String("Zip"),
			Architectures: pulumi.StringArray{
				pulumi.String("x86_64"),
			},
			CodeSha256: pulumi.String("ZfkUx8xRoNXNRTvM4x8cbOgkY57a+iVkb+rju7/ugzw="),
			Environment: &lambda.FunctionEnvironmentArgs{
				Variables: pulumi.StringMap{
					"DYNAMODB_TABLE_NAME":  stravaTable.Name,
					"SNS_TOPIC_ARN":        stravaNotifications.Arn,
					"STRAVA_CLIENT_ID":     pulumi.String("233218"),
					"STRAVA_CLIENT_SECRET": pulumi.String("e1e180c771372efbb9fdd87722ff718605f19f89"),
					"STRAVA_REFRESH_TOKEN": pulumi.String("51afe0827a69bb7536aae664f5910de24fd910cc"),
				},
			},
			EphemeralStorage: &lambda.FunctionEphemeralStorageArgs{
				Size: pulumi.Int(512),
			},
			LoggingConfig: &lambda.FunctionLoggingConfigArgs{
				LogFormat: pulumi.String("Text"),
				LogGroup:  pulumi.String("/aws/lambda/stravaWebhookHandler"),
			},
			TracingConfig: &lambda.FunctionTracingConfigArgs{
				Mode: pulumi.String("PassThrough"),
			},
		}, pulumi.Protect(true))
		if err != nil {
			return err
		}

		// Export resource references for other stacks
		ctx.Export("stravaTableName", stravaTable.Name)
		ctx.Export("stravaTableArn", stravaTable.Arn)
		ctx.Export("stravaNotificationsArn", stravaNotifications.Arn)

		return nil
	})
}
