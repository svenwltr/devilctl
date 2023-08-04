package homie

import (
	"errors"
	"fmt"
	"path"
	"strings"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/sirupsen/logrus"
	"golang.org/x/exp/slices"
)

const (
	QOSAtMostOnce  = 0
	QOSAtLeastOnce = 1
	QOSExactlyOnce = 2
)

type Device struct {
	client    mqtt.Client
	baseTopic string

	Nodes         Nodes
	ActionHandler func(string, string, string) error
}

func New(broker string) (*Device, error) {
	baseTopic := "homie/raumfeld-bridge"

	opts := mqtt.NewClientOptions()
	opts.AddBroker(broker)
	opts.SetAutoReconnect(true)
	opts.SetWill(path.Join(baseTopic, "$state"), "lost", QOSAtLeastOnce, true)

	client := mqtt.NewClient(opts)
	token := client.Connect()
	token.Wait()
	if token.Error() != nil {
		return nil, token.Error()
	}

	device := &Device{
		client:    client,
		baseTopic: baseTopic,
	}

	_ = client.Subscribe(path.Join(baseTopic, "+", "+", "set"), QOSAtMostOnce, device.handleAction)
	// not sure what to do which the token

	return device, nil
}

func (d *Device) handleAction(client mqtt.Client, message mqtt.Message) {
	if d.ActionHandler == nil {
		message.Ack()
		return
	}

	topic := message.Topic()
	topic = strings.TrimPrefix(topic, d.baseTopic)
	topic = strings.TrimSuffix(topic, "set")
	topic = strings.Trim(topic, "/")

	nodeID, propertyID, ok := strings.Cut(topic, "/")
	if !ok {
		logrus.Errorf("invalid topic %#v", message.Topic())
		return
	}

	err := d.ActionHandler(nodeID, propertyID, string(message.Payload()))
	if err != nil {
		logrus.Error(err)
		return
	}

	message.Ack()
}

func (d *Device) MustClose() {
	err := d.Close()
	if err != nil {
		logrus.Error(err)
	}
}

func (d *Device) Close() error {
	err := d.publish("$state", "disconnected")
	if err != nil {
		return err
	}

	d.client.Disconnect(1000)
	return nil
}

func (d *Device) PublishAll() error {
	nodeIDs := []string{}
	for nodeID, node := range d.Nodes {
		nodeIDs = append(nodeIDs, nodeID)

		propertyIDs := []string{}
		for propertyID, property := range node.Properties {
			propertyIDs = append(propertyIDs, propertyID)

			err := errors.Join(
				d.publish(path.Join(nodeID, propertyID, "$name"), property.Name),
				d.publish(path.Join(nodeID, propertyID, "$datatype"), property.DataType),
				d.publish(path.Join(nodeID, propertyID, "$format"), property.Format),
				d.publish(path.Join(nodeID, propertyID, "$unit"), property.Unit),
				d.publish(path.Join(nodeID, propertyID, "$settable"), fmt.Sprintf("%t", property.Settable)),
				d.publish(path.Join(nodeID, propertyID, "$retained"), fmt.Sprintf("%t", property.Retained)),
			)
			if err != nil {
				return err
			}
		}

		slices.Sort(propertyIDs)
		err := errors.Join(
			d.publish(path.Join(nodeID, "$name"), node.Name),
			d.publish(path.Join(nodeID, "$type"), node.Type),
			d.publish(path.Join(nodeID, "$properties"), strings.Join(propertyIDs, ",")),
		)
		if err != nil {
			return err
		}
	}

	err := errors.Join(
		d.publish("$homie", "4.0.0"),
		d.publish("$name", "Homie Raumfeld Bridge"),
		d.publish("$state", "ready"),
		d.publish("$implementation", "github.com/svenwltr/devilctl"),
		d.publish("$nodes", strings.Join(nodeIDs, ",")),
	)

	if err != nil {
		return err
	}

	return nil
}

func (d *Device) Value(nodeID, propertyID string, value any) error {
	return d.publish(path.Join(nodeID, propertyID), fmt.Sprint(value))
}

func (d *Device) publish(topic string, message string) error {
	fullTopic := path.Join(d.baseTopic, topic)

	token := d.client.Publish(fullTopic, QOSAtLeastOnce, true, message)
	token.Wait()
	if token.Error() != nil {
		return fmt.Errorf("publish %q: %w", fullTopic, token.Error())
	}

	return nil
}
