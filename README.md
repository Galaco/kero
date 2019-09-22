[![GoDoc](https://godoc.org/github.com/Galaco/kero?status.svg)](https://godoc.org/github.com/Galaco/kero)
[![Go report card](https://goreportcard.com/badge/github.com/galaco/kero)](https://goreportcard.com/badge/github.com/galaco/kero)
[![GolangCI](https://golangci.com/badges/github.com/galaco/kero.svg)](https://golangci.com)
[![codecov](https://codecov.io/gh/Galaco/kero/branch/master/graph/badge.svg)](https://codecov.io/gh/Galaco/kero)
[![CircleCI](https://circleci.com/gh/Galaco/kero.svg?style=svg)](https://circleci.com/gh/Galaco/kero)

# Kero
Lambda Client is a game engine written in golang designed that loads Valve's Source Engine projects. Put simply, pointing this projects configuration at
a Source Engine game installation directory will allow for loading that targets .bsp maps and contents.


## Current state
You can build this right now, and, assuming you set the configuration to point to an existing Source game installation (this is tested primarily against CS:S):
* Loads game data files from projects gameinfo.txt
* Load BSP map
* Load high-resolution texture data for bsp faces, including pakfile entries
* Full visibility data support
* Staticprop loading (working, but is incomplete)
* Basic entdata loading (dynamic and physics props)

##### Counterstrike: Source de_dust2.bsp
![de_dust2](https://raw.githubusercontent.com/Galaco/kero/master/docs/de_dust2.jpg)


## What will this do?
The end goal is to be able to point this application at a source engine game, with a bsp as the target, and be able to
load and play that map. Where this progresses beyond that, needs to be decided. Most likely this will be come either a thin client for multiple
source games with game specific code layered on top (target multiplayer as priority), or the full server simulation for single player games
would be written (targeting single player as priority).


## Getting started
There is a small amount of configuration required to get this project running, beyond `dep ensure`.
* For best results, you need a source engine game installed already.
* Copy `config.example.json` to `config.json`, and update the `gameDirectory` property to point to whatever game installation
you are targeting (e.g. HL2 would be `<steam_dir>/steamapps/common/hl2`).

## Contributing
There is loads to do! Right now there are a few core issues that need fixing, and loads of fundamental features to add. Here
are just a few!
* StudioModel library needs finishing before props can be properly added. There are some issues around multiple stripgroups per mesh, multiple
materials per prop, mdl data not fully loaded, and likely more
* Implement physics (probably bullet physics? Accurate VPhysics is probably not worthwhile, but needs investigation)
* A vulkan renderer would be a huge step forward, particularly this early on. Abstracting a mesh away from ogl would also help
* Displacement support incomplete - generation is buggy, and visibility checks cull displacements always (visible when outside of world only)
* Additional game support/testing in BSP library


#### Additional examples
##### Counterstrike: Source ze_FFVII_Mako_Reactor_v5_3.bsp
![de_dust2](https://raw.githubusercontent.com/Galaco/kero/master/docs/ze_FFVII_Mako_Reactor_v5_3.jpg)
