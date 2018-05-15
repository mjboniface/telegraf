package systemd

import (
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
)

type Systemd struct {
	SampleRate int
}

type StateSample struct {
	StateName string
	SampleTime int
}

type State struct {
	StateName string
	StateDuration int
}

const sampleConfig = `
	sample_rate = 2
`

func (s *Systemd) SampleConfig() string {
	return sampleConfig
}

func (h *Systemd) Description() string {
	return "Systemd state monitor"
}

func (h *Systemd) Gather(acc telegraf.Accumulator) error {

	tags := map[string]string{"server": "host", "port": "port"}	
	fields := map[string]interface{}{
		"active":   10,
		"accepts":  20,
	}

	acc.AddFields("systemd", fields, tags)

	return nil
}

func (h *Systemd) CreateMeasurement(samples []StateSample, initial_state_time int, current_time int) map[string]int {

	states := h.AggregateSamples(samples) 
	fields := h.AggregateStates(states, initial_state_time)

	return fields
}

func (h *Systemd) AggregateSamples(samples []StateSample) []State {
//	states := make([]State, len(samples))

	return nil
}

func (h *Systemd) AggregateStates(states []State, initial_state_time int) map[string]int {
	fields := map[string]int{
		"active":   10,
		"accepts":  20,
	}

	return fields
}

func (h *Systemd) Start(acc telegraf.Accumulator) error {
	return nil
}

func (h *Systemd) Stop(acc telegraf.Accumulator) error {
	return nil
}

func init() {
	inputs.Add("systemd", func() telegraf.Input {
		return &Systemd{}
	})
}

