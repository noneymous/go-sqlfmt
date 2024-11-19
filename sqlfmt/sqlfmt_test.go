package sqlfmt

import (
	"fmt"
	"github.com/noneymous/go-sqlfmt/sqlfmt/formatters"
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
    active = true
    AND attr2 = true
    AND attr3 = true
    AND attr4 = true
    AND attr5 = true
    AND attr6 IN (
      SELECT
        *
      FROM attributes
    )
    AND attr6 = true
    AND attr7 = true
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
			name: "With ... as",
			sql:  `WITH cte_quantity AS (SELECT SUM(Quantity) as Total FROM OrderDetails GROUP BY ProductID) SELECT AVG(Total) average_product_quantity FROM cte_quantity;`,
			want: `WITH cte_quantity AS (
  SELECT
    SUM(Quantity) AS Total
  FROM OrderDetails
  GROUP BY ProductID
)
SELECT
  AVG(Total) average_product_quantity
FROM cte_quantity;`,
		},

		/*
		 * Parenthesis variations
		 */
		{
			name: "Parenthesis variations",
			sql:  `SELECT DISTINCT ON (col1, col2) a0, t1.attr1 AS a1, t1.attr2 AS a2, pg_catalog.not_a_function (db.oid, 'CREATE') AS a3, pg_catalog.HAS_DATABASE_PRIVILEGE(t1.oid, 'CREATE') AS a4, DATE_TRUNC('DAY', TO_TIMESTAMP('2022-01-01')) AS a5 FROM tble1 t1 LEFT JOIN tble2 t2 ON (t2.id = t1.id) WHERE db IN ('postgres', 'edb') AND ANY ((SELECT * FROM tble3 WHERE rel IN ('r', 's', 't', 'p') AND oid = 33176310:: oid AND col3 = '' OR ((port = 80 AND port = 443) OR (port = 80 AND port = 443))):: oid []) ORDER BY attr, attr2`,
			want: `SELECT DISTINCT ON (col1, col2) a0,
  t1.attr1 AS a1,
  t1.attr2 AS a2,
  pg_catalog.not_a_function (db.oid, 'CREATE') AS a3,
  pg_catalog.HAS_DATABASE_PRIVILEGE(t1.oid, 'CREATE') AS a4,
  DATE_TRUNC('DAY', TO_TIMESTAMP('2022-01-01')) AS a5
FROM tble1 t1
LEFT JOIN tble2 t2 ON (t2.id = t1.id)
WHERE db IN ('postgres', 'edb') AND ANY (
  (
    SELECT
      *
    FROM tble3
    WHERE
      rel IN ('r', 's', 't', 'p')
      AND oid = 33176310:: oid
      AND col3 = ''
      OR (
        (
          port = 80
          AND port = 443
        )
        OR (
          port = 80
          AND port = 443
        )
      )
  ):: oid []
)
ORDER BY attr, attr2`,
		},

		/*
		 * AND / OR clauses
		 */
		{
			name: "AND / OR clauses",
			sql:  `select * from all_hosts where ip = "127.0.0.1" OR dns_name = "localhost"`,
			want: `SELECT
  *
FROM all_hosts
WHERE ip = "127.0.0.1" OR dns_name = "localhost"`,
		},
		{
			name: "AND / OR clauses long",
			sql:  `select * from all_hosts where ip = "127.0.0.1" OR ip = "127.0.0.2" OR ip = "127.0.0.3" OR dns_name = "localhost" AND protocol = "tcp"`,
			want: `SELECT
  *
FROM all_hosts
WHERE
  ip = "127.0.0.1"
  OR ip = "127.0.0.2"
  OR ip = "127.0.0.3"
  OR dns_name = "localhost"
  AND protocol = "tcp"`,
		},
		{
			name: "AND / OR clauses grouped",
			sql:  `select *  from all_hosts where ip = "127.0.0.1" AND ( port = 80 OR port = 443) AND protocol = "tcp" limit 1`,
			want: `SELECT
  *
FROM all_hosts
WHERE
  ip = "127.0.0.1"
  AND (
    port = 80
    OR port = 443
  )
  AND protocol = "tcp"
LIMIT 1`,
		},
		{
			name: "AND / OR clauses grouped nested",
			sql:  `SELECT * FROM all_hosts WHERE ip = "127.0.0.1" AND ((port = 80 AND port = 443) OR (port = 80 AND port = 443)) AND protocol = "tcp" LIMIT 1`,
			want: `SELECT
  *
FROM all_hosts
WHERE
  ip = "127.0.0.1"
  AND (
    (
      port = 80
      AND port = 443
    )
    OR (
      port = 80
      AND port = 443
    )
  )
  AND protocol = "tcp"
LIMIT 1`,
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
  a = 1
  AND b != 2
  AND c <> 3
  AND d > 4
  AND e < 5
  AND f >= 6
  AND g <= 7`,
		},
		{
			name: "Basic comparators formatted into one line",
			sql:  `SELECT * FROM tble WHERE a=1 AND b> 2 AND c   <   3`,
			want: `SELECT
  *
FROM tble
WHERE
  a = 1
  AND b > 2
  AND c < 3`,
		},
		{
			name: "Advanced comparators",
			sql:  `SELECT * FROM tble WHERE a~~1 AND b~~*2 AND c!~~3 AND d!~~*4`,
			want: `SELECT
  *
FROM tble
WHERE
  a ~~ 1
  AND b ~~* 2
  AND c !~~ 3
  AND d !~~* 4`,
		},
		{
			name: "Invalid comparators",
			sql:  `SELECT * FROM tble WHERE a~~~1 AND b!2 AND c!== 3 AND d ===4`,
			want: `SELECT
  *
FROM tble
WHERE
  a~~~1
  AND b!2
  AND c!== 3
  AND d ===4`,
		},
		{
			name: "Invalid comparators 2",
			sql:  `SELECT * FROM tble WHERE a    !=*  1 AND b<>>2 AND c><3 AND d <==4 `,
			want: `SELECT
  *
FROM tble
WHERE
  a !=* 1
  AND b<>>2
  AND c><3
  AND d <==4`,
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
ORDER BY datname`,
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
		 * CASE WHEN ELSE samples
		 */
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
			name: "CASE WHEN ELSE with AND/OR",
			sql:  `SELECT CASE WHEN usesuper AND pg_catalog.pg_is_in_recovery() OR pg_catalog.pg_is_in_recovery() THEN pg_is_wal_replay_paused () ELSE FALSE END as isreplaypaused FROM pg_catalog.pg_user WHERE usename=current_user`,
			want: `SELECT
  CASE
    WHEN usesuper AND pg_catalog.PG_IS_IN_RECOVERY() OR pg_catalog.PG_IS_IN_RECOVERY() THEN pg_is_wal_replay_paused ()
    ELSE FALSE
  END AS isreplaypaused
FROM pg_catalog.pg_user
WHERE usename = CURRENT_USER`,
		},

		/*
		 * Unconventional queries
		 */
		{
			name: "SET query",
			sql:  `SET client_encoding TO 'UTF8'`,
			want: `SET client_encoding TO 'UTF8'`,
		},
		{
			name: "SET query with unformatted equal 1",
			sql:  `SET client_encoding= 'UNICODE'`,
			want: `SET client_encoding = 'UNICODE'`,
		},
		{
			name: "SET query with unformatted equal 2",
			sql:  `SET client_min_messages=notice`,
			want: `SET client_min_messages = notice`,
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
ORDER BY datname`,
		},

		/*
		 * The following samples don't return a perfect result yet
		 * TODO
		 */
		{
			name: "distinct from",
			sql:  `SELECT foo, bar FROM "table" WHERE foo IS NOT DISTINCT FROM bar;`,
			want: `SELECT
  foo,
  bar
FROM "table"
WHERE foo IS NOT DISTINCT
FROM bar;`,
		},
		{
			name: "within group",
			sql:  `SELECT PERCENTILE_DISC(0.5) WITHIN GROUP (ORDER BY temperature) FROM city_data;`,
			want: `SELECT
  PERCENTILE_DISC(0.5) WITHIN GROUP (
    ORDER BY temperature
  )
FROM city_data;`,
		},
		{
			name: "distinct on 1",
			sql:  `SELECT DISTINCT ON (col1, col2) col1, col2 FROM Tabellenname ORDER BY col1, col2;`,
			want: `SELECT DISTINCT ON (col1, col2) col1,
  col2
FROM Tabellenname
ORDER BY col1, col2;`,
		},
		{
			name: "distinct on 2",
			sql:  `SELECT DISTINCT ON (col1, col2) col1 FROM Tabellenname ORDER BY col1, col2;`,
			want: `SELECT DISTINCT ON (col1, col2) col1
FROM Tabellenname
ORDER BY col1, col2;`,
		},
		{
			name: "nested no function",
			sql:  `SELECT sum(customfn(xxx)) FROM "table"`,
			want: `SELECT
  SUM(customfn (xxx))
FROM "table"`,
		},
		{
			name: "nested functions",
			sql:  `SELECT sum(avg(xxx)) FROM "table"`,
			want: `SELECT
  SUM( AVG(xxx))
FROM "table"`,
		},
		{
			name: "nested functions 2",
			sql:  `SELECT test, sum(avg(xxx)) FROM "table"`,
			want: `SELECT
  test,
  SUM( AVG(xxx))
FROM "table"`,
		},
		{
			name: "multidimensional array",
			sql:  `SELECT [[xx], xx] FROM "table"`,
			want: `SELECT
  [[ xx], xx]
FROM "table"`,
		},
		{
			name: "select with line comment",
			sql: `select xxxx, --comment
        xxxx`,
			want: `SELECT
  xxxx,
  --comment xxxx`,
		},

		/*
		 * INSERT query
		 */
		{
			name: "Insert VALUES one",
			sql:  `INSERT INTO Customers (CustomerName, ContactName, Address, City, PostalCode, Country) VALUES ('Cardinal', 'Tom B. Erichsen', 'Skagen 21', 'Stavanger', '4006', 'Norway')`,
			want: `INSERT INTO Customers
  (CustomerName, ContactName, Address, City, PostalCode, Country)
VALUES
  ('Cardinal', 'Tom B. Erichsen', 'Skagen 21', 'Stavanger', '4006', 'Norway')`,
		},
		{
			name: "Insert VALUES mutliple",
			sql:  `INSERT INTO Customers(first_name, last_name, age, country) VALUES ('Harry', 'Potter', 31, 'USA'), ('Chris', 'Hemsworth', 43, 'USA'), ('Tom', 'Holland', 26, 'UK')`,
			want: `INSERT INTO Customers
  (first_name, last_name, age, country)
VALUES
  ('Harry', 'Potter', 31, 'USA'),
  ('Chris', 'Hemsworth', 43, 'USA'),
  ('Tom', 'Holland', 26, 'UK')`,
		},
		{
			name: "Insert SET short",
			sql:  `INSERT INTO actor SET first_name = 'Tom', last_name  = 'Hanks'`,
			want: `INSERT INTO actor
SET first_name = 'Tom', last_name = 'Hanks'`,
		},
		{
			name: "Insert SET long",
			sql:  `INSERT INTO actor SET first_name = 'Tom', last_name  = 'Hanks', gender  = 'M', country  = 'US'`,
			want: `INSERT INTO actor
SET
  first_name = 'Tom',
  last_name = 'Hanks',
  gender = 'M',
  country = 'US'`,
		},

		/*
		 * UPDATE query
		 */
		{
			name: "UPDATE SET short",
			sql:  `UPDATE Customers SET ContactName = 'Alfred Schmidt', City= 'Frankfurt' WHERE CustomerID = 1`,
			want: `UPDATE Customers
SET ContactName = 'Alfred Schmidt', City = 'Frankfurt'
WHERE CustomerID = 1`,
		},
		{
			name: "UPDATE SET long",
			sql:  `UPDATE Customers SET ContactName = 'Alfred Schmidt', City= 'Frankfurt', Gender= 'M', Country= 'Germany' WHERE CustomerID = 1 AND Active = true AND Show = true AND Accepted = true`,
			want: `UPDATE Customers
SET
  ContactName = 'Alfred Schmidt',
  City = 'Frankfurt',
  Gender = 'M',
  Country = 'Germany'
WHERE
  CustomerID = 1
  AND Active = true
  AND Show = true
  AND Accepted = true`,
		},
		{
			name: "UPDATE complex",
			sql:  `UPDATE books SET books.primary_author = authors.name, books.primary_author_surname = authors.surname, books.primary_author_gender = authors.gender FROM books INNER JOIN authors ON books.author_id = authors.id WHERE books.title = 'The Hobbit'`,
			want: `UPDATE books
SET
  books.primary_author = authors.name,
  books.primary_author_surname = authors.surname,
  books.primary_author_gender = authors.gender
FROM books
INNER JOIN authors ON books.author_id = authors.id
WHERE books.title = 'The Hobbit'`,
		},
		{
			name: "UPDATE complex 2",
			sql:  `Update C Set C.Name = CAST(p.Number as varchar(10)) + '|'+ C.Name FROM Catelog.Component C JOIN Catelog.ComponentPart cp ON p.ID = cp.PartID JOIN Catelog.Component c ON cp.ComponentID = c.ID where p.BrandID = 1003 AND ct.Name='Door' + '|'+ C.Name`,
			want: `UPDATE C
SET
  C.Name = CAST(p.Number AS VARCHAR(10)) + '|' + C.Name
FROM Catelog.Component C
JOIN Catelog.ComponentPart cp ON p.ID = cp.PartID
JOIN Catelog.Component c ON cp.ComponentID = c.ID
WHERE p.BrandID = 1003 AND ct.Name = 'Door' + '|' + C.Name`,
		},

		/*
		 * CREATE / ALTER / DELETE / DROP
		 */
		{
			name: "CREATE DATABASE",
			sql:  `CREATE DATABASE IF NOT EXISTS db_name`,
			want: `CREATE DATABASE IF NOT EXISTS db_name`,
		},
		{
			name: "CREATE TABLE",
			sql:  `CREATE TABLE IF NOT EXISTS table_name`,
			want: `CREATE TABLE IF NOT EXISTS table_name`,
		},
		{
			name: "CREATE TABLE with types",
			sql:  `CREATE TABLE IF NOT EXISTS table_name(column1 TEXT, column2 INTEGER, column3 BOOLEAN, column4 BOOLEAN, PRIMARY KEY(column1,column2))`,
			want: `CREATE TABLE IF NOT EXISTS table_name (
  column1 TEXT,
  column2 INTEGER,
  column3 BOOLEAN,
  column4 BOOLEAN,
  PRIMARY KEY (column1, column2)
)`,
		},
		{
			name: "CREATE TABLE with types complex",
			sql:  `CREATE TABLE Companies (id int, name varchar(50), address text, email varchar(50), phone varchar(10))`,
			want: `CREATE TABLE Companies (
  id INT,
  name VARCHAR(50),
  address TEXT,
  email VARCHAR(50),
  phone VARCHAR(10)
)`,
		},
		{
			name: "CREATE TABLE AS",
			sql:  `CREATE TABLE CustomersBackup AS (SELECT * FROM Customers)`,
			want: `CREATE TABLE CustomersBackup AS (
  SELECT
    *
  FROM Customers
)`,
		},
		{
			name: "CREATE TABLE AS without parentheses",
			sql:  `CREATE TABLE TestTable AS SELECT customername, contactname FROM customers WHERE active = true`,
			want: `CREATE TABLE TestTable AS
  SELECT
    customername,
    contactname
  FROM customers
  WHERE active = true`,
		},
		{
			name: "ALTER TABLE ADD",
			sql:  `alter table table_name add column_name boolean`,
			want: `ALTER TABLE table_name ADD column_name BOOLEAN`,
		},
		{
			name: "ALTER TABLE DROP",
			sql:  `alter table table_name drop column column_name`,
			want: `ALTER TABLE table_name DROP COLUMN column_name`,
		},
		{
			name: "ALTER TABLE RENAME",
			sql:  `alter table table_name rename column column_name to new_name`,
			want: `ALTER TABLE table_name RENAME COLUMN column_name TO new_name`,
		},
		{
			name: "ALTER TABLE type",
			sql:  `alter table table_name alter column column_name INTEGER`,
			want: `ALTER TABLE table_name ALTER COLUMN column_name INTEGER`,
		},
		{
			name: "ALTER TABLE MODIFY",
			sql:  `alter table table_name modify column column_name INTEGER`,
			want: `ALTER TABLE table_name MODIFY COLUMN column_name INTEGER`,
		},
		{
			name: "ALTER TABLE MODIFY 2",
			sql:  `alter table table_name modify column_name INTEGER`,
			want: `ALTER TABLE table_name MODIFY column_name INTEGER`,
		},
		{
			name: "DELETE",
			sql:  `DELETE FROM "table" t1 WHERE t1.V1 > t1.V2 and t1.V3 > t1.V4 and EXISTS (SELECT * FROM "table" t2 WHERE t2.V1 = t1.V2  and t2.V2 = t1.V1 AND t2.V3 = t1.V3)`,
			want: `DELETE FROM "table" t1
WHERE
  t1.V1 > t1.V2
  AND t1.V3 > t1.V4
  AND EXISTS (
    SELECT
      *
    FROM "table" t2
    WHERE
      t2.V1 = t1.V2
      AND t2.V2 = t1.V1
      AND t2.V3 = t1.V3
  )`,
		},
		{
			name: "DELETE short",
			sql:  `DELETE FROM Customers`,
			want: `DELETE FROM Customers`,
		},
		{
			name: "DROP TABLE",
			sql:  `DROP TABLE Customers`,
			want: `DROP TABLE Customers`,
		},

		/*
		 * Comment variations
		 */
		{
			name: "Comment in simple select",
			sql:  `select xxxx, /* comment */ xxxx`,
			want: `SELECT
  xxxx, /* comment */
  xxxx`,
		},
		{
			name: "Comment variations",
			sql: `SELECT
  col1, // The first column
  col2, /* The second column */
  col3 /* Column in between */,
  col4
FROM (
  SELECT DISTINCT // This is a test one-line comment
    * 
  FROM table1 // This defines the table
  WHERE // This is a where clause
    // This is a where clause
    col1 > 0
    /* 
     * It starts with some comments
     */
    AND col2 != ""
    AND col4 /* important */ = 2
  ORDER BY col1, col2 DESC // Sort by those clauses
) t2
WHERE a = 1 /* first clause */ AND b = 2 // second clause
ORDER BY col1 DESC, col2 ASC /* Final comment.
                                Multi line.
                                Multi multi. */`,
			want: `SELECT
  col1, // The first column
  col2, /* The second column */
  col3 /* Column in between */,
  col4
FROM (
  SELECT DISTINCT // This is a test one-line comment
    *
  FROM table1 // This defines the table
  WHERE // This is a where clause
    // This is a where clause
    col1 > 0 /* 
     * It starts with some comments
     */
    AND col2 != ""
    AND col4 /* important */ = 2
  ORDER BY
    col1,
    col2 DESC // Sort by those clauses
) t2
WHERE
  a = 1 /* first clause */
  AND b = 2 // second clause
ORDER BY
  col1 DESC,
  col2 ASC /* Final comment.
                                Multi line.
                                Multi multi. */`,
		},
		{
			name: "Comment outside SELECT/WHERE",
			sql: `select xxx from yyy where a = 1 order by
  col1, // important column
  col2`,
			want: `SELECT
  xxx
FROM yyy
WHERE a = 1
ORDER BY
  col1, // important column
  col2`,
		},
		{
			name: "Comment in all lines",
			sql: `select // comment line 1
  max(*), // comment line 2
  sum(*), // comment line 3
  col1, // comment line 4
  col2, // comment line 5
  ( // comment line 6
    select // comment line 7
      1 // comment line 8
  ) // comment line 9
from ( // comment line 10
  select // comment line 11
    * // comment line 12
  from tble // comment line 13
) // comment line 14
where // comment line 15
  col1 = 2 // comment line 16
group by // comment line 17
  col1, // comment line 18
  col2, // comment line 19
order by // comment line 20
  col1, // comment line 21
  col2 // comment line 22
limit 1 // comment line 23`,
			want: `SELECT // comment line 1
  MAX(*), // comment line 2
  SUM(*), // comment line 3
  col1, // comment line 4
  col2, // comment line 5
  ( // comment line 6
    SELECT // comment line 7
      1 // comment line 8
  ) // comment line 9
FROM ( // comment line 10
  SELECT // comment line 11
    * // comment line 12
  FROM tble // comment line 13
) // comment line 14
WHERE // comment line 15
  col1 = 2 // comment line 16
GROUP BY // comment line 17
  col1, // comment line 18
  col2, // comment line 19
ORDER BY // comment line 20
  col1, // comment line 21
  col2 // comment line 22
LIMIT 1 // comment line 23`,
		},
		{
			name: "Comment at beginning",
			sql:  `/*pga4dash*/ SELECT pid, datname, usename, application_name, client_addr, pg_catalog.to_char(backend_start, 'YYYY-MM-DD HH24:MI:SS TZ') AS backend_start, state, wait_event_type || ': ' || wait_event AS wait_event, array_to_string(pg_catalog.pg_blocking_pids(pid), ', ') AS blocking_pids, query, pg_catalog.to_char(state_change, 'YYYY-MM-DD HH24:MI:SS TZ') AS state_change, pg_catalog.to_char(query_start, 'YYYY-MM-DD HH24:MI:SS TZ') AS query_start, pg_catalog.to_char(xact_start, 'YYYY-MM-DD HH24:MI:SS TZ') AS xact_start, backend_type, CASE WHEN state = 'active' THEN ROUND((extract(epoch from now() - query_start) / 60)::numeric, 2) ELSE 0 END AS active_since FROM pg_catalog.pg_stat_activity WHERE datname = (SELECT datname FROM pg_catalog.pg_database WHERE oid = 23500)ORDER BY pid`,
			want: `/*pga4dash*/
SELECT
  pid,
  datname,
  usename,
  application_name,
  client_addr,
  pg_catalog.TO_CHAR(backend_start, 'YYYY-MM-DD HH24:MI:SS TZ') AS backend_start,
  state,
  wait_event_type || ': ' || wait_event AS wait_event,
  ARRAY_TO_STRING(pg_catalog.pg_blocking_pids (pid), ', ') AS blocking_pids,
  query,
  pg_catalog.TO_CHAR(state_change, 'YYYY-MM-DD HH24:MI:SS TZ') AS state_change,
  pg_catalog.TO_CHAR(query_start, 'YYYY-MM-DD HH24:MI:SS TZ') AS query_start,
  pg_catalog.TO_CHAR(xact_start, 'YYYY-MM-DD HH24:MI:SS TZ') AS xact_start,
  backend_type,
  CASE
    WHEN state = 'active' THEN ROUND(
    ( EXTRACT(epoch
      FROM NOW() - query_start) /60
    ):: NUMERIC, 2)
    ELSE 0
  END AS active_since
FROM pg_catalog.pg_stat_activity
WHERE datname = (
  SELECT
    datname
  FROM pg_catalog.pg_database
  WHERE oid = 23500
)
ORDER BY pid`,
		},

		/*
		 * END
		 */
	}

	/*
	 * Execute test cases
	 */
	for _, tt := range formatTestingData {
		options := formatters.DefaultOptions()
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

func TestCompareSemantic(t *testing.T) {
	test := struct {
		before string
		after  string
		want   bool
	}{
		before: "select * from xxx",
		after:  "select\n  *\nFROM xxx",
		want:   true,
	}
	if got := CompareSemantic(test.before, test.after); got != test.want {
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
