package main

import (
	"fmt"
	"log"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

// Download downloads the given objectkey from s3
func Download(sess *session.Session, bucketname, objectkey, filename string) (err error) {
	downloader := s3manager.NewDownloader(sess)

	// Create a file to write the S3 Object contents to.
	var f *os.File
	if f, err = os.Create(filename); err != nil {
		return fmt.Errorf("failed to create file %q, %v", filename, err)
	}
	defer f.Close()

	var n int64
	// Write the contents of S3 Object to the file
	if n, err = downloader.Download(f, &s3.GetObjectInput{
		Bucket: aws.String(bucketname),
		Key:    aws.String(objectkey),
	}); err != nil {
		return fmt.Errorf("failed to download file, %v", err)
	}
	log.Printf("file downloaded, %d bytes\n", n)
	return nil
}
