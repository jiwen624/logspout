# LogSpout

![circleci](https://circleci.com/gh/jiwen624/logspout.svg?&style=shield&circle-token=03cbb9928f598c18e45b96161e4bb254ac90bfab "circleci")
[![GitHub license](https://img.shields.io/badge/license-Apache%202-blue.svg)](https://github.com/jiwen624/logspout/blob/master/LICENSE)
[![GoDoc](https://godoc.org/github.com/jiwen624/logspout?status.svg)](https://godoc.org/github.com/jiwen624/logspout)


**Please checkout the ![latest release](https://github.com/jiwen624/logspout/releases) rather than the master branch. The master branch is under active development, which
may not be working properly all the time.'

LogSpout is a easy-to-use tool to generate machine logs in specified format. It needs only a sample log and the
configuration file in json format. With the value replacement rules defined in the configuration file, logspout
 can produce log events in various speed (EPS: Events Per Second).
 
The features of logspout include:

1. Configurable `think time` in milliseconds, or keeps producing log events with `hightide` mode set to `true`.
2. The logs can be sent to `stdout` or a regular file, with rotation support, or syslog.
3. Multi-line log support.
4. Straight-forward replacement rules defined with regular patterns and json KV pairs.
5. Various data format support: IPv4/IPv6 addresses, Email addresses, Names, Countries, User Agents, Timestamp, Fixed-lists, etc.
6. Transaction mode (A transaction may contain two or more events, each has its own think time).
7. Multi-threaded log events support.
8. Easy hands-on experience with docker support

And many others.

### TODO List

There are a couple of new features on the roadmap now.


### How To Use

A simple way to use it is through Docker, in just a few steps:

1. Clone the repository to your local server/laptop:

```git clone git@github.com:jiwen624/logspout.git```

2. Build the docker image and run the servicei (You need to have docker and docker-compose installed):

```docker-compose up```

Note that you can also use `docker-compose up -d` to make it run in background, and use `docker-compose logs -f` to check out the debug logs (the debug logs of logspout itself, not the simulated machine logs)

Stop it with Ctrl-C if you run it foreground, or `docker-compose down` for background services.

You may change the configuration file `logspout-docker.json` or use your own one by modifying the docker-compose.yml file.

Enjoy it.

### Tutorial

The tutorial is still working in progress. Leave me a message if there is someone really using it.

### Issues

Please open an issue and post the `logspout.json` and log sample you are using if you find a bug. 

### Contributing

Pull requests are welcome. For major changes, please open an issue first to discuss what you would like to change. Please make sure to update tests as appropriate.

