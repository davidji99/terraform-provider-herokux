package herokux

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/davidji99/simpleresty"
	"github.com/davidji99/terraform-provider-herokux/api/kafka"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func resourceHerokuxKafkaConsumerGroup() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceHerokuxKafkaConsumerGroupCreate,
		ReadContext:   resourceHerokuxKafkaConsumerGroupRead,
		DeleteContext: resourceHerokuxKafkaConsumerGroupDelete,

		Importer: &schema.ResourceImporter{
			StateContext: resourceHerokuxKafkaConsumerGroupImport,
		},

		Schema: map[string]*schema.Schema{
			"kafka_id": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.IsUUID,
			},

			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},
	}
}

func resourceHerokuxKafkaConsumerGroupImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	d.SetId(d.Id())

	readErr := resourceHerokuxKafkaConsumerGroupRead(ctx, d, meta)
	if readErr.HasError() {
		return nil, fmt.Errorf("unable to import existing MTLS configuration")
	}

	return []*schema.ResourceData{d}, nil
}

func resourceHerokuxKafkaConsumerGroupCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*Config)
	client := config.API
	opts := kafka.NewConsumerGroupRequest()

	kafkaID := getKafkaID(d)

	if v, ok := d.GetOk("name"); ok {
		opts.Name = v.(string)
		log.Printf("[DEBUG] consumer group name: %s", opts.Name)
	}

	log.Printf("[DEBUG] Creating Kafka consumer group %s", opts.Name)
	_, _, createErr := client.Kafka.CreateConsumerGroup(kafkaID, opts)
	if createErr != nil {
		return diag.FromErr(createErr)
	}

	log.Printf("[DEBUG] Waiting for Kafka consumer group %s to be ready", opts.Name)
	stateConf := &resource.StateChangeConf{
		Pending: []string{kafka.ConsumerGroupStatuses.PENDING.ToString()},
		Target:  []string{kafka.ConsumerGroupStatuses.CREATED.ToString()},
		Refresh: kafkaConsumerGroupStateRefreshFunc(kafkaID, opts.Name,
			kafka.ConsumerGroupStatuses.CREATED, client.Kafka.WasConsumerGroupCreated),
		Timeout:      time.Duration(config.KafkaCGCreateTimeout) * time.Minute,
		PollInterval: StateRefreshPollInterval,
	}

	if _, err := stateConf.WaitForStateContext(ctx); err != nil {
		return diag.Errorf("error waiting for consumer group to be ready on %s: %s", opts.Name, err.Error())
	}

	// Set the resource ID to be a composite of the kafka ID and group name.
	d.SetId(fmt.Sprintf("%s:%s", kafkaID, opts.Name))

	return resourceHerokuxKafkaConsumerGroupRead(ctx, d, meta)
}

func resourceHerokuxKafkaConsumerGroupRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*Config).API

	result, parseErr := parseCompositeID(d.Id(), 2)
	if parseErr != nil {
		return diag.FromErr(parseErr)
	}

	group, _, getErr := client.Kafka.GetConsumerGroupByName(result[0], result[1])
	if getErr != nil {
		return diag.FromErr(getErr)
	}

	d.Set("kafka_id", result[0])
	d.Set("name", group.GetName())

	return nil
}

func resourceHerokuxKafkaConsumerGroupDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*Config)
	client := config.API

	result, parseErr := parseCompositeID(d.Id(), 2)
	if parseErr != nil {
		return diag.FromErr(parseErr)
	}

	kafkaID := result[0]
	groupName := result[1]

	opts := kafka.NewConsumerGroupRequest()
	opts.Name = groupName

	log.Printf("[DEBUG] Deleting consumer group %s from %s", groupName, kafkaID)
	_, _, deleteErr := client.Kafka.DeleteConsumerGroup(kafkaID, opts)
	if deleteErr != nil {
		return diag.FromErr(deleteErr)
	}

	log.Printf("[DEBUG] Waiting for Kafka consumer group %s to be deleted from %s", groupName, kafkaID)
	stateConf := &resource.StateChangeConf{
		Pending: []string{kafka.ConsumerGroupStatuses.PENDING.ToString()},
		Target:  []string{kafka.ConsumerGroupStatuses.DELETED.ToString()},
		Refresh: kafkaConsumerGroupStateRefreshFunc(result[0], result[1],
			kafka.ConsumerGroupStatuses.DELETED, client.Kafka.WasConsumerGroupDeleted),
		Timeout:      time.Duration(config.KafkaCGDeleteTimeout) * time.Minute,
		PollInterval: StateRefreshPollInterval,
	}

	if _, err := stateConf.WaitForStateContext(ctx); err != nil {
		return diag.Errorf("error waiting for consumer group to be deleted on %s: %s", opts.Name, err.Error())
	}

	d.SetId("")

	return nil
}

func kafkaConsumerGroupStateRefreshFunc(kafkaID,
	groupName string, targetState kafka.ConsumerGroupStatus,
	checker func(id, n string) (bool, *simpleresty.Response, error)) resource.StateRefreshFunc {

	return func() (interface{}, string, error) {
		result, _, err := checker(kafkaID, groupName)
		if err != nil {
			return nil, kafka.ConsumerGroupStatuses.PENDING.ToString(), err
		}

		if result {
			return result, targetState.ToString(), nil
		}

		return result, kafka.ConsumerGroupStatuses.PENDING.ToString(), nil
	}
}
