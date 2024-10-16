package sqlfmt

import (
	"testing"
)

func TestCompare(t *testing.T) {
	test := struct {
		before string
		after  string
		want   bool
	}{
		before: "select * from xxx",
		after:  "select\n  *\nFROM xxx",
		want:   true,
	}
	if got := compare(test.before, test.after); got != test.want {
		t.Errorf("want %v#v got %#v", test.want, got)
	}
}

func TestRemove(t *testing.T) {
	got := removeSpace("select xxx from xxx")
	want := "selectxxxfromxxx"
	if got != want {
		t.Errorf("want %#v, got %#v", want, got)
	}
}

func TestFormat(t *testing.T) {
	for _, tt := range formatTestingData {
		opt := &Options{}
		t.Run(tt.src, func(t *testing.T) {
			got, err := Format(tt.src, opt)
			if err != nil {
				t.Errorf("should be nil, got %v", err)
			}
			if tt.want != got {
				t.Errorf("\nwant %#v, \ngot %#v", tt.want, got)
			}
		})
	}
}

var formatTestingData = []struct {
	src  string
	want string
}{
	{
		src: `select true from m where t < date_trunc('DAY', to_timestamp('2022-01-01'))`,
		want: `
SELECT
  true
FROM m
WHERE t < DATE_TRUNC('DAY', TO_TIMESTAMP('2022-01-01'))`,
	},
}
