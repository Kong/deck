# Installation

decK is entirely written in Go. The build process builds a single static binary,
which makes it easy and convenient to install decK.

You can follow along installation instructions based on your
Operating System(OS):

### macOS

If you are on macOS, install decK using brew:

```shell
$ brew tap hbagdi/deck
$ brew install deck
```

### Linux

If you are Linux, you can either use the Debian or RPM archive from
the Github [release page](https://github.com/hbagdi/deck/releases)
or install by downloading a compressed archive, which contains the binary:

```shel
$ curl -sL https://github.com/hbagdi/deck/releases/download/v0.7.1/deck_0.7.1_linux_amd64.tar.gz -o deck.tar.gz
$ tar -xf deck.tar.gz -C /tmp
$ sudo cp /tmp/deck /usr/local/bin/
```

### Docker image

If your workflow requires a Docker image, then you can use `hbagdi/deck` Docker
image from the official Docker hub:

```
docker pull hbagdi/deck
```

You will have to mount the state files into the container as volumes so that
decK can read the files during diff/sync procedures.

If you're integrating decK into your CI system, you can either install decK
into the system itself, use the Docker based environment or pull the binaries
hosted on Github in each job.

