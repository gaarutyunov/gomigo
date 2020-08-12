# gomigo

Migration engine for PostgreSQL written in Go.

### Features

- Migrations are written with sql
- Migrate up and down to specific version
- Migrations are built into one executable
- All sql files are embedded resources

### How it works

When migration is added gomigo does two things:
- generates a folder with following structure:
```
├── migrations
    └── 20200812012658_initial
        ├── up.sql
        ├── down.sql
        └── migrate.go

```
- registers migration in database

up.sql is executed to update database

down.sql is executed to downgrade database

### Usage

1. Install dependencies

2. Install gomigo

3. Create Go Module

4. Execute in command line:

```bash
Usage:

  gomigo [options] [command]

Commands:

   init
         initializes migrations
   clean
         cleans migrations
   add
         adds new migration (requires -name option)
   remove
         removes a migration (requires -name option)
   up
         upgrades to specific version (requires -version option)
   down
         downgrades to specific version (requires -version option)
   update
         updates migrations


Options:

  -db string
        database connection string
  -module string
        module name
  -name string
        migration name
  -version int
        version to upgrade/downgrade

```