package main

import (
	"os"
	"testing"
	"time"
)

func Test_timename(t *testing.T) {
	loc = time.UTC

	tests := []struct {
		name      string
		mode      string
		path      string
		wantTname string
		wantErr   bool
	}{
		{"", modeEXIF, "testdata/canon_40d.jpg", "20080530075601", false},
		{"", modeEXIF, "testdata/jolla.jpg", "20140921080056", false},
		{"", modeEXIF, "testdata/lengtoo.jpg", "", true},

		{"", modeMtime, "testdata/canon_40d.jpg", "20200413094815", false},
		{"", modeMtime, "testdata/jolla.jpg", "20200413094831", false},
		{"", modeMtime, "testdata/lengtoo.jpg", "20200413094637", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mode = tt.mode
			info, err := os.Stat(tt.path)
			if err != nil {
				t.Errorf("test file not exist: %v", tt.path)
				return
			}
			gotTname, err := timename(tt.path, info)
			if (err != nil) != tt.wantErr {
				t.Errorf("timename() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotTname != tt.wantTname {
				t.Errorf("timename() = %v, want %v", gotTname, tt.wantTname)
			}
		})
	}
}
