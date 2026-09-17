package cronRepository

import "testing"

func TestHourField(t *testing.T) {
	var none [24]bool

	all := [24]bool{}
	for i := range all {
		all[i] = true
	}

	subset := [24]bool{}
	subset[0] = true
	subset[6] = true
	subset[12] = true
	subset[18] = true

	single := [24]bool{}
	single[9] = true

	cases := []struct {
		name      string
		sendtimes [24]bool
		want      string
	}{
		{"none set falls back to midnight", none, "0"},
		{"all set is hourly", all, "*"},
		{"subset is a sorted comma list", subset, "0,6,12,18"},
		{"single hour", single, "9"},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := hourField(c.sendtimes); got != c.want {
				t.Errorf("hourField() = %q, want %q", got, c.want)
			}
		})
	}
}

func TestBuildFTPUploadLine(t *testing.T) {
	all := [24]bool{}
	for i := range all {
		all[i] = true
	}

	cases := []struct {
		name      string
		enabled   bool
		sendtimes [24]bool
		want      string
	}{
		{"disabled is commented out", false, all, "#0 * * * * smartpi  /usr/local/bin/smartpiftpupload"},
		{"enabled hourly", true, all, "0 * * * * smartpi  /usr/local/bin/smartpiftpupload"},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := buildFTPUploadLine(c.enabled, c.sendtimes); got != c.want {
				t.Errorf("buildFTPUploadLine() = %q, want %q", got, c.want)
			}
		})
	}
}
