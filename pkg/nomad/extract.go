package nomad

import (
	"encoding/json"
	"fmt"

	"github.com/falcosecurity/plugin-sdk-go/pkg/sdk"
	"github.com/hashicorp/nomad/api"
)

// Fields return the list of extractor fields exported by this plugin.
// This method is mandatory the field extraction capability.
// If the Fields method is defined, the framework expects an Extract method
// to be specified too.
func (p *Plugin) Fields() []sdk.FieldEntry {
	return []sdk.FieldEntry{
		{
			// we currently support uint64 and string
			Type:    "uint64",
			Name:    "nomad.index",
			Display: "Event index",
			Desc:    "The index of the Nomad event",
			// fields can extract a single value, or a list of values
			IsList: false,
			// fields can have an argument
			Arg: sdk.FieldEntryArg{
				IsRequired: false,
			},
		},
		{
			Type:    "string",
			Name:    "nomad.topic",
			Display: "Event topic",
			Desc:    "The topic of the Nomad event",
		},
		{
			Type:    "string",
			Name:    "nomad.type",
			Display: "Event type",
			Desc:    "The type of the Nomad event",
		},
		{
			Type:    "string",
			Name:    "nomad.job.images",
			Display: "Job images",
			Desc:    "Docker images used in the Nomad job",
			IsList:  true,
		},
	}
}

// func getJobImageName(img string) string {
// 	tokens := strings.Split(img, ":")
// 	if len(tokens) == 2 {
// 		return tokens[0]
// 	}
// 	return ""
// }

// func getJobImageVersion(img string) string {
// 	tokens := strings.Split(img, ":")
// 	if len(tokens) == 2 {
// 		return tokens[1]
// 	}
// 	return ""
// }

func getJobImages(evt *api.Event) ([]string, error) {
	job, err := evt.Job()
	if err != nil {
		return nil, err
	}

	var images []string
	for _, group := range job.TaskGroups {
		for _, task := range group.Tasks {
			if task.Driver == "docker" {
				image, ok := task.Config["image"].(string)
				if ok {
					images = append(images, image)
				}
			}
		}
	}

	return images, nil
}

// This method is mandatory the field extraction capability.
// If the Extract method is defined, the framework expects an Fields method
// to be specified too.
func (p *Plugin) Extract(req sdk.ExtractRequest, evt sdk.EventReader) error {
	if p.lastEvtNum != evt.EventNum() {
		encoder := json.NewDecoder(evt.Reader())
		if err := encoder.Decode(&p.evt); err != nil {
			return err
		}
	}

	switch req.Field() {
	case "nomad.index":
		req.SetValue(p.evt.Index)

	case "nomad.topic":
		req.SetValue(string(p.evt.Topic))

	case "nomad.type":
		req.SetValue(p.evt.Type)

	case "nomad.job.images":
		images, err := getJobImages(&p.evt)
		if err == nil {
			req.SetValue(images)
		}
	default:
		return fmt.Errorf("unsupported field: %s", req.Field())
	}

	return nil
}
