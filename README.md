# no: simple nodejs containers

`no` can containerize nodejs projects without needing a `Dockerfile` or docker installed on your system.

[![asciicast](https://asciinema.org/a/gMggXsWgL3Hg3ypvRWzyomH8e.svg)](https://asciinema.org/a/gMggXsWgL3Hg3ypvRWzyomH8e)

## Roadmap

- `no run <project_root>` to containerize and run a node script via locally installed runtime such as `docker` or `podman`. `done`
- `no build <project_root>` to containerize and save to local daemon. `done`
- `no publish <project_root>` to containerize and publish the application image to a remote registry. `tbd`

### Heavily inspired from [ko](https://github.com/google/ko)

