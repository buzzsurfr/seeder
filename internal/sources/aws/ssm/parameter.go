package ssm

import (
	"fmt"
	"io"
	"io/ioutil"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	awsSsm "github.com/aws/aws-sdk-go/service/ssm"
)

type Parameter struct {
	Name  string
	Value string
	r     io.ReadCloser
}

func NewParameter(sess *session.Session, name string) *Parameter {
	ssmSvc := awsSsm.New(sess)

	result, err := ssmSvc.GetParameter(&awsSsm.GetParameterInput{
		Name:           aws.String(name),
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
		return &Parameter{}
	}

	return &Parameter{
		Name:  name,
		Value: aws.StringValue(result.Parameter.Value),
		r:     ioutil.NopCloser(strings.NewReader(aws.StringValue(result.Parameter.Value))),
	}
}

func (param *Parameter) Read(b []byte) (int, error) {
	return param.r.Read(b)
}

func (param *Parameter) Close() error {
	return param.r.Close()
}
