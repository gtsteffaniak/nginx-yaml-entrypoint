package main

import (
	"bytes"
	"reflect"
	"testing"
)

func Test_mergeLocations(t *testing.T) {
	type args struct {
		keep  Location
		merge Location
	}
	tests := []struct {
		name string
		args args
		want Location
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := mergeLocations(tt.args.keep, tt.args.merge); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("mergeLocations() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_mergeStringMaps(t *testing.T) {
	type args struct {
		keep  map[string]string
		merge map[string]string
	}
	tests := []struct {
		name string
		args args
		want map[string]string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := mergeStringMaps(tt.args.keep, tt.args.merge); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("mergeStringMaps() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_mergeConditions(t *testing.T) {
	type args struct {
		keep  []Condition
		merge []Condition
	}
	tests := []struct {
		name string
		args args
		want []Condition
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := mergeConditions(tt.args.keep, tt.args.merge); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("mergeConditions() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_mergeLimitReq(t *testing.T) {
	type args struct {
		keep  LimitReqZone
		merge LimitReqZone
	}
	tests := []struct {
		name string
		args args
		want LimitReqZone
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := mergeLimitReq(tt.args.keep, tt.args.merge); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("mergeLimitReq() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_getGlobalZone(t *testing.T) {
	type args struct {
		name string
	}
	tests := []struct {
		name string
		args args
		want LimitReqZone
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getGlobalZone(tt.args.name); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getGlobalZone() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_setKeys(t *testing.T) {
	type args struct {
		m   map[string]string
		key string
	}
	tests := []struct {
		name string
		args args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setKeys(tt.args.m, tt.args.key)
		})
	}
}

func Test_setKey(t *testing.T) {
	type args struct {
		key string
		s   string
	}
	tests := []struct {
		name string
		args args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setKey(tt.args.key, tt.args.s)
		})
	}
}

func TestWriteOutput(t *testing.T) {
	type args struct {
		output bytes.Buffer
	}
	tests := []struct {
		name string
		args args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			WriteOutput(tt.args.output)
		})
	}
}

func Test_setConditions(t *testing.T) {
	type args struct {
		conditions []Condition
	}
	tests := []struct {
		name string
		args args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setConditions(tt.args.conditions)
		})
	}
}
