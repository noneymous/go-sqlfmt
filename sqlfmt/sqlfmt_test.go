package sqlfmt

import (
	"fmt"
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
		name string
		sql  string
		want string
	}{
		{
			name: "func in func",
			sql:  `select true from m where t < date_trunc('DAY', to_timestamp('2022-01-01'))`,
			want: `SELECT
  true
FROM m
WHERE t < DATE_TRUNC('DAY', TO_TIMESTAMP('2022-01-01'))`,
		},
		{
			name: "simple query mixed elements",
			sql:  `SELECT db.oid as did, db.datname as name, ta.spcname as spcname, db.datallowconn, db.datistemplate AS is_template, pg_catalog.has_database_privilege(db.oid, 'CREATE') as cancreate, datdba as owner FROM pg_catalog.pg_database db LEFT OUTER JOIN pg_catalog.pg_tablespace ta ON db.dattablespace = ta.oid WHERE db.oid > 16383::OID OR db.datname IN ('postgres', 'edb')  ORDER BY datname`,
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
			name: "functions vs none-functions vs column names",
			sql:  `SELECT pg_catalog.not_a_function (db.oid, 'CREATE') AS cancreate1, pg_catalog.NOT_A_FUNCTION (db.oid, 'CREATE') AS cancreate2, PG_CATALOG.NOT_A_FUNCTION (db.oid, 'CREATE') AS cancreate3, pg_catalog.some_column_name AS cancreate4, pg_catalog.SOME_COLUMN_NAME AS cancreate5, PG_CATALOG.SOME_COLUMN_NAME AS cancreate6, pg_catalog.has_database_privilege(db.oid, 'CREATE') AS cancreate7, pg_catalog.HAS_DATABASE_PRIVILEGE(db.oid, 'CREATE') AS cancreate8, PG_CATALOG.HAS_DATABASE_PRIVILEGE(db.oid, 'CREATE') AS cancreate9 FROM pg_catalog.pg_database db WHERE db.oid > 16383::OID ORDER BY datname`,
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
			name: "simple short",
			sql:  `SELECT '1', '2'`,
			want: `SELECT
  '1',
  '2'`,
		},
		{
			name: "simple short with union incomplete",
			sql:  `SELECT * FROM xxx UNION`,
			want: `SELECT
  *
FROM xxx
UNION`, // Invalid query but still formatted well
		},
		{
			name: "where exists",
			sql:  `SELECT has_table_privilege( 'pgagent.pga_job', 'INSERT, SELECT, UPDATE' ) has_priviledge WHERE EXISTS( SELECT has_schema_privilege('pgagent', 'USAGE') WHERE EXISTS( SELECT cl.oid FROM pg_catalog.pg_class cl LEFT JOIN pg_catalog.pg_namespace ns ON ns.oid=relnamespace WHERE relname='pga_job' AND nspname='pgagent' ) )`,
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
			name: "simple join",
			sql:  `SELECT a.xxx, a.yyy, b.zzz FROM a LEFT JOIN b ON a.id = b.id WHERE b.column > 2`,
			want: `SELECT
  a.xxx,
  a.yyy,
  b.zzz
FROM a
LEFT JOIN b ON a.id = b.id
WHERE b.column > 2`,
		},
		{
			name: "sub query in select 1",
			sql:  `select (select col1 from tble2 where tble2.col3 = tble1.col3 limit 1), col2 from tble1 limit 1`,
			want: `SELECT
  (
    SELECT
      col1
    FROM tble2
    WHERE tble2.col3 = tble1.col3
    LIMIT 1
  ),
  col2
FROM tble1
LIMIT 1`,
		},
		{
			name: "sub query in select 2",
			sql:  `select col0, (select col1 from tble2 where tble2.col3 = tble1.col3 limit 1) from tble1 limit 1`,
			want: `SELECT
  col0,
  (
    SELECT
      col1
    FROM tble2
    WHERE tble2.col3 = tble1.col3
    LIMIT 1
  )
FROM tble1
LIMIT 1`,
		},
		{
			name: "select function",
			sql:  `SELECT version()`,
			want: `SELECT
  version ()`,
		},
		{
			name: "set query",
			sql:  `SET client_encoding TO 'UTF8'`,
			want: `SET
  client_encoding TO 'UTF8'`,
		},
		{
			name: "select any",
			sql:  `select any ( select xxx from xxx ) from xxx where xxx limit xxx`,
			want: `SELECT
  ANY (
    SELECT
      xxx
    FROM xxx
  )
FROM xxx
WHERE xxx
LIMIT xxx`,
		},
		{
			name: "with as",
			sql:  `WITH cte_quantity AS (SELECT SUM(Quantity) as Total FROM OrderDetails GROUP BY ProductID) SELECT AVG(Total) average_product_quantity FROM cte_quantity;`,
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
		{
			name: "oid type cast",
			sql:  `SELECT con.conkey FROM pg_catalog.pg_class rel LEFT OUTER JOIN pg_catalog.pg_constraint con ON con.conrelid=rel.oid AND con.contype='p' WHERE rel.relkind IN ('r','s','t', 'p') AND rel.oid = 33176310::oid`,
			want: `SELECT
  con.conkey
FROM pg_catalog.pg_class rel
LEFT OUTER JOIN pg_catalog.pg_constraint con ON con.conrelid=rel.oid AND con.contype= 'p'
WHERE rel.relkind IN ('r', 's', 't', 'p') AND rel.oid = 33176310:: oid`,
		},
		{
			name: "complex query with inner select and oid type cast",
			sql:  `SELECT at.attname, at.attnum, ty.typname FROM pg_catalog.pg_attribute at LEFT JOIN pg_catalog.pg_type ty ON (ty.oid = at.atttypid) WHERE attrelid=33176310::oid AND attnum = ANY ((SELECT con.conkey FROM pg_catalog.pg_class rel LEFT OUTER JOIN pg_catalog.pg_constraint con ON con.conrelid=rel.oid AND con.contype='p' WHERE rel.relkind IN ('r','s','t', 'p') AND rel.oid = 33176310::oid)::oid[])`,
			want: `SELECT
  at.attname,
  at.attnum,
  ty.typname
FROM pg_catalog.pg_attribute AT
LEFT JOIN pg_catalog.pg_type ty ON (ty.oid = at.atttypid)
WHERE attrelid=33176310:: oid AND attnum = ANY (
  (
    SELECT
      con.conkey
    FROM pg_catalog.pg_class rel
    LEFT OUTER JOIN pg_catalog.pg_constraint con ON con.conrelid=rel.oid AND con.contype= 'p'
    WHERE rel.relkind IN ('r', 's', 't', 'p') AND rel.oid = 33176310:: oid
  ):: oid []
)`,
		},

		//
		// TODO The following samples don't return a perfect result yet
		//
		{
			name: "distinct from",
			sql:  `SELECT foo, bar FROM table WHERE foo IS NOT DISTINCT FROM bar;`,
			want: `SELECT
  foo,
  bar
FROM table
WHERE foo IS NOT DISTINCT
FROM bar;`,
		},
		{
			name: "within group",
			sql:  `SELECT PERCENTILE_DISC(0.5) WITHIN GROUP (ORDER BY temperature) FROM city_data;`,
			want: `SELECT
  PERCENTILE_DISC(0.5) WITHIN
GROUP (ORDER BY temperature)
FROM city_data;`,
		},
		{
			name: "distinct on",
			sql:  `SELECT DISTINCT ON (Spalte1, Spalte2) Spalte1, Spalte2 FROM Tabellenname ORDER BY Spalte1, Spalte2;`,
			want: `SELECT DISTINCT ON
  (Spalte1, Spalte2) Spalte1,
  Spalte2
FROM Tabellenname
ORDER BY
  Spalte1,
  Spalte2;`,
		},
		{
			name: "nested no function",
			sql:  `SELECT sum(customfn(xxx)) FROM table`,
			want: `SELECT
  SUM(customfn (xxx))
FROM table`,
		},
		{
			name: "nested functions",
			sql:  `SELECT sum(avg(xxx)) FROM table`,
			want: `SELECT
  SUM( AVG(xxx))
FROM table`,
		},
		{
			name: "multidimensional array",
			sql:  `SELECT [[xx], xx] FROM table`,
			want: `SELECT
  [[ xx], xx]
FROM table`,
		},
		{
			name: "select with line comment",
			sql: `select xxxx, --comment
        xxxx`,
			want: `SELECT
  xxxx,
  --comment xxxx`,
		},
		{
			name: "select with multi line comment",
			sql:  `select xxxx, /* comment */ xxxx`,
			want: `SELECT
  xxxx,
  /* comment */ xxxx`,
		},
	}

	for _, tt := range formatTestingData {
		opt := &Options{}
		t.Run(tt.name, func(t *testing.T) {
			got, err := Format(tt.sql, opt)
			if err != nil {
				t.Errorf("should be nil, got %v", err)
			} else {
				if tt.want != got {
					t.Errorf("\n=======================\n=== WANT =============>\n%s\n=======================\n=== GOT ==============>\n%s\n=======================", tt.want, got)
				} else {
					fmt.Println(fmt.Sprintf("%s\n%s", got, "========================================================================"))
				}
			}
		})
	}
}
