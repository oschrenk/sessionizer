# Development

**Requirements**

- [go](https://go.dev/) `brew install go`
- [air](https://github.com/cosmtrek/air) `go install github.com/air-verse/air@latest`
- [staticcheck]() `go install honnef.co/go/tools/cmd/staticcheck@latest`

## Tasks

- `task build` Build project
- `task run` Run example
- `task test` Run tests
- `task tidy` Ensure all imports are satisfied
- `task lint` Lint
- `task install` Install app in `$GOBIN/`
- `task uninstall` Removed app from `$GOBIN/`
- `task artifacts` Produces artifact in `./`
- `task tag` Pushes git tag from `VERSION`
- `task release` Creates GitHub release from artifacts
- `task sha` Prints hashes from artifacts
- `task clean` Removes build directory `.build`
- `task updates` Find dependency updates

## Release

1. Increase version number in `VERSION`
2. `task release` to tag and push
3. `task sha` to print hashes to `stdout`
4. Make changes in [homebrew-made](https://github.com/oschrenk/homebrew-made) and push
5. `brew update` to update taps
6. `brew upgrade` to upgrade formula
