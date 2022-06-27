package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sns"
)

type SNSPublishAPI interface {
	Publish(ctx context.Context,
		params *sns.PublishInput,
		optFns ...func(*sns.Options)) (*sns.PublishOutput, error)
}

func PublishMessage(c context.Context, api SNSPublishAPI, input *sns.PublishInput) (*sns.PublishOutput, error) {
	return api.Publish(c, input)
}

func HandleRequest() {
	apiKey := os.Getenv("SHOPIFY_API_KEY")
	password := os.Getenv("SHOPIPY_API_PASSWORD")
	domain := os.Getenv("SHOPIFY_SHOPIFY_DOMAIN")

	url := fmt.Sprintf("https://%s:%s@%s.myshopify.com/admin/api/2022-04/products.json", apiKey, password, domain)

	resp, err := http.Get(url)
	if err != nil {
		panic(err)
	}

	defer resp.Body.Close()

	bytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}

	var products map[string]interface{}
	if err := json.NewDecoder(strings.NewReader(string(bytes))).Decode(&products); err != nil {
		panic(err)
	}

	skuToStock := []string{}

	for _, productData := range products {
		for _, products := range productData.([]interface{}) {
			for key, product := range products.(map[string]interface{}) {
				if key == "variants" {
					for _, variant := range product.([]interface{}) {
						if variant.(map[string]interface{})["inventory_quantity"].(float64) <= 10 {
							sku := variant.(map[string]interface{})["sku"]
							skuToStock = append(skuToStock, sku.(string))
						}
					}
				}
			}
		}
	}

	if len(skuToStock) > 0 {
		snsMessage := strings.Join(skuToStock, ",")
		topicARN := "<lambda-arn>"

		cfg, err := config.LoadDefaultConfig(context.TODO())

		if err != nil {
			panic("configuration error, " + err.Error())
		}

		client := sns.NewFromConfig(cfg)
		input := &sns.PublishInput{
			Message:  &snsMessage,
			TopicArn: &topicARN,
		}

		result, err := PublishMessage(context.TODO(), client, input)
		if err != nil {
			fmt.Println("Got an error publishing the message:")
			fmt.Println(err)
		}

		fmt.Println("Message ID: " + *result.MessageId)
	}
}

func main() {
	lambda.Start(HandleRequest)
}
