package s3url

import (
	"github.com/pkg/errors"
	"net/url"
	"regexp"
	"strings"
)

type S3Config struct {
	AccessKeyId  string
	SecretKey    string
	Bucket       string
	Prefix       string
	Endpoint     string
	EndpointHost string
	Params       url.Values
}

// Validate the S3Config
func (s3Config S3Config) Validate() error {
	if s3Config.AccessKeyId == "" {
		return errors.New("s3Config.AccessKeyId must not be empty")
	}
	if s3Config.SecretKey == "" {
		return errors.New("s3Config.SecretKey must not be empty")
	}
	if s3Config.Bucket == "" {
		return errors.New("s3Config.BucketName must not be empty")
	}
	if s3Config.Endpoint == "" {
		return errors.New("s3Config.Endpoint must not be empty")
	}
	return nil
}

// Parse takes a s3://accesskey:secretket@endpoint/bucket/...prefix and returns a S3Config
// the accesskey and secret key can be wrapped with [ and ] to allow for special characters
func Parse(value string) (S3Config, error) {
	var s3Config S3Config
	var err error

	// Don't get cute with parsing, just swap the custom stuff (bracket pairs) with url encoded values and hand off to url.Parse
	accessKeyRegex := regexp.MustCompile(`s3://(\[.+?\]):`)
	secretKeyRegex := regexp.MustCompile(`s3://.+?:(\[.+?\])@`)
	encodedUrn := accessKeyRegex.ReplaceAllStringFunc(value, func(wrappedKey string) string {
		res := accessKeyRegex.FindStringSubmatch(wrappedKey)
		if len(res) < 2 {
			return wrappedKey
		}
		key := strings.Replace(wrappedKey, res[1], url.QueryEscape(strings.Trim(res[1], "[]")), 1)
		return key
	})
	encodedUrn = secretKeyRegex.ReplaceAllStringFunc(encodedUrn, func(wrappedKey string) string {
		res := secretKeyRegex.FindStringSubmatch(wrappedKey)
		if len(res) < 2 {
			return wrappedKey
		}
		key := strings.Replace(wrappedKey, res[1], url.QueryEscape(strings.Trim(res[1], "[]")), 1)
		return key
	})

	// Parse the URN using the url package.
	parsedUrl, err := url.Parse(encodedUrn)
	if err != nil {
		return s3Config, errors.Wrap(err, "failed to parse the URN")
	}

	if parsedUrl.Scheme != "s3" {
		return s3Config, errors.New("invalid scheme in the URN. Expecting s3://")
	}

	// Extract credentials from the URL
	accessKeyID, secretKey := "", ""
	if parsedUrl.User != nil {
		accessKeyID = parsedUrl.User.Username()
		var isSet bool
		secretKey, isSet = parsedUrl.User.Password()
		if !isSet {
			return s3Config, errors.New("missing secret key in the URN")
		}
	}

	// Split the path to separate the bucket name and the prefix
	pathParts := strings.SplitN(strings.Trim(parsedUrl.Path, "/"), "/", 2)
	if len(pathParts) == 0 {
		return s3Config, errors.New("missing bucket name in the URN")
	}
	bucketName := pathParts[0]
	bucketPrefix := ""
	if len(pathParts) > 1 {
		bucketPrefix = pathParts[1]
		// add back prefix trailing slash if it was there
		if parsedUrl.Path[len(parsedUrl.Path)-1] == '/' {
			bucketPrefix += "/"
		}
	}

	values := parsedUrl.Query()

	// check that prefix ends in a slash
	if bucketPrefix != "" && bucketPrefix[len(bucketPrefix)-1] != '/' {
		if values.Get("anyPrefix") == "" {
			return s3Config, errors.New("prefix must end with a slash, set anyPrefix=1 on the url to allow it.")
		}
	}

	values.Del("anyPrefix")
	// Populate the S3Config struct with the extracted and decoded values
	s3Config = S3Config{
		AccessKeyId:  accessKeyID,
		SecretKey:    secretKey,
		Bucket:       bucketName,
		Prefix:       bucketPrefix,
		Endpoint:     "https://" + parsedUrl.Host,
		EndpointHost: parsedUrl.Host,
		Params:       values,
	}

	// Validate the configuration
	if err = s3Config.Validate(); err != nil {
		return s3Config, err
	}

	return s3Config, nil
}
