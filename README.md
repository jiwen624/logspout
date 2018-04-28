# logspout

![circleci](https://circleci.com/gh/jiwen624/logspout.svg?&style=shield&circle-token=03cbb9928f598c18e45b96161e4bb254ac90bfab "circleci")
[![GitHub license](https://img.shields.io/badge/license-Apache%202-blue.svg)](https://github.com/jiwen624/logspout/blob/master/LICENSE)
[![GoDoc](https://godoc.org/github.com/jiwen624/logspout?status.svg)](https://godoc.org/github.com/jiwen624/logspout)

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

### Tutorial

The tutorial is still working in progress. Leave me a message if there is someone really using it.

### Issues

Please open an issue and post the `logspout.json` and log sample you are using if you find a bug. 

### Contributing

Pull requests are welcome. For major changes, please open an issue first to discuss what you would like to change. Please make sure to update tests as appropriate.

