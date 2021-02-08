# Contributing to kVDI

If you are familiar with other `operator-sdk` projects, then this code base should be relatively easy to navigate. 
The controllers and APIs were originally generated using it, so the overall structure of those areas remains in tact.

Majority of types and K8s API definitions can be found in `pkg/apis`. These are still open to many change unless the project reaches a "stable release" phase.

You can use the [`godoc`](https://pkg.go.dev/github.com/tinyzimmer/kvdi) for navigating the code base and hopefully most of the documentation will be informative.

## Submitting a PR

I might draw up some templates later, but for now if you'd like to submit a PR, fork this repository and then open a PR into the `main` branch of this repo.

## Current TODOs

If you are wanting to contribute I am open to discussions in an issue, PRs, whatever. 
If you do intend to open a code-related PR, maybe just give a heads up in an Issue first since I might be playing around in the same part of the code base. 
I occasionally go down the rabbit hole of things I wrote months ago and end up completely reworking them.

Here are just some of the things I know could use more work.

- Tests everywhere
  - Backend has some, needs way more coverage. 
  - UI has none, primarily because I have no idea how to do it so need to learn

- More elaborate docs. The APIs, backend methods, and app configurations are well documented for the most part, and I want to stick to doc generation where I can.
  It's things like more snapshots and user docs that I need to make.

- I suck at graphic design so if someone wants to make a badass logo that would be awesome. Right now the UI just uses the quasar logo.

- I'm currently in the middle of looking into "app profiles". Meaning, the ability to have a `DesktopTemplate` which just launches a single app on the display server and fills the viewport in the UI with it. I see two potential ways to accomplish this so far that don't require any major reworks.

  - Base desktop images already install `Xpra` for `XRANDR` support in `Xvnc` (dynamic screen resizing). `Xpra` can also launch a display server on any socket and there exist websocket client implementations in JS that can be experimented with on the frontend. At time of writing this is the one I am exploring further. 
    - There is a semi-working implementation of this on `main` branch at the moment. It requires the `DesktopTemplate` set `spec.config.socketType: xpra`.
    - Current issues:
      - App sometimes launches to the far corner of the viewport or off it entirely. 
      - Need logic in UI to handle auto DPI scaling
      - Needs iframe to the official xpra html5 client, would rather have a built-in implementation
      - Some text fuziness, may be due to scaling issues.