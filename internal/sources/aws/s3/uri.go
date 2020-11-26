package s3

import (
	"errors"
	"fmt"
	"net/url"
	"regexp"
	"strings"
)

// DefaultRegion contains a default region for an S3 bucket, when a region
// cannot be determined, for example when the s3:// schema is used or when
// path style URL has been given without the region component in the
// fully-qualified domain name.
//
// Based on https://gist.github.com/kwilczynski/f6e626990d6d2395b42a12721b165b86#file-main-go
const DefaultRegion = "us-east-1"

var (
	// ErrBucketNotFound is an error when the S3 bucket could not be found
	ErrBucketNotFound = errors.New("bucket name could not be found")
	// ErrHostnameNotFound is an error where the hostname could not be found
	ErrHostnameNotFound = errors.New("hostname could not be found")
	// ErrInvalidS3Endpoint is an error where the S3 endpoint is an invalid URL
	ErrInvalidS3Endpoint = errors.New("an invalid S3 endpoint URL")

	// Pattern used to parse multiple path and host style S3 endpoint URLs.
	s3URLPattern = regexp.MustCompile(`^(.+\.)?s3[.-](?:(accelerated|dualstack|website)[.-])?([a-z0-9-]+)\.`)
)

// URIOpt is the functional options set for a URI
type URIOpt func(*URI)

// WithScheme is a functional options to add a scheme
func WithScheme(s string) URIOpt {
	return func(uri *URI) {
		uri.Scheme = String(s)
	}
}

// WithBucket is a functional options to add a bucket
func WithBucket(s string) URIOpt {
	return func(uri *URI) {
		uri.Bucket = String(s)
	}
}

// WithKey is a functional options to add a key
func WithKey(s string) URIOpt {
	return func(uri *URI) {
		uri.Key = String(s)
	}
}

// WithVersionID is a functional options to add a version ID
func WithVersionID(s string) URIOpt {
	return func(uri *URI) {
		uri.VersionID = String(s)
	}
}

// WithRegion is a functional options to specify the region
func WithRegion(s string) URIOpt {
	return func(uri *URI) {
		uri.Region = String(s)
	}
}

// WithNormalizedKey is a functional options to add a normalized key
func WithNormalizedKey(b bool) URIOpt {
	return func(uri *URI) {
		uri.normalize = Bool(b)
	}
}

// URI is a S3 bucket/key definition based upon the S3 URI
type URI struct {
	uri       *url.URL
	options   []URIOpt
	normalize *bool

	HostStyle   *bool
	PathStyle   *bool
	Accelerated *bool
	DualStack   *bool
	Website     *bool

	Scheme    *string
	Bucket    *string
	Key       *string
	VersionID *string
	Region    *string
}

// NewURI creates a new URI
func NewURI(opts ...URIOpt) *URI {
	return &URI{options: opts}
}

// Reset resets the URI
func (uri *URI) Reset() *URI {
	return reset(uri)
}

// Parse reads the values and transforms into a URI
func (uri *URI) Parse(v interface{}) (*URI, error) {
	return parse(uri, v)
}

// ParseURL reads the URL and parses to a URI
func (uri *URI) ParseURL(u *url.URL) (*URI, error) {
	return parse(uri, u)
}

// ParseString reads the string and parses to a URI
func (uri *URI) ParseString(s string) (*URI, error) {
	return parse(uri, s)
}

// URI returns the URI as a URL
func (uri *URI) URI() *url.URL {
	return uri.uri
}

// Parse reads the values and transforms into a URI
func Parse(v interface{}) (*URI, error) {
	return NewURI().Parse(v)
}

// ParseURL reads the URL and parses to a URI
func ParseURL(u *url.URL) (*URI, error) {
	return NewURI().ParseURL(u)
}

// ParseString reads the string and parses to a URI
func ParseString(s string) (*URI, error) {
	return NewURI().ParseString(s)
}

// MustParse either returns the URI or panics
func MustParse(uri *URI, err error) *URI {
	if err != nil {
		panic(err)
	}
	return uri
}

// Validate specifies whether the passed arguments are a valid URI
func Validate(v interface{}) bool {
	_, err := NewURI().Parse(v)
	return err == nil
}

// ValidateURL specifies whether the URL passed is a valid URI
func ValidateURL(u *url.URL) bool {
	_, err := NewURI().Parse(u)
	return err == nil
}

// ValidateString specifies whether the string passed is a valid URI
func ValidateString(s string) bool {
	_, err := NewURI().Parse(s)
	return err == nil
}

