package homie

import (
	"errors"
	"fmt"
	"path"
	"strings"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/sirupsen/logrus"
)

const (
	QOSAtMostOnce  = 0
	QOSAtLeastOnce = 1
	QOSExactlyOnce = 2
)

type Broker struct {
	client    mqtt.Client
	baseTopic string

	ActionHandler func(string, string, string) error
}

func New(server string) (*Broker, error) {
	baseTopic := "homie/raumfeld-bridge"

	opts := mqtt.NewClientOptions()
	opts.AddBroker(server)
	opts.SetAutoReconnect(true)
	opts.SetWill(path.Join(baseTopic, "$state"), "lost", QOSAtLeastOnce, true)

	client := mqtt.NewClient(opts)
	token := client.Connect()
	token.Wait()
	if token.Error() != nil {
		return nil, token.Error()
	}

	broker := &Broker{
		client:    client,
		baseTopic: baseTopic,
	}

	_ = client.Subscribe(path.Join(baseTopic, "+", "+", "set"), QOSAtMostOnce, broker.handleAction)
	// not sure what to do which the token

	return broker, nil
}

func (b *Broker) handleAction(client mqtt.Client, message mqtt.Message) {
	if b.ActionHandler == nil {
		message.Ack()
		return
	}

	topic := message.Topic()
	topic = strings.TrimPrefix(topic, b.baseTopic)
	topic = strings.TrimSuffix(topic, "set")
	topic = strings.Trim(topic, "/")

	nodeID, propertyID, ok := strings.Cut(topic, "/")
	if !ok {
		logrus.Errorf("invalid topic %#v", message.Topic())
		return
	}

	err := b.ActionHandler(nodeID, propertyID, string(message.Payload()))
	if err != nil {
		logrus.Error(err)
		return
	}

	message.Ack()
}

func (b *Broker) MustClose() {
	err := b.Close()
	if err != nil {
		logrus.Error(err)
	}
}

func (b *Broker) Close() error {
	err := b.publish("$state", "disconnected")
	if err != nil {
		return err
	}

	b.client.Disconnect(1000)
	return nil
}

func (d *Broker) PublishDevice(device Device) error {
	return errors.Join(
		d.publish("$homie", "4.0.0"),
		d.publish("$name", device.Name),
		d.publish("$state", "ready"),
		d.publish("$implementation", device.Implementation),
		d.publish("$nodes", strings.Join(device.NodeIDs, ",")),
	)
}

func (d *Broker) PublishNode(node Node) error {
	return errors.Join(
		d.publish(path.Join(node.NodeID, "$name"), node.Name),
		d.publish(path.Join(node.NodeID, "$type"), node.Type),
		d.publish(path.Join(node.NodeID, "$properties"), strings.Join(node.PropertyIDs, ",")),
	)
}

func (d *Broker) PublishProperty(property Property) error {
	prefix := path.Join(property.NodeID, property.PropertyID)
	return errors.Join(
		d.publish(path.Join(prefix, "$name"), property.Name),
		d.publish(path.Join(prefix, "$datatype"), property.DataType),
		d.publish(path.Join(prefix, "$format"), property.Format),
		d.publish(path.Join(prefix, "$unit"), property.Unit),
		d.publish(path.Join(prefix, "$settable"), fmt.Sprintf("%t", property.Settable)),
		d.publish(path.Join(prefix, "$retained"), fmt.Sprintf("%t", property.Retained)),
	)
}

func (d *Broker) PublishValue(nodeID, propertyID string, value any) error {
	return d.publish(path.Join(nodeID, propertyID), fmt.Sprint(value))
}

func (d *Broker) publish(topic string, message string) error {
	fullTopic := path.Join(d.baseTopic, topic)

	token := d.client.Publish(fullTopic, QOSAtLeastOnce, true, message)
	token.Wait()
	if token.Error() != nil {
		return fmt.Errorf("publish %q: %w", fullTopic, token.Error())
	}

	return nil
}
