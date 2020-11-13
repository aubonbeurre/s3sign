package main

import (
	"fmt"
	"log"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

// Upload uploads the given filename to s3.
func Upload(sess *session.Session, bucketname, objectkey, filename string, kmskey string) error {
	uploader := s3manager.NewUploader(sess, func(u *s3manager.Uploader) {
		u.Concurrency = 10
	})

	f, err := os.Open(filename)
	if err != nil {
		return fmt.Errorf("failed to open file %q, %v", filename, err)
	}
	defer f.Close()

	uploadInput := s3manager.UploadInput{
		Bucket: aws.String(bucketname),
		Key:    aws.String(objectkey),
		Body:   f,
	}

	// aws kms key set, use aws kms server side encryption
	if kmskey != "" {
		uploadInput.SSEKMSKeyId = aws.String("aws:kms")
		uploadInput.SSEKMSKeyId = aws.String(kmskey)
	}

	// Upload the file to S3.
	result, err := uploader.Upload(&uploadInput)
	if err != nil {
		return fmt.Errorf("failed to upload file, %v", err)
	}

	log.Printf("file uploaded to, %s\n", result.Location)
	return nil
}