func parse(uri *URI, s interface{}) (*URI, error) {
	var (
		u   *url.URL
		err error
	)

	switch s := s.(type) {
	case string:
		u, err = url.Parse(s)
	case *url.URL:
		u = s
	default:
		return nil, fmt.Errorf("unable to parse unknown type: %T", s)
	}
	if err != nil {
		return nil, fmt.Errorf("unable to parse given S3 endpoint URL: %w", err)
	}

	reset(uri)
	uri.uri = u

	switch u.Scheme {
	case "s3", "http", "https":
		uri.Scheme = String(u.Scheme)
	default:
		return nil, fmt.Errorf("unable to parse schema type: %s", u.Scheme)
	}

	// Handle S3 endpoint URL with the schema s3:// that is neither
	// the host style nor the path style.
	if u.Scheme == "s3" {
		if u.Host == "" {
			return nil, ErrBucketNotFound
		}
		uri.Bucket = String(u.Host)

		if u.Path != "" && u.Path != "/" {
			uri.Key = String(u.Path[1:len(u.Path)])
		}
		uri.Region = String(DefaultRegion)

		return uri, nil
	}

	if u.Host == "" {
		return nil, ErrHostnameNotFound
	}

	matches := s3URLPattern.FindStringSubmatch(u.Host)
	if matches == nil || len(matches) < 1 {
		return nil, ErrInvalidS3Endpoint
	}

	prefix := matches[1]
	usage := matches[2] // Type of the S3 bucket.
	region := matches[3]

	if prefix == "" {
		uri.PathStyle = Bool(true)

		if u.Path != "" && u.Path != "/" {
			u.Path = u.Path[1:len(u.Path)]

			index := strings.Index(u.Path, "/")
			switch {
			case index == -1:
				uri.Bucket = String(u.Path)
			case index == len(u.Path)-1:
				uri.Bucket = String(u.Path[:index])
			default:
				uri.Bucket = String(u.Path[:index])
				uri.Key = String(u.Path[index+1:])
			}
		}
	} else {
		uri.HostStyle = Bool(true)
		uri.Bucket = String(prefix[:len(prefix)-1])

		if u.Path != "" && u.Path != "/" {
			uri.Key = String(u.Path[1:len(u.Path)])
		}
	}

	const (
		// Used to denote type of the S3 bucket.
		accelerated = "accelerated"
		dualStack   = "dualstack"
		website     = "website"

		// Part of the amazonaws.com domain name.  Set when no region
		// could be ascertain correctly using the S3 endpoint URL.
		amazonAWS = "amazonaws"

		// Part of the query parameters.  Used when retrieving S3
		// object (key) of a particular version.
		versionID = "versionId"
	)

	// An S3 bucket can be either accelerated or website endpoint,
	// but not both.
	if usage == accelerated {
		uri.Accelerated = Bool(true)
	} else if usage == website {
		uri.Website = Bool(true)
	}

	// An accelerated S3 bucket can also be dualstack.
	if usage == dualStack || region == dualStack {
		uri.DualStack = Bool(true)
	}

	// Handle the special case of an accelerated dualstack S3
	// endpoint URL:
	//   <BUCKET>.s3-accelerated.dualstack.amazonaws.com/<KEY>.
	// As there is no way to accertain the region solely based on
	// the S3 endpoint URL.
	if usage != accelerated {
		uri.Region = String(DefaultRegion)
		if region != amazonAWS {
			uri.Region = String(region)
		}
	}

	// Query string used when requesting a particular version of a given
	// S3 object (key).
	if s := u.Query().Get(versionID); s != "" {
		uri.VersionID = String(s)
	}

	// Apply options that serve as overrides after the initial parsing
	// is completed.  This allows for bucket name, key, version ID, etc.,
	// to be overridden at the parsing stage.
	for _, o := range uri.options {
		o(uri)
	}

	// Remove trailing slash from the key name, so that the "key/" will
	// become "key" and similarly "a/complex/key/" will simply become
	// "a/complex/key" afer being normalized.
	if BoolValue(uri.normalize) && uri.Key != nil {
		k := StringValue(uri.Key)
		if k[len(k)-1] == '/' {
			k = k[:len(k)-1]
		}
		uri.Key = String(k)
	}

	return uri, nil
}

// Reset fields in the URI type, and set boolean values to false.
func reset(uri *URI) *URI {
	*uri = URI{
		HostStyle:   Bool(false),
		PathStyle:   Bool(false),
		Accelerated: Bool(false),
		DualStack:   Bool(false),
		Website:     Bool(false),
	}
	return uri
}

// String returns a pointer to the string
func String(s string) *string {
	return &s
}

// Bool returns a pointer to the bool
func Bool(b bool) *bool {
	return &b
}

// StringValue returns the value of a string pointer
func StringValue(s *string) string {
	if s != nil {
		return *s
	}
	return ""
}

// BoolValue returns the value of a bool pointer
func BoolValue(b *bool) bool {
	if b != nil {
		return *b
	}
	return false
}
