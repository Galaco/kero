[![GoDoc](https://godoc.org/github.com/Galaco/kero?status.svg)](https://godoc.org/github.com/Galaco/kero)
[![Go report card](https://goreportcard.com/badge/github.com/galaco/kero)](https://goreportcard.com/report/github.com/galaco/kero)
[![codecov](https://codecov.io/gh/Galaco/kero/branch/master/graph/badge.svg)](https://codecov.io/gh/Galaco/kero)
[![CircleCI](https://circleci.com/gh/Galaco/kero.svg?style=svg)](https://circleci.com/gh/Galaco/kero)

# Kero
> Kero is a Source Engine game engine implementation written in Go.



### Installation
To build the project, all you need to do is run:
`go build main.go`

To run it, copy `config.example.json` to `config.json`, then change the `gameDirectory` value to a Source Engine game installation
on your machine. After that, just run the binary you compiled.

**This is based on another project of mine: [https://github.com/galaco/lambda-client](https://github.com/galaco/lambda-client)
This is meant to be an attempt to create a somewhat more modular and reusable and high-quality codebase. Lambda-Client could
be considered more of an experimentation ground for feature implementations.**