package main

import (
	"fmt"
	"github.com/davidji99/terraform-provider-herokux/api"
	"github.com/davidji99/terraform-provider-herokux/api/kafka"
	"github.com/davidji99/terraform-provider-herokux/api/pkg/config"
)

func main() {
	token := "SOME_TOKEN"
	api, clientInitErr := api.New(config.APIToken(token),
		config.BasicAuth("some_user_name", token))

	if clientInitErr != nil {
		panic(clientInitErr)
	}

	opts := kafka.NewTopicRequest("SOME_TOPIC_NAME", 3, "3d")
	opts.Compaction = true
	opts.Partitions = 6

	r, response, err := api.Kafka.CreateTopic("SOME_CLUSTER_ID", opts)
	if err != nil {
		panic(err)
	}

	fmt.Println(response.StatusCode)
	fmt.Println(r.GetMessage())
}
