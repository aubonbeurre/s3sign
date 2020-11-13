package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/jessevdk/go-flags"
)

var gOpts struct {
	Verbose  []bool `short:"v" long:"verbose" description:"Show verbose debug information"`
	List     string `short:"l" long:"list" description:"List bucket content"`
	Upload   string `short:"u" long:"upload" description:"Upload dir to bucket"`
	Delete   string `short:"D" long:"delete" description:"Delete bucket"`
	Download string `short:"d" long:"download" description:"Download bucket"`
	Region   string `short:"r" long:"region" description:"Region bucket" default:"eu-west-1"`
}

var parser = flags.NewParser(&gOpts, flags.Default)

func isDir(path string) bool {
	if info, err := os.Stat(path); err == nil && info.IsDir() {
		return true
	}
	return false
}

func getListOfFiles(dirName string) (fileNames []string, err error) {

	if isDir(dirName) {
		var files []os.FileInfo
		if files, err = ioutil.ReadDir(dirName); err != nil {
			return nil, err
		}
		for _, file := range files {
			if !file.IsDir() {
				fileNames = append(fileNames, filepath.Join(dirName, file.Name()))
			} else {
				var subFilesNames []string
				if subFilesNames, err = getListOfFiles(filepath.Join(dirName, file.Name())); err != nil {
					return nil, err
				}
				fileNames = append(fileNames, subFilesNames...)
			}
		}
	} else {
		fileNames = append(fileNames, dirName)
	}
	return fileNames, nil
}

func deleteAllBucket(sess *session.Session, bucketName string) error {
	svc := s3.New(sess)
	iter := s3manager.NewDeleteListIterator(svc, &s3.ListObjectsInput{
		Bucket: aws.String(bucketName + "/"),
		//MaxKeys: aws.Int64(100),
	})

	return s3manager.NewBatchDeleteWithClient(svc).Delete(aws.BackgroundContext(), iter)
}

func makeAWSSession() (sess *session.Session, err error) {
	cfg := aws.NewConfig()
	cfg.Region = aws.String(gOpts.Region)
	if len(gOpts.Verbose) > 0 {
		cfg.WithCredentialsChainVerboseErrors(true)
	}

	// Initialize a session in us-west-2 that the SDK will use to load
	// credentials from the shared credentials file ~/.aws/credentials.

	return session.NewSession(cfg)
}

func parsePrefix(s string) (params *s3.ListObjectsV2Input) {
	if strings.Contains(s, "/") {
		ar := strings.Split(s, "/")
		prefix := strings.Join(ar[1:], "/") + "/"
		params = &s3.ListObjectsV2Input{
			Bucket: &ar[0],
			Prefix: &prefix,
		}
	} else {
		params = &s3.ListObjectsV2Input{
			Bucket: &s,
		}
	}
	return params
}

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
	log.SetOutput(os.Stdout)

	// Initialize a session in us-west-2 that the SDK will use to load
	// credentials from the shared credentials file ~/.aws/credentials.
	var sess *session.Session
	if sess, err = makeAWSSession(); err != nil {
		panic(err)
	}

	// Create S3 service client
	svc := s3.New(sess)

	if len(gOpts.Delete) > 0 {
		if err = deleteAllBucket(sess, gOpts.Delete); err != nil {
			panic(err)
		}
		log.Println("deleteAllBucket done!")
	}

	if len(gOpts.List) > 0 {
		params := parsePrefix(gOpts.List)

		if err = svc.ListObjectsV2Pages(params, func(resp *s3.ListObjectsV2Output, lastPage bool) bool {
			for _, key := range resp.Contents {
				fmt.Println(*key.Key)
			}
			return true
		}); err != nil {
			panic(err)
		}
		return
	}

	if len(gOpts.Download) > 0 {
		params := parsePrefix(gOpts.Download)

		if err = svc.ListObjectsV2Pages(params, func(resp *s3.ListObjectsV2Output, lastPage bool) bool {
			for _, key := range resp.Contents {
				if info, err := os.Stat(*key.Key); !os.IsNotExist(err) && !info.IsDir() {
					log.Printf("Skipping %s", *key.Key)
					continue
				}
				fmt.Println(*key.Key)
				dir := filepath.Dir(*key.Key)
				os.MkdirAll(dir, 0600)
				bucket := gOpts.Download
				if strings.Contains(bucket, "/") {
					bucket = strings.Split(bucket, "/")[0]
				}
				if err = Download(sess, bucket, *key.Key, *key.Key); err != nil {
					panic(err)
				}
			}
			return true
		}); err != nil {
			panic(err)
		}
		return
	}

	if len(gOpts.Upload) > 0 {
		for _, arg := range args {
			var fileNames []string
			if fileNames, err = getListOfFiles(arg); err != nil {
				panic(err)
			}
			root := filepath.Dir(arg)
			for _, f := range fileNames {
				var path string
				if root != "." {
					path = f[len(root)+1:]
				} else {
					path = f
				}
				log.Printf("%s %s", filepath.Dir(path), filepath.Base(path))
				var prefix string = ""
				if filepath.Dir(path) != "." {
					prefix = filepath.Dir(path) + "/"
				}
				bucket := gOpts.Upload + "/" + prefix
				bucket = strings.Replace(bucket, "\\", "/", -1)
				if err = Upload(sess, bucket, filepath.Base(path), f, ""); err != nil {
					panic(err)
				}
			}
		}
		return
	}

	for i := 0; i < len(args); i += 2 {
		req, _ := svc.GetObjectRequest(&s3.GetObjectInput{
			Bucket: aws.String(args[i]),
			Key:    aws.String(args[i+1]),
		})
		if urlStr, err := req.Presign(5 * 24 * time.Hour); err != nil {
			log.Println("Failed to sign request", err)
		} else {
			log.Println("The URL is", urlStr)
		}
	}
}
