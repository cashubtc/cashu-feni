# cashu-feni

[![made-with-Go](https://img.shields.io/badge/Made%20with-Go-1f425f.svg)](https://go.dev/)
[![GoReportCard example](https://goreportcard.com/badge/github.com/cashubtc/cashu-feni)](https://goreportcard.com/report/github.com/cashubtc/cashu-feni)
[![Docker](https://badgen.net/badge/icon/docker?icon=docker&label)](https://https://docker.com/)
[![Github tag](https://badgen.net/github/tag/cashubtc/cashu-feni)](https://github.com/cashubtc/cashu-feni/tags/)
[![codecov](https://codecov.io/gh/cashubtc/cashu-feni/branch/master/graph/badge.svg)](https://codecov.io/gh/cashubtc/cashu-feni)

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

Please read the [Cashu](https://github.com/callebtc/cashu) documentation for more detailed information.

This project aims to replicate the python mint implementation of cashu.

# Install

<p align="center">
<a href="#from-source">Source</a> ·
<a href="#download">Download</a> ·
<a href="#docker">Docker</a> 
</p>

## From source

These steps will help you installing cashu-feni from source. This project has two parts, a Cashu command line wallet and a mint.

### Requirements

* [golang](https://go.dev/dl/)

### Building

```bash 
git clone https://github.com/cashubtc/cashu-feni && cd cashu-feni
```

Copy `config_example.yaml` to `config.yaml` and update configuration values.

#### Wallet

```bash
go build -o feni cmd/cashu/feni.go && ./feni
```

#### Mint

```bash
go build -v -o cashu-feni cmd/mint/mint.go && ./cashu-feni
```

## Download

Download the latest binary from [releases](https://github.com/cashubtc/cashu-feni/releases)

## Using Docker

Start cashu-feni using docker. Please provide a local volume path to the data folder.

```bash 
docker pull cashubtc/cashu-feni
```

```bash
docker run -it -p 3338:3338 \
-v $(pwd)/data/:/app/data/ \
cashubtc/cashu-feni
```

Mounting custom `config.yaml` to `/app/config.yaml`

```bash
docker run -it -p 3338:3338 \
-v $(pwd)/config.yaml:/app/config.yaml \
-v $(pwd)/data/:/app/data/ \
cashubtc/cashu-feni
```

### Build image

```bash 
git clone https://github.com/cashubtc/cashu-feni && cd cashu-feni
docker build -t cashu -f Dockerfile .
```
