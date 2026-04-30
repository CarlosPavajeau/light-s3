package awsclient

import (
	"context"
	"fmt"
	"time"

	"light-s3/internal/config"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/cognitoidentity"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// cognitoCredsProvider fetches temporary AWS credentials from a Cognito Identity Pool
// using unauthenticated (guest) access.
type cognitoCredsProvider struct {
	identityPoolID string
	accountID      string
	client         *cognitoidentity.Client
}

func (p *cognitoCredsProvider) Retrieve(ctx context.Context) (aws.Credentials, error) {
	input := &cognitoidentity.GetIdInput{
		IdentityPoolId: aws.String(p.identityPoolID),
	}
	if p.accountID != "" {
		input.AccountId = aws.String(p.accountID)
	}

	idResult, err := p.client.GetId(ctx, input)
	if err != nil {
		return aws.Credentials{}, fmt.Errorf("cognito GetId: %w", err)
	}

	credsResult, err := p.client.GetCredentialsForIdentity(ctx, &cognitoidentity.GetCredentialsForIdentityInput{
		IdentityId: idResult.IdentityId,
	})
	if err != nil {
		return aws.Credentials{}, fmt.Errorf("cognito GetCredentialsForIdentity: %w", err)
	}

	c := credsResult.Credentials
	return aws.Credentials{
		AccessKeyID:     aws.ToString(c.AccessKeyId),
		SecretAccessKey: aws.ToString(c.SecretKey),
		SessionToken:    aws.ToString(c.SessionToken),
		Expires:         aws.ToTime(c.Expiration),
		CanExpire:       true,
		Source:          "CognitoIdentityPool",
	}, nil
}

func NewPresignClient(cfg *config.Config) (*s3.PresignClient, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Anonymous config to bootstrap the Cognito client — unauthenticated pools
	// don't require prior AWS credentials to call GetId/GetCredentialsForIdentity.
	anonymousCfg, err := awsconfig.LoadDefaultConfig(ctx,
		awsconfig.WithRegion(cfg.AWSRegion),
		awsconfig.WithCredentialsProvider(aws.AnonymousCredentials{}),
	)
	if err != nil {
		return nil, fmt.Errorf("load anonymous aws config: %w", err)
	}

	cognitoClient := cognitoidentity.NewFromConfig(anonymousCfg)

	credsProvider := aws.NewCredentialsCache(&cognitoCredsProvider{
		identityPoolID: cfg.IdentityPoolID,
		accountID:      cfg.AWSAccountID,
		client:         cognitoClient,
	})

	s3Cfg, err := awsconfig.LoadDefaultConfig(ctx,
		awsconfig.WithRegion(cfg.AWSRegion),
		awsconfig.WithCredentialsProvider(credsProvider),
	)
	if err != nil {
		return nil, fmt.Errorf("load s3 aws config: %w", err)
	}

	return s3.NewPresignClient(s3.NewFromConfig(s3Cfg)), nil
}
