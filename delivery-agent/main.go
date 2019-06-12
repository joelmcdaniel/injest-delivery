package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/go-redis/redis"
	"github.com/subosito/gotenv"
)

// postback object which will contain
// the HTTP method and endpoing URL
type postback struct {
	Method string `json:"method"`
	URL    string `json:"url"`
}

// consumer will pop postback json object
// messages off of the queue and send them
// into the postback channel
type consumer struct {
	redisClient *redis.Client
	queueName   string
	pbChannel   chan string
}

// deliverer will create postback objects
// from the json object messages received
// off the postback channel and deliver them
// to the endpoint url capturing delivery
// response information and logging it
type deliverer struct {
	pbChannel    chan string
	deliveryInfo deliveryInfo
}

// deliveryInfo is used to store the postback
// delivery response info that will be logged
type deliveryInfo struct {
	postback     postback
	deliveryTime time.Time
	responseCode int
	responseTime time.Duration
	responseBody string
}

var c consumer
var d deliverer
var concurrencyLevel int

func init() {
	gotenv.Load()

	// initialize redis client
	rc := redis.NewClient(&redis.Options{
		Addr:     os.Getenv("REDIS_HOST") + ":" + os.Getenv("REDIS_PORT"),
		Password: os.Getenv("REDIS_PASSWORD"),
		DB: func(v string) int {
			i, _ := strconv.Atoi(v)
			return i
		}(os.Getenv("REDIS_DB")),
	})

	// open and set log file
	f, err := os.OpenFile(os.Getenv("LOGFILE_NAME"), os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("Error: Open log file failure. %v", err)
	}
	log.SetOutput(f)

	// test redis connection and log if connection failed.
	_, err = rc.Ping().Result()
	if err != nil {
		log.Fatalf("Error: Redis connection failure. %v", err)
	}

	// channel for postback objects
	ch := make(chan string)

	// initialize consumer and deliverer
	c = consumer{
		redisClient: rc,
		queueName:   os.Getenv("QUEUE_NAME"),
		pbChannel:   ch,
	}

	d = deliverer{
		pbChannel:    ch,
		deliveryInfo: deliveryInfo{},
	}

	concurrencyLevel = func(v string) int {
		i, _ := strconv.Atoi(v)
		return i
	}(os.Getenv("CONCURRENCY_LEVEL"))
}

func main() {

	var wg sync.WaitGroup

	for i := 1; i <= concurrencyLevel; i++ {
		wg.Add(1)
		// consume and deliver postbacks
		go c.consume()
		go d.deliver(&wg)
	}
	wg.Wait()
}

// consumes messages from queue and sends them into postback channel
func (c consumer) consume() {
	for {
		msg, err := c.redisClient.RPop(c.queueName).Result()
		if err != nil {
			fmt.Println("Sleeping...")
			time.Sleep(5 * time.Second)
		} else {
			fmt.Println("Message popped!")
			c.pbChannel <- msg
		}
	}
}

// deliver postback objects recieved on postback channel to endpoint url
// log delivery response info returned from the send request
func (d deliverer) deliver(wg *sync.WaitGroup) {
	defer wg.Done()

	for msg := range d.pbChannel {
		pb := postback{}
		json.Unmarshal([]byte(msg), &pb)
		d.deliveryInfo = deliverPostback(pb, time.Now())
		d.logDelivery()
	}
}

// send a postback to endpoint url, return postback delivery response info
func deliverPostback(pb postback, startTime time.Time) deliveryInfo {

	switch pb.Method {
	//TODO: maybe handle POST case in the future
	default:
		resp, err := http.Get(pb.URL)
		if err != nil {
			log.Fatalln(err)
		}

		defer resp.Body.Close()
		body, _ := ioutil.ReadAll(resp.Body)
		fmt.Println("Postback delivered!")

		return deliveryInfo{
			postback:     pb,
			deliveryTime: startTime,
			responseCode: resp.StatusCode,
			responseTime: time.Since(startTime),
			responseBody: string(body),
		}
	}
}

// Logs postback delivery response info
func (d deliverer) logDelivery() {
	di := d.deliveryInfo

	values := []string{"Postback Delivery|",
		"Method: " + di.postback.Method,
		"URL: " + di.postback.URL,
		"Delivery Time: " + di.deliveryTime.String(),
		"Response Code: " + strconv.Itoa(int(di.responseCode)),
		"Response Time: " + di.responseTime.String(),
		"Response Body: " + di.responseBody}

	for i := range values {
		log.Printf("|%-10v\n", values[i])
	}
}
