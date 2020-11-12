package s3

import (
	"fmt"
	"io"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	awsS3 "github.com/aws/aws-sdk-go/service/s3"
)

type Object struct {
	Bucket string
	Key    string
	r      io.ReadCloser
}

func NewObject(sess *session.Session, location string) *Object {
	s3u, parseErr := ParseString(location)
	if parseErr != nil {
		return parseErr
	}

	s3Svc := awsS3.New(sess)

	result, err := s3Svc.GetObject(&awsS3.GetObjectInput{
		Bucket: aws.String(s3u.Bucket),
		Key:    aws.String(s3u.Key),
	})
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case awsS3.ErrCodeNoSuchKey:
				fmt.Println(awsS3.ErrCodeNoSuchKey, aerr.Error())
			case awsS3.ErrCodeInvalidObjectState:
				fmt.Println(awsS3.ErrCodeInvalidObjectState, aerr.Error())
			default:
				fmt.Println(aerr.Error())
			}
		} else {
			// Print the error, cast err to awserr.Error to get the Code and
			// Message from an error.
			fmt.Println(err.Error())
		}
		return &Object{}
	}

	return &Object{
		Bucket: s3u.Bucket,
		Key:    s3u.Key,
		r:      result.Body,
	}

}

func (obj *Object) Read(b []byte) (int, error) {
	return obj.r.Read(b)
}

func (obj *Object) Close() error {
	return obj.r.Close()
}
