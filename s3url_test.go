package s3url

import (
	"github.com/stretchr/testify/require"
	"net/url"
	"testing"
)

func TestParseS3Urn(t *testing.T) {
	tests := []struct {
		name      string
		urn       string
		expect    S3Config
		expectErr bool
	}{
		{
			name: "Valid URN with no url encoding",
			urn:  "s3://accessKey123:secretKey123@endpoint/bucket/prefix/",
			expect: S3Config{
				AccessKeyId:  "accessKey123",
				SecretKey:    "secretKey123",
				Bucket:       "bucket",
				Prefix:       "prefix/",
				Endpoint:     "https://endpoint",
				EndpointHost: "endpoint",
				Params:       make(url.Values),
			},
		},
		{
			name: "Valid URN with url encoded access key and secret key",
			urn:  "s3://%61%63%63%65%73%73%4B%65%79:%73%65%63%72%65%74%4B%65%79@endpoint/bucket/prefix/",
			expect: S3Config{
				AccessKeyId:  "accessKey",
				SecretKey:    "secretKey",
				Bucket:       "bucket",
				Prefix:       "prefix/",
				Endpoint:     "https://endpoint",
				EndpointHost: "endpoint",
				Params:       make(url.Values),
			},
		},
		{
			name: "Valid URN with unsafe URL characters",
			urn:  "s3://[ac=@\\c:e/ss]:[k=?e&y@123]@endpoint/bucket/prefix?anyPrefix=1",
			expect: S3Config{
				AccessKeyId:  "ac=@\\c:e/ss",
				SecretKey:    "k=?e&y@123",
				Bucket:       "bucket",
				Prefix:       "prefix",
				Endpoint:     "https://endpoint",
				EndpointHost: "endpoint",
				Params:       make(url.Values),
			},
		},
		{
			name: "Prefix trailing slash preserved",
			urn:  "s3://[ac=@\\c:e/ss]:[k=?e&y@123]@endpoint/bucket/prefix/",
			expect: S3Config{
				AccessKeyId:  "ac=@\\c:e/ss",
				SecretKey:    "k=?e&y@123",
				Bucket:       "bucket",
				Prefix:       "prefix/",
				Endpoint:     "https://endpoint",
				EndpointHost: "endpoint",
				Params:       make(url.Values),
			},
		},
		{
			name: "Valid URN with no prefix",
			urn:  "s3://accessKey123:secretKey123@endpoint/bucket",
			expect: S3Config{
				AccessKeyId:  "accessKey123",
				SecretKey:    "secretKey123",
				Bucket:       "bucket",
				Prefix:       "",
				Endpoint:     "https://endpoint",
				EndpointHost: "endpoint",
				Params:       make(url.Values),
			},
		},
		{
			name: "Valid URN with multiple prefixes",
			urn:  "s3://accessKey123:secretKey123@endpoint/bucket/prefix/subprefix?anyPrefix=1",
			expect: S3Config{
				AccessKeyId:  "accessKey123",
				SecretKey:    "secretKey123",
				Bucket:       "bucket",
				Prefix:       "prefix/subprefix",
				Endpoint:     "https://endpoint",
				EndpointHost: "endpoint",
				Params:       make(url.Values),
			},
		},
		{
			name: "Valid URN with multiple prefixes and trialing slash",
			urn:  "s3://accessKey123:secretKey123@endpoint/bucket/prefix/subprefix/",
			expect: S3Config{
				AccessKeyId:  "accessKey123",
				SecretKey:    "secretKey123",
				Bucket:       "bucket",
				Prefix:       "prefix/subprefix/",
				Endpoint:     "https://endpoint",
				EndpointHost: "endpoint",
				Params:       make(url.Values),
			},
		},
		{
			name: "Valid URN with special characters in bucket and prefix",
			urn:  "s3://accessKey123:secretKey123@endpoint/bucket-name/prefix-name/",
			expect: S3Config{
				AccessKeyId:  "accessKey123",
				SecretKey:    "secretKey123",
				Bucket:       "bucket-name",
				Prefix:       "prefix-name/",
				Endpoint:     "https://endpoint",
				EndpointHost: "endpoint",
				Params:       make(url.Values),
			},
		},
		{
			name: "Valid URN with port in endpoint",
			urn:  "s3://accessKey123:secretKey123@endpoint:1234/bucket/prefix/",
			expect: S3Config{
				AccessKeyId:  "accessKey123",
				SecretKey:    "secretKey123",
				Bucket:       "bucket",
				Prefix:       "prefix/",
				Endpoint:     "https://endpoint:1234",
				EndpointHost: "endpoint:1234",
				Params:       make(url.Values),
			},
		},
		{
			name: "URN with encoded special chars in path",
			urn:  "s3://accessKey123:secretKey123@endpoint/bucket/%70r%65fix/",
			expect: S3Config{
				AccessKeyId:  "accessKey123",
				SecretKey:    "secretKey123",
				Bucket:       "bucket",
				Prefix:       "prefix/", // assuming auto decoding
				Endpoint:     "https://endpoint",
				EndpointHost: "endpoint",
				Params:       make(url.Values),
			},
		},
		{
			name: "URN with query parameters",
			urn:  "s3://accessKey123:secretKey123@endpoint/bucket/prefix?versionId=123&anyPrefix=1",
			expect: S3Config{
				AccessKeyId:  "accessKey123",
				SecretKey:    "secretKey123",
				Bucket:       "bucket",
				Prefix:       "prefix",
				Endpoint:     "https://endpoint",
				EndpointHost: "endpoint",
				Params: map[string][]string{
					"versionId": {"123"},
				},
			},
		},
		{
			name: "Valid URN with bracketed credentials and query parameters",
			urn:  "s3://[accessKey123]:[secretKey123]@endpoint/bucket/prefix/?versionId=123",
			expect: S3Config{
				AccessKeyId:  "accessKey123",
				SecretKey:    "secretKey123",
				Bucket:       "bucket",
				Prefix:       "prefix/",
				Endpoint:     "https://endpoint",
				EndpointHost: "endpoint",
				Params: map[string][]string{
					"versionId": {"123"},
				},
			},
		},
		{
			name: "Brackets also allowed in the access key and secretkey",
			urn:  "s3://[acc[essK[e[]y123]:[secret[K[e[[y123]@endpoint/bucket/prefix/?versionId=123",
			expect: S3Config{
				AccessKeyId:  "acc[essK[e[]y123",
				SecretKey:    "secret[K[e[[y123",
				Bucket:       "bucket",
				Prefix:       "prefix/",
				Endpoint:     "https://endpoint",
				EndpointHost: "endpoint",
				Params: map[string][]string{
					"versionId": {"123"},
				},
			},
		},
		{
			name: "Valid URN with encoded special chars in credentials and query parameters",
			urn:  "s3://%61%63%63%65%73%73%4B%65%79:[s%65%63r%65tKey123]@endpoint/bucket/prefix/?lifetime=3600",
			expect: S3Config{
				AccessKeyId:  "accessKey",
				SecretKey:    "s%65%63r%65tKey123",
				Bucket:       "bucket",
				Prefix:       "prefix/",
				Endpoint:     "https://endpoint",
				EndpointHost: "endpoint",
				Params: map[string][]string{
					"lifetime": {"3600"},
				},
			},
		},
		{
			name: "Valid URN with bracketed and special encoded combined in credentials",
			urn:  "s3://[%61%63%63%65%73%73%4B%65%79]:[%73%65%63%72%65%74%4B%65%82]@endpoint/bucket/prefix/?versionId=123&mode=strict",
			expect: S3Config{
				AccessKeyId:  "%61%63%63%65%73%73%4B%65%79",
				SecretKey:    "%73%65%63%72%65%74%4B%65%82",
				Bucket:       "bucket",
				Prefix:       "prefix/",
				Endpoint:     "https://endpoint",
				EndpointHost: "endpoint",
				Params: map[string][]string{
					"versionId": {"123"},
					"mode":      {"strict"},
				},
			},
		},
		{
			name:      "URN with missing protocol",
			urn:       "accessKey123:secretKey123@endpoint/bucket/prefix/",
			expectErr: true,
		},
		{
			name:      "URN with extra slashes",
			urn:       "s3:///accessKey123:secretKey123@endpoint/bucket/prefix/",
			expectErr: true,
		},
		{
			name:      "URN with no access key",
			urn:       "s3://:secretKey123@endpoint/bucket/prefix/",
			expectErr: true,
		},
		{
			name:      "URN with no secret key",
			urn:       "s3://accessKey123:@endpoint/bucket/prefix/",
			expectErr: true,
		},
		{
			name:      "URN with empty credentials",
			urn:       "s3://:@endpoint/bucket/prefix/",
			expectErr: true,
		},
		{
			name:      "URN with no bucket",
			urn:       "s3://accessKey123:secretKey123@endpoint/",
			expectErr: true,
		},
		{
			name:      "URN with only endpoint",
			urn:       "s3://endpoint",
			expectErr: true,
		},
		{
			name:      "Invalid URN with missing credentials",
			urn:       "s3://@endpoint/bucket/prefix/",
			expectErr: true,
		},
		{
			name:      "Invalid URN with missing endpoint",
			urn:       "s3://accessKey123:secretKey123@",
			expectErr: true,
		},
		{
			name:      "Invalid URN with incorrect format",
			urn:       "s3:/accessKey123:secretKey123@endpoint/bucket/prefix/",
			expectErr: true,
		},
		{
			name:      "URN with bracketed but incomplete credentials and query parameters",
			urn:       "s3://[accessKey123]:[]@endpoint/bucket/prefix/?logging=true",
			expectErr: true,
		},
		{
			name:      "URN with invalidly placed query parameters and bracketed credentials",
			urn:       "s3://[accessKey123]?apiKey=123:[secretKey123]@endpoint/bucket/prefix/",
			expectErr: true,
		},
		{
			name:      "urn with https protocol",
			urn:       "https://accessKey123:secretKey123@endpoint/bucket/prefix/",
			expectErr: true,
		},
		{
			name:      "dangling prefix",
			urn:       "https://accessKey123:secretKey123@endpoint/bucket/prefix-not-finished",
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config, err := Parse(tt.urn)
			if tt.expectErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.expect, config, "urn %s", tt.urn)
			}
		})
	}
}

func BenchmarkParseS3Urn(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, err := Parse("s3://accessKey123:secretKey123@endpoint/bucket/prefix")
		if err != nil {
			b.Fatal(err)
		}
	}
}
