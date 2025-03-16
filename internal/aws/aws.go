package aws

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

var S3Client *s3.Client

func init() {
	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion("us-west-2"))
	if err != nil {
		panic(fmt.Errorf("unable to load AWS config: %w", err))
	}

	fmt.Println("Successfully loaded credentials...")
	S3Client = s3.NewFromConfig(cfg)
	fmt.Println("Successfully created new s3Client from config...")
}
