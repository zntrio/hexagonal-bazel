# Bazel / Golang / Hexagonal Architecture

A little bit over-engineered urlshortener service to demonstrate Hexagonal 
Architecture usage with Bazel.

## Build

You need to install bazelisk

```sh
$ brew install bazelisk
or
$ nix-shell -p bazelisk
```

Compile the world

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

Start the GraphQL API in a different console

```sh
bazelisk run //cmd/urlshortener-graphql
```
## Tests

### Public links

Create a shortened URL.

```sh
$ curl -d '{"url":"https://blog.zenithar.org"}' http://localhost:3000/api/v1/links
{"link":{"id":"TwkS2HS7s","url":"https://blog.zenithar.org"}}
```

```sh
$ curl http://localhost:3000/api/v1/links/TwkS2HS7s
{"link":{"id":"TwkS2HS7s","url":"https://blog.zenithar.org"}}
```

### Secret links

Create a *secret* shortened URL.

```sh
$ curl -d '{"url":"https://blog.zenithar.org","secret":"my-very-secret-password"}' http://localhost:3000/api/v1/links
{"link":{"id":"pjuesJ9YX","url":"https://blog.zenithar.org","secret_required":true}}
```

Resolve the *secret* shortened URL (POST)

```sh
$ curl -d '{"secret":"my-very-secret-password"}' http://localhost:3000/api/v1/links/pjuesJ9YX
{"link":{"id":"pjuesJ9YX","url":"https://blog.zenithar.org"}}
```

If you try to access the *secret* shortened url with an invalid secret (POST)

```sh
$ curl -d '{"secret":"my-secret"}' http://localhost:3000/api/v1/links/pjuesJ9YX
{"error_message":"This shortened url requires a valid secret to be revealed.","error_code":403}
```

If you try to access the *secret* shortened url without the secret (GET)

```sh
$ curl http://localhost:3000/api/v1/links/pjuesJ9YX
{"error_message":"This shortened url requires a secret to be revealed.","error_code":406}
```

> This should return a 404 to prevent the information leak about the fact that 
> the given shortened ID exists. But be aware that it could also be identified 
> using time based analysis. 