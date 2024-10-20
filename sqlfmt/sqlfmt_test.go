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
	got := removeSymbol("select xxx from xxx")
	want := "selectxxxfromxxx"
	if got != want {
		t.Errorf("want %#v, got %#v", want, got)
	}
}

func TestFormat(t *testing.T) {
	var formatTestingData = []struct {
		src  string
		want string
	}{
		{
			src: `select true from m where t < date_trunc('DAY', to_timestamp('2022-01-01'))`,
			want: `SELECT
  true
FROM m
WHERE t < DATE_TRUNC('DAY', TO_TIMESTAMP('2022-01-01'))`,
		},
		{
			src: `SELECT db.oid as did, db.datname as name, ta.spcname as spcname, db.datallowconn, db.datistemplate AS is_template, pg_catalog.has_database_privilege(db.oid, 'CREATE') as cancreate, datdba as owner FROM pg_catalog.pg_database db LEFT OUTER JOIN pg_catalog.pg_tablespace ta ON db.dattablespace = ta.oid WHERE db.oid > 16383::OID OR db.datname IN ('postgres', 'edb')  ORDER BY datname`,
			want: `SELECT
  db.oid AS did,
  db.datname AS name,
  ta.spcname AS spcname,
  db.datallowconn,
  db.datistemplate AS is_template,
  pg_catalog.HAS_DATABASE_PRIVILEGE(db.oid, 'CREATE') AS cancreate,
  datdba AS owner
FROM pg_catalog.pg_database db
LEFT OUTER JOIN pg_catalog.pg_tablespace ta ON db.dattablespace = ta.oid
WHERE db.oid > 16383:: OID OR db.datname IN ('postgres', 'edb')
ORDER BY
  datname`,
		},
		{
			src: `SELECT pg_catalog.not_a_function (db.oid, 'CREATE') AS cancreate1, pg_catalog.NOT_A_FUNCTION (db.oid, 'CREATE') AS cancreate2, PG_CATALOG.NOT_A_FUNCTION (db.oid, 'CREATE') AS cancreate3, pg_catalog.some_column_name AS cancreate4, pg_catalog.SOME_COLUMN_NAME AS cancreate5, PG_CATALOG.SOME_COLUMN_NAME AS cancreate6, pg_catalog.has_database_privilege(db.oid, 'CREATE') AS cancreate7, pg_catalog.HAS_DATABASE_PRIVILEGE(db.oid, 'CREATE') AS cancreate8, PG_CATALOG.HAS_DATABASE_PRIVILEGE(db.oid, 'CREATE') AS cancreate9 FROM pg_catalog.pg_database db WHERE db.oid > 16383::OID ORDER BY datname`,
			want: `SELECT
  pg_catalog.not_a_function (db.oid, 'CREATE') AS cancreate1,
  pg_catalog.NOT_A_FUNCTION (db.oid, 'CREATE') AS cancreate2,
  PG_CATALOG.NOT_A_FUNCTION (db.oid, 'CREATE') AS cancreate3,
  pg_catalog.some_column_name AS cancreate4,
  pg_catalog.SOME_COLUMN_NAME AS cancreate5,
  PG_CATALOG.SOME_COLUMN_NAME AS cancreate6,
  pg_catalog.HAS_DATABASE_PRIVILEGE(db.oid, 'CREATE') AS cancreate7,
  pg_catalog.HAS_DATABASE_PRIVILEGE(db.oid, 'CREATE') AS cancreate8,
  PG_CATALOG.HAS_DATABASE_PRIVILEGE(db.oid, 'CREATE') AS cancreate9
FROM pg_catalog.pg_database db
WHERE db.oid > 16383:: OID
ORDER BY
  datname`,
		},
		{
			src: `SELECT '1', '2'`,
			want: `SELECT
  '1',
  '2'`,
		},
		{
			src: `SELECT * FROM xxx UNION`,
			want: `SELECT
  *
FROM xxx
UNION`, // Invalid query but still formatted well
		},
		{
			src: `SELECT has_table_privilege( 'pgagent.pga_job', 'INSERT, SELECT, UPDATE' ) has_priviledge WHERE EXISTS( SELECT has_schema_privilege('pgagent', 'USAGE') WHERE EXISTS( SELECT cl.oid FROM pg_catalog.pg_class cl LEFT JOIN pg_catalog.pg_namespace ns ON ns.oid=relnamespace WHERE relname='pga_job' AND nspname='pgagent' ) )`,
			want: `SELECT
  HAS_TABLE_PRIVILEGE('pgagent.pga_job', 'INSERT, SELECT, UPDATE') has_priviledge
WHERE EXISTS (
  SELECT
    HAS_SCHEMA_PRIVILEGE('pgagent', 'USAGE')
  WHERE EXISTS (
    SELECT
      cl.oid
    FROM pg_catalog.pg_class cl
    LEFT JOIN pg_catalog.pg_namespace ns ON ns.oid=relnamespace
    WHERE relname= 'pga_job' AND nspname= 'pgagent'
  )
)`,
		},
		{
			src: `SELECT a.xxx, a.yyy, b.zzz FROM a LEFT JOIN b ON a.id = b.id WHERE b.column > 2`,
			want: `SELECT
  a.xxx,
  a.yyy,
  b.zzz
FROM a
LEFT JOIN b ON a.id = b.id
WHERE b.column > 2`,
		},
		{
			src: `SELECT version()`,
			want: `SELECT
  version ()`,
		},
		{
			src: `SET client_encoding TO 'UTF8'`,
			want: `SET
  client_encoding TO 'UTF8'`,
		},
		{
			src: `select any ( select xxx from xxx ) from xxx where xxx limit xxx`,
			want: `SELECT
  any (
    SELECT
      xxx
    FROM xxx
  )
FROM xxx
WHERE xxx
LIMIT xxx`,
		},
		{
			src: `WITH cte_quantity AS (SELECT SUM(Quantity) as Total FROM OrderDetails GROUP BY ProductID) SELECT AVG(Total) average_product_quantity FROM cte_quantity;`,
			want: `WITH cte_quantity AS (
  SELECT
    SUM(Quantity) AS Total
  FROM OrderDetails
  GROUP BY
    ProductID
)
SELECT
  AVG(Total) average_product_quantity
FROM cte_quantity;`,
		},
	}

	for _, tt := range formatTestingData {
		opt := &Options{}
		t.Run(tt.src, func(t *testing.T) {
			got, err := Format(tt.src, opt)
			if err != nil {
				t.Errorf("should be nil, got %v", err)
			} else {
				if tt.want != got {
					t.Errorf("\nwant %#v, \ngot  %#v", tt.want, got)
				}
			}
		})
	}
}

