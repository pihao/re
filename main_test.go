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
		{"", modeExif, "testdata/canon_40d.jpg", "20080530075601", false},
		{"", modeExif, "testdata/jolla.jpg", "20140921080056", false},
		{"", modeExif, "testdata/lengtoo.jpg", "", true},

		// mtime 只验证结果长度, 因为这个时间会变化
		{"", modeMtime, "testdata/canon_40d.jpg", "", false},
		{"", modeMtime, "testdata/jolla.jpg", "", false},
		{"", modeMtime, "testdata/lengtoo.jpg", "", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			option.mode = tt.mode

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

			switch option.mode {
			case modeExif:
				if gotTname != tt.wantTname {
					t.Errorf("timename() = %v, want %v", gotTname, tt.wantTname)
				}
			case modeMtime:
				if len(gotTname) != 14 {
					t.Errorf("timename().len = %v, want %v", len(gotTname), 14)
				}
			default:
				t.Errorf("invalid mode: %s", tt.mode)
			}
		})
	}
}
