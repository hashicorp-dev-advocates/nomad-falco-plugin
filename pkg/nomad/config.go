package nomad

// Defining a type for the plugin configuration.
// In this simple example, users can define the starting value the event
// counter. the `jsonschema` tags is used to automatically generate a
// JSON Schema definition, so that the framework can perform automatic
// validations.
type PluginConfig struct {
	// this will contain any configuration variables used by the plugin
	// and passed-in by the user
	Address string `json:"address" jsonschema:"title=Nomad API Address,Default=http://localhost:4646"`
}

// Resets sets the configuration to its default values
func (p *PluginConfig) Reset() {
	// reset the config values
}