// TODO The following samples don't return a perfect result yet
func TestFormat_Unsupported(t *testing.T) {
	var formatTestingData = []struct {
		src  string
		want string
	}{
		{
			src: `SELECT foo, bar FROM table WHERE foo IS NOT DISTINCT FROM bar;`,
			want: `SELECT
  foo,
  bar
FROM table
WHERE foo IS NOT DISTINCT
FROM bar;`,
		},
		{
			src: `SELECT PERCENTILE_DISC(0.5) WITHIN GROUP (ORDER BY temperature) FROM city_data;`,
			want: `SELECT
  PERCENTILE_DISC(0.5) WITHIN
GROUP (ORDER BY temperature)
FROM city_data;`,
		},
		{
			src: `SELECT DISTINCT ON (Spalte1, Spalte2) Spalte1, Spalte2 FROM Tabellenname ORDER BY Spalte1, Spalte2;`,
			want: `SELECT DISTINCT ON
  (Spalte1, Spalte2) Spalte1,
  Spalte2
FROM Tabellenname
ORDER BY
  Spalte1,
  Spalte2;`,
		},
		{
			src: `SELECT sum(customfn(xxx)) FROM table`,
			want: `SELECT
  SUM(customfn (xxx))
FROM table`,
		},
		{
			src: `SELECT sum(avg(xxx)) FROM table`,
			want: `SELECT
  SUM( AVG(xxx))
FROM table`,
		},
		{
			src: `SELECT [[xx], xx] FROM table`,
			want: `SELECT
  [[ xx], xx]
FROM table`,
		},
		{
			src: `select xxxx, --comment
        xxxx`,
			want: `SELECT
  xxxx,
  --comment xxxx`,
		},
		{
			src: `select xxxx, /* comment */ xxxx`,
			want: `SELECT
  xxxx,
  /* comment */ xxxx`,
		},
	}

	for _, tt := range formatTestingData {
		opt := &Options{}
		t.Run(tt.src, func(t *testing.T) {
			got, err := Format(tt.src, opt)
			if err != nil {
				t.Errorf("should be nil, got %v", err)
			} else {
				if tt.want != got {
					t.Errorf("\nwant %#v, \ngot  %#v", tt.want, got)
				}
			}
		})
	}
}
