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

```bash
tree
.
├── go.mod
├── main.go
├── statement.txt
├── test
│ ├── basic.js
│ ├── basic_testcases.json
│ ├── in
│ │ └── basic
│ │ ├── 00_sample_00.in
│ │ ├── 00_sample_01.in
│ │ ├── 20_gen_00.in
│ │ ├── 20_gen_01.in
│ │ ├── 20_gen_02.in
│ │ ├── 20_gen_03.in
│ │ ├── 20_gen_04.in
│ │ ├── 20_gen_05.in
│ │ ├── 20_gen_06.in
│ │ ├── 20_gen_07.in
│ │ ├── 20_gen_08.in
│ │ ├── 20_gen_09.in
│ │ ├── 20_gen_10.in
│ │ ├── 20_gen_11.in
│ │ ├── 20_gen_12.in
│ │ ├── 20_gen_13.in
│ │ ├── 20_gen_14.in
│ │ ├── 20_gen_15.in
│ │ ├── 20_gen_16.in
│ │ ├── 60_gen_00.in
│ │ ├── 60_gen_01.in
│ │ ├── 60_gen_02.in
│ │ ├── 60_gen_03.in
│ │ ├── 60_gen_04.in
│ │ ├── 60_gen_05.in
│ │ ├── 60_gen_06.in
│ │ ├── 60_gen_07.in
│ │ ├── 60_gen_08.in
│ │ ├── 60_gen_09.in
│ │ ├── 60_gen_10.in
│ │ ├── 60_gen_11.in
│ │ ├── 60_gen_12.in
│ │ ├── 60_gen_13.in
│ │ ├── 60_gen_14.in
│ │ ├── 60_gen_15.in
│ │ └── 60_gen_16.in
│ ├── out
│ │ └── basic
│ │ ├── 00_sample_00.out
│ │ ├── 00_sample_01.out
│ │ ├── 20_gen_00.out
│ │ ├── 20_gen_01.out
│ │ ├── 20_gen_02.out
│ │ ├── 20_gen_03.out
│ │ ├── 20_gen_04.out
│ │ ├── 20_gen_05.out
│ │ ├── 20_gen_06.out
│ │ ├── 20_gen_07.out
│ │ ├── 20_gen_08.out
│ │ ├── 20_gen_09.out
│ │ ├── 20_gen_10.out
│ │ ├── 20_gen_11.out
│ │ ├── 20_gen_12.out
│ │ ├── 20_gen_13.out
│ │ ├── 20_gen_14.out
│ │ ├── 20_gen_15.out
│ │ ├── 20_gen_16.out
│ │ ├── 60_gen_00.out
│ │ ├── 60_gen_01.out
│ │ ├── 60_gen_02.out
│ │ ├── 60_gen_03.out
│ │ ├── 60_gen_04.out
│ │ ├── 60_gen_05.out
│ │ ├── 60_gen_06.out
│ │ ├── 60_gen_07.out
│ │ ├── 60_gen_08.out
│ │ ├── 60_gen_09.out
│ │ ├── 60_gen_10.out
│ │ ├── 60_gen_11.out
│ │ ├── 60_gen_12.out
│ │ ├── 60_gen_13.out
│ │ ├── 60_gen_14.out
│ │ ├── 60_gen_15.out
│ │ └── 60_gen_16.out
│ └── settings.json
├── typeCheckIn.go
├── typeFees.go
├── typeSeats.go
└── typeShower.go

5 directories, 82 files
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

1. Build

  ```bash
  go build -o main .
  ```

1. Run `00_sample_00.in`

  ```bash
  cat test/in/basic/00_sample_00.in | ./main
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
