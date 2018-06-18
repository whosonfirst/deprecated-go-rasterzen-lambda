# go-rasterzen-lambda

## Install

You will need to have both `Go` (specifically a version of Go more recent than 1.7 so let's just assume you need [Go 1.9](https://golang.org/dl/) or higher) and the `make` programs installed on your computer. Assuming you do just type:

```
make bin
```

All of this package's dependencies are bundled with the code in the `vendor` directory.

## Lambda

```
aws lambda create-function --region {REGION} --function-name Rasterzen \
    --zip-file fileb://./deployment.zip --runtime go1.x --tracing-config Mode=Active \
    --role {ROLE} --handler main
```

_See notes on (IAM) roles below

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

## IAM Roles

* `AWSLambdaExecute`
* `AWSXRayFullAccess`
* A role that allows your function to read/write to the S3 bucket defined in `RASTERZEN_S3_DSN`

## API Gateway

_Okay, so this is still all a bit slippery for me, meaning I can barely ever keep track of all the buttons you have to press to accomplish things in AWS..._

* Once you've created your API, you want to "Create Resource" from the `Actions` menu.
* Configure it as a `proxy resource`
* Set the `Resource Name` to be "/" (or whatever you need it to be)
* Set the `Resource Path` to be "{proxy+}" (... I have no idea what's going on here)
* Enable the CORS gateway if you want...

### Testing

In the `Method` menu select 

* GET

In the `Path {proxy}` field add something like:

* /svg/13/1315/3171.json

And in the `Query Strings {proxy} field add:

* api_key={NEXTZEN_APIKEY}

In principle at this point you should get back a `200 OK` response and there should be tiles in your S3 bucket. If it doesn't then... uh... AWS???

## See also

* https://github.com/whosonfirst/go-rasterzen
* https://github.com/whosonfirst/go-whosonfirst-aws
* https://artem.krylysov.com/blog/2018/01/18/porting-go-web-applications-to-aws-lambda/
* https://github.com/akrylysov/algnhsa

### AWS

* https://aws.amazon.com/blogs/compute/announcing-go-support-for-aws-lambda/
* https://docs.aws.amazon.com/lambda/latest/dg/lambda-go-how-to-create-deployment-package.html
* https://docs.aws.amazon.com/lambda/latest/dg/env_variables.html
* https://docs.aws.amazon.com/cli/latest/reference/sts/get-session-token.html
