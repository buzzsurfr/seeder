package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/aws/aws-sdk-go/service/ssm"
	"github.com/buzzsurfr/seeder/internal/s3uri"
)

var (
	outputDir = "/tmp/certificates"
)

const (
	chainFilename = "chain.pem"
	keyFilename   = "key.pem"
)

func downloadFromParameterStore(sess *session.Session, parameterName, filename string) error {
	ssmSvc := ssm.New(sess)

	result, err := ssmSvc.GetParameter(&ssm.GetParameterInput{
		Name:           aws.String(parameterName),
		WithDecryption: aws.Bool(true),
	})
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case ssm.ErrCodeInvalidKeyId:
				fmt.Println(ssm.ErrCodeInvalidKeyId, aerr.Error())
			case ssm.ErrCodeParameterNotFound:
				fmt.Println(ssm.ErrCodeParameterNotFound, aerr.Error())
			case ssm.ErrCodeParameterVersionNotFound:
				fmt.Println(ssm.ErrCodeParameterVersionNotFound, aerr.Error())
			default:
				fmt.Println(aerr.Error())
			}
		} else {
			fmt.Println(err.Error())
		}
		return err
	}
	writeErr := ioutil.WriteFile(filename, []byte(aws.StringValue(result.Parameter.Value)), 0644)
	if writeErr != nil {
		fmt.Println(writeErr)
	}

	return writeErr
}

func downloadFromS3Uri(sess *session.Session, location, filename string) error {
	s3u, parseErr := s3uri.ParseString(location)
	if parseErr != nil {
		return parseErr
	}

	downloader := s3manager.NewDownloader(sess)

	file, createFileErr := os.Create(filename)
	if createFileErr != nil {
		return createFileErr
	}
	_, downloadErr := downloader.Download(file, &s3.GetObjectInput{
		Bucket: s3u.Bucket,
		Key:    s3u.Key,
	})
	return downloadErr
}

func poll(sess *session.Session) {
	// Get certificate chain from SSM and store in file
	if chainParameterStoreName, ok := os.LookupEnv("CHAIN_PARAMETER_STORE_NAME"); ok {
		downloadFromParameterStore(sess, chainParameterStoreName, filepath.Join(outputDir, chainFilename))
	}

	// Get certificate chain from S3 and store in file
	if chainS3Uri, ok := os.LookupEnv("CHAIN_S3URI"); ok {
		downloadFromS3Uri(sess, chainS3Uri, filepath.Join(outputDir, chainFilename))
	}

	// Get private key from SSM and store in file
	if keyParameterStoreName, ok := os.LookupEnv("KEY_PARAMETER_STORE_NAME"); ok {
		downloadFromParameterStore(sess, keyParameterStoreName, filepath.Join(outputDir, keyFilename))
	}

	// Get private key from S3 and store in file
	if keyS3Uri, ok := os.LookupEnv("KEY_S3URI"); ok {
		downloadFromS3Uri(sess, keyS3Uri, filepath.Join(outputDir, keyFilename))
	}
}

func run() {
	// Timer
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	// AWS Session
	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))

	for {
		select {
		case <-ticker.C:
			poll(sess)
		}
	}
}

func main() {
	// Create output directory if it doesn't exist
	if outputDirEnv, ok := os.LookupEnv("OUTPUT_DIR"); ok {
		outputDir = outputDirEnv
	}
	_ = os.Mkdir(outputDir, os.ModeDir|0755)

	// AWS Session
	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))

	poll(sess)

	run()
}
