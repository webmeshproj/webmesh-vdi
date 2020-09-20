package gst

// PipelineConfig represents a list of elements and their configurations
// to be used with NewPipelineFromConfig.
type PipelineConfig struct {
	Plugins []*Plugin
}

// GetPluginByName returns the Plugin configuration for the given name.
func (p *PipelineConfig) GetPluginByName(name string) *Plugin {
	for _, plugin := range p.Plugins {
		if name == plugin.GetName() {
			return plugin
		}
	}
	return nil
}

// PluginNames returns a string slice of the names of all the plugins.
func (p *PipelineConfig) PluginNames() []string {
	names := make([]string, 0)
	for _, plugin := range p.Plugins {
		names = append(names, plugin.GetName())
	}
	return names
}

// Plugin represents an `GstElement` in a `GstPipeline` when building a Pipeline with `NewPipelineFromConfig`.
// The Name should coorespond to a valid gstreamer plugin name. The data are additional
// fields to set on the element. If SinkCaps is non-nil, they are applied to the sink of this
// element.
//
// If `InternalSink` is true, the `Name` field is ignored and the sink is tied to
// the Pipeline's read-buffer. Likewise, if `InternalSource` is true, the source
// is tied to the Pipeline's write-buffer. Plugins with these fields must be set
// first and last respectively.
type Plugin struct {
	Name                         string
	SinkCaps                     *Caps
	Data                         map[string]interface{}
	InternalSource, InternalSink bool
}

// GetName returns the name to use when creating Elements from this configuration.
func (p *Plugin) GetName() string {
	if p.InternalSource {
		return "fdsrc"
	}
	if p.InternalSink {
		return "fdsink"
	}
	return p.Name
}
