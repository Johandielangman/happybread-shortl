![banner](/docs/images/banner.png)

# Happy Bread's Simple, Cheap URL Shortener!

![architecture](/docs/images/architecture.svg)

This repository offers a simple solution for something I've always wanted to self-host: a URL shortener!

**🌟 Give it a Try! 🌟**

https://bun.happybread.net/ef480

If everything works in this mini demo, it should take you to the documentation for defining a Lambda handler in Go. How cool is that? A URL shortener is essentially just a lookup table for long links. The "ef480" path parameter is the key in our hash map. In layman's terms, "ef480" points to `https://docs.aws.amazon.com/lambda/latest/dg/golang-handler.html`. The "ef480" was created using the first 5 characters from a sha256 hash. It will always be the result from the long link we created.

The architecture diagram might look complicated, but it's actually very straightforward! We have three key components:

1. **API Gateway (and Cloudflare)**  
   These protect me, the person hosting the tool. They provide rate limiting, WAF, usage plans, SSL, and other important cybersecurity features.

2. **AWS Lambda**  
   This uses Go under the hood for super-fast response times. There are two handlers: one for creating shortened links and another for retrieving the original links.

3. **Upstash**  
   A low-latency serverless data platform with a *VERY GENEROUS* [free tier](https://upstash.com/pricing), offering 10,000 commands per day. Since my links probably won’t reach that level of popularity, it’s a perfect fit! I use their Redis solution, which is ideal for our hash table lookups.

In summary, everything here is serverless. Thanks to API Gateway's rate limiting, it’s a very affordable, pay-as-you-go solution for me. If no one clicks on the links, I don’t have to worry about feeding the AWS beast with money!

## From Go to Lambda

There are two `main.go` files involved:
- [One to handle the 'create new link' request](https://github.com/Johandielangman/happybread-shortl/blob/main/new/main.go)
- [One to handle the 'get link' request](https://github.com/Johandielangman/happybread-shortl/blob/main/get/main.go)

### The Problem? 🤔

My operating system differs from the one used by the Lambda function that will execute the compiled Go code. Therefore, we need to compile the code using the correct [runtime family](https://docs.aws.amazon.com/lambda/latest/dg/golang-package.html):

```bash
GOOS=linux GOARCH=arm64 go build -tags lambda.norpc -o bootstrap main.go
```

Once compiled, we need to create a deployment package:

```bash
zip myFunction.zip bootstrap
```

### The Challenge

With two files involved, running these commands manually every time something changes can be cumbersome. This is where [GNU Make](https://www.gnu.org/software/make/manual/make.html) becomes incredibly helpful! It allows us to define how to compile and zip our executables using a simple set of rules.

After setting up the [Makefile](https://github.com/Johandielangman/happybread-shortl/blob/main/Makefile), you can simplify the process by running a single command:

```bash
make compile
```

## The Makefile

Here’s a look at what the Makefile looks like:


```Makefile
.ONESHELL:

varGoos=linux
varGoosArch=arm64
varExeName=bootstrap
varEntryFileName=main.go
varDeploymentFolder=deployments

compile-get:
	@echo "Compiling get package"
	@cd ./get/
	GOOS=$(varGoos) GOARCH=$(varGoosArch) go build -tags lambda.norpc -o $(varExeName) $(varEntryFileName)

compile-new:
	@echo "Compiling new package"
	@cd ./new/
	GOOS=$(varGoos) GOARCH=$(varGoosArch) go build -tags lambda.norpc -o $(varExeName) $(varEntryFileName)

zip:
	zip -j ./$(varDeploymentFolder)/get_deployment.zip ./get/bootstrap
	zip -j ./$(varDeploymentFolder)/new_deployment.zip ./new/bootstrap

compile:
	mkdir -p deployments
	@echo "Compiling..."
	$(MAKE) compile-get
	$(MAKE) compile-new
	@echo "Zip..."
	$(MAKE) zip
	@echo "Done 🚀"
```

This structure keeps the workflow efficient and minimizes repetitive tasks, making development and deployment much smoother!


## From API Gateway to Lambda

I decided to go for a simple [REST API](https://docs.aws.amazon.com/apigateway/latest/developerguide/apigateway-rest-api.html). It allows me to protect my endpoints and create basic usage plans in case anyone wants to use the shortener to create their own shortened links.

### The Two Endpoints

The Gateway is actually fairly simple. There are two endpoints, each connected to its respective lambda handler:

- **POST to `/new`**  
  This POST request accepts the following payload:
  ```json
  {
      "link": "https://foo.com"
  }
  ```
  The `link` value is used to create a SHA256 hash. The first 5 characters of this hash are used as the Redis key. The POST request also requires a `x-api-key` header with your API key for authentication.

- **GET to `/{link}`**  
  This is a simple GET request with a path parameter called `link`. It’s important to set the *Lambda proxy integration* to `true` so that we can receive critical information from the [proxy payload](https://docs.aws.amazon.com/apigateway/latest/developerguide/set-up-lambda-proxy-integrations.html), such as the incoming IP address. If someone is attacking our endpoint, we can ban their IP using Cloudflare. For Redis hits, we use a clever trick to perform the redirect.

### The Redirect

The `/{link}` endpoint always returns a `text/html` *Content-Type*.  

- A Redis miss will return a 404 and display a small "not found" HTML page.  
- A Redis hit will use the [HTML `<meta>` http-equiv Attribute](https://www.w3schools.com/tags/att_meta_http_equiv.asp) to perform a redirect.

Here’s an example of the HTML returned for a Redis hit:

```html
<!DOCTYPE html>
<html>
  <head>
    <meta http-equiv="refresh" content="0; url=%s">
  </head>
</html>
```

Once you land on this page, the `http-equiv` attribute will immediately redirect you to the provided `url`. I think this is pretty cool!

### A custom domain

The main idea here is to create a **Client Certificate** in Cloudflare and upload it to AWS. I found this amazing blog post: https://bun.happybread.net/edf5d . Ha! Look at me using my own product. Once the Client Certificate is created, I simply map **bun.happybread.net** to the API Gateway.


## Conclusion

I hope you really enjoyed this! Feel free to reach out to me if you're interested in an API key: toast@happybread.net