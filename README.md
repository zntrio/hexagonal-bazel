# Bazel / Golang / Hexagonal Architecture

## Build

You need to install bazelisk

```sh
$ make bazel-build
```

## Run

Start the datastore

```sh
bazelisk run //cmd/urlshortener-datastore
```

Start the API in a different console

```sh
bazelisk run //cmd/urlshortener-api
```

## Test

Create a shortened URL.

```sh
$ curl -d '{"url":"https://blog.zenithar.org"}' http://localhost:3000/api/v1/links
{"link":{"id":"TwkS2HS7s","url":"https://blog.zenithar.org"}}
```

Resolve the shortened URL

```sh
$ curl http://localhost:3000/api/v1/links/TwkS2HS7s
{"link":{"id":"TwkS2HS7s","url":"https://blog.zenithar.org"}}
```