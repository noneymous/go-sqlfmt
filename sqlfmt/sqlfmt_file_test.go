package sqlfmt

import (
	"io"
	"os"
	"testing"
)

func TestFormatFile(t *testing.T) {
	tests := []struct {
		name     string
		filename string
		want     string
		wantErr  bool
	}{
		{
			name:     "sample1",
			filename: "testdata/testing_gofile.go",
			want: `package testdata

import (
	"database/sql"
)

func sendSQL() int {
	var id int
	var db *sql.DB
	db.QueryRow(` + "`" + `SELECT
  ANY (
    SELECT
      xxx
    FROM xxx
  )
FROM xxx
WHERE xxx
LIMIT xxx` + "`" + `).Scan(&id)
	return id
}
`,
			wantErr: false,
		},
		{
			name:     "sample2",
			filename: "testdata/testing_gofile_url_query.go",
			want: `package testdata

import (
	"net/url"
)

func parseQuery() int {
	u := url.Parse("https://example.org/?a=1&a=2&b=&=3&&&&")
	u.Query()
}
`,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		opt := &Options{}
		t.Run(tt.name, func(t *testing.T) {
			f, errOpen := os.Open(tt.filename)
			if errOpen != nil {
				t.Errorf("FormatFile() error = %v", errOpen)
				return
			}
			src, errRead := io.ReadAll(f)
			if errRead != nil {
				t.Errorf("FormatFile() error = %v", errRead)
				return
			}
			got, err := FormatFile(tt.filename, src, opt)
			if (err != nil) != tt.wantErr {
				t.Errorf("FormatFile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !(string(got) == tt.want) {
				t.Errorf("FormatFile() error = \nwant %#v, \ngot  %#v", tt.want, string(got))
			}
		})
	}
}
