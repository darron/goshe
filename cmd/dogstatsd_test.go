package cmd

import (
	"testing"
)

func TestDogStatsdSetup(t *testing.T) {
	connection := DogStatsdSetup()
	if connection == nil {
		t.Error("Did not setup DogStatsd connection.")
	}
}
