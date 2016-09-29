# meiru

Minimal batteries-included mail server for small deployments

## What?

Mail servers are a pain to setup. They do too much because they're meant to handle many case scenarios.

**meiru** is a minimal SMTP/IMAP server meant to be downloaded and deployed in minutes. It features:

 - Single binary server (for both IMAP and SMTP)
 - Human readable configuration format ([sample](https://github.com/Hamcha/meiru/blob/master/conf/meiru.conf.sample))

To keep meiru from becoming the next mail server behemoth (and to keep me sane), meiru will have the following limitations:

 - IMAP, SMTP and **nothing else**. No POP3, no SMAP, no antispam, no firewall, no antivirus..
 - Configuration files for configuration only, no DB backend for user credentials or other config parameters
 - No support for relays
 - Only one database backend ([BoltDB](https://github.com/boltdb/bolt))
 - Protocol extensions (ESMTP, IMAP capabilities) will be implemented only when there is a strong argument for them<sup>1</sup>
 - Single node only (for now)
 - Simplicity over performance
 
<sup>
1. Unless they take little to no extra work (ie. `SMTPUTF8`, `PIPELINING`)
</sup>

Note that these limitations are not set in stone and may vary with time.

## Getting started

**Meiru currently does not work**

### Requirements

- Go 1.5+

### Installation

`go get github.com/hamcha/meiru/...`

Refer to the sample configuration in `conf/` on how to configure a simple system

## Anime tax

Meiru sounds like.. Meiru (Mayl) from Megaman Battle Network!

![Meiru](http://www.therockmanexezone.com/gallery/albums/userpics/10002/Meiru_OSSsitez.PNG)
