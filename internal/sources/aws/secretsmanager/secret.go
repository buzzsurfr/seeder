package secretsmanager

import (
	"fmt"
	"io"
	"io/ioutil"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/secretsmanager"
)

// Secret represents a seed that sources from an AWS Secrets Manager secret
type Secret struct {
	Name        string
	value       string
	sess        *session.Session
	r           io.ReadCloser
	lastUpdated time.Time
}

// NewSecret creates a new Secret seed
func NewSecret(sess *session.Session, name string) *Parameter {
	secret := Secret{
		Name: name,
		sess: sess,
	}
	secret.fetch()

	return &param
}

func (s *Secret) Read(b []byte) (int, error) {
	s.fetch()

	// Trap io.EOF and reset reader (so that the reader is always ready)
	n, err := s.r.Read(b)
	if err == io.EOF {
		s.r = ioutil.NopCloser(strings.NewReader(s.value))
	}
	return n, err
}

// Close is a wrapper function to meet io.Closer (but is not needed)
func (s *Secret) Close() error {
	return s.r.Close()
}

func (s *Secret) fetch() {
	secretsmanagerSvc := secretsmanager.New(s.sess)

	result, err := secretsmanagerSvc.GetSecretValue(&secretsmanager.GetSecretValueInput{
		Name:           aws.String(s.Name),
	})
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case secretsmanager.ErrCodeResourceNotFoundException:
				fmt.Println(secretsmanager.ErrCodeResourceNotFoundException, aerr.Error())
			case secretsmanager.ErrCodeInvalidParameterException:
				fmt.Println(secretsmanager.ErrCodeInvalidParameterException, aerr.Error())
			case secretsmanager.ErrCodeInvalidRequestException:
				fmt.Println(secretsmanager.ErrCodeInvalidRequestException, aerr.Error())
			case secretsmanager.ErrCodeDecryptionFailure:
				fmt.Println(secretsmanager.ErrCodeDecryptionFailure, aerr.Error())
			case secretsmanager.ErrCodeInternalServiceError:
				fmt.Println(secretsmanager.ErrCodeInternalServiceError, aerr.Error())
			default:
				fmt.Println(aerr.Error())
			}
		} else {
			fmt.Println(err.Error())
		}
	}

	lastModifiedDate := aws.TimeValue(result.CreatedDate)
	if lastModifiedDate.After(s.lastUpdated) {
		s.value = aws.StringValue(result.SecretString)
		s.lastUpdated = lastModifiedDate
		s.r = ioutil.NopCloser(strings.NewReader(s.value))
	}
}
