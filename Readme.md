# Event Sourcing

[![standard-readme compliant](https://img.shields.io/badge/standard--readme-OK-green.svg?style=flat-square)](https://github.com/RichardLitt/standard-readme)

> Event Sourcing for Golang

## Table of Contents

- [Security](#security)
- [Background](#background)
- [Install](#install)
- [Usage](#usage)
- [Maintainers](#maintainers)
- [Contributing](#contributing)
- [License](#license)

## Security

## Background

## Install
This is currently under development install it via go modules `go mod edit -require github.com/florianlenz/event-sourcing-go@$COMMIT_HASH`

## Usage

In order to use this library you need to create an new instance of `EventSourcing`.
Once you have the instance you are able to commit events. If you commit an event it will get persisted and passed to the processor.
The Processor will take care to apply the event to your projectors as well as passing it to the reactors. Don't forget to register your events in the event registry. 

Replaying events is done via the replay method. Make sure that the processor is NOT running while you replay events. 


## Maintainers

## Contributing

## License
