package systemctl

import (
	"bytes"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
)

// DurFieldSuffix is the suffix appended to create unique field names for state duration
const DurFieldSuffix string = "_dur"

// CountFieldSuffix is the suffix appended to create unique field names for state count
const CountFieldSuffix string = "_count"

// Systemctl is an telegraf serive input plugin for systemctl service state
type Systemctl struct {
	ServiceName string
	SampleRate  int

	mux                  sync.Mutex
	AggState             map[string]uint64
	CurrentState         string
	CurrentStateDuration uint64
	Sampler              StateSampler
	running              bool
	done                 chan bool
	collect              chan bool
	out                  chan []Sample
}

// Sample is a single sample of systemctl service state with timestamp
type Sample struct {
	name      string
	timestamp uint64
}

// State is a systemctl service state and it's duration
type State struct {
	name     string
	duration uint64
}

// StateSampler is an interface to collect state samples
type StateSampler interface {
	Sample() Sample
}

const sampleConfig = `
	sample_rate = 2
	# service_name = "nginx"
`

// SampleConfig returns the sample configuration for the input plugin
func (s *Systemctl) SampleConfig() string {
	return sampleConfig
}

// Description returns a short description of the input pluging
func (s *Systemctl) Description() string {
	return "Systemctl state monitor"
}

// Gather collects the samples and adds the AggState to the telegraf accumulator
func (s *Systemctl) Gather(acc telegraf.Accumulator) error {
	s.mux.Lock()
	defer s.mux.Unlock()

	fmt.Printf("Gathering samples in goroutine\n")
	// notify sampler of collection
	s.collect <- true
	// read samples
	samples := <-s.out
	// aggregate samples into states
	states, err := s.AggregateSamples(samples)
	if err != nil {
		return err
	}
	// add states to aggregation
	s.AggregateStates(states, s.CurrentStateDuration)
	// create fields
	fields := map[string]interface{}{
		"current_state_time": s.CurrentStateDuration,
		"current_state":      s.CurrentState,
	}
	for k := range s.AggState {
		fields[k] = s.AggState[k]
	}
	// create tag
	tags := map[string]string{"resource": s.ServiceName}
	acc.AddFields("service_config_state", fields, tags)
	fmt.Printf("Added fields\n")
	return nil
}

// AggregateSamples creates states and their duration from a set of samples
func (s *Systemctl) AggregateSamples(samples []Sample) ([]State, error) {
	sampleCount := len(samples)
	// error if no samples to aggregate
	if sampleCount < 2 {
		return nil, errors.New("2 or more samples needed for aggregation")
	}

	states := make([]State, 0)
	var stateTime uint64
	var stateStartTime uint64
	// for the 1st sample we set the current state and state_start_time
	currentState := samples[0].name
	stateStartTime = samples[0].timestamp
	lastIndex := sampleCount - 1
	for i := 1; i < sampleCount; i++ {
		if currentState != samples[i].name {
			// calc duration in current state
			stateTime = samples[i].timestamp - stateStartTime
			states = append(states, State{name: currentState, duration: stateTime})
			// set the new start time and current state
			stateStartTime = stateStartTime + stateTime
			currentState = samples[i].name
		}
		// if the last sample
		if i == lastIndex {
			// if transition in the last sample, add last state with zero duration
			if currentState != samples[i].name {
				// add next state with zero transition
				states = append(states, State{name: samples[i].name, duration: 0})
			} else {
				// calc duration in current state
				stateTime = samples[i].timestamp - stateStartTime
				states = append(states, State{name: currentState, duration: stateTime})
			}
		}
	}
	return states, nil
}

// GetKeyName creates concatinates two strings to create a key name for the AggState map
func GetKeyName(currentState string, suffix string) string {
	var b bytes.Buffer

	b.WriteString(currentState)
	b.WriteString(suffix)

	return b.String()
}

// AggregateStates creates AggState from a set of states
func (s *Systemctl) AggregateStates(states []State, currentStateDuration uint64) {
	var stateDurKey string
	var stateCountKey string
	var containsState bool

	// set initial state to the 1st sample
	initialState := states[0].name
	// set the current state as the last state sampled
	stateCount := len(states)
	currentState := states[stateCount-1].name
	// if no change in state  take the initial state time and add current state time
	if initialState == currentState && stateCount == 1 {
		currentStateDuration += states[stateCount-1].duration
		stateDurKey = GetKeyName(currentState, DurFieldSuffix)
		stateCountKey = GetKeyName(currentState, CountFieldSuffix)
		// initialise the number of transitions if it's the 1st time
		_, containsState = s.AggState[stateDurKey]
		if !containsState {
			s.AggState[stateDurKey] = currentStateDuration
			s.AggState[stateCountKey] = 1
		}
	} else {
		// current state time is the last state time
		currentStateDuration = states[stateCount-1].duration
		// calc the total duration and number of transitions in each state.
		for i := 0; i < stateCount; i++ {
			// if first occurance of state add with initial duration and a single transition
			stateDurKey = GetKeyName(states[i].name, DurFieldSuffix)
			stateCountKey = GetKeyName(states[i].name, CountFieldSuffix)
			_, containsState = s.AggState[stateDurKey]
			if !containsState {
				s.AggState[stateCountKey] = 1
				s.AggState[stateDurKey] = states[i].duration
			} else {
				// increment number of times in the state
				s.AggState[stateCountKey]++
				// add state time to aggregate total
				s.AggState[stateDurKey] += states[i].duration
			}
		}
	}

	s.CurrentState = currentState
	s.CurrentStateDuration = currentStateDuration
}

// Start starts the sampling of systemctl state for the configured services
func (s *Systemctl) Start(acc telegraf.Accumulator) {
	s.mux.Lock()
	defer s.mux.Unlock()
	if s.running {
		return
	}

	go s.CollectSamples(s.done, s.collect, s.out)
	s.running = true
}

// CollectSamples samples system control state
func (s *Systemctl) CollectSamples(done chan bool, collect chan bool, out chan []Sample) {
	samples := make([]Sample, 0)
	for {
		select {
		default:
			fmt.Printf("Getting sample in go routine\n")
			sample := s.Sampler.Sample()
			samples = append(samples, sample)
			time.Sleep(time.Duration(s.SampleRate) * time.Second)
		case <-collect:
			fmt.Printf("Collecting samples in goroutine\n")
			out <- samples
			lastSample := samples[len(samples)-1]
			samples = make([]Sample, 1)
			samples[0] = lastSample
		case <-done:
			fmt.Printf("retuning from goroutine\n")
			return
		}
	}
}

// Stop stops the sampling of systemctrl state for the configured services
func (s *Systemctl) Stop(acc telegraf.Accumulator) {
	s.mux.Lock()
	defer s.mux.Unlock()
	if !s.running {
		return
	}

	s.done <- false
	s.running = false
}

// MyStateSampler is a sampler for systemctl states
type MyStateSampler struct{}

// Sample implements state sampling
func (t *MyStateSampler) Sample() Sample {
	timestamp := time.Now().UnixNano()
	fmt.Printf("Time stamp %d\n", timestamp)
	s := Sample{name: "active", timestamp: uint64(timestamp)}
	return s
}

func init() {
	inputs.Add("systemctl", func() telegraf.Input {
		return &Systemctl{
			ServiceName:          "nginx",
			SampleRate:           2,
			AggState:             make(map[string]uint64),
			CurrentState:         "unknown",
			CurrentStateDuration: 0,
			Sampler:              &MyStateSampler{},

			done:    make(chan bool),
			collect: make(chan bool),
			out:     make(chan []Sample),
		}
	})
}
