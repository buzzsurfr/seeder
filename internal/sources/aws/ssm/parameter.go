package ssm

import (
	"fmt"
	"io"
	"io/ioutil"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	awsSsm "github.com/aws/aws-sdk-go/service/ssm"
)

// Parameter represents a seed that sources from an AWS SSM Parameter
type Parameter struct {
	Name        string
	value       string
	sess        *session.Session
	r           io.ReadCloser
	lastUpdated time.Time
}

// NewParameter creates a new Parameter seed
func NewParameter(sess *session.Session, name string) *Parameter {
	param := Parameter{
		Name: name,
		sess: sess,
	}
	param.fetch()

	return &param
}

func (param *Parameter) Read(b []byte) (int, error) {
	param.fetch()

	// Trap io.EOF and reset reader (so that the reader is always ready)
	n, err := param.r.Read(b)
	if err == io.EOF {
		param.r = ioutil.NopCloser(strings.NewReader(param.value))
	}
	return n, err
}

// Close is a wrapper function to meet io.Closer (but is not needed)
func (param *Parameter) Close() error {
	return param.r.Close()
}

func (param *Parameter) fetch() {
	ssmSvc := awsSsm.New(param.sess)

	result, err := ssmSvc.GetParameter(&awsSsm.GetParameterInput{
		Name:           aws.String(param.Name),
		WithDecryption: aws.Bool(true),
	})
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case awsSsm.ErrCodeInvalidKeyId:
				fmt.Println(awsSsm.ErrCodeInvalidKeyId, aerr.Error())
			case awsSsm.ErrCodeParameterNotFound:
				fmt.Println(awsSsm.ErrCodeParameterNotFound, aerr.Error())
			case awsSsm.ErrCodeParameterVersionNotFound:
				fmt.Println(awsSsm.ErrCodeParameterVersionNotFound, aerr.Error())
			default:
				fmt.Println(aerr.Error())
			}
		} else {
			fmt.Println(err.Error())
		}
	}

	lastModifiedDate := aws.TimeValue(result.Parameter.LastModifiedDate)
	if lastModifiedDate.After(param.lastUpdated) {
		param.value = aws.StringValue(result.Parameter.Value)
		param.lastUpdated = lastModifiedDate
		param.r = ioutil.NopCloser(strings.NewReader(param.value))
	}
}
