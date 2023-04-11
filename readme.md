# Shortcut-go-client

A command line client for [shortcut-pages](https://github.com/mt-empty/shortcut-pages), written in Golang.

![](https://github.com/mt-empty/shortcut-c-client/blob/master/shortcut.gif)


## Installing

Install from source:
```bash
sudo curl -sSL https://github.com/mt-empty/shortcut-go-client/releases/latest/download/shortcut -o /usr/local/bin/shortcut \
  && sudo chmod +x /usr/local/bin/shortcut \
  && sudo /usr/local/bin/shortcut update 
```


## Usage

```
Usage:
  shortcut [flags]
  shortcut [command]

Available Commands:
  completion  Generate the autocompletion script for the specified shell
  help        Help about any command
  list        List all available shortcut pages in the cache
  update      Update the local cache

Flags:
  -h, --help        help for shortcut
  -n, --no-colour   Remove colours from the output (default true)
  -v, --version     version for shortcut
```


## Contributing

Contributions are most welcome!

Bugs: open an issue here.

New features: open an issue here or feel free to send a pull request with the included feature.
