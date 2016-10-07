# meiru

Minimal batteries-included mail server for small deployments

## What?

Mail servers are a pain to setup. They do too much because they're meant to handle many case scenarios.

**meiru** is a minimal SMTP/IMAP server meant to be downloaded and deployed in minutes. It features:

 - Single binary server (for both IMAP and SMTP)
 - Human readable configuration format ([sample](https://github.com/Hamcha/meiru/blob/master/conf/meiru.conf.sample))

To keep meiru from becoming the next mail server behemoth (and to keep me sane), meiru will have the following limitations:

 - IMAP, SMTP and **nothing else**. No POP3, no SMAP, no antispam, no firewall, no antivirus..
 - Configuration files only, no DB backend for user credentials or things like that
 - No support for relays
 - Protocol extensions (ESMTP, IMAP capabilities) will be implemented only when there is a strong argument for them<sup>1</sup>
 - Single node only (for now)
 - Simplicity over performance

<sup>
1. Unless they take little to no extra work (ie. `SMTPUTF8`, `PIPELINING`)
</sup>

Note that these limitations are not set in stone and may vary with time.

## Getting started

**meiru currently does not work**

### Requirements

- Go 1.5+

### Installation

`go get github.com/hamcha/meiru/...`

Refer to the sample configuration in `conf/` on how to configure a simple system

## License

All of meiru source code is licensed under MIT

The full text for the license can be found in the [LICENSE file](https://raw.githubusercontent.com/Hamcha/meiru/master/LICENSE) included with the source code

## Anime tax

meiru sounds like.. Meiru (Mayl) from Megaman Battle Network!

![Meiru](http://www.therockmanexezone.com/gallery/albums/userpics/10002/Meiru_OSSsitez.PNG)
