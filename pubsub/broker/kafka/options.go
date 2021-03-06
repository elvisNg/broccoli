package kafka

import (
	"context"

	"github.com/Shopify/sarama"
	"github.com/micro/go-micro/broker"
)

var (
	DefaultBrokerConfig  = sarama.NewConfig()
	DefaultClusterConfig = sarama.NewConfig()
)

type brokerConfigKey struct{}
type clusterConfigKey struct{}

func BrokerConfig(c *sarama.Config) broker.Option {
	return setBrokerOption(brokerConfigKey{}, c)
}

func ClusterConfig(c *sarama.Config) broker.Option {
	return setBrokerOption(clusterConfigKey{}, c)
}

type subscribeContextKey struct{}

// SubscribeContext set the context for broker.SubscribeOption
func SubscribeContext(ctx context.Context) broker.SubscribeOption {
	return setSubscribeOption(subscribeContextKey{}, ctx)
}

// consumerGroupHandler is the implementation of sarama.ConsumerGroupHandler
type consumerGroupHandler struct {
	handler broker.Handler
	subopts broker.SubscribeOptions
	kopts   broker.Options
	cg      sarama.ConsumerGroup
	sess    sarama.ConsumerGroupSession
}

func (*consumerGroupHandler) Setup(_ sarama.ConsumerGroupSession) error   { return nil }
func (*consumerGroupHandler) Cleanup(_ sarama.ConsumerGroupSession) error { return nil }
func (h *consumerGroupHandler) ConsumeClaim(sess sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for msg := range claim.Messages() {
		var m broker.Message
		if err := h.kopts.Codec.Unmarshal(msg.Value, &m); err != nil {
			continue
		}
		if _, ok := m.Header["Micro-Id"]; !ok {
			// {"Header":{"Content-Type":"application/json","Micro-Id":"990001ec-d996-4da7-b04b-ba5e6e5c7236","Micro-Topic":"dev.sample.broccoli"}
			m.Header = map[string]string{
				"Content-Type": "application/json",
				"Broker":       "kafka",
				"Topic":        msg.Topic,
				// "Micro-Id":     "kafka-broker",
				// "Micro-Topic":  msg.Topic,
			}
			m.Body = msg.Value
		}
		m.Header["GroupID"] = h.subopts.Queue
		if err := h.handler(&publication{
			m:    &m,
			t:    msg.Topic,
			km:   msg,
			cg:   h.cg,
			sess: sess,
		}); err == nil && h.subopts.AutoAck {
			sess.MarkMessage(msg, "")
		}
	}
	return nil
}
