[![GoDoc](https://godoc.org/github.com/Galaco/kero?status.svg)](https://godoc.org/github.com/Galaco/kero)
[![Go report card](https://goreportcard.com/badge/github.com/galaco/kero)](https://goreportcard.com/report/github.com/galaco/kero)
[![codecov](https://codecov.io/gh/Galaco/kero/branch/master/graph/badge.svg)](https://codecov.io/gh/Galaco/kero)
[![CircleCI](https://circleci.com/gh/Galaco/kero.svg?style=svg)](https://circleci.com/gh/Galaco/kero)

# Kero

> Kero is a Source Engine game client implementation written in Go.

## Current Features

* BSP rendering with visdata support
* Skybox rendering
* Lightmap support (incomplete, BSP geometry only)
* Staticprop rendering
* Prop entity rendering (incomplete, models with bones unsupported)
* Bullet physics for brush:physics entity collisions

###### Build Kero, run it by pointing it to a Source Engine game installation, and it *should* just work!


<p align="center">
  <img width="100%" height="auto" src="https://raw.githubusercontent.com/Galaco/kero/dev/.github/readme/de_dust2.gif">
</p>

## Building

### Prerequisites
This project is tested against Go 1.14+, although will probably build on Go 1.12 or later. CGo is required for Imgui and Bullet.
To compile with the physics module, `Bullet` is required; 
* On Mac OS it can be installed with `brew install bullet`. 
* See Bullet documentation for other platforms

### Build
To build the project on Windows, Mac OS or Linux, all you need to do is run (assuming you have Go 1.12 or later 
installed) in the directory `samples/demo`:
`go build .`

The demo targets Counterstrike: Source entities. To target a different game, you will need to update `samples/demo/gameDef.go`.

## Running
First, you will need to have a source engine game installed, unless you are loading a map that has all its content
bspzipped.

* Run the built executable with this flag: `-game="<GameDir>/<ContentDir>"`, where `<GameDir>` is the root directory 
of the game, and `<ContentDir>` is the sub-folder where the game content is located (e.g. `cstrike`, `hl2`, `csgo` etc).
For example, a default Counterstrike: Source installation would be specified like this: 
`-game="C:\Program Files (x86)\Steam\Steamapps\common\Counterstrike Source\cstrike"`

## What's the end goal?

* An accurate-as-possible renderer compared to the original source engine
* A physics environment for client-side simulation (Bullet will perform as "close enough")
* Sound/audio playback
* Expose an interface for controlling/passing game state in/out (e.g. demo files, netcode, etc)
* Expose an interface for game specific implementations to be built on top
* Headless mode. Be able to run this without a renderer or audio output
* Interface for querying game data at runtime (e.g. LoS calculations between entities)
* As little reliance on CGo as possible
* 0 reliance on any valve code or libraries

## Contributing
1. Fork it (<https://github.com/galaco/kero/fork>)
2. Create your feature branch (`git checkout -b feature/fooBar`)
3. Commit your changes (`git commit -am 'Add some fooBar'`)
4. Push to the branch (`git push origin feature/fooBar`)
5. Create a new Pull Request
