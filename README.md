# sqlfmt

[![Go Report Card](https://goreportcard.com/badge/github.com/noneymous/go-sqlfmt)](https://goreportcard.com/report/github.com/noneymous/go-sqlfmt)

## Description

Sqlfmt formats SQL queries or SQL statements in  `.go` files into a consistent format.

This is a fork of https://github.com/kanmu/go-sqlfmt which seems to abandoned. 
Some of its open pull requests were applied too. 
The complete code was simplified, cleaned up and restructured.

## Example Direct Call

```go

import "github.com/noneymous/go-sqlfmt/sqlfmt"

sql := `
  select xxx ,xxx ,xxx
  , case
  when xxx is null then xxx
  else true
  end as xxx
  from xxx as xxx join xxx on xxx = xxx join xxx as xxx on xxx = xxx
  left outer join xxx as xxx
  on xxx = xxx
  where xxx in ( select xxx from ( select xxx from xxx ) as xxx where xxx = xxx )
  order by xxx
`

sqlFormatted, errFormat := sqlfmt.Format(sql, &sqlfmt.Options{})
if errFormat != nil {
    return
}

```

## Example .go File

_Unformatted SQL in a `.go` file_

```go
package main

import (
	"database/sql"
)


func sendSQL() int {
	var id int
	var db *sql.DB
	db.QueryRow(`
	select xxx ,xxx ,xxx
	, case
	when xxx is null then xxx
	else true
end as xxx
from xxx as xxx join xxx on xxx = xxx join xxx as xxx on xxx = xxx
left outer join xxx as xxx
on xxx = xxx
where xxx in ( select xxx from ( select xxx from xxx ) as xxx where xxx = xxx )
and xxx in ($2, $3) order by xxx`).Scan(&id)
	return id
}
```

The above will be formatted into the following:

```go
package main

import (
	"database/sql"
)

func sendSQL() int {
	var id int
	var db *sql.DB
	db.QueryRow(`
SELECT
  xxx
  , xxx
  , xxx
  , CASE
       WHEN xxx IS NULL THEN xxx
       ELSE true
    END AS xxx
FROM xxx AS xxx
JOIN xxx
ON xxx = xxx
JOIN xxx AS xxx
ON xxx = xxx
LEFT OUTER JOIN xxx AS xxx
ON xxx = xxx
WHERE xxx IN (
  SELECT
    xxx
  FROM (
    SELECT
      xxx
    FROM xxx
  ) AS xxx
  WHERE xxx = xxx
)
AND xxx IN ($2, $3)
ORDER BY
  xxx`).Scan(&id)
	return id
}
```

## Installation

```bash
go get github.com/noneymous/go-sqlfmt/cmd/sqlfmt
```
## Usage

Provide flags and input files or directory  
  ```bash
  $ sqlfmt -w input_file.go 
  ```

## Flags
```
  -l
		Do not print reformatted sources to standard output.
		If a file's formatting is different from src, print its name
		to standard output.
  -d
		Do not print reformatted sources to standard output.
		If a file's formatting is different than src, print diffs
		to standard output.
  -w
                Do not print reformatted sources to standard output.
                If a file's formatting is different from src, overwrite it
                with gofmt style.
  -distance     
                Write the distance from the edge to the begin of SQL statements
```

## Limitations

The `sqlfmt` is currently only able to format SQL statements in **`QueryRow`**, **`Query`**, **`Exec`**  functions from the `"database/sql"` package.


The following SQL statements will be formatted

  ```go
  func sendSQL() int {
  	var id int
  	var db *sql.DB
  	db.QueryRow(`select xxx from xxx`).Scan(&id)
  	return id
  }
  ```
  
  to

  ```go
  func sendSQL() int {
  	var id int
  	var db *sql.DB
  	db.QueryRow(`
SELECT
  xxx
FROM xxx`).Scan(&id)
  	return id
  }
  ```


## Not perfectly supported

- `IS DISTINCT FROM`
- `WITHIN GROUP`
- `DISTINCT ON(xxx)`
- `select(array)`
- Comments after comma such as
  ```
  select 
         xxxx, --comment
         xxxx
  
  SELECT
         xxxx, /* comment */
         xxxx
  ```
- Nested square brackets or braces such as `[[xx], xx]`
  - Currently, being formatted into this: `[[ xx], xx]`
  - Ideally, it should be formatted into this: `[[xx], xx]`

- Nested functions such as `sum(average(xxx))`
  - Currently, being formatted into this: `SUM( AVERAGE(xxx))`
  - Ideally, it should be formatted into this: `SUM(AVERAGE(xxx))`
  
  
## Contribution

Thank you for thinking of contributing to the sqlfmt!
Pull Requests are welcome!

## License

MIT
