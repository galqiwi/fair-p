# Limiter

[![GoDev](https://img.shields.io/static/v1?label=godev&message=reference&color=00add8)][godev]

[godev]: https://pkg.go.dev/github.com/sunshineplan/limiter

The `limiter` package provides reader/writer capabilities with transfer speed limiting in Go.

## Overview
The `Limiter` is designed to limit the total transfer speed of readers and writers simultaneously.

## Features
- Transfer Speed Limiting: Control the speed at which operations can be performed.
- Dynamic Configuration: Set and update the transfer speed limit and burst size as needed.
- Contextual Support: Perform operations with transfer speed limits within a specific context.

## Usage
Import the package in your Go code:

```go
import "github.com/sunshineplan/limiter"
```

Create a new `Limiter` by calling the `New` function:

```go
limiter := limiter.New(100*1024)
```

The above code creates a `Limiter` with a total transfer speed limit of 100KB per second and a burst size of 100KB.

### Transfer Speed Limiting

To perform limited operations, you can use the provided `Reader` and `Writer` methods. Here's an example of using the `Reader` method:

```go
reader := limiter.Reader(yourReader)
```

The `Reader` method takes an `io.Reader` and returns a new reader that performs limited reads.

Similarly, you can use the `Writer` method to create a limited writer:

```go
writer := limiter.Writer(yourWriter)
```

### Updating Configuration

You can update the limit and burst size of the `Limiter` by calling the respective `Set` methods:

```go
limiter.SetLimit(200*1024) // 200KB
limiter.SetBurst(20*1024) // 20KB
```

These methods allow you to dynamically adjust the limiting behavior according to your requirements.

## Contributing
Contributions are welcome! If you find a bug or have a suggestion for improvement, please open an issue or submit a pull request.

## License
This package is released under the [MIT License](https://github.com/sunshineplan/limiter/blob/main/LICENSE).
