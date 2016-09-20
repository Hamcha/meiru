# Meiru

(WIP) Minimal mail server for small deployments

## What?

Mail servers are a pain to setup. They do too much because they're meant to handle many case scenarios.

I like my stuff to be simple and limited in scope, because of this, meiru will have the following limitations:

 - No user authentication data or configuration in the database. The database is only used for metadata. Only one database backend
 - IMAP, SMTP and nothing else. No POP3, no SMAP, no antispam, no firewall, no antivirus
 - Protocol extensions (ESMTP, IMAP capabilities) will be implemented only when there is a strong argument for them
 - Simplicity over performance

## Getting started

**Meiru currently does not work**

### Requirements

- Go 1.5+

### Installation

`go get github.com/hamcha/meiru`

Refer to the sample configuration in `conf/` on how to configure a simple system

## Anime tax

Meiru sounds like.. Meiru (Mayl) from Megaman Battle Network!

![Meiru](http://www.therockmanexezone.com/gallery/albums/userpics/10002/Meiru_OSSsitez.PNG)