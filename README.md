# Injest-Delivery

Injest-Delivery is a service that functions as a small scale simulation of how to distribute data to third parties in real time. It consist of two small applications, Injestion Agent (PHP7) and Delivery Agent (Go) which communicate between the two through a Redis job queue.

**Data flow​ :**
1. Web request (see sample request) >
2. "Ingestion Agent" (php7) >
3. "Delivery Queue" (redis)
4. "Delivery Agent" (go) >
5. Web response (see sample response)

## **Injestion Agent (PHP7)**
 Injests http requests in json format and pushes a "postback" object to a Redis queue for each data object contained in the raw post data. Placeholders {} for query string values in endpoint url will be replaced with their corresponding data property values.

### **Sample Request​ :**
(POST) 

http://{server_ip}/ingest.php

(RAW POST DATA)
```json
{        
	"endpoint": {
		"method":"GET",         
		"url":"http://sample_domain_endpoint.com/data?title={mascot}&image={location}&foo={bar}"
	},
	"data":[
		{
			"mascot":"Gopher",
			"location":"https://blog.golang.org/gopher/gopher.png"
		},
		{
			"mascot":"Gopher",
			"location":"https://blog.golang.org/gopher/gopher.png",
			"bar":"bar"
		}
	]
}
```

## **Delivery Agent (Go)**
 Continuously pulls the "postback" objects from the Redis delivery queue delivering them to the http endpoint url specified in the "postback" object. Logs the delivery time, response code, response time and response body of each "postback" object sent. 

### **Sample Response ​ (Postback):**
GET

`http://sample_domain_endpoint.com/data?title=Gopher&image=https%3A%2F%2Fblog.golang.org%2Fgopher%2Fgopher.png&foo=`

## **Stack and Installation Steps**
1. [Ubuntu](https://www.ubuntu.com/download) (preferably 18.04) or some variation of (I'm running [Peppermint 9](https://peppermintos.com/2019/01/peppermint-9-respin-2-released/))
2. [Redis](https://redis.io/download)
3. [Apache2](https://httpd.apache.org/download.cgi#apache24)
4. [PHP 7](https://www.php.net/downloads.php)
5. [PhpRedis](https://github.com/phpredis/phpredis)
6. [go-redis](https://github.com/go-redis/redis)
7. [Go](https://golang.org/dl/)

### 1. Install Redis
```bash
sudo apt update
sudo apt install redis
```

Check installation:
```bash
redis-cli --version
```
### 2. Install Apache Web Server
```bash
sudo apt update
sudo apt install apache2
```

Verify installation:
```bash
apache2 -version
```

### 3. Install PHP 7.2
```bash
sudo apt update
sudo apt-get install php libapache2-mod-php
```
Verify installation:
```bash
php -v
```

Restart the Apache service to apply the changes:
```bash
sudo systemctl restart apache2
```

### 4. Install Redis command line client and PhpRedis extension
```bash
 sudo apt-get update
 sudo apt-get install redis-tools php-redis
```
### 5. Install Go
Follow the instructions here: https://golang.org/doc/install

### 6. Install go-redis redis client for Go
go get -u `github.com/go-redis/redis`

### 7. Install gotenv
go get `github.com/subosito/gotenv`

## **Steps to Run**

1. Copy config.php, injest.php and Postback.php from ingestion-agent directory to /var/www/html to run at http://{server_ip}/ingest.php. Or, alternatively copy whole injestion-agent directory to /var/www/html to run at http://{server_ip}/injestion-agent/ingest.php.
2. Copy delivery-agent directory to /go/src
3. From /go/src/delivery-agent run go build
4. From /go/src/delivery-agent run go run delivery-agent
5. Using Postman (or some similar tool) copy web request test json data from ingestion-agent/tests-json to begin posting web requests to http://{server_ip}/ingest.php or alternatively http://{server_ip}/injestion-agent/ingest.php
6. Open delivery-log under /go/src/delivery-agent to view the logged delivery response info. 

