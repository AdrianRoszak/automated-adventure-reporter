package main

import (
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/apigatewayv2"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/dynamodb"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/lambda"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/s3"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/sns"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi/config"
)

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {
		// =========================================================================
		// Storage Layer
		// =========================================================================

		// DynamoDB table for Strava activities
		stravaTable, err := dynamodb.NewTable(ctx, "stravaActivities", &dynamodb.TableArgs{
			Name:        pulumi.String("StravaActivities"),
			BillingMode: pulumi.String("PAY_PER_REQUEST"),
			HashKey:     pulumi.String("pk"),
			RangeKey:    pulumi.String("sk"),
			Attributes: dynamodb.TableAttributeArray{
				&dynamodb.TableAttributeArgs{Name: pulumi.String("pk"), Type: pulumi.String("S")},
				&dynamodb.TableAttributeArgs{Name: pulumi.String("sk"), Type: pulumi.String("S")},
			},
			PointInTimeRecovery: &dynamodb.TablePointInTimeRecoveryArgs{
				Enabled: pulumi.Bool(false),
			},
			TableClass: pulumi.String("STANDARD"),
		}, pulumi.Protect(true))
		if err != nil {
			return err
		}

		// S3 bucket for raw audio inbound from Telegram
		audioInboundRaw, err := s3.NewBucketV2(ctx, "audioInboundRaw", &s3.BucketV2Args{
			Bucket: pulumi.String("weirdo-audio-inbound-raw"),
			Region: pulumi.String("eu-central-1"),
		}, pulumi.Protect(true))
		if err != nil {
			return err
		}

		// S3 bucket for Deepgram transcription results
		transcriptsStorage, err := s3.NewBucketV2(ctx, "transcriptsStorage", &s3.BucketV2Args{
			Bucket: pulumi.String("weirdo-transcripts-storage"),
			Region: pulumi.String("eu-central-1"),
		}, pulumi.Protect(true))
		if err != nil {
			return err
		}

		// =========================================================================
		// Messaging Layer
		// =========================================================================

		// SNS topic for Strava activity notifications
		stravaNotifications, err := sns.NewTopic(ctx, "stravaNotifications", &sns.TopicArgs{
			Name: pulumi.String("StravaNotifications"),
		}, pulumi.Protect(true))
		if err != nil {
			return err
		}

		// =========================================================================
		// Compute Layer - Lambda Functions
		// =========================================================================

		// Strava webhook handler - receives activity data from Strava
		_, err = lambda.NewFunction(ctx, "stravaWebhookHandler", &lambda.FunctionArgs{
			Name:          pulumi.String("arn:aws:lambda:eu-central-1:891377403759:function:stravaWebhookHandler"),
			Role:          pulumi.String("arn:aws:iam::891377403759:role/service-role/stravaWebhookHandler-role-4sow3cz9"),
			Runtime:       pulumi.String(lambda.RuntimePython3d14),
			Handler:       pulumi.String("lambda_function.lambda_handler"),
			PackageType:   pulumi.String("Zip"),
			Architectures: pulumi.StringArray{pulumi.String("x86_64")},
			CodeSha256:    pulumi.String("ZfkUx8xRoNXNRTvM4x8cbOgkY57a+iVkb+rju7/ugzw="),
			Environment: &lambda.FunctionEnvironmentArgs{
				Variables: pulumi.StringMap{
					"DYNAMODB_TABLE_NAME": stravaTable.Name,
					"SNS_TOPIC_ARN":       stravaNotifications.Arn,
				},
			},
			EphemeralStorage: &lambda.FunctionEphemeralStorageArgs{Size: pulumi.Int(512)},
			LoggingConfig: &lambda.FunctionLoggingConfigArgs{
				LogFormat: pulumi.String("Text"),
				LogGroup:  pulumi.String("/aws/lambda/stravaWebhookHandler"),
			},
			TracingConfig: &lambda.FunctionTracingConfigArgs{Mode: pulumi.String("PassThrough")},
		}, pulumi.Protect(true))
		if err != nil {
			return err
		}

		// Telegram voice downloader - downloads voice messages from Telegram
		_, err = lambda.NewFunction(ctx, "telegramVoiceDownloader", &lambda.FunctionArgs{
			Name:          pulumi.String("arn:aws:lambda:eu-central-1:891377403759:function:telegramVoiceDownloader"),
			Role:          pulumi.String("arn:aws:iam::891377403759:role/service-role/telegramVoiceDownloader-role-kgpao31o"),
			Runtime:       pulumi.String(lambda.RuntimePython3d14),
			Handler:       pulumi.String("lambda_function.lambda_handler"),
			PackageType:   pulumi.String("Zip"),
			Architectures: pulumi.StringArray{pulumi.String("x86_64")},
			CodeSha256:    pulumi.String("UdYcQr5EJhfWbAL/x0VVEzw7NFqFfwOYpwfzD1k2uXk="),
			Timeout:       pulumi.Int(120),
			Environment: &lambda.FunctionEnvironmentArgs{
				Variables: pulumi.StringMap{
					"BUCKET_NAME": audioInboundRaw.Bucket,
				},
			},
			EphemeralStorage: &lambda.FunctionEphemeralStorageArgs{Size: pulumi.Int(512)},
			LoggingConfig: &lambda.FunctionLoggingConfigArgs{
				LogFormat: pulumi.String("Text"),
				LogGroup:  pulumi.String("/aws/lambda/telegramVoiceDownloader"),
			},
			TracingConfig: &lambda.FunctionTracingConfigArgs{Mode: pulumi.String("PassThrough")},
		}, pulumi.Protect(true))
		if err != nil {
			return err
		}

		// Telegram bot token is stored in Pulumi config as a secret.
		telegramToken := config.New(ctx, "").RequireSecret("telegramToken")

		// Telegram to S3 ingestor - receives webhook from Telegram bot
		_, err = lambda.NewFunction(ctx, "telegramToS3Ingestor", &lambda.FunctionArgs{
			Name:          pulumi.String("arn:aws:lambda:eu-central-1:891377403759:function:telegramToS3Ingestor"),
			Role:          pulumi.String("arn:aws:iam::891377403759:role/service-role/telegramToS3Ingestor-role-z25shvc7"),
			Runtime:       pulumi.String(lambda.RuntimePython3d14),
			Handler:       pulumi.String("lambda_function.lambda_handler"),
			PackageType:   pulumi.String("Zip"),
			Architectures: pulumi.StringArray{pulumi.String("x86_64")},
			CodeSha256:    pulumi.String("7pDqad8KRd8dHhxyW/I6MnADhPmknl5SGwBuAXV5b04="),
			Environment: &lambda.FunctionEnvironmentArgs{
				Variables: pulumi.StringMap{
					"BUCKET_NAME":    audioInboundRaw.Bucket,
					"TELEGRAM_TOKEN": telegramToken,
				},
			},
			EphemeralStorage: &lambda.FunctionEphemeralStorageArgs{Size: pulumi.Int(512)},
			LoggingConfig: &lambda.FunctionLoggingConfigArgs{
				LogFormat: pulumi.String("Text"),
				LogGroup:  pulumi.String("/aws/lambda/telegramToS3Ingestor"),
			},
			TracingConfig: &lambda.FunctionTracingConfigArgs{Mode: pulumi.String("PassThrough")},
		}, pulumi.Protect(true))
		if err != nil {
			return err
		}

		// Deepgram API key is stored in Pulumi config as a secret.
		deepgramApiKey := config.New(ctx, "").RequireSecret("deepgramApiKey")

		// Deepgram transcriber - transcribes audio using Deepgram API
		_, err = lambda.NewFunction(ctx, "deepgramTranscriber", &lambda.FunctionArgs{
			Name:          pulumi.String("arn:aws:lambda:eu-central-1:891377403759:function:deepgramTranscriber"),
			Role:          pulumi.String("arn:aws:iam::891377403759:role/service-role/deepgramTranscriber-role-rcsyet8a"),
			Runtime:       pulumi.String(lambda.RuntimePython3d14),
			Handler:       pulumi.String("lambda_function.lambda_handler"),
			PackageType:   pulumi.String("Zip"),
			Architectures: pulumi.StringArray{pulumi.String("x86_64")},
			CodeSha256:    pulumi.String("0cHOR7AWuBCMrXG38Uh97J45RwtIEZm7ikVoGblI1FU="),
			Environment: &lambda.FunctionEnvironmentArgs{
				Variables: pulumi.StringMap{
					"DEEPGRAM_API_KEY":  deepgramApiKey,
					"TRANSCRIPT_BUCKET": transcriptsStorage.Bucket,
				},
			},
			EphemeralStorage: &lambda.FunctionEphemeralStorageArgs{Size: pulumi.Int(512)},
			LoggingConfig: &lambda.FunctionLoggingConfigArgs{
				LogFormat: pulumi.String("Text"),
				LogGroup:  pulumi.String("/aws/lambda/deepgramTranscriber"),
			},
			TracingConfig: &lambda.FunctionTracingConfigArgs{Mode: pulumi.String("PassThrough")},
		}, pulumi.Protect(true))
		if err != nil {
			return err
		}

		// =========================================================================
		// API Layer - API Gateway v2 (HTTP APIs)
		// =========================================================================

		// API Gateway for Strava webhook
		_, err = apigatewayv2.NewApi(ctx, "stravaWebhookApi", &apigatewayv2.ApiArgs{
			Name:          pulumi.String("stravaWebhookHandler-API"),
			ProtocolType:  pulumi.String("HTTP"),
			Description:   pulumi.String("Created by AWS Lambda"),
			IpAddressType: pulumi.String("ipv4"),
		}, pulumi.Protect(true))
		if err != nil {
			return err
		}

		// API Gateway for Telegram ingestor
		_, err = apigatewayv2.NewApi(ctx, "telegramIngestorApi", &apigatewayv2.ApiArgs{
			Name:          pulumi.String("telegramToS3Ingestor-API"),
			ProtocolType:  pulumi.String("HTTP"),
			Description:   pulumi.String("Created by AWS Lambda"),
			IpAddressType: pulumi.String("ipv4"),
		}, pulumi.Protect(true))
		if err != nil {
			return err
		}

		// =========================================================================
		// Stack Outputs
		// =========================================================================

		ctx.Export("stravaTableName", stravaTable.Name)
		ctx.Export("stravaTableArn", stravaTable.Arn)
		ctx.Export("stravaNotificationsArn", stravaNotifications.Arn)
		ctx.Export("audioInboundRawBucket", audioInboundRaw.Bucket)
		ctx.Export("transcriptsStorageBucket", transcriptsStorage.Bucket)

		return nil
	})
}
