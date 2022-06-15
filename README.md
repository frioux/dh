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
mkdir -p dh/001
cat <<EOF > dh/001/dh.sql
CREATE TABLE users (
        "id",
        "name",
        "email"
);
EOF

cat <<EOF > dh/001/shortlinks.sql
CREATE TABLE shortlinks (
        "from",
        "to",
        "deleted"
        PRIMARY KEY ("from")
);
EOF

cat <<EOF > dh/001/history.sql
CREATE TABLE IF NOT EXISTS history (
        "from",
        "to",
        "when",
        "who"
);
EOF
```

Add that directory to your migration plan:

```
echo 000-sqlite > dh/plan.txt
echo 001       >> dh/plan.txt
```

Now let's create a migration to add another column:

```
mkdir dh/002

cat <<SQL >dh/002/shortlinks.sql
ALTER TABLE shortlinks ADD COLUMN "description" NOT NULL DEFAULT ''
SQL

cat <<SQL > dh/002/history.sql
ALTER TABLE history ADD COLUMN "description" NOT NULL DEFAULT ''
SQL
```

And add it to your plan:

```
echo 002 >> dh/plan.txt
```

And here's how you'd use it:

```golang
dbh := sql.Connect(...)
m := dh.NewMigrator()
if err := m.MigrateOne(dbh, dh.DHMigrations, "000-sqlite"); err != nil {
        panic(err)
}
if err := m.MigrateAll(dbh, os.DirFS("dh")); err != nil {
        return err
}
```

---

Out of the box `dh` can apply SQL or lists of SQL from JSON files, but the
migration interface is extensible so you could wire up
[gopher-lua](https://github.com/yuin/gopher-lua) to do more advanced
migrations.
