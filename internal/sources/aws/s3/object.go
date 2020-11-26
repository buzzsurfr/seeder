package s3

import (
	"fmt"
	"io"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	awsS3 "github.com/aws/aws-sdk-go/service/s3"
)

// Object is a S3 object seed
type Object struct {
	Bucket      string
	Key         string
	Value       string
	sess        *session.Session
	r           io.ReadCloser
	isRead      bool
	lastUpdated time.Time
}

// NewFromURI creates a new object from a S3 URI
func NewFromURI(sess *session.Session, location string) *Object {
	u, parseErr := ParseString(location)
	if parseErr != nil {
		return &Object{}
	}
	return NewObject(sess, *u.Bucket, *u.Key)
}

// NewObject creates a new object from a bucket and key
func NewObject(sess *session.Session, bucket, key string) *Object {
	obj := Object{
		Bucket: bucket,
		Key:    key,
	}

	obj.fetch()

	return &obj
}

// Read is a wrapper for an io.Reader
func (obj *Object) Read(b []byte) (int, error) {
	if obj.isRead {
		obj.fetch()
	}

	// Trap io.EOF and reset reader (so that the reader is always ready)
	n, err := obj.r.Read(b)
	if err == io.EOF {
		obj.isRead = true
	}
	return n, err
}

// Close is a wrapper for an io.Closer
func (obj *Object) Close() error {
	obj.isRead = true
	return obj.r.Close()
}

func (obj *Object) fetch() {
	s3Svc := awsS3.New(obj.sess)

	result, err := s3Svc.GetObject(&awsS3.GetObjectInput{
		Bucket: aws.String(obj.Bucket),
		Key:    aws.String(obj.Key),
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
			fmt.Println(err.Error())
		}
	}

	lastModifiedDate := aws.TimeValue(result.LastModified)
	if lastModifiedDate.After(obj.lastUpdated) {
		obj.lastUpdated = lastModifiedDate
		obj.r = result.Body
		obj.isRead = false
	}
}
