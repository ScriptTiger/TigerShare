[![Say Thanks!](https://img.shields.io/badge/Say%20Thanks-!-1EAEDB.svg)](https://docs.google.com/forms/d/e/1FAIpQLSfBEe5B_zo69OBk19l3hzvBmz3cOV6ol1ufjh0ER1q3-xd2Rg/viewform)

**DISCLAIMER!!!: THIS APP IS STILL IN ITS EARLY DEVELOPMENT AND HAS NOT BEEN AUDITED FOR SECURITY, SO USE AT YOUR OWN RISK!**

# TigerShare (https://scripttiger.github.io/tigershare/)
TigerShare is a simple peer-to-peer file-sharing app following a "share to anyone from anywhere" principle and runs directly from within your browser, doesn't need to be installed, requires no accounts, no logins, nor has any inherent limitations, other than those imposed by your Internet service provider, any TURN service you may be using, and/or hardware limitations. And with that being said, in cases where networks may be preventing peer-to-peer connections from being established, TigerShare is easily configurable for TURN. However, even though every connection is inherently encrypted by WebRTC, TigerShare is not intended to be a secure, anonymous file-sharing app. It's just intended as a way to quickly and easily share files amongst friends.

If you don't already have access to a TURN server, you can sign up for ExpressTURN (https://www.expressturn.com) absloutely free, go to your dashboard, and you'll be presented with your TURN information which you can use to configure TigerShare for TURN. There are obviously a lot of different options out there for this, but I've found this to be the simplest for folks who are not too tech-savvy, as you can literally sign up and get your information for TURN access all in just a couple of minutes or less.

TigerShare is written in Go and transpiled to JavaScript via GopherJS. And to smooth out some of the clunkiness of interacting with the JS API in Go, the `jsGo` package has also been used in TigerShare in order to simplify DOM manipulation slightly, but also to simplify interacting with the native JS API in general to avoid using the Go standard library wherever possible and keep TigerShare as lightweight as possible.

For additional notes on `jsGo`, please refer to its documentation:  
https://github.com/ScriptTiger/jsGo

# More About ScriptTiger

For more ScriptTiger scripts and goodies, check out ScriptTiger's GitHub Pages website:  
https://scripttiger.github.io/
