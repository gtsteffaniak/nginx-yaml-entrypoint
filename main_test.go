package main

import "testing"

func Test_createMainNginxConfig(t *testing.T) {
	tests := []struct {
		name string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			createMainNginxConfig()
		})
	}
}

func Test_setInfo(t *testing.T) {
	type args struct {
		location Location
		indent   string
	}
	tests := []struct {
		name string
		args args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setInfo(tt.args.location, tt.args.indent)
		})
	}
}

func Test_createServer(t *testing.T) {
	type args struct {
		i      int
		server NginxServer
	}
	tests := []struct {
		name string
		args args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			createServer(tt.args.i, tt.args.server)
		})
	}
}

func Test_main(t *testing.T) {
	tests := []struct {
		name string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			main()
		})
	}
}
