package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/jessevdk/go-flags"
)

var gOpts struct {
	Verbose []bool `short:"v" long:"verbose" description:"Show verbose debug information"`
}

var parser = flags.NewParser(&gOpts, flags.Default)

// Downloads an item from an S3 Bucket
//
// Usage:
//    go run s3_download.go BUCKET ITEM
func main() {
	var err error
	var args []string
	if args, err = parser.Parse(); err != nil {
		if flagsErr, ok := err.(*flags.Error); ok && flagsErr.Type == flags.ErrHelp {
			os.Exit(0)
		} else {
			fmt.Printf("%v\n", flagsErr)
			os.Exit(1)
		}
	}

	cfg := aws.NewConfig()
	cfg.Region = aws.String("eu-west-1")
	if len(gOpts.Verbose) > 0 {
		cfg.WithCredentialsChainVerboseErrors(true)
	}

	// Initialize a session in us-west-2 that the SDK will use to load
	// credentials from the shared credentials file ~/.aws/credentials.
	sess, _ := session.NewSession(cfg)

	// Create S3 service client
	svc := s3.New(sess)

	for i := 0; i < len(args); i += 2 {

		req, _ := svc.GetObjectRequest(&s3.GetObjectInput{
			Bucket: aws.String(args[i]),
			Key:    aws.String(args[i+1]),
		})
		urlStr, err := req.Presign(5 * 24 * time.Hour)

		if err != nil {
			log.Println("Failed to sign request", err)
		}

		log.Println("The URL is", urlStr)
	}
}
