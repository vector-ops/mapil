# Mapil - Store and access simple data from the command line

[![Go Report Card](https://goreportcard.com/badge/github.com/vector-ops/mapil)](https://goreportcard.com/report/github.com/vector-ops/mapil)
[![GoDoc](https://godoc.org/github.com/vector-ops/mapil?status.svg)](https://pkg.go.dev/github.com/vector-ops/mapil)
[![GitHub release](https://img.shields.io/github/v/release/vector-ops/mapil)](<(https://img.shields.io/github/v/release/vector-ops/mapil)>)

Mapil is a command-line tool built with Golang, designed to simplify the management of your data. With Mapil, you can store, retrieve, update, and delete small pieces of data as key-value or key-list pairs with ease, all from the command line. You can use Mapil to store bookmarks, passwords, API keys, URLs or any other key-value data you need to manage.

Mapil is still in development, and more features (like data encryption, authentication, data sync etc.) will be added in the future. If you have any suggestions or feedback, please feel free to open an issue.

## Installation

To get started, make sure you have Go installed on your system. Then, install Mapil globally using go install:

**So far it is only tested on linux machine. It may not work as intended on other platforms.**

If you install from the releases page make sure you place the binary in a directory that is in your $PATH so you can run the commands easily.

If you install using the following command, make sure that your $PATH includes the $GOPATH/bin directory so your commands can be easily run.

```bash
go install github.com/vector-ops/mapil@latest
```

## Usage

Use the help command
```bash
mapil -h
```
