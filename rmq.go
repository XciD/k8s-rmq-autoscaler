package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
)

type rmq struct {
	URL      string
	User     string
	Password string
}

type queueStatSample struct {
	Timestamp int32 `json:"timestamp"`
	Sample    int32 `json:"sample"`
}

type queueStatDetails struct {
	Samples []queueStatSample `json:"samples"`
}

type queueMessageStats struct {
	PublishDetails queueStatDetails `json:"publish_details"`
}

type queueResponse struct {
	Consumers    int32             `json:"consumers"`
	Messages     int32             `json:"messages"`
	MessageStats queueMessageStats `json:"message_stats"`
}

func newRmq(rmqURL string, rmqUser string, rmqPassword string) (*rmq, error) {

	if len(rmqURL) == 0 || len(rmqUser) == 0 || len(rmqPassword) == 0 {
		return nil, errors.New("missing rmq information")
	}

	return &rmq{
		URL:      rmqURL,
		User:     rmqUser,
		Password: rmqPassword,
	}, nil
}

func (rmq *rmq) getQueueInformation(queue string, vhost string, cooldown int32) (int32, int32, int32, error) {
	client := &http.Client{}
	var statsAge int32
	statsAge = max(cooldown, 1)
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/api/queues/%s/%s?columns=messages,consumers,message_stats.publish_details&msg_rates_age=%d&msg_rates_incr=%d", rmq.URL, vhost, queue, statsAge, statsAge), nil)
	if err != nil {
		return 0, 0, 0, err
	}
	req.SetBasicAuth(rmq.User, rmq.Password)
	resp, err := client.Do(req)

	if err != nil {
		return 0, 0, 0, err
	}

	if resp.StatusCode != 200 {
		return 0, 0, 0, errors.New(resp.Status)
	}

	defer resp.Body.Close()
	var data queueResponse
	json.NewDecoder(resp.Body).Decode(&data)

	var messagesPublished int32
	if cooldown == 0 || len(data.MessageStats.PublishDetails.Samples) < 2 {
		messagesPublished = 0
	} else {
		messagesPublished = data.MessageStats.PublishDetails.Samples[0].Sample - data.MessageStats.PublishDetails.Samples[1].Sample
	}

	return data.Consumers, data.Messages, messagesPublished, nil
}
