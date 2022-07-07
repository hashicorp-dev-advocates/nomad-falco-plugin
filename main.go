package main

import (
	"bytes"
	"context"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"time"

	"github.com/alecthomas/jsonschema"
	"github.com/falcosecurity/plugin-sdk-go/pkg/sdk"
	"github.com/falcosecurity/plugin-sdk-go/pkg/sdk/plugins"
	"github.com/falcosecurity/plugin-sdk-go/pkg/sdk/plugins/source"
	"github.com/hashicorp/nomad/api"
)

type PluginConfig struct {
	Address      string `json:"address" jsonschema:"title=address,description=The address of the Nomad API endpoint"`
	Token        string `json:"token" jsonschema:"title=token,description=The token to authenticate with against the Nomad API endpoint"`
	Index        uint64 `json:"index" jsonschema:"title=index,description=The index to start from"`
	UseAsync     bool   `json:"useAsync" jsonschema:"description=If true then async extraction optimization is enabled (Default: true)"`
	MaxEventSize uint64
}

func (c *PluginConfig) Reset() {
	c.Address = "http://localhost:4646"
	c.Token = ""
	c.Index = 0
	c.UseAsync = true
	c.MaxEventSize = uint64(sdk.DefaultEvtSize)
}

type Plugin struct {
	plugins.BasePlugin
	config PluginConfig
}

type PluginInstance struct {
	source.BaseInstance
}

func init() {
	plugins.SetFactory(func() plugins.Plugin {
		p := &Plugin{}
		source.Register(p)
		// extractor.Register(p)
		return p
	})
}

func (p *Plugin) Info() *plugins.Info {
	return &plugins.Info{
		ID:          999,
		Name:        "nomad",
		Description: "A Plugin that sources and extracts events from the Nomad event stream",
		Contact:     "github.com/hashicorp-dev-advocates/nomad-falco-plugin/",
		Version:     "0.1.0",
		EventSource: "nomad",
	}
}

func (p *Plugin) InitSchema() *sdk.SchemaInfo {
	schema, err := jsonschema.Reflect(&PluginConfig{}).MarshalJSON()
	if err == nil {
		return &sdk.SchemaInfo{
			Schema: string(schema),
		}
	}
	return nil
}

func (p *Plugin) Init(config string) error {
	p.config.Reset()
	json.Unmarshal([]byte(config), &p.config)
	return nil
}

// // Fields return the list of extractor fields exported by this plugin.
// // This method is mandatory the field extraction capability.
// // If the Fields method is defined, the framework expects an Extract method
// // to be specified too.
// func (p *Plugin) Fields() []sdk.FieldEntry {
// 	return []sdk.FieldEntry{
// 		{Type: "uint64", Name: "example.count", Display: "Counter value", Desc: "Current value of the internal counter"},
// 		{Type: "string", Name: "example.countstr", Display: "Counter string value", Desc: "String represetation of current value of the internal counter"},
// 	}
// }

// // This method is mandatory the field extraction capability.
// // If the Extract method is defined, the framework expects an Fields method
// // to be specified too.
// func (p *Plugin) Extract(req sdk.ExtractRequest, evt sdk.EventReader) error {
// 	var value uint64
// 	encoder := gob.NewDecoder(evt.Reader())
// 	if err := encoder.Decode(&value); err != nil {
// 		return err
// 	}

// 	switch req.FieldID() {
// 	case 0:
// 		req.SetValue(value)
// 		return nil
// 	case 1:
// 		req.SetValue(fmt.Sprintf("%d", value))
// 		return nil
// 	default:
// 		return fmt.Errorf("unsupported field: %s", req.Field())
// 	}
// }

func (p *Plugin) Open(params string) (source.Instance, error) {
	ctx := context.Background()
	eventCh := make(chan source.PushEvent)

	client, err := api.NewClient(&api.Config{
		Address: p.config.Address,
	})
	if err != nil {
		return nil, err
	}

	index := p.config.Index

	topics := map[api.Topic][]string{
		api.TopicAll: {"*"},
	}

	streamCh, err := client.EventStream().Stream(ctx, topics, index, &api.QueryOptions{
		WaitIndex: index,
		WaitTime:  30 * time.Second,
	})
	if err != nil {
		return nil, err
	}

	go func() {
		defer close(eventCh)
		for event := range streamCh {
			p.parseEventsAndPush(event, eventCh)
		}
	}()

	return source.NewPushInstance(
		eventCh,
		source.WithInstanceClose(func() { client.Close() }),
		source.WithInstanceEventSize(uint32(p.config.MaxEventSize)),
	)
}

func (p *Plugin) parseEventsAndPush(events *api.Events, output chan<- source.PushEvent) {
	for _, event := range events.Events {
		var buffer bytes.Buffer
		gob.NewEncoder(&buffer).Encode(event)
		output <- source.PushEvent{
			Err:       nil,
			Data:      buffer.Bytes(),
			Timestamp: time.Now(),
		}
	}
}

func (p *Plugin) String(evt sdk.EventReader) (string, error) {
	var value uint64
	encoder := gob.NewDecoder(evt.Reader())
	if err := encoder.Decode(&value); err != nil {
		return "", err
	}
	return fmt.Sprintf("event: %v", value), nil
}

func main() {}
