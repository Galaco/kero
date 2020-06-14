[![GoDoc](https://godoc.org/github.com/Galaco/kero?status.svg)](https://godoc.org/github.com/Galaco/kero)
[![Go report card](https://goreportcard.com/badge/github.com/galaco/kero)](https://goreportcard.com/report/github.com/galaco/kero)
[![codecov](https://codecov.io/gh/Galaco/kero/branch/master/graph/badge.svg)](https://codecov.io/gh/Galaco/kero)
[![CircleCI](https://circleci.com/gh/Galaco/kero.svg?style=svg)](https://circleci.com/gh/Galaco/kero)

# Kero

> Kero is a Source Engine game engine implementation written in Go.

<p align="center">
  <img width="640" height="480" src="https://raw.githubusercontent.com/Galaco/kero/dev/.github/readme/de_dust2.gif">
</p>

## Building
To build the project on Windows, Mac OS or Linux, all you need to do is run (assuming you have Go 1.12 or later 
installed) in the directory `samples/demo`:
`go build .`

The demo targets Counterstrike: Source. To target a different game, you will need to modify `ContentDirectory()` in
`samples/demo/gameDef.go`.

## Running
First, you will need to have a source engine game installed, unless you are loading a map that has all its content
bspzipped.

* Run the built executable with this flag: `-game="<GameDir>"`, where `<GameDir>` is the root directory of the game.


## Contributing
1. Fork it (<https://github.com/galaco/kero/fork>)
2. Create your feature branch (`git checkout -b feature/fooBar`)
3. Commit your changes (`git commit -am 'Add some fooBar'`)
4. Push to the branch (`git push origin feature/fooBar`)
5. Create a new Pull Request
