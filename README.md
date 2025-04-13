# Environment Setup

This project is based on the [Hugo](https://gohugo.io/) framework.
To work with it, we need to do following steps

1. Install these tools: [Golang](https://go.dev/), [Hugo](https://gohugo.io/), [D2](https://d2lang.com/).
Configure environment variables correctly in other that they can be called with commands: `go`, `hugo` and `d2` respectively.

2. Pull the [hextra](themes/hextra/) theme module:

- Initialize for the first time: `git submodule update --init --recursive`
- Update the module if necessary: `git submodule update --recursive --remote`

3. Diagrams in this project are built with [D2](https://d2lang.com/) not natively supported.
Therefore, we need to run a local renderer server in a standalone process: `go run tools/d2_renderer.go`

4. Now, we can build the project (flags can be found [here](https://gohugo.io/commands/hugo_server/)): `hugo server`
