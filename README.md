# Elephant - cuz it's phat.

`elephant` is a service providing data and actions via various data-providers. It is meant to be a backend to create f.e. custom launchers.

[![Discord](https://img.shields.io/discord/1402235361463242964?logo=discord)](https://discord.gg/mGQWBQHASt)

## Current State

The project just started and is therefore highly wip.

## Communication

Communicating with `elephant` is done via unix-sockets and protobuf messages.

## Current Providers

- `desktopapplications`
- `files`

## Quick-Guide

1. You need: `elephant`
2. ... a provider
3. something to make unix socket calls with (or use `elephant query/activate` for testing)

```
mkdir ~/.config/elephant
mkdir ~/.config/elephant/providers
git clone https://github.com/abenz1267/elephant && cd elephant/cmd && go install elephant.go
cd ../internals/providers/desktopapplications
go build -buildmode=plugin && cp desktopapplications.so ~/.config/elephant/providers/
```

Once you have this setup, you can start using `elephant`.

### Using `elephant` as client

`elephant` has a built-in tiny client which is meant for testing purpose only.

Querying: `elephant query "files;somefile;5;false"` => the arguments position correlate to their respective protobuf file.

Activating: `elephant activate "1;desktopapplications;/usr/share/applications/firefox-developer-edition.desktop:new-private-window;"`
