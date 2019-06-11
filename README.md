# Mobell

This is a very basic application that wraps Pushover.net's API 
to make it easy to create ad-hoc scripts that send notifications 
about status.

This application requires a [pushover.net](https://pushover.net) account and API key to work.

### Why the name?

It is a portmanteau of mobile and bell :bell: (for Terminal bell).

### How to install

The exact installations instructions vary for different environments. This is the gist:

```bash
$ git clone https://github.com/Pear0/mobell.git
$ go get
$ go build -o mobell
```

Then use `mobell init` to create a config file.

### Commands

`$ mobel` will send a simple message with just the current time.

`$ mobel init` will interactively create a new user-level config file.

`$ mobel push "Custom message"` will send a notification with a custom message.
