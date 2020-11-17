package herokux

import (
	"context"
	"fmt"
	"log"
	"regexp"
	"time"

	"github.com/davidji99/terraform-provider-herokux/api"
	"github.com/davidji99/terraform-provider-herokux/api/kafka"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func resourceHerokuxKafkaTopic() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceHerokuxKafkaTopicCreate,
		ReadContext:   resourceHerokuxKafkaTopicRead,
		UpdateContext: resourceHerokuxKafkaTopicUpdate,
		DeleteContext: resourceHerokuxKafkaTopicDelete,

		Importer: &schema.ResourceImporter{
			StateContext: resourceHerokuxKafkaTopicImport,
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

			"partitions": {
				Type:     schema.TypeInt,
				Required: true,
				ForceNew: true,
			},

			"replication_factor": {
				Type:         schema.TypeInt,
				Optional:     true,
				Default:      3,
				ValidateFunc: validation.IntAtLeast(3),
			},

			"retention_time": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "1d",
				ValidateFunc: validateRetentionTime,
			},

			"compaction": {
				Type:     schema.TypeBool,
				Optional: true,
				Computed: true,
			},

			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"cleanup_policy": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func validateRetentionTime(v interface{}, k string) (ws []string, errors []error) {
	duration := v.(string)

	// First validate the retention time is defined in a supported value.
	if !regexp.MustCompile(kafka.RetentionTimeDuragionRegexStricterWithDisable).MatchString(duration) {
		errors = append(errors, fmt.Errorf(
			"Unsupported retention time %s. Format needs to be in `##ms|s|m|h|d|w` or `disable`", duration))
		return
	}

	// Do not test for retention time minimum if the value is 'disable'
	if duration != kafka.RetentionTimeDisableVal {
		// Then verify that the retention time is set to a value that's at least 24 hours.
		minRetentionTime, _ := kafka.ConvertDurationToMilliseconds("24h")
		durationInt, _ := kafka.ConvertDurationToMilliseconds(duration)

		if durationInt < minRetentionTime {
			errors = append(errors, fmt.Errorf("you must specify a retention time that is at least 24 hours equivalent or greater"))
		}
	}

	return
}

func resourceHerokuxKafkaTopicImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	d.SetId(d.Id())

	readErr := resourceHerokuxKafkaTopicRead(ctx, d, meta)
	if readErr.HasError() {
		return nil, fmt.Errorf("unable to import resource")
	}

	return []*schema.ResourceData{d}, nil
}

func resourceHerokuxKafkaTopicCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*Config)
	client := config.API
	opts := &kafka.TopicRequest{}

	kafkaID := getKakfaID(d)

	if v, ok := d.GetOk("name"); ok {
		vs := v.(string)
		log.Printf("[DEBUG] topic name is : %v", vs)
		opts.Name = vs
	}

	if v, ok := d.GetOk("partitions"); ok {
		vs := v.(int)
		log.Printf("[DEBUG] topic partitions is : %v", vs)
		opts.Partitions = vs
	}

	if v, ok := d.GetOk("replication_factor"); ok {
		vs := v.(int)
		log.Printf("[DEBUG] topic replication_factor is : %v", vs)
		opts.ReplicationFactor = vs
	}

	if v, ok := d.GetOk("retention_time"); ok {
		duration := v.(string)
		log.Printf("[DEBUG] topic retention_time (string) is : %v", duration)

		// Convert the duration to milliseconds as that's what is supported by the API
		// unless the attribute value is `disable`. If it is disable,
		// then set the opts.RetentionTimeMS to nil.
		if duration == kafka.RetentionTimeDisableVal {
			opts.RetentionTimeMS = nil
			log.Printf("[DEBUG] topic new retention_time (int) is : %v", "nil")
		} else {
			ms, conErr := kafka.ConvertDurationToMilliseconds(duration)
			if conErr != nil {
				return diag.FromErr(conErr)
			}
			opts.RetentionTimeMS = &ms
			log.Printf("[DEBUG] topic new retention_time (int) is : %v", ms)
		}
	}

	if v, ok := d.GetOk("compaction"); ok {
		vs := v.(bool)
		log.Printf("[DEBUG] topic compaction is : %v", vs)
		opts.Compaction = vs
	}

	log.Printf("[DEBUG] Creating Kafka topic %s", opts.Name)

	_, _, createErr := client.Kafka.CreateTopic(kafkaID, opts)
	if createErr != nil {
		return diag.FromErr(createErr)
	}

	log.Printf("[DEBUG] Waiting for Kafka topic %s to be ready", opts.Name)

	stateConf := &resource.StateChangeConf{
		Pending:      []string{kafka.TopicStatuses.PENDING.ToString()},
		Target:       []string{kafka.TopicStatuses.READY.ToString()},
		Refresh:      topicCreationStateRefreshFunc(client, kafkaID, opts.Name, opts.Partitions),
		Timeout:      time.Duration(config.KafkaTopicCreateTimeout) * time.Minute,
		PollInterval: StateRefreshPollInterval,
	}

	if _, err := stateConf.WaitForStateContext(ctx); err != nil {
		return diag.Errorf("error waiting for topic to be ready on %s: %s", opts.Name, err.Error())
	}

	d.SetId(fmt.Sprintf("%s:%s", kafkaID, opts.Name))

	return resourceHerokuxKafkaTopicRead(ctx, d, meta)
}

