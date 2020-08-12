package main

import (
	"flag"
	"fmt"
	"github.com/gaarutyunov/gomigo"
	log "github.com/sirupsen/logrus"
	"os"
)

var commands = map[string]string{
	"init": "initializes migrations",
	"clean": "cleans migrations",
	"add": "adds new migration (requires -name option)",
	"remove": "removes a migration (requires -name option)",
	"up": "upgrades to specific version (requires -version option)",
	"down": "downgrades to specific version (requires -version option)",
}

func main() {
	var args gomigo.Args

	gomigo.Parse(&args)

	if args.Command == "" {
		usage()
		os.Exit(1)
	}

	doMain(&args)
}

func usage() {
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println()
	fmt.Println("  gomigo [options] [command]")
	fmt.Println()
	fmt.Println("Commands:")
	fmt.Println()
	for name := range commands {
		printCommand(name)
	}
	fmt.Println()
	fmt.Println("Options:")
	fmt.Println()
	flag.PrintDefaults()
	fmt.Println()
}

func printCommand(name string) {
	fmt.Println("  ", name)
	fmt.Println("\t", commands[name])
}

func doMain(args *gomigo.Args) {
	conn, err := gomigo.Connect(&gomigo.MigratorConfig{
		Module:  args.Module,
		ConnStr: args.ConnStr,
	})

	if err != nil {
		log.Fatalf("error connecting to db: %v", err)
	}

	defer conn.Close()

	switch args.Command {
	case "init":
		doInit(conn)
	case "clean":
		doClean(conn)
	case "up":
		if args.Version == -1 {
			printCommand(args.Command)
			os.Exit(1)
		}
		doUp(conn, args.Version)
	case "down":
		if args.Version == -1 {
			printCommand(args.Command)
			os.Exit(1)
		}
		doDown(conn, args.Version)
	case "add":
		if args.Name == "" {
			printCommand(args.Command)
			os.Exit(1)
		}
		doAdd(conn, args.Name)
	case "remove":
		if args.Name == "" {
			printCommand(args.Command)
			os.Exit(1)
		}
		doRemove(conn, args.Name)
	}
}

func doInit(conn *gomigo.Migrator) {
	if err := conn.Init(); err != nil {
		log.Fatalf("error initializing migrations: %v", err)
	}
}

func doClean(conn *gomigo.Migrator) {
	if err := conn.Clean(); err != nil {
		log.Fatalf("error cleaning up migrations: %v", err)
	}
}

func doUp(conn *gomigo.Migrator, version int) {
	v, err := conn.UpV(version)

	if err != nil {
		log.Fatalf("error upgrading migrations: %v", err)
	}

	log.Infof("new version: %d", v)
}

func doDown(conn *gomigo.Migrator, version int) {
	v, err := conn.DownV(version)

	if err != nil {
		log.Fatalf("error downgrading migrations: %v", err)
	}

	log.Infof("new version: %d", v)
}

func doAdd(conn *gomigo.Migrator, name string) {
	if err := conn.Add(name, "migrations"); err != nil {
		log.Fatalf("error adding migration %s: %v", name, err)
	}

	log.Infof("new migration: %s", name)
}

func doRemove(conn *gomigo.Migrator, name string) {
	if err := conn.Remove(name, "migrations"); err != nil {
		log.Fatalf("error adding migration %s: %v", name, err)
	}

	log.Infof("removed migration: %s", name)
}
