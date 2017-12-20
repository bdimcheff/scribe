# scribe

This is a little tiny program that takes logs in the format that Olark uses and sticks them on the network in remote syslog format.  It's pretty specialized, so it's unlikely it will be usable out-of-the-box, but it's here as an example in case someone needs it.  There's a slightly modified version of go's syslog library to improve compatibility with RFC5424 that may be useful.

This is also no longer maintained, as we are gradually switching to use standard logging formats and existing log aggregation software.

Sticks stdout into syslog in a logcentral-compatible way

![scribe](http://i.imgur.com/UxnBsTy.jpg)

Owners
------
 1. brandon

Usage
-----
1. Check out the general [go docs](https://github.com/olark/supportdocumentation/blob/master/drafts/dev_golang.md) to set up your env first.
2. Put scribe in the proper place in your $GOPATH, probably `src/github.com/olark/scribe`
3. run `make`
4. You should have `scribe` in `$GOPATH/bin` now, which should be in your `$PATH` if you followed the directions above.
5. run `scribe --help`
