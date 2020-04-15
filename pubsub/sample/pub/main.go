package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/elvisNg/broccoli/config"
	brokerpb "github.com/elvisNg/broccoli/pubsub/proto"
	zpub "github.com/elvisNg/broccoli/pubsub/pub"
	structpb "github.com/golang/protobuf/ptypes/struct"
	"github.com/micro/go-micro/metadata"
)

func main() {
	pub()
}

type Request struct {
	Message  string `json:"message"`
	Count    int32  `json:"count"`
	Finished bool   `json:"finished"`
}

func pub() {
	conf := &config.Broker{}

	// conf.Type = "redis"
	// conf.TopicPrefix = "dev"

	conf.Type = "kafka"
	conf.TopicPrefix = "dev"
	conf.Hosts = []string{"10.1.8.14:9094"}

	zpub.InitDefault(conf)
	for {
		h := new(brokerpb.Header)
		h.Id = fmt.Sprint(time.Now().Unix())
		h.Category = "sample"
		h.Source = "broccoli"
		body := new(brokerpb.Sample_Body)
		body.Id = "body-" + h.Id
		body.Name = "test Sample"
		msg := new(brokerpb.Sample)
		msg.Header = h
		msg.Body = body
		log.Printf("Publish Sample message: %+v\n", msg)
		err := zpub.Publish(context.Background(), h, msg)
		if err != nil {
			fmt.Println("err==>>", err)
		}

		sreq := &brokerpb.RequestSample{Message: "hello world!", Count: 100, Finished: true}
		log.Printf("Publish SampleRequest message: %+v\n", sreq)
		err = zpub.Publish(context.Background(), &brokerpb.Header{Id: fmt.Sprint(time.Now().Unix()), Category: "samplerequest", Source: "broccoli"}, sreq)
		if err != nil {
			fmt.Println("err==>>", err)
		}

		sMsg := &structpb.Struct{
			Fields: map[string]*structpb.Value{
				"message": &structpb.Value{
					Kind: &structpb.Value_StringValue{StringValue: "hello world!"},
				},
				"count": &structpb.Value{
					Kind: &structpb.Value_NumberValue{NumberValue: 100},
				},
				"finished": &structpb.Value{
					Kind: &structpb.Value_BoolValue{BoolValue: true},
				},
			},
		}
		log.Printf("Publish pbstruct message: %+v\n", sMsg)
		err = zpub.Publish(context.Background(), &brokerpb.Header{Id: fmt.Sprint(time.Now().Unix()), Category: "pbstruct", Source: "broccoli"}, sMsg)
		if err != nil {
			fmt.Println("err==>>", err)
		}

		req := &Request{Message: "hello world!", Count: 100, Finished: true}
		log.Printf("Publish JSON Request message: %+v\n", req)
		err = zpub.Publish(metadata.NewContext(context.Background(), metadata.Metadata{"for-handler": "json-request"}), &brokerpb.Header{Id: fmt.Sprint(time.Now().Unix()), Category: "jsonrequest", Source: "broccoli"}, req)
		if err != nil {
			fmt.Println("err==>>", err)
		}

		time.Sleep(time.Second)
	}
}
