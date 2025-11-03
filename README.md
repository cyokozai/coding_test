# Coding test with golang

## Environment

|tool|version|
|--|--|
|go|1.25.3|

## Development

```bash
go mod init github.com/cyokozai/coding_test.git

go run main.go
```

## Test

```bash
go build -o main .

cat test.in | ./main
```

- Result

  ```bash
  checkin: userid = 1, seatid = 1
  checkin: userid = 2, seatid = 3
  ```

## Git Rules

### Branch

main and dev branches are protected.
Please create a feature branch from dev branch.

```bash
git checkout -b feature/***
```

### Commit

Please follow the conventional commits specification.

- feat: A new feature
- fix: A bug fix
- docs: Documentation only changes
- style: Changes that do not affect the meaning of the code (white-space, formatting, missing semi-colons, etc)
- refactor: A code change that neither fixes a bug nor adds a feature
- perf: A code change that improves performance
- test: Adding missing tests or correcting existing tests
- chore: Changes to the build process or auxiliary tools and libraries such as documentation generation

### Pull Request

When you create a pull request, please push your feature branch and set the upstream.

```bash
git push --set-upstream origin feature/***
```

The pull request flow is as follows:

```planetext
feature/*** -> dev -> main
```
