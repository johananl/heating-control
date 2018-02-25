package controller

import "testing"

var c Controller = Controller{}

func TestParseReading(t *testing.T) {
	raw := `{
  "sensorID": "sensor-1",
  "type": "temperature",
  "value": 25.3
}`
	r, err := c.ParseReading([]byte(raw))
	if err != nil {
		t.Error(err)
	}

	if r.SensorID != "sensor-1" {
		t.Errorf("Wrong sensor ID: got %v want %v", r.SensorID, "sensor-1")
	}
	if r.ReadingType != "temperature" {
		t.Errorf("Wrong reading type: got %v want %v", r.ReadingType, "temperature")
	}
	if r.Value != 25.3 {
		t.Errorf("Wrong reading value: got %v want %v", r.Value, 25.3)
	}
}
