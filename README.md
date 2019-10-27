![Kero](https://github.com/galaco/kero/blob/master/docs/images/banner.jpg)
[![GoDoc](https://godoc.org/github.com/Galaco/kero?status.svg)](https://godoc.org/github.com/Galaco/kero)
[![Go report card](https://goreportcard.com/badge/github.com/galaco/kero)](https://goreportcard.com/report/github.com/galaco/kero)
[![codecov](https://codecov.io/gh/Galaco/kero/branch/master/graph/badge.svg)](https://codecov.io/gh/Galaco/kero)
[![CircleCI](https://circleci.com/gh/Galaco/kero.svg?style=svg)](https://circleci.com/gh/Galaco/kero)

# Kero

> Kero is a Source Engine game engine implementation written in Go.

<p align="center">
  <img width="640" height="480" src="https://cdn.galaco.me/github/kero/readme/de_dust2.gif">
</p>

## Installation
To build the project on Windows, Mac OS or Linux, all you need to do is run (assuming you have Go 1.12 or later installed)
in the directory `samples/demo`:
`go build .`

NOTE: You may need to change the `const` `GameDirectory` in `samples/demo/main.go` to point to a valid game installation
directory on your machine.


## Contributing
1. Fork it (<https://github.com/galaco/lambda-client/fork>)
2. Create your feature branch (`git checkout -b feature/fooBar`)
3. Commit your changes (`git commit -am 'Add some fooBar'`)
4. Push to the branch (`git push origin feature/fooBar`)
5. Create a new Pull Request


### Notes
**This is based on another project of mine: [https://github.com/galaco/lambda-client](https://github.com/galaco/lambda-client)
This is meant to be an attempt to create a somewhat more modular and reusable and high-quality codebase. Lambda-Client could
be considered more of an experimentation ground for feature implementations.**
