package systemd

import (
	"testing"

	"github.com/stretchr/testify/assert"

	
//	"github.com/influxdata/telegraf/testutil"
)

func TestAgg(t *testing.T) {
	var samples = []StateSample{
		StateSample {
			StateName: "active",
			SampleTime: 0, 
		},
		StateSample {
			StateName: "active", 
			SampleTime: 2,
		},
	}

	s := &Systemd{
		SampleRate: 2,
	}

	//var acc testutil.Accumulator

	//s.Gather(&acc)

	var out_fields = s.CreateMeasurement(samples, 10, 12)
	
	fields := make(map[string]interface{})
    fields["current_state"] = "active"
    fields["current_state_time"] = 12
    fields["active_sum"] = 12
	fields["active_count"] = 1	
	
    assert.Equal(t, out_fields["current_state"], "active")
    assert.Equal(t, out_fields["current_state_time"], 12)
	assert.Equal(t, out_fields["active_sum"], 12)
	assert.Equal(t, out_fields["active_count"], 1)
 	

	//acc.AssertContainsFields(t, "systemd", fields)

/*
	t = ConfigCollector(get_sample_test, write_output, "resource")    
    measurement = t.create_measurement(samples[0], 10, 12)
    assert measurement[0]['fields']['current_state'] == 'active'
    assert measurement[0]['fields']['current_state_time'] == 12
    assert measurement[0]['fields']['active_sum'] == 12
    assert measurement[0]['fields']['active_count'] == 1
    assert measurement[0]['time'] == 12000000000    

    t = ConfigCollector(get_sample_test, write_output, "resource")
    measurement = t.create_measurement(samples[1], 10, 14)
    assert measurement[0]['fields']['current_state'] == 'active'
    assert measurement[0]['fields']['current_state_time'] == 14
    assert measurement[0]['fields']['active_sum'] == 14
    assert measurement[0]['fields']['active_count'] == 1
    assert measurement[0]['time'] == 14000000000    

    t = ConfigCollector(get_sample_test, write_output, "resource")
    measurement = t.create_measurement(samples[2], 8, 10)
    assert measurement[0]['fields']['current_state'] == 'failed'
    assert measurement[0]['fields']['current_state_time'] == 0
    assert measurement[0]['fields']['active_sum'] == 2
    assert measurement[0]['fields']['active_count'] == 1
    assert measurement[0]['fields']['failed_sum'] == 0
    assert measurement[0]['fields']['failed_count'] == 1
    assert measurement[0]['time'] == 10000000000    

    t = ConfigCollector(get_sample_test, write_output, "resource")
    measurement = t.create_measurement(samples[3], 2, 12)
    assert measurement[0]['fields']['current_state'] == 'inactive'
    assert measurement[0]['fields']['current_state_time'] == 0
    assert measurement[0]['fields']['active_sum'] == 6
    assert measurement[0]['fields']['active_count'] == 2
    assert measurement[0]['fields']['inactive_sum'] == 2
    assert measurement[0]['fields']['inactive_count'] == 2
    assert measurement[0]['fields']['failed_sum'] == 2
    assert measurement[0]['fields']['failed_count'] == 1
    assert measurement[0]['time'] == 12000000000       
    
    t = ConfigCollector(get_sample_test, write_output, "resource")    
    measurement = t.create_measurement(samples[4], 4, 14)
    assert measurement[0]['fields']['current_state'] == 'failed'
    assert measurement[0]['fields']['current_state_time'] == 0
    assert measurement[0]['fields']['active_sum'] == 4
    assert measurement[0]['fields']['active_count'] == 2
    assert measurement[0]['fields']['inactive_sum'] == 4
    assert measurement[0]['fields']['inactive_count'] == 2
    assert measurement[0]['fields']['failed_sum'] == 2
    assert measurement[0]['fields']['failed_count'] == 2
	assert measurement[0]['time'] == 14000000000   	
	*/
}