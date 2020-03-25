# RPIHomeServer [![Build Status](https://travis-ci.org/Alberto-Izquierdo/RPIHomeServer-go.svg?branch=master)](https://travis-ci.org/Alberto-Izquierdo/RPIHomeServer-go)

Personal raspberry pi home server. Based on the [C++ project with the same purpose](https://github.com/Alberto-Izquierdo/RPIHomeServer) but implemented in Golang.

Like the original project, it allows managing the GPIO interface of the raspberry pi from a Telegram bot. As the original, it allows having "automatic" messages that are launched at the specified time.

It also includes a few new features:

- Communication between multiple devices: we can set a device as server (it does not have to be a raspberry) and connect the rest of devices as clients. As new clients are connected, their actions are available in the telegram interface.
- Automatic messages editor: we can now remove and create automatic actions from the telegram interface appart from the ones set in the configuration file. We can set this messages to be repeated every 24 hours or just once.
