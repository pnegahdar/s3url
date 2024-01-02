S3_URL
======

Like `DATABASE_URL` but for s3 buckets 🪣. `S3_URL`.

## Motivation

`DATABASE_URL` is a well known environment variable that is used to configure database connections.

Before that we had to configure database connections with a bunch of environment variables like `DB_HOST`, `DB_PORT`, `DB_USER`, `DB_PASSWORD`, `DB_NAME`, etc.

It freaking sucked.

I still find that most people configure s3 like this as well and you end up needing quite a few envars to configure a bucket. `STORAGE_MAIN_S3_ENDPOINT`, `STORAGE_MAIN_S3_ACCESS_KEY_ID`, `STORAGE_MAIN_S3_SECRET_KEY`, `STORAGE_MAIN_S3_BUCKET`, `STORAGE_MAIN_S3_PREFIX`,

It still sucks.

```
S3_URL="s3://access_key:secret_key@endpoint/bucket/prefix"
```

Thats nice and clean 😎.

## Format

Format is basically like http urls with basic auth and bucket name. You can wrap the username/password parts in `[]` if they contain non-url-safe parameters.

```
# Some s3 bucket
s3://user1:secret_key@s3.us-east-2.amazonaws.com/my_bucket/my_prefix

# Url unsafe password will parse as 'sec&ret_key'
s3://user1:[sec&ret_key]@s3.us-east-2.amazonaws.com/my_bucket/my_prefix

# Some cloudflare bucket
s3://29d929fjsd:29592950@my.r2.cloudflarestorage.com/bucket_2"

# Some params can be passed in too
s3://29d929fjsd:29592950@my.r2.cloudflarestorage.com/bucket_2?a=1"
```


## Usage

Simple parsing:

```go 
package main

import (
    "github.com/pnegahdar/s3url"
)


func main(){
    s3Config, err := s3url.Parse("s3://user1:secret_key@s3...")
    if err != nil {
        panic(err)  // invalid url
    }
    println(s3Config.Bucket) // my_bucket
}
```

To Initialize an s3 client using a `S3Config` you might do the following: 

```go
package main

import (
    "github.com/aws/aws-sdk-go-v2/aws"
    "github.com/aws/aws-sdk-go-v2/config"
    "github.com/aws/aws-sdk-go-v2/credentials"
    "github.com/aws/aws-sdk-go-v2/service/s3"
    "github.com/pnegahdar/s3url"
)   


func MkS3Client(s3Config S3Config) (*s3.Client, error) {
	r2Resolver := aws.EndpointResolverWithOptionsFunc(func(service, region string, options ...interface{}) (aws.Endpoint, error) {
		return aws.Endpoint{
			URL:               s3Config.Endpoint,
			HostnameImmutable: true,
			Source:            aws.EndpointSourceCustom,
		}, nil
	})

	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithEndpointResolverWithOptions(r2Resolver),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(s3Config.AccessKeyId, s3Config.SecretKey, "")),
		config.WithRegion("auto"),
	)
	if err != nil {
		return nil, errors.Wrap(err, "unable to load s3 config")
	}
	return s3.NewFromConfig(cfg), nil
}

func main() {
	s3Config, err := s3url.Parse("s3://user1:secret_key@s3...")
	if err != nil {
		panic(err)  // invalid url
	}

    s3Client, err := MkS3Client(s3Config)
    if err != nil {
        panic(err)  
    }

}

```

## Limitations

The format is currently the bare minimum to serve my common usecase. Will need to adjust as we see more cases in the wild.


