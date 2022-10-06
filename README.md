# cashu-feni
[![made-with-Go](https://img.shields.io/badge/Made%20with-Go-1f425f.svg)](https://go.dev/)
[![GoReportCard example](https://goreportcard.com/badge/github.com/gohumble/cashu-feni)](https://goreportcard.com/report/github.com/gohumble/cashu-feni)
[![Docker](https://badgen.net/badge/icon/docker?icon=docker&label)](https://https://docker.com/)
[![Github tag](https://badgen.net/github/tag/gohumble/cashu-feni)](https://github.com/gohumble/cashu-feni/tags/)

<html>
<simple-boost amount="2100" address="hello@getalby.com"></simple-boost>
</html>

Cashu is a Chaumian Ecash wallet and mint with Bitcoin Lightning support.

*Disclaimer: The author is NOT a cryptographer and this work has not been reviewed. This means that there is very likely
a fatal flaw somewhere. Cashu is still experimental and not production-ready.*

Cashu is an Ecash implementation based on David Wagner's variant of Chaumian blinding. Token logic based
on [minicash](https://github.com/phyro/minicash) ([description](https://gist.github.com/phyro/935badc682057f418842c72961cf096c))
which implements a [Blind Diffie-Hellman Key Exchange](https://cypherpunks.venona.com/date/1996/03/msg01848.html) scheme
written down by Ruben Somsen [here](https://gist.github.com/RubenSomsen/be7a4760dd4596d06963d67baf140406). The database
mechanics and the Lightning backend uses parts from [LNbits](https://github.com/lnbits/lnbits-legend).

Please read the [Cashu](https://github.com/callebtc/cashu) documentation.

# Install

<p align="center">
<a href="#from-source">Source</a> ·
<a href="#download">Download</a> ·
<a href="#docker">Docker</a> 
</p>

## From source

These steps will help you installing cashu-feni from source.

### Requirements

* [golang](https://go.dev/dl/)

### Building

```bash 
git clone https://github.com/gohumble/cashu-feni && cd cashu-feni
```

copy `config_example.yaml` to `config.yaml` and update configuration values

```bash
go build . && ./cashu-feni
```

## Download

Download the latest binary from [releases](https://github.com/gohumble/cashu-feni/releases)

## Using Docker

Start cashu-feni using docker. Please provide a local volume path to the data folder.

```bash
docker run -it -p 3338:3338 \
-v /home/user/cashu-feni/data/:/app/data/ \
gohumble/cashu-feni
```
Mounting custom `config.yaml` to `/app/config.yaml`
```bash
docker run -it -p 3338:3338 \
-v /home/user/config.yaml:/app/config.yaml \
-v /home/user/cashu-feni/data/:/app/data/ \
gohumble/cashu-feni
```

### Build image

```bash 
git clone https://github.com/gohumble/cashu-feni && cd cashu-feni
docker build -t cashu -f Dockerfile .
```
