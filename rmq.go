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

type queueResponse struct {
	Consumers int32 `json:"consumers"`
	Messages  int32 `json:"messages"`
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

func (rmq *rmq) getQueueInformation(queue string, vhost string) (int32, int32, error) {
	client := &http.Client{}
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/api/queues/%s/%s", rmq.URL, vhost, queue), nil)
	if err != nil {
		return 0, 0, err
	}
	req.SetBasicAuth(rmq.User, rmq.Password)
	resp, err := client.Do(req)

	if err != nil {
		return 0, 0, err
	}

	if resp.StatusCode != 200 {
		return 0, 0, errors.New(resp.Status)
	}

	defer resp.Body.Close()
	var data queueResponse
	json.NewDecoder(resp.Body).Decode(&data)

	return data.Consumers, data.Messages, nil
}
