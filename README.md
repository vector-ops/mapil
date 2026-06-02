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

Mapil provides the following commands:

### 1. `mapil add`

Use this command to add a new key value pair to your Mapil keyring. You will be prompted to enter the name and key.
To add multiple values to the same key use comma separated values.
It trims the spaces at the start and end of the values.

```bash
mapil add
```

### 3. `mapil list`

List all the data currently stored in your Mapil keyring.

```bash
mapil list
```

### 4. `mapil upd`

Update an existing key in your Mapil keyring. You'll be presented with a list of key names, and you can choose the one you want to update.
Currently you will not be able to update individual values in a list.

```bash
mapil upd
```

### 5. `mapil del`

Delete a key from your Mapil keyring. You'll be presented with a list of key names, and you can choose the one you want to delete. You can also use the `-a` flag to delete all keys.

```bash
mapil del
mapil del -a
```

## Example

```bash
$ mapil add
? Enter a name for the key: MyKey
? Enter the value: MyValue
MyKey successfully added to Mapil keyring.

$ mapil add
? Enter a name for the key: MyListKey
? Enter the value: MyValue1, MyValue2, MyValue3, MyValue4
MyKey successfully added to Mapil keyring.

$ mapil key list
MyKey: MyValue
MyListKey: MyValue1, MyValue2, MyValue3, MyValue4

$ mapil key upd
? Choose a key to update: MyKey
? Enter the new value: MyNewValue
MyKey updated.

$ mapil key del
? Choose a key to delete: MyKey
MyKey deleted.

$ mapil key list
Data store empty.
```

## TODO

-   [ ] Terminal UI
-   [ ] Namespaces
-   [ ] Data synchronization with multiple devices
-   [ ] Optional data encryption
-   [ ] Import and store bookmarks and passwords
-   [x] New datatype: List
-   [ ] Authentication
-   [ ] Optional cloud storage
-   [ ] Data backup and restore
