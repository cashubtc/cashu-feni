# cashu-feni
Cashu is a Chaumian Ecash wallet and mint with Bitcoin Lightning support.

*Disclaimer: The author is NOT a cryptographer and this work has not been reviewed. This means that there is very likely a fatal flaw somewhere. Cashu is still experimental and not production-ready.*

Cashu is an Ecash implementation based on David Wagner's variant of Chaumian blinding. Token logic based on [minicash](https://github.com/phyro/minicash) ([description](https://gist.github.com/phyro/935badc682057f418842c72961cf096c)) which implements a [Blind Diffie-Hellman Key Exchange](https://cypherpunks.venona.com/date/1996/03/msg01848.html) scheme written down by Ruben Somsen [here](https://gist.github.com/RubenSomsen/be7a4760dd4596d06963d67baf140406). The database mechanics and the Lightning backend uses parts from [LNbits](https://github.com/lnbits/lnbits-legend).

Please read the [Cashu](https://github.com/callebtc/cashu) documentation.

# Installation
<p align="center">
<a href="#from-source">Source</a> ·
<a href="#download">Download</a> ·
<a href="#docker">Docker</a> 
</p>

## From source 
* `git clone https://github.com/gohumble/cashu-feni`
* `cd cashu-feni`
* copy `config_example.yaml` to `config.yaml` and update configuration values 
* `go build .`
* `./cashu-feni`

## Download
Download the latest binary from [releases](https://github.com/gohumble/cashu-feni/releases)
## Using Docker
### Build image
### Download image 