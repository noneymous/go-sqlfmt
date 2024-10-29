package sqlfmt

import (
	"fmt"
	"github.com/noneymous/go-sqlfmt/sqlfmt/reindenters"
	"testing"
)

func TestFormat(t *testing.T) {
	var formatTestingData = []struct {
		name string
		sql  string
		want string
	}{

		/*
		 * Simple query with a wide variety of cases
		 */
		{
			name: "Simple query with high variety",
			sql:  `SELECT tble1.col1 AS a, ( SELECT unit FROM tble2 WHERE col4 > 14 AND col5 = 'test' ) AS b, tble1.col3 AS c FROM ( SELECT col1, col2, col3, col4 FROM contents WHERE active = true AND attr2 = true AND attr3 = true AND attr4 = true AND attr5 = true AND attr6 IN ( SELECT * FROM attributes ) AND attr6 = true AND attr7 = true ) AS tble1 WHERE col3 ILIKE '%substr%' AND col4 > ( SELECT MAX(salary) FROM employees ) LIMIT 1`,
			want: `SELECT
  tble1.col1 AS a,
  (
    SELECT
      unit
    FROM tble2
    WHERE col4 > 14 AND col5 = 'test'
  ) AS b,
  tble1.col3 AS c
FROM (
  SELECT
    col1,
    col2,
    col3,
    col4
  FROM contents
  WHERE
    active = true AND
    attr2 = true AND
    attr3 = true AND
    attr4 = true AND
    attr5 = true AND
    attr6 IN (
      SELECT
        *
      FROM attributes
    ) AND
    attr6 = true AND
    attr7 = true
) AS tble1
WHERE col3 ILIKE '%substr%' AND col4 > (
  SELECT
    MAX(salary)
  FROM employees
)
LIMIT 1`,
		},

		/*
		 * Some simple query samples
		 */
		{
			name: "Fixed values",
			sql:  `SELECT '1', '2'`,
			want: `SELECT
  '1',
  '2'`,
		},
		{
			name: "Select function",
			sql:  `SELECT version()`,
			want: `SELECT
  VERSION()`,
		},
		{
			name: "Nested functions",
			sql:  `select true from m where t < date_trunc('DAY', to_timestamp('2022-01-01'))`,
			want: `SELECT
  true
FROM m
WHERE t < DATE_TRUNC('DAY', TO_TIMESTAMP('2022-01-01'))`,
		},
		{
			name: "Union select",
			sql:  `SELECT * FROM xxx UNION`,
			want: `SELECT
  *
FROM xxx
UNION`, // Invalid query but still formatted well
		},
		{
			name: "Join select",
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
			name: "Sub query in first column",
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
			name: "Sub query in last column",
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
			name: "Select ANY",
			sql:  `select any ( select xxx from xxx ) from xxx where xxx limit xxx`,
			want: `SELECT ANY (
  SELECT
    xxx
  FROM xxx
)
FROM xxx
WHERE xxx
LIMIT xxx`,
		},
		{
			name: "With ... as",
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

		/*
		 * Comparators
		 */
		{
			name: "Basic comparators",
			sql:  `SELECT * FROM tble WHERE a=1 AND b!=2 AND c<>3 AND d>4 AND e<5 AND f>=6 AND g<=7`,
			want: `SELECT
  *
FROM tble
WHERE
  a = 1 AND
  b != 2 AND
  c <> 3 AND
  d > 4 AND
  e < 5 AND
  f >= 6 AND
  g <= 7`,
		},
		{
			name: "Basic comparators formatted into one line",
			sql:  `SELECT * FROM tble WHERE a=1 AND b> 2 AND c   <   3`,
			want: `SELECT
  *
FROM tble
WHERE a = 1 AND b > 2 AND c < 3`,
		},
		{
			name: "Advanced comparators",
			sql:  `SELECT * FROM tble WHERE a~~1 AND b~~*2 AND c!~~3 AND d!~~*4`,
			want: `SELECT
  *
FROM tble
WHERE
  a ~~ 1 AND
  b ~~* 2 AND
  c !~~ 3 AND
  d !~~* 4`,
		},
		{
			name: "Invalid comparators",
			sql:  `SELECT * FROM tble WHERE a~~~1 AND b!2 AND c!== 3 AND d ===4`,
			want: `SELECT
  *
FROM tble
WHERE
  a~~~1 AND
  b!2 AND
  c!== 3 AND
  d ===4`,
		},
		{
			name: "Invalid comparators 2",
			sql:  `SELECT * FROM tble WHERE a    !=*  1 AND b<>>2 AND c><3 AND d <==4 `,
			want: `SELECT
  *
FROM tble
WHERE
  a !=* 1 AND
  b<>>2 AND
  c><3 AND
  d <==4`,
		},

		/*
		 * Some complex real-world queries
		 */
		{
			name: "All sorts mixed elements",
			sql:  `SELECT db.oid as did, db.datname as name, ta.spcname as spcname, db.datallowconn, db.datistemplate AS is_template, pg_catalog.has_database_privilege(db.oid, 'CREATE') as cancreate, datdba as owner, DATE_TRUNC('DAY', TO_TIMESTAMP('2022-01-01')) FROM pg_catalog.pg_database db LEFT OUTER JOIN pg_catalog.pg_tablespace ta ON db.dattablespace = ta.oid WHERE db.oid > 16383::OID OR db.datname IN ('postgres', 'edb')  ORDER BY datname`,
			want: `SELECT
  db.oid AS did,
  db.datname AS name,
  ta.spcname AS spcname,
  db.datallowconn,
  db.datistemplate AS is_template,
  pg_catalog.HAS_DATABASE_PRIVILEGE(db.oid, 'CREATE') AS cancreate,
  datdba AS owner,
  DATE_TRUNC('DAY', TO_TIMESTAMP('2022-01-01'))
FROM pg_catalog.pg_database db
LEFT OUTER JOIN pg_catalog.pg_tablespace ta ON db.dattablespace = ta.oid
WHERE db.oid > 16383:: OID OR db.datname IN ('postgres', 'edb')
ORDER BY
  datname`,
		},
		{
			name: "All sorts mixed elements with WHERE EXISTS",
			sql:  `SELECT has_table_privilege( 'pgagent.pga_job', 'INSERT, SELECT, UPDATE' ) has_priviledge WHERE EXISTS( SELECT has_schema_privilege('pgagent', 'USAGE') WHERE EXISTS( SELECT cl.oid FROM pg_catalog.pg_class cl LEFT JOIN pg_catalog.pg_namespace ns ON ns.oid = relnamespace WHERE relname = 'pga_job' AND nspname ='pgagent' ) )`,
			want: `SELECT
  HAS_TABLE_PRIVILEGE('pgagent.pga_job', 'INSERT, SELECT, UPDATE') has_priviledge
WHERE EXISTS (
  SELECT
    HAS_SCHEMA_PRIVILEGE('pgagent', 'USAGE')
  WHERE EXISTS (
    SELECT
      cl.oid
    FROM pg_catalog.pg_class cl
    LEFT JOIN pg_catalog.pg_namespace ns ON ns.oid = relnamespace
    WHERE relname = 'pga_job' AND nspname = 'pgagent'
  )
)`,
		},
		{
			name: "All sorts mixed elements with nested WHERE EXISTS",
			sql:  `SELECT HAS_TABLE_PRIVILEGE('pgagent.pga_job', 'INSERT, SELECT, UPDATE') has_priviledge WHERE EXISTS (SELECT HAS_SCHEMA_PRIVILEGE('pgagent', 'USAGE') WHERE EXISTS (SELECT cl.oid FROM pg_catalog.pg_class cl LEFT JOIN pg_catalog.pg_namespace ns ON ns.oid = relnamespace WHERE EXISTS (SELECT cl.oid FROM pg_catalog.pg_class cl LEFT JOIN pg_catalog.pg_namespace ns ON ns.oid = relnamespace WHERE relname = 'pga_job' AND nspname = 'pgagent')))`,
			want: `SELECT
  HAS_TABLE_PRIVILEGE('pgagent.pga_job', 'INSERT, SELECT, UPDATE') has_priviledge
WHERE EXISTS (
  SELECT
    HAS_SCHEMA_PRIVILEGE('pgagent', 'USAGE')
  WHERE EXISTS (
    SELECT
      cl.oid
    FROM pg_catalog.pg_class cl
    LEFT JOIN pg_catalog.pg_namespace ns ON ns.oid = relnamespace
    WHERE EXISTS (
      SELECT
        cl.oid
      FROM pg_catalog.pg_class cl
      LEFT JOIN pg_catalog.pg_namespace ns ON ns.oid = relnamespace
      WHERE relname = 'pga_job' AND nspname = 'pgagent'
    )
  )
)`,
		},
		{
			name: "All sorts mixed elements with sub query and OID type cast",
			sql:  `SELECT at.attname, at.attnum, ty.typname FROM pg_catalog.pg_attribute at LEFT JOIN pg_catalog.pg_type ty ON (ty.oid = at.atttypid) WHERE attrelid=33176310::oid AND attnum = ANY ((SELECT con.conkey FROM pg_catalog.pg_class rel LEFT OUTER JOIN pg_catalog.pg_constraint con ON con.conrelid = rel.oid AND con.contype = 'p' WHERE rel.relkind IN ('r','s','t', 'p') AND rel.oid = 33176310::oid)::oid[])`,
			want: `SELECT
  at.attname,
  at.attnum,
  ty.typname
FROM pg_catalog.pg_attribute AT
LEFT JOIN pg_catalog.pg_type ty ON (ty.oid = at.atttypid)
WHERE attrelid = 33176310:: oid AND attnum = ANY (
  (
    SELECT
      con.conkey
    FROM pg_catalog.pg_class rel
    LEFT OUTER JOIN pg_catalog.pg_constraint con ON con.conrelid = rel.oid AND con.contype = 'p'
    WHERE rel.relkind IN ('r', 's', 't', 'p') AND rel.oid = 33176310:: oid
  ):: oid []
)`,
		},
		{
			name: "All sorts mixed elements with multiple nested sub queries",
			sql:  `SELECT at.attname, at.attnum, ty.typname FROM pg_catalog.pg_attribute AT LEFT JOIN pg_catalog.pg_type ty ON (ty.oid = at.atttypid) WHERE attrelid=33176310:: oid AND attnum = ANY ( ( SELECT con.conkey FROM pg_catalog.pg_class rel LEFT OUTER JOIN pg_catalog.pg_constraint con ON con.conrelid = rel.oid AND con.contype = 'p' WHERE attnum = ANY ( ( SELECT con.conkey FROM pg_catalog.pg_class rel LEFT OUTER JOIN pg_catalog.pg_constraint con ON con.conrelid = rel.oid AND con.contype = 'p' WHERE attnum = ANY ( ( SELECT con.conkey FROM pg_catalog.pg_class rel LEFT OUTER JOIN pg_catalog.pg_constraint con ON con.conrelid = rel.oid AND con.contype = 'p' WHERE rel.relkind IN ('r', 's', 't', 'p') AND rel.oid = 33176310:: oid ):: oid [] ) ):: oid [] ) ):: oid [])`,
			want: `SELECT
  at.attname,
  at.attnum,
  ty.typname
FROM pg_catalog.pg_attribute AT
LEFT JOIN pg_catalog.pg_type ty ON (ty.oid = at.atttypid)
WHERE attrelid = 33176310:: oid AND attnum = ANY (
  (
    SELECT
      con.conkey
    FROM pg_catalog.pg_class rel
    LEFT OUTER JOIN pg_catalog.pg_constraint con ON con.conrelid = rel.oid AND con.contype = 'p'
    WHERE attnum = ANY (
      (
        SELECT
          con.conkey
        FROM pg_catalog.pg_class rel
        LEFT OUTER JOIN pg_catalog.pg_constraint con ON con.conrelid = rel.oid AND con.contype = 'p'
        WHERE attnum = ANY (
          (
            SELECT
              con.conkey
            FROM pg_catalog.pg_class rel
            LEFT OUTER JOIN pg_catalog.pg_constraint con ON con.conrelid = rel.oid AND con.contype = 'p'
            WHERE rel.relkind IN ('r', 's', 't', 'p') AND rel.oid = 33176310:: oid
          ):: oid []
        )
      ):: oid []
    )
  ):: oid []
)`,
		},

		{
			name: "CASE WHEN ELSE with complex case",
			sql:  `SELECT roles.rolsuper AS is_superuser, CASE WHEN roles.rolsuper THEN true WHEN roles.roladmin THEN true ELSE roles.rolcreaterole END AS can_create_role, CASE WHEN 'pg_signal_backend' = ANY ( array ( with recursive cte AS ( SELECT pg_roles.oid, pg_roles.rolname FROM pg_roles WHERE pg_roles.oid = roles.oid UNION ALL SELECT m.roleid, pgr.rolname FROM cte cte_1 JOIN pg_auth_members m ON m.member = cte_1.oid JOIN pg_roles pgr ON pgr.oid = m.roleid ) SELECT rolname FROM cte ) ) THEN true ELSE false END AS can_signal_backend FROM pg_catalog.pg_roles AS roles WHERE rolname = CURRENT_USER`,
			want: `SELECT
  roles.rolsuper AS is_superuser,
  CASE
    WHEN roles.rolsuper THEN true
    WHEN roles.roladmin THEN true
    ELSE roles.rolcreaterole
  END AS can_create_role,
  CASE
    WHEN 'pg_signal_backend' = ANY (
      ARRAY (
        WITH recursive cte AS (
          SELECT
            pg_roles.oid,
            pg_roles.rolname
          FROM pg_roles
          WHERE pg_roles.oid = roles.oid
          UNION ALL
          SELECT
            m.roleid,
            pgr.rolname
          FROM cte cte_1
          JOIN pg_auth_members m ON m.member = cte_1.oid
          JOIN pg_roles pgr ON pgr.oid = m.roleid
        )
        SELECT
          rolname
        FROM cte
      )
    ) THEN true
    ELSE false
  END AS can_signal_backend
FROM pg_catalog.pg_roles AS roles
WHERE rolname = CURRENT_USER`,
		},
		{
			name: "ANY and ARRAY",
			sql: `SELECT *
			FROM tble
			WHERE 'value' = ANY (ARRAY (SELECT col FROM tble2))`,
			want: `SELECT
  *
FROM tble
WHERE 'value' = ANY (
  ARRAY (
    SELECT
      col
    FROM tble2
  )
)`,
		},

		/*
		 * Unconventional queries
		 */
		{
			name: "SET query",
			sql:  `SET client_encoding TO 'UTF8'`,
			want: `SET
  client_encoding TO 'UTF8'`,
		},
		{
			name: "SET query with unformatted equal 1",
			sql:  `SET client_encoding= 'UNICODE'`,
			want: `SET
  client_encoding = 'UNICODE'`,
		},
		{
			name: "SET query with unformatted equal 2",
			sql:  `SET client_min_messages=notice`,
			want: `SET
  client_min_messages = notice`,
		},
		{
			name: "Type cast OID",
			sql:  `SELECT con.conkey FROM pg_catalog.pg_class rel LEFT OUTER JOIN pg_catalog.pg_constraint con ON con.conrelid = rel.oid AND con.contype = 'p' WHERE rel.relkind IN ('r','s','t', 'p') AND rel.oid = 33176310::oid`,
			want: `SELECT
  con.conkey
FROM pg_catalog.pg_class rel
LEFT OUTER JOIN pg_catalog.pg_constraint con ON con.conrelid = rel.oid AND con.contype = 'p'
WHERE rel.relkind IN ('r', 's', 't', 'p') AND rel.oid = 33176310:: oid`,
		},

		/*
		 * Test whether function names are properly capitalized. Function names are usually indicated by a matching
		 * name PLUS a subsequent parenthesis. However, there are a few Postgres functions that behave more like
		 * a keyword, in that they don't have arguments and therefore don't use parenthesis.
		 * Tests need to distinguish between
		 * 		- a normal column name (doesn't need to be in double quotes) <- Don't edit case
		 * 		- an ambiguous column name colliding with a parenthesis-less function (must be quoted to resolve ambiguity)  <- Don't edit case
		 * 		- a normal function name (indicated by a subsequent parenthesis)  <- Edit case to UPPER
		 * 		- a parenthesis-less function name  <- Edit case to UPPER
		 */
		{
			name: "Postgres functions without arguments/parenthesis",
			sql:  `select current_catalog, current_user, session_user, user, current_date, current_time, current_timestamp, localtime, localtimestamp`,
			want: `SELECT
  CURRENT_CATALOG,
  CURRENT_USER,
  SESSION_USER,
  USER,
  CURRENT_DATE,
  CURRENT_TIME,
  CURRENT_TIMESTAMP,
  LOCALTIME,
  LOCALTIMESTAMP`,
		},
		{
			name: "Postgres parenthesis-less function name colldiging with column name",
			sql:  `select localtime, "localtime" from test`,
			want: `SELECT
  LOCALTIME,
  "localtime"
FROM test`,
		},
		{
			name: "Postgres parenthesis-less function colliding with normal SQL function, colliding with column name",
			sql:  `select current_timestamp, current_timestamp(), "current_timestamp" from test`,
			want: `SELECT
  CURRENT_TIMESTAMP,
  CURRENT_TIMESTAMP(),
  "current_timestamp"
FROM test`,
		},
		{
			name: "Postgres parenthesis-less function nested within normal function",
			sql:  `select * from tble where t < date_trunc('DAY', current_timestamp) AND  t < date_trunc('DAY', current_timestamp())`,
			want: `SELECT
  *
FROM tble
WHERE t < DATE_TRUNC('DAY', CURRENT_TIMESTAMP) AND t < DATE_TRUNC('DAY', CURRENT_TIMESTAMP())`,
		},
		{
			name: "Functions vs none-functions vs column names",
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

		/*
		 * The following samples don't return a perfect result yet
		 * TODO
		 */
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
			name: "distinct on 1",
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
			name: "distinct on 2",
			sql:  `SELECT DISTINCT ON (Spalte1, Spalte2) Spalte1 FROM Tabellenname ORDER BY Spalte1, Spalte2;`,
			want: `SELECT DISTINCT ON
  (Spalte1, Spalte2) Spalte1
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
			name: "nested functions 2",
			sql:  `SELECT test, sum(avg(xxx)) FROM table`,
			want: `SELECT
  test,
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

	/*
	 * Execute test cases
	 */
	for _, tt := range formatTestingData {
		options := reindenters.DefaultOptions()
		t.Run(tt.name, func(t *testing.T) {
			got, err := Format(tt.sql, options)
			if err != nil {
				t.Errorf("%v", err)
			} else {
				if tt.want != got {
					t.Errorf("\n=======================\n=== GOT ==============>\n%s\n=======================\n=== WANT =============>\n%s\n=======================", got, tt.want)
				} else {
					fmt.Println(fmt.Sprintf("%s\n%s", got, "========================================================================"))
				}
			}
		})
	}
}

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
	got := removeSymbols("select xxx from xxx")
	want := "selectxxxfromxxx"
	if got != want {
		t.Errorf("want %#v, got %#v", want, got)
	}
}
