package main

import (
	"log"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go/aws/session"
)

func TestUpload(t *testing.T) {
	var sess *session.Session
	var err error
	if sess, err = makeAWSSession(); err != nil {
		t.Fatal(err)
	}

	if err = deleteAllBucket(sess, "aubonbeurref"); err != nil {
		t.Fatal(err)
	}

	if err = Upload(sess, "aubonbeurref/test/", "README.md", "README.md", ""); err != nil {
		t.Fatal(err)
	}

	var dir string
	if dir, err = os.Getwd(); err != nil {
		t.Fatal(err)
	}

	var fileNames []string
	if fileNames, err = getListOfFiles(dir); err != nil {
		t.Fatal(err)
	}

	for _, f := range fileNames {
		if strings.Contains(f, ".git") {
			continue
		}
		path := f[len(dir)+1:]
		log.Printf("%s %s", filepath.Dir(path), filepath.Base(path))
	}
}
