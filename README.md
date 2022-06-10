# dh - A simple Database Deployment Handler in Go

`dh` is a reaction to
[DBIx::Class::DeploymentHandler](https://metacpan.org/pod/DBIx::Class::DeploymentHandler),
a database migration system I built in in 2010.  While the original DBICDH was built on
top of an existing ORM and schema transformation tooling, this version is intentionally
simpler, and allows users to drop in ORM support if needed.

## Getting Started

The primary users of this program are expected to be Go programmers, using `dh`
as a library.  First, create your deployment directory:

```
mkdir -p dh/deploy/001
cat <<EOF > dh/deploy/001/dh.sql
CREATE TABLE dh_migrations (
        "id",
        "version",
        "sql"
);
EOF

cat <<EOF > dh/deploy/001/shortlinks.sql
CREATE TABLE shortlinks (
        "from",
        "to",
        "deleted"
        PRIMARY KEY ("from")
);
EOF

cat <<EOF > dh/deploy/001/history.sql
CREATE TABLE IF NOT EXISTS history (
        "from",
        "to",
        "when",
        "who"
);
EOF
```

Now let's create a migration to add another column:

```
mkdir -p dh/upgrade/002

cat <<SQL >dh/upgrade/002/shortlinks.sql
ALTER TABLE shortlinks ADD COLUMN "description" NOT NULL DEFAULT ''
SQL

cat <<SQL > dh/upgrade/002/history.sql
ALTER TABLE history ADD COLUMN "description" NOT NULL DEFAULT ''
SQL
```

And here's how you'd use it:

```golang
dbh := sql.Connect(...)
m := dh.Basic(dh.BasicOptions{
        Migrations: os.OpenDir("./dh/"),
        DB: dbh,
})
if err := m.Migrate(); err != nil {
        return err
}
```

The Basic configuration knows how to read an fs.FS, deploys the maximum
available `deploy`, and then deploys each version `upgrade` after that.  Each
stage will be wrapped in a transaction.  You can store statements in a SQL file,
or a JSON file to clearly separate statements, or you can create a Migrator:

```golang
type Migrator interface {
        Name() string
        Migrate(sqlx.DB) (done bool, err error)
}
```

```golang
m := dh.Basic(dh.BasicOptions{
        Migrations: os.OpenDir("./dh/"),
        MigrationEngine: dh.BasicEngine(map[string]dh.Migrator{
                "hash-passwords.migr": &migr{},
        }),
        DB: dbh,
})
```

You can then call the migrator by either creating an empty file named `hash-passwords.migr` in
your upgrade directory, or putting a string like this in a sql file:

```SQL
MIGRATE hash-passwords.migr;
```
