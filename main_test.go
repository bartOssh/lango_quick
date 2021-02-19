package main

import (
	"fmt"
	"testing"

	"go.uber.org/goleak"
)

func TestMain(m *testing.M) {
	goleak.VerifyTestMain(m)
}

func Test_main(t *testing.T) {
	tests := []struct {
		name string
	}{
		{name: "Test1"},
		{name: "Test2"},
		{name: "Test3"},
		{name: "Test4"},
		{name: "Test5"},
	}
	for _, tt := range tests {
		fmt.Printf(" test: %s \n", tt.name)
		t.Run(tt.name, func(t *testing.T) {
			main()
		})
	}
}