// topicCreationStateRefreshFunc checks if the topic is ready. 'Ready' state is determined by two things:
//  1) the topic is present when retrieving from all topics
//  2) the number of partitions matches the specified count.
func topicCreationStateRefreshFunc(client *api.Client, kafkaID, topicName string, partitionCount int) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		topic, response, getErr := client.Kafka.GetTopicByName(kafkaID, topicName)
		if getErr != nil {
			if response.StatusCode == 404 {
				// this means the topic hasn't been created just yet
				return nil, kafka.TopicStatuses.PENDING.ToString(), nil
			}
			return nil, kafka.TopicStatuses.UNKNOWN.ToString(), getErr
		}

		if topic.GetPartitions() != partitionCount {
			log.Printf("[DEBUG] topic created but partitions not provisioned. Count is %d", topic.GetPartitions())
			return topic, kafka.TopicStatuses.PENDING.ToString(), nil
		}

		return topic, kafka.TopicStatuses.READY.ToString(), nil
	}
}

func resourceHerokuxKafkaTopicRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*Config).API

	result, parseErr := parseCompositeID(d.Id(), 2)
	if parseErr != nil {
		return diag.FromErr(parseErr)
	}

	kafkaID := result[0]
	name := result[1]

	topic, _, getErr := client.Kafka.GetTopicByName(kafkaID, name)
	if getErr != nil {
		return diag.FromErr(getErr)
	}

	// Convert the remote retention time from milliseconds to duration unless it is disabled.
	var retentiontimeDuration string
	if topic.GetRetentionTimeInMS() == 0 {
		retentiontimeDuration = "disable"
	} else {
		var convErr error
		retentiontimeDuration, convErr = kafka.ConvertMillisecondstoDuration(topic.GetRetentionTimeInMS())
		if convErr != nil {
			return diag.FromErr(convErr)
		}
	}

	d.Set("kafka_id", kafkaID)
	d.Set("name", topic.GetName())
	d.Set("partitions", topic.GetPartitions())
	d.Set("replication_factor", topic.GetReplicationFactor())
	d.Set("retention_time", retentiontimeDuration)
	d.Set("compaction", topic.GetCompaction())
	d.Set("status", topic.GetStatus())
	d.Set("cleanup_policy", topic.GetCleanupPolicy())

	return nil
}

func resourceHerokuxKafkaTopicUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*Config)
	client := config.API

	opts := &kafka.TopicRequest{}
	kafkaID := getKakfaID(d)
	checkFuncs := make([]func(t *kafka.Topic) bool, 0)

	if v, ok := d.GetOk("name"); ok {
		vs := v.(string)
		log.Printf("[DEBUG] topic name is : %v", vs)
		opts.Name = vs
	}

	if v, ok := d.GetOk("compaction"); ok {
		vs := v.(bool)
		log.Printf("[DEBUG] topic compaction is : %v", vs)
		opts.Compaction = vs
	}

	if v, ok := d.GetOk("retention_time"); ok {
		duration := v.(string)
		log.Printf("[DEBUG] topic retention_time (string) is : %v", duration)

		// Convert the duration to milliseconds as that's what is supported by the API
		// unless the attribute value is `disable`. If it is disable,
		// then set the opts.RetentionTimeMS to nil.
		var targetRetentionTime int
		if duration == kafka.RetentionTimeDisableVal {
			opts.RetentionTimeMS = nil
			targetRetentionTime = 0
			log.Printf("[DEBUG] topic new retention_time (int) is : %v", "nil")
		} else {
			ms, conErr := kafka.ConvertDurationToMilliseconds(duration)
			if conErr != nil {
				return diag.FromErr(conErr)
			}
			opts.RetentionTimeMS = &ms
			targetRetentionTime = ms
			log.Printf("[DEBUG] topic new retention_time (int) is : %v", ms)
		}

		// The API generally requires retention_time being set in the PUT request body but if there's an update,
		// the resource needs to poll to check that the new retention time is applied correctly.
		if ok := d.HasChange("retention_time"); ok {
			// Setting checkFunc so the resource knows what to check for
			checkFuncs = append(checkFuncs, func(t *kafka.Topic) bool {
				return t.GetRetentionTimeInMS() == targetRetentionTime
			})
		}
	}

	if ok := d.HasChange("replication_factor"); ok {
		vs := d.Get("replication_factor").(int)
		log.Printf("[DEBUG] topic new replication_factor is : %v", vs)
		opts.ReplicationFactor = vs
	}

	log.Printf("[DEBUG] updating topic %s with %v", opts.Name, opts)

	_, _, updateErr := client.Kafka.UpdateTopic(kafkaID, opts)
	if updateErr != nil {
		return diag.FromErr(updateErr)
	}

	log.Printf("[DEBUG] Waiting for Kafka topic %s to be updated", opts.Name)
	stateConf := &resource.StateChangeConf{
		Pending:      []string{kafka.TopicStatuses.UPDATING.ToString()},
		Target:       []string{kafka.TopicStatuses.UPDATED.ToString()},
		Refresh:      topicUpdateStateRefreshFunc(client, kafkaID, opts.Name, checkFuncs),
		Timeout:      time.Duration(config.KafkaTopicCreateTimeout) * time.Minute,
		PollInterval: StateRefreshPollInterval,
	}

	if _, err := stateConf.WaitForStateContext(ctx); err != nil {
		return diag.Errorf("error waiting for topic to be updated on %s: %s", opts.Name, err.Error())
	}

	log.Printf("[DEBUG] updated topic %s with %v", opts.Name, opts)

	return resourceHerokuxKafkaTopicRead(ctx, d, meta)
}

// topicUpdateStateRefreshFunc checks if certain topic fields were updated remotely
// by executing custom functions passed in as function argument.
func topicUpdateStateRefreshFunc(client *api.Client, kafkaID, topicName string, checkFuncs []func(t *kafka.Topic) bool) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		topic, _, getErr := client.Kafka.GetTopicByName(kafkaID, topicName)
		if getErr != nil {
			return nil, kafka.TopicStatuses.UNKNOWN.ToString(), getErr
		}

		// Loop through all the checkFuncs. Return UPDATING if any of the functions return false
		for _, cf := range checkFuncs {
			if !cf(topic) {
				log.Printf("[DEBUG] topic not updated yet")
				return topic, kafka.TopicStatuses.UPDATING.ToString(), nil
			}
		}

		return topic, kafka.TopicStatuses.UPDATED.ToString(), nil
	}
}

func resourceHerokuxKafkaTopicDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*Config).API

	result, parseErr := parseCompositeID(d.Id(), 2)
	if parseErr != nil {
		return diag.FromErr(parseErr)
	}

	kafkaID := result[0]
	name := result[1]

	log.Printf("[DEBUG] Deleting Kafka topic %s", name)

	_, _, deleteErr := client.Kafka.DeleteTopic(kafkaID, name)
	if deleteErr != nil {
		return diag.FromErr(deleteErr)
	}

	log.Printf("[DEBUG] Deleted Kafka topic %s", name)

	d.SetId("")

	return nil
}
