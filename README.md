# go-rasterzen-lambda

Run the `go-rasterzen` code in an AWS Lambda function.

## Install

You will need to have both `Go` (specifically a version of Go more recent than 1.7 so let's just assume you need [Go 1.9](https://golang.org/dl/) or higher) and the `make` programs installed on your computer. Assuming you do just type:

```
make bin
```

All of this package's dependencies are bundled with the code in the `vendor` directory.

## Important

This works, or more specifically appears to work _for me_ so it will probably
work for you too, right? 

There are a couple important things to keep in mind:

* Tiles are only cached in S3
* There is still no cache invalidation logic (in `go-rasterzen`) so you should
  be ready to manually purge or otherwise investigate S3 data 
* There are no hooks (in `go-rasterzen`) for enabling chatty logging which means
  introspecting any kind of errors in the Lambda function will be... "fun"

What follows is a small thumderstorm of AWS alphabet (services) soup that you'll
need to power your way through to get this to work.

## Lambda

```
aws lambda create-function --region {REGION} --function-name Rasterzen \
    --zip-file fileb://./deployment.zip --runtime go1.x --tracing-config Mode=Active \
    --role {ROLE} --handler main
```

_See notes on (IAM) roles below_

### Environment variables

You will need to configure your Lamdba function to set the following  environment variables:

### RASTERZEN_S3_DSN

```
bucket={BUCKET} prefix={PREFIX} region={REGION} credentials=iam:
```

### RASTERZEN_CACHE_OPTIONS

```
ACL=public-read
```

_Or you can leave this empty if you don't want the cached tiles to be public._

### Basic settings

It is unlikely that the default execution timeout settings (3 seconds) will be
enough to fetch and process some tiles. You should adjust this as necessary.

## IAM Roles

* `AWSLambdaExecute`
* `AWSXRayFullAccess` (I don't know why... maybe it's the tracing flag above?)
* A role that allows your function to read/write to the S3 bucket defined in `RASTERZEN_S3_DSN`

## CloudFront

Can't be done without routing all the requests though API Gateway,
yet. Apparently Lambda functions written in Go do not work as CloudFront
triggers.

If you are going to use API Gateway (more below) you should make sure to read
the [Lambda gotchas](https://github.com/tilezen/tapalcatl-py#lambda-gotchas)
section of the Tilezen `tapalcatl-py` package.

## API Gateway

_Okay, so this is still all a bit slippery for me, meaning I can barely ever
keep track of all the buttons you have to press to accomplish things in
AWS. It's very possible, still, that I've missed something or gotten something
else wrong. Gentle cluebats are welcome and encouraged._

* Once you've created your API, you want to "Create Resource" from the `Actions` menu.
* Configure it as a `proxy resource`
* Set the `Resource Name` to be "/" (or whatever you need it to be)
* Set the `Resource Path` to be "{proxy+}" (... I have no idea what's going on here)
* Enable the CORS gateway if you want...

## Really important

In order for requests to produce PNG output (rather than a base64 encoded string) you will need to do a few things:

1. Make sure your API Gateway settings list `image/png` as a known and valid binary type:

![](docs/images/20180625-agw-binary.png)

2. If you've put a CloudFront distribution in front of your API Gateway then you
will to ensure that you blanket enable all HTTP headers or whitelist the
`Accept:` header , via the `Cache Based on Selected Request Headers` option (for
the CloudFront behaviour that points to your gateway):

![](docs/images/20180625-cf-cache.png)

3. Make sure you pass an `Accept: image/png` header when you request the PNG rendering.

### Testing

In the `Method` menu select 

* GET

In the `Path {proxy}` field add something like:

* /svg/13/1315/3171.json

And in the `Query Strings {proxy} field add:

* api_key={NEXTZEN_APIKEY}

In principle at this point you should get back a `200 OK` response and there should be tiles in your S3 bucket. If it doesn't then... uh... AWS???

## See also

### Package specific

* https://github.com/whosonfirst/go-rasterzen
* https://github.com/whosonfirst/go-whosonfirst-aws
* https://artem.krylysov.com/blog/2018/01/18/porting-go-web-applications-to-aws-lambda/
* https://github.com/akrylysov/algnhsa

### AWS

* https://aws.amazon.com/blogs/compute/announcing-go-support-for-aws-lambda/
* https://docs.aws.amazon.com/lambda/latest/dg/lambda-go-how-to-create-deployment-package.html
* https://docs.aws.amazon.com/lambda/latest/dg/env_variables.html
* https://docs.aws.amazon.com/cli/latest/reference/sts/get-session-token.html

### General

* https://apimeister.com/2017/05/09/hosting-a-cloudfront-site-with-s3-and-api-gateway.html
* https://github.com/tilezen/tapalcatl-py#lambda-gotchas
