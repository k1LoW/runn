## [v0.70.0](https://github.com/k1LoW/runn/compare/v0.69.1...v0.70.0) - 2023-05-10
### Breaking Changes 🛠
- Support for JSON type columns by @xande0812 in https://github.com/k1LoW/runn/pull/499

## [v0.69.1](https://github.com/k1LoW/runn/compare/v0.69.0...v0.69.1) - 2023-05-07
### Fix bug 🐛
- Fix an issue that the JSON value in the request body becomes a string  when it contains newlines. by @k1LoW in https://github.com/k1LoW/runn/pull/497

## [v0.69.0](https://github.com/k1LoW/runn/compare/v0.68.1...v0.69.0) - 2023-04-25
### Breaking Changes 🛠
- If the loop is run a specified number of times and there is an error, the runbook run will result in an error ( for simple loop ) by @k1LoW in https://github.com/k1LoW/runn/pull/494
### Fix bug 🐛
- Renew CDP runners on every root loop by @k1LoW in https://github.com/k1LoW/runn/pull/493

## [v0.68.1](https://github.com/k1LoW/runn/compare/v0.68.0...v0.68.1) - 2023-04-25
### Other Changes
- Fix output of `runn run` by @k1LoW in https://github.com/k1LoW/runn/pull/491
- Increase cdpTimeoutByStep by @k1LoW in https://github.com/k1LoW/runn/pull/492

## [v0.68.0](https://github.com/k1LoW/runn/compare/v0.67.0...v0.68.0) - 2023-04-24
### Breaking Changes 🛠
- Fix `runn run` output ( Show runbook path / Add `--verbose` option ) by @k1LoW in https://github.com/k1LoW/runn/pull/485
- Fix output JSON of `runn run --format json` by @k1LoW in https://github.com/k1LoW/runn/pull/487
- Fix output of `runn run` by @k1LoW in https://github.com/k1LoW/runn/pull/488
- Further fix output of `runn run` by @k1LoW in https://github.com/k1LoW/runn/pull/489

## [v0.67.0](https://github.com/k1LoW/runn/compare/v0.66.0...v0.67.0) - 2023-04-06
### Breaking Changes 🛠
- Generate `outcome` store values even if runbook is skipped by @k1LoW in https://github.com/k1LoW/runn/pull/483
### Other Changes
- Bump github.com/docker/docker from 20.10.7+incompatible to 20.10.24+incompatible by @dependabot in https://github.com/k1LoW/runn/pull/479
- Add Cloud Spanner support without `xo/dburl` by @BIwashi in https://github.com/k1LoW/runn/pull/482
- Fix `README.md` for spanner by @BIwashi in https://github.com/k1LoW/runn/pull/484

## [v0.66.0](https://github.com/k1LoW/runn/compare/v0.65.0...v0.66.0) - 2023-04-03
### Breaking Changes 🛠
- Change function signature of GrpcRunner and Add GrpcRunnerWithOptions by @k1LoW in https://github.com/k1LoW/runn/pull/475
### New Features 🎉
- Support `force:` section to force all steps to run. by @k1LoW in https://github.com/k1LoW/runn/pull/478
### Other Changes
- Bump github.com/opencontainers/runc from 1.1.2 to 1.1.5 by @dependabot in https://github.com/k1LoW/runn/pull/477

## [v0.65.0](https://github.com/k1LoW/runn/compare/v0.64.1...v0.65.0) - 2023-03-28
### Breaking Changes 🛠
- Make to record the outcome of each step to store. by @k1LoW in https://github.com/k1LoW/runn/pull/472
### Fix bug 🐛
- Fix recording bug by @k1LoW in https://github.com/k1LoW/runn/pull/473
### Other Changes
- Bump up httpstub/grpcstub version by @k1LoW in https://github.com/k1LoW/runn/pull/467
- Add sudo test using `keepSession: true` by @k1LoW in https://github.com/k1LoW/runn/pull/469
- Record step result by @k1LoW in https://github.com/k1LoW/runn/pull/470
- Change store key to const by @k1LoW in https://github.com/k1LoW/runn/pull/471
- Fix gRPC runner dial strategy by @k1LoW in https://github.com/k1LoW/runn/pull/474

## [v0.64.1](https://github.com/k1LoW/runn/compare/v0.64.0...v0.64.1) - 2023-03-11
### Breaking Changes 🛠
- Fix type of value for concurrent by @k1LoW in https://github.com/k1LoW/runn/pull/466

## [v0.64.0](https://github.com/k1LoW/runn/compare/v0.63.2...v0.64.0) - 2023-03-11
### Breaking Changes 🛠
- Rename `Parallel` to `Concurrent` by @k1LoW in https://github.com/k1LoW/runn/pull/462
### New Features 🎉
- Add `concurrency:` for ensuring that only a single runbook using the same group will run at a time. by @k1LoW in https://github.com/k1LoW/runn/pull/464

## [v0.63.2](https://github.com/k1LoW/runn/compare/v0.63.1...v0.63.2) - 2023-03-06
### Fix bug 🐛
- Fix handling skip count by @k1LoW in https://github.com/k1LoW/runn/pull/459
### Other Changes
- Use SetLimit by @k1LoW in https://github.com/k1LoW/runn/pull/461

## [v0.63.1](https://github.com/k1LoW/runn/compare/v0.63.0...v0.63.1) - 2023-03-05
### New Features 🎉
- Add `--shard-n/--shard-index` options for sharding runbooks by @k1LoW in https://github.com/k1LoW/runn/pull/456
### Fix bug 🐛
- Fix `run list` not working with options by @k1LoW in https://github.com/k1LoW/runn/pull/454
### Other Changes
- Rename countOfSteps to numberOfSteps by @k1LoW in https://github.com/k1LoW/runn/pull/457
- Add `interval:` to sshd_keep_session.yml by @k1LoW in https://github.com/k1LoW/runn/pull/458

## [v0.63.0](https://github.com/k1LoW/runn/compare/v0.62.0...v0.63.0) - 2023-03-04
### Breaking Changes 🛠
- Update k1LoW/expand to v0.7.0 by @k1LoW in https://github.com/k1LoW/runn/pull/452
### Fix bug 🐛
- Escape `\n` if parameter is a string literal by @k1LoW in https://github.com/k1LoW/runn/pull/450

## [v0.62.0](https://github.com/k1LoW/runn/compare/v0.61.0...v0.62.0) - 2023-02-27
### Breaking Changes 🛠
- Allow to list somewhat broken runbooks in the `runn list` command by @k1LoW in https://github.com/k1LoW/runn/pull/445
- Fix output of `runn list` ( Add count of steps ) by @k1LoW in https://github.com/k1LoW/runn/pull/446
### Fix bug 🐛
- Fix handling for SSH connection by @k1LoW in https://github.com/k1LoW/runn/pull/447
### Other Changes
- Shorten path of `runn list` output by @k1LoW in https://github.com/k1LoW/runn/pull/448

## [v0.61.0](https://github.com/k1LoW/runn/compare/v0.60.1...v0.61.0) - 2023-02-24
### Breaking Changes 🛠
- The values of `runners:` and `vars:` in the included runbook accept variables from the parent runbook. by @k1LoW in https://github.com/k1LoW/runn/pull/443
### Other Changes
- If `keyboardInteractive:` is not set and the user is prompted for key input, runn will prompt the user for input. by @k1LoW in https://github.com/k1LoW/runn/pull/440

## [v0.60.1](https://github.com/k1LoW/runn/compare/v0.60.0...v0.60.1) - 2023-02-20
### Other Changes
- Fix handling context by @k1LoW in https://github.com/k1LoW/runn/pull/438

## [v0.60.0](https://github.com/k1LoW/runn/compare/v0.59.4...v0.60.0) - 2023-02-19
### New Features 🎉
- Support Windows by @k1LoW in https://github.com/k1LoW/runn/pull/436
### Other Changes
- Use `modernc.org/sqlite` package by @k1LoW in https://github.com/k1LoW/runn/pull/434
- Add test on Windows and macOS by @k1LoW in https://github.com/k1LoW/runn/pull/437

## [v0.59.4](https://github.com/k1LoW/runn/compare/v0.59.3...v0.59.4) - 2023-02-19
### Other Changes
- Check file of sshConfig: exists by @k1LoW in https://github.com/k1LoW/runn/pull/430
- Bump golang.org/x/net from 0.5.0 to 0.7.0 by @dependabot in https://github.com/k1LoW/runn/pull/432
- Update pkgs by @k1LoW in https://github.com/k1LoW/runn/pull/433

## [v0.59.3](https://github.com/k1LoW/runn/compare/v0.59.2...v0.59.3) - 2023-02-13
### New Features 🎉
- Support `identityKey:` for SSH Runner by @k1LoW in https://github.com/k1LoW/runn/pull/427

## [v0.59.2](https://github.com/k1LoW/runn/compare/v0.59.1...v0.59.2) - 2023-02-11
### Other Changes
- Add link of runn cookbook by @k1LoW in https://github.com/k1LoW/runn/pull/424
- Append basename function by @k2tzumi in https://github.com/k1LoW/runn/pull/423
- Fixing goreleaser deprecated by @k2tzumi in https://github.com/k1LoW/runn/pull/426

## [v0.59.1](https://github.com/k1LoW/runn/compare/v0.59.0...v0.59.1) - 2023-02-11
### Fix bug 🐛
- Fix to be able to parse different return value of query in case of SSH port forwarding by @k1LoW in https://github.com/k1LoW/runn/pull/421
### Other Changes
- Add SSH port fowarding test by @k1LoW in https://github.com/k1LoW/runn/pull/420

## [v0.59.0](https://github.com/k1LoW/runn/compare/v0.58.3...v0.59.0) - 2023-02-10
### Breaking Changes 🛠
- [BREAKING] Override req.Host when set "Host" in `headers:` by @k1LoW in https://github.com/k1LoW/runn/pull/416
### New Features 🎉
- Support SSH port forwarding ( `localForward:` ) for SSH Runner by @k1LoW in https://github.com/k1LoW/runn/pull/414
- Add `keyboardInteractive:` for SSH Runner by @k1LoW in https://github.com/k1LoW/runn/pull/418
### Fix bug 🐛
- Fix buildTree output by @k1LoW in https://github.com/k1LoW/runn/pull/419

## [v0.58.3](https://github.com/k1LoW/runn/compare/v0.58.2...v0.58.3) - 2023-02-09
### Fix bug 🐛
- Fix buildTree with builtin function by @k1LoW in https://github.com/k1LoW/runn/pull/411
### Other Changes
- Implement intersect function by @IzumiSy in https://github.com/k1LoW/runn/pull/413

## [v0.58.2](https://github.com/k1LoW/runn/compare/v0.58.1...v0.58.2) - 2023-02-07
### Fix bug 🐛
- Fix nil pointer dereference at httpRunner by @k1LoW in https://github.com/k1LoW/runn/pull/406
- Add test by @k1LoW in https://github.com/k1LoW/runn/pull/408

## [v0.58.1](https://github.com/k1LoW/runn/compare/v0.58.0...v0.58.1) - 2023-02-05
### Other Changes
- Fix setup go on macos-latest by @k1LoW in https://github.com/k1LoW/runn/pull/404

## [v0.58.0](https://github.com/k1LoW/runn/compare/v0.57.2...v0.58.0) - 2023-02-04
### New Features 🎉
- Support expand vars to output path of dump by @k2tzumi in https://github.com/k1LoW/runn/pull/401
### Other Changes
- [BREAKING] Bump up github.com/antonmedv/expr version by @k1LoW in https://github.com/k1LoW/runn/pull/403
- Bump up go and pkgs version by @k2tzumi in https://github.com/k1LoW/runn/pull/400

## [v0.57.2](https://github.com/k1LoW/runn/compare/v0.57.1...v0.57.2) - 2023-02-03
### New Features 🎉
- Support custom CA and certificates for HTTP Runner by @k1LoW in https://github.com/k1LoW/runn/pull/399
### Other Changes
- Use x509.SystemCertPool instead of x509.NewCertPool by @k1LoW in https://github.com/k1LoW/runn/pull/394
- Synchronize handling of HTTPRunner and parseHTTPRunnerWithDetailed configs by @k1LoW in https://github.com/k1LoW/runn/pull/396

## [v0.57.1](https://github.com/k1LoW/runn/compare/v0.57.0...v0.57.1) - 2023-01-29
### Other Changes
- Print step description on failure if available by @nobuyo in https://github.com/k1LoW/runn/pull/390
- Bump up version of `kin-openapi` to `v0.113.1-0.20230128122015-6e233af317f2` by @k2tzumi in https://github.com/k1LoW/runn/pull/391

## [v0.57.0](https://github.com/k1LoW/runn/compare/v0.56.2...v0.57.0) - 2023-01-27
### New Features 🎉
- Add `latestTab` to cdpfn by @k1LoW in https://github.com/k1LoW/runn/pull/388
### Fix bug 🐛
- Freeze loadt result by @k1LoW in https://github.com/k1LoW/runn/pull/389
### Other Changes
- Specify the Category of the release note by @k2tzumi in https://github.com/k1LoW/runn/pull/383
- Fix test case by @k1LoW in https://github.com/k1LoW/runn/pull/386
- chore: bump up k1LoW/ghfs 0.7.0 to 0.8.0 by @miseyu in https://github.com/k1LoW/runn/pull/384

## [v0.56.2](https://github.com/k1LoW/runn/compare/v0.56.1...v0.56.2) - 2023-01-22
- Bump up version of `getkin/kin-openapi` by @k2tzumi in https://github.com/k1LoW/runn/pull/379

## [v0.56.1](https://github.com/k1LoW/runn/compare/v0.56.0...v0.56.1) - 2023-01-22
- Fixing `Frame not found for the given storage id` by @k2tzumi in https://github.com/k1LoW/runn/pull/380

## [v0.56.0](https://github.com/k1LoW/runn/compare/v0.55.0...v0.56.0) - 2023-01-08
- [BREAKING] Fix default interval for loop by @k1LoW in https://github.com/k1LoW/runn/pull/376

## [v0.55.0](https://github.com/k1LoW/runn/compare/v0.54.5...v0.55.0) - 2023-01-05
- Bonsai by @k1LoW in https://github.com/k1LoW/runn/pull/372
- Add `--threshold` option for checking result of loadt by @k1LoW in https://github.com/k1LoW/runn/pull/375

## [v0.54.5](https://github.com/k1LoW/runn/compare/v0.54.4...v0.54.5) - 2022-12-29
- Prefer step desc if exist in debug printing by @nobuyo in https://github.com/k1LoW/runn/pull/369
- Add `--cache-dir` and `--retain-cache-dir` option by @k1LoW in https://github.com/k1LoW/runn/pull/371

## [v0.54.4](https://github.com/k1LoW/runn/compare/v0.54.3...v0.54.4) - 2022-12-25
- Fix handling usage of flags by @k1LoW in https://github.com/k1LoW/runn/pull/366
- Support fetch files when using external vars with `http://` or `github://` by @k1LoW in https://github.com/k1LoW/runn/pull/368

## [v0.54.3](https://github.com/k1LoW/runn/compare/v0.54.2...v0.54.3) - 2022-12-25
- Fix panic when --concurrent > 1 by @k1LoW in https://github.com/k1LoW/runn/pull/362
- Fix loadt result by @k1LoW in https://github.com/k1LoW/runn/pull/364

## [v0.54.2](https://github.com/k1LoW/runn/compare/v0.54.1...v0.54.2) - 2022-12-24
- Remove debug code by @k1LoW in https://github.com/k1LoW/runn/pull/360

## [v0.54.1](https://github.com/k1LoW/runn/compare/v0.54.0...v0.54.1) - 2022-12-23
- Support fetch files when using `http://` or `github://` by @k1LoW in https://github.com/k1LoW/runn/pull/357

## [v0.54.0](https://github.com/k1LoW/runn/compare/v0.53.4...v0.54.0) - 2022-12-23
- Support fetching runbooks via `https://` or `github://` by @k1LoW in https://github.com/k1LoW/runn/pull/355

## [v0.53.4](https://github.com/k1LoW/runn/compare/v0.53.3...v0.53.4) - 2022-12-20
- runn examples by @atsushi-ishibashi in https://github.com/k1LoW/runn/pull/351
- Fix to be able to get trailers in Bidirectional streaming by @k1LoW in https://github.com/k1LoW/runn/pull/353
- Fix sshd_keep_session.yml by @k1LoW in https://github.com/k1LoW/runn/pull/354

## [v0.53.3](https://github.com/k1LoW/runn/compare/v0.53.2...v0.53.3) - 2022-12-09
- openapi3filter.RegisterBodyDecoder only needs once. by @k1LoW in https://github.com/k1LoW/runn/pull/349

## [v0.53.2](https://github.com/k1LoW/runn/compare/v0.53.1...v0.53.2) - 2022-12-05
- Fix duplicate drainBody by @k1LoW in https://github.com/k1LoW/runn/pull/347

## [v0.53.1](https://github.com/k1LoW/runn/compare/v0.53.0...v0.53.1) - 2022-12-05
- Fix capture/runbook.go reading out request body when multipart/form-data by @k1LoW in https://github.com/k1LoW/runn/pull/345

## [v0.53.0](https://github.com/k1LoW/runn/compare/v0.52.3...v0.53.0) - 2022-12-02
- [BREAKING] Support for loop run of runbook by @k1LoW in https://github.com/k1LoW/runn/pull/340

## [v0.52.3](https://github.com/k1LoW/runn/compare/v0.52.2...v0.52.3) - 2022-11-30
- Fix handling result by @k1LoW in https://github.com/k1LoW/runn/pull/337
- Add path resolution for setUploadFile.path by @k1LoW in https://github.com/k1LoW/runn/pull/339

## [v0.52.2](https://github.com/k1LoW/runn/compare/v0.52.1...v0.52.2) - 2022-11-30
- Support array for multipart body by @atsushi-ishibashi in https://github.com/k1LoW/runn/pull/332
- Fix sshOutTimeout by @k1LoW in https://github.com/k1LoW/runn/pull/333
- Add Stdout Stderr opiton by @k1LoW in https://github.com/k1LoW/runn/pull/335
- Support capturing multiple files (capture.runbook) by @k1LoW in https://github.com/k1LoW/runn/pull/336

## [v0.52.1](https://github.com/k1LoW/runn/compare/v0.52.0...v0.52.1) - 2022-11-29
- Show index of CDP actions when error by @k1LoW in https://github.com/k1LoW/runn/pull/327
- Fix Dump runner output by @k1LoW in https://github.com/k1LoW/runn/pull/329
- Reset headers as well by @atsushi-ishibashi in https://github.com/k1LoW/runn/pull/331

## [v0.52.0](https://github.com/k1LoW/runn/compare/v0.51.1...v0.52.0) - 2022-11-29
- Support SSH runner with detailed by @k1LoW in https://github.com/k1LoW/runn/pull/322
- Add SSHRunnerWithOptions by @k1LoW in https://github.com/k1LoW/runn/pull/324
- [BREAKING] Add KeepSession Option / Set keepSession to default false by @k1LoW in https://github.com/k1LoW/runn/pull/325

## [v0.51.1](https://github.com/k1LoW/runn/compare/v0.51.0...v0.51.1) - 2022-11-28
- Bump up k1LoW/sshc version by @k1LoW in https://github.com/k1LoW/runn/pull/320

## [v0.51.0](https://github.com/k1LoW/runn/compare/v0.50.0...v0.51.0) - 2022-11-28
- Update .octocov.yml ( report to GitHub Actions Summary ) by @k1LoW in https://github.com/k1LoW/runn/pull/311
- Add pronunciation by @k1LoW in https://github.com/k1LoW/runn/pull/313
- [BREAKING] Change function signature of BeforeFunc args  by @k1LoW in https://github.com/k1LoW/runn/pull/314
- Support SSH by @k1LoW in https://github.com/k1LoW/runn/pull/315
- [BREAKING] Support capturing SSH command stdout and stderr by @k1LoW in https://github.com/k1LoW/runn/pull/316
- Add doc about SSH Runner by @k1LoW in https://github.com/k1LoW/runn/pull/317
- Fix handling stdout/stderr of operator and runners by @k1LoW in https://github.com/k1LoW/runn/pull/318
- Update pkgs by @k1LoW in https://github.com/k1LoW/runn/pull/319

## [v0.50.0](https://github.com/k1LoW/runn/compare/v0.49.3...v0.50.0) - 2022-11-25
- Support built-in func on Dump runner by @k1LoW in https://github.com/k1LoW/runn/pull/303
- Bonsai dummy images by @k1LoW in https://github.com/k1LoW/runn/pull/305
- Export `eval*` functions for afterFuncs by @k1LoW in https://github.com/k1LoW/runn/pull/306
- Fix handling store keys by @k1LoW in https://github.com/k1LoW/runn/pull/308
- Add AfterFuncIf Option by @k1LoW in https://github.com/k1LoW/runn/pull/309
- Add `localStorage` `sessionStorage` to cdpfn by @atsushi-ishibashi in https://github.com/k1LoW/runn/pull/307

## [v0.49.3](https://github.com/k1LoW/runn/compare/v0.49.2...v0.49.3) - 2022-11-21
- Bump up k1LoW/expand version by @k1LoW in https://github.com/k1LoW/runn/pull/300
- Handle expand error by @k1LoW in https://github.com/k1LoW/runn/pull/302

## [v0.49.2](https://github.com/k1LoW/runn/compare/v0.49.1...v0.49.2) - 2022-11-20
- Add `input` `secret` `select` to built-in function for interactive input by @k1LoW in https://github.com/k1LoW/runn/pull/297

## [v0.49.1](https://github.com/k1LoW/runn/compare/v0.49.0...v0.49.1) - 2022-11-19
- Bonsai by @k1LoW in https://github.com/k1LoW/runn/pull/292
- Support marshal runbook with mapped steps by @k1LoW in https://github.com/k1LoW/runn/pull/294
- Support for mapped step runbook on `runn new`. by @k1LoW in https://github.com/k1LoW/runn/pull/295

## [v0.49.0](https://github.com/k1LoW/runn/compare/v0.48.0...v0.49.0) - 2022-11-18
- [BREAKING] Rename CaptureFailed to CaptureFailure by @k1LoW in https://github.com/k1LoW/runn/pull/280
- [BREAKING] Support JSON output of RunN result by @k1LoW in https://github.com/k1LoW/runn/pull/279
- Rename `userAgent` to `setUserAgent` by @k1LoW in https://github.com/k1LoW/runn/pull/282
- Set timeout to CDP Runner for each step by @k1LoW in https://github.com/k1LoW/runn/pull/284
- Aggregate runbook parsing process by @k1LoW in https://github.com/k1LoW/runn/pull/285
- Support multipart/form-data via http by @atsushi-ishibashi in https://github.com/k1LoW/runn/pull/283
- Tiny fix test for multipart/form-data by @k1LoW in https://github.com/k1LoW/runn/pull/288
- Support relative file path by @k1LoW in https://github.com/k1LoW/runn/pull/289
- base64encode & decode as built-in func by @atsushi-ishibashi in https://github.com/k1LoW/runn/pull/290
- Support parsing body of multipart/form-data by @k1LoW in https://github.com/k1LoW/runn/pull/291

## [v0.48.0](https://github.com/k1LoW/runn/compare/v0.47.3...v0.48.0) - 2022-11-14
- [BREAKING] AfterFuncs will run even if the scenario fails by @k1LoW in https://github.com/k1LoW/runn/pull/273
- [BREAKING] Make AfterFuncs receive runbook run error by @k1LoW in https://github.com/k1LoW/runn/pull/275
- Make AfterFuncs receive runbook run result ( contains error ) by @k1LoW in https://github.com/k1LoW/runn/pull/276
- Bump up k1LoW/expand version by @k1LoW in https://github.com/k1LoW/runn/pull/277
- Add `userAgent` for setting User-Agent header by @k1LoW in https://github.com/k1LoW/runn/pull/278

## [v0.47.3](https://github.com/k1LoW/runn/compare/v0.47.2...v0.47.3) - 2022-11-10
- Add git for using `actions/checkout@v3` on `container:` workflow by @k1LoW in https://github.com/k1LoW/runn/pull/271

## [v0.47.2](https://github.com/k1LoW/runn/compare/v0.47.1...v0.47.2) - 2022-11-09
- Fix panic: reflect: Call using zero Value argument by @k1LoW in https://github.com/k1LoW/runn/pull/269

## [v0.47.1](https://github.com/k1LoW/runn/compare/v0.47.0...v0.47.1) - 2022-11-08
- Add `fullHTML` for getting full HTML of current page. by @k1LoW in https://github.com/k1LoW/runn/pull/262
- Fix timing of record step value on Dump runner and Bind runner. by @k1LoW in https://github.com/k1LoW/runn/pull/264
- Support `current` and `previous` key on Dump runner by @k1LoW in https://github.com/k1LoW/runn/pull/265
- Fix Dump runner ( support file output ) by @k1LoW in https://github.com/k1LoW/runn/pull/266
- Add `screenshot` for take screenshot of page. by @k1LoW in https://github.com/k1LoW/runn/pull/267
- Add `scroll` for scroll window by @k1LoW in https://github.com/k1LoW/runn/pull/268

## [v0.47.0](https://github.com/k1LoW/runn/compare/v0.46.0...v0.47.0) - 2022-11-08
- Enable gosec by @k1LoW in https://github.com/k1LoW/runn/pull/258
- Support Chrome DevTools Protocol (CDP) by @k1LoW in https://github.com/k1LoW/runn/pull/257
- Add `setUploadFile` for uploading file via browser. by @k1LoW in https://github.com/k1LoW/runn/pull/260
- Add Chrominium to Docker image for supporting CDP by @k1LoW in https://github.com/k1LoW/runn/pull/261

## [v0.46.0](https://github.com/k1LoW/runn/compare/v0.45.1...v0.46.0) - 2022-11-02
- [BREAKING] Support `--random` option for CLI and `RunRandom` option for test helper by @k1LoW in https://github.com/k1LoW/runn/pull/253
- Add number of runbooks to loadt result by @k1LoW in https://github.com/k1LoW/runn/pull/255
- Re fix loadt result by @k1LoW in https://github.com/k1LoW/runn/pull/256

## [v0.45.1](https://github.com/k1LoW/runn/compare/v0.45.0...v0.45.1) - 2022-11-01
- Fix rprof --sort by @k1LoW in https://github.com/k1LoW/runn/pull/251

## [v0.45.0](https://github.com/k1LoW/runn/compare/v0.44.1...v0.45.0) - 2022-10-31
- [BREAKING] Add ID for record id and context by @k1LoW in https://github.com/k1LoW/runn/pull/248
- Add `runn rprof` command for reading runn profile file by @k1LoW in https://github.com/k1LoW/runn/pull/250

## [v0.44.1](https://github.com/k1LoW/runn/compare/v0.44.0...v0.44.1) - 2022-10-31
- Add `--profile/--profile-out` option for CLI by @k1LoW in https://github.com/k1LoW/runn/pull/245
- Copy ca-certificates.crt from builder image by @k1LoW in https://github.com/k1LoW/runn/pull/247

## [v0.44.0](https://github.com/k1LoW/runn/compare/v0.43.0...v0.44.0) - 2022-10-28
- Fix test by @k1LoW in https://github.com/k1LoW/runn/pull/239
- Support `--var foo.bar.key:value` option for CLI by @k1LoW in https://github.com/k1LoW/runn/pull/241
- Fix `runn new` when no `--and-run` by @k1LoW in https://github.com/k1LoW/runn/pull/242
- Support `--runner req:https://example.com/api/v1` option for CLI by @k1LoW in https://github.com/k1LoW/runn/pull/243
- Add `runn loadt` for load test using runbooks by @k1LoW in https://github.com/k1LoW/runn/pull/244

## [v0.43.0](https://github.com/k1LoW/runn/compare/v0.42.1...v0.43.0) - 2022-10-24
- [BREAKING] Support `*sql.Tx` with DBRunner by @k1LoW in https://github.com/k1LoW/runn/pull/237

## [v0.42.1](https://github.com/k1LoW/runn/compare/v0.42.0...v0.42.1) - 2022-10-23
- Fix docker image build pipeline by @k1LoW in https://github.com/k1LoW/runn/pull/234
- Set `steps.*.run` to true when step is run by @k1LoW in https://github.com/k1LoW/runn/pull/236

## [v0.42.0](https://github.com/k1LoW/runn/compare/v0.41.0...v0.42.0) - 2022-10-23
- Append runn new using access log to "Quickstart" section by @k1LoW in https://github.com/k1LoW/runn/pull/227
- Add comment for vars by @k1LoW in https://github.com/k1LoW/runn/pull/229
- Support for appending steps via `runn new` by @k1LoW in https://github.com/k1LoW/runn/pull/230
- Bump up go and pkgs version by @k1LoW in https://github.com/k1LoW/runn/pull/231
- Fix typo by @k1LoW in https://github.com/k1LoW/runn/pull/232
- Support `previous` variable by @k1LoW in https://github.com/k1LoW/runn/pull/233

## [v0.41.0](https://github.com/k1LoW/runn/compare/v0.40.0...v0.41.0) - 2022-10-22
- Add "Quickstart" section by @k1LoW in https://github.com/k1LoW/runn/pull/221
- Support `runn new` via STDIN by @k1LoW in https://github.com/k1LoW/runn/pull/224
- Support create runbook from access log by @k1LoW in https://github.com/k1LoW/runn/pull/225
- Fix host handling by @k1LoW in https://github.com/k1LoW/runn/pull/226

## [v0.40.0](https://github.com/k1LoW/runn/compare/v0.39.0...v0.40.0) - 2022-10-20
- Add command `runn new` to create new runbook by @k1LoW in https://github.com/k1LoW/runn/pull/216
- Support gRPCurl command by @k1LoW in https://github.com/k1LoW/runn/pull/218
- [BREAKING] gRPC runner use TLS by default by @k1LoW in https://github.com/k1LoW/runn/pull/219
- Support `--grpc-no-tls` option for CLI and `GRPCNoTLS` option for test helper by @k1LoW in https://github.com/k1LoW/runn/pull/220

## [v0.39.0](https://github.com/k1LoW/runn/compare/v0.38.0...v0.39.0) - 2022-10-16
- Update grpcstub to v0.6.1 by @k1LoW in https://github.com/k1LoW/runn/pull/200
- Append capture points by @k1LoW in https://github.com/k1LoW/runn/pull/205
- Make it possible to get counts of run results by @k1LoW in https://github.com/k1LoW/runn/pull/206
- [BREAKING] Change the custom operation of `runn run` command to `operators.RunN()` by @k1LoW in https://github.com/k1LoW/runn/pull/207
- Fix logic of listing in `runn list` by @k1LoW in https://github.com/k1LoW/runn/pull/208
- Support `--shuffle` option for CLI and `Shuffle` option for test helper by @k1LoW in https://github.com/k1LoW/runn/pull/209
- Support `--parallel` option for CLI and `Parallel` option for test helper by @k1LoW in https://github.com/k1LoW/runn/pull/211
- Fix handling of Option by @k1LoW in https://github.com/k1LoW/runn/pull/212
- [BREAKING] Support `--skip-included` option for CLI by @k1LoW in https://github.com/k1LoW/runn/pull/213
- Support `--shuffle` `--overlay` `--underlay` for `runn list` by @k1LoW in https://github.com/k1LoW/runn/pull/214
- Support `--sample` option for CLI by @k1LoW in https://github.com/k1LoW/runn/pull/215

## [v0.38.0](https://github.com/k1LoW/runn/compare/v0.37.4...v0.38.0) - 2022-10-06
- Add `notFollowRedirect` option for HTTP Runner by @k1LoW in https://github.com/k1LoW/runn/pull/198

## [v0.37.4](https://github.com/k1LoW/runn/compare/v0.37.3...v0.37.4) - 2022-10-05
- Wrap expr.Eval() with eval() by @k1LoW in https://github.com/k1LoW/runn/pull/190
- Move functions related to function eval() to eval.go by @k1LoW in https://github.com/k1LoW/runn/pull/192
- Fix timing of trimming comments of string to eval() by @k1LoW in https://github.com/k1LoW/runn/pull/193
- Fix installing Go by @k1LoW in https://github.com/k1LoW/runn/pull/194
- Move logic form o.expand() to evalExpand() by @k1LoW in https://github.com/k1LoW/runn/pull/195
- [BREAKING] Update k1LoW/expand to v0.5.2 by @k1LoW in https://github.com/k1LoW/runn/pull/196

## [v0.37.3](https://github.com/k1LoW/runn/compare/v0.37.2...v0.37.3) - 2022-10-01
- Support for comment statements within the test syntax by @k2tzumi in https://github.com/k1LoW/runn/pull/188

## [v0.37.2](https://github.com/k1LoW/runn/compare/v0.37.1...v0.37.2) - 2022-10-01
- migrate builtin functions to package by @k2tzumi in https://github.com/k1LoW/runn/pull/185
- Append diff function by @k2tzumi in https://github.com/k1LoW/runn/pull/186

## [v0.37.1](https://github.com/k1LoW/runn/compare/v0.37.0...v0.37.1) - 2022-10-01
- Fixed compare function always returning true by @k2tzumi in https://github.com/k1LoW/runn/pull/183

## [v0.37.0](https://github.com/k1LoW/runn/compare/v0.36.2...v0.37.0) - 2022-10-01
- Support solo-single-line comment by @k1LoW in https://github.com/k1LoW/runn/pull/176
- Support single-line comment by @k1LoW in https://github.com/k1LoW/runn/pull/179
- Add doc about structure of request and response by @k1LoW in https://github.com/k1LoW/runn/pull/181
- Show number of loop times when an error occurs during loop. by @k1LoW in https://github.com/k1LoW/runn/pull/180
- Fix buildTree() by @k1LoW in https://github.com/k1LoW/runn/pull/182

## [v0.36.2](https://github.com/k1LoW/runn/compare/v0.36.1...v0.36.2) - 2022-09-28
- Re-fix release pipeline by @k1LoW in https://github.com/k1LoW/runn/pull/174

## [v0.36.1](https://github.com/k1LoW/runn/compare/v0.36.0...v0.36.1) - 2022-09-28
- Fix release pipeline by @k1LoW in https://github.com/k1LoW/runn/pull/172

## [v0.35.3](https://github.com/k1LoW/runn/compare/v0.35.2...v0.35.3) - 2022-09-28
- Use tagpr by @k1LoW in https://github.com/k1LoW/runn/pull/166
- [README]The `*sql.DB` used by the test target and the `*sql.DB` used by runn should be separated. by @k1LoW in https://github.com/k1LoW/runn/pull/169
- Improve post-processing of YAML parsed results by @k1LoW in https://github.com/k1LoW/runn/pull/168
- Add badge by @k1LoW in https://github.com/k1LoW/runn/pull/171
- Support `--overlay/--underlay` option for CLI and `Overlay/Underlay` option for test helper by @k1LoW in https://github.com/k1LoW/runn/pull/170

## [v0.35.2](https://github.com/k1LoW/runn/compare/v0.35.1...v0.35.2) (2022-09-26)

* Fix book path [#165](https://github.com/k1LoW/runn/pull/165) ([k1LoW](https://github.com/k1LoW))

## [v0.35.1](https://github.com/k1LoW/runn/compare/v0.35.0...v0.35.1) (2022-09-26)

* Use gopkg.in/yaml.v2 only for the first YAML parsing to keep the parsing stable [#159](https://github.com/k1LoW/runn/pull/159) ([k1LoW](https://github.com/k1LoW))
* Fix error handling [#162](https://github.com/k1LoW/runn/pull/162) ([k1LoW](https://github.com/k1LoW))
* Decimal point was not considered in the parsing of duration. [#163](https://github.com/k1LoW/runn/pull/163) ([k1LoW](https://github.com/k1LoW))
* Change loop index variable name [#160](https://github.com/k1LoW/runn/pull/160) ([k1LoW](https://github.com/k1LoW))
* Fix bind reserved key [#158](https://github.com/k1LoW/runn/pull/158) ([k2tzumi](https://github.com/k2tzumi))
* Allow time units in the `*interval:` section [#157](https://github.com/k1LoW/runn/pull/157) ([k1LoW](https://github.com/k1LoW))
* Move the parsing operation of runbook(YAML) from operator{} to book{} to clarify responsibilities [#156](https://github.com/k1LoW/runn/pull/156) ([k1LoW](https://github.com/k1LoW))
* Rename `Array` to `List` [#154](https://github.com/k1LoW/runn/pull/154) ([k1LoW](https://github.com/k1LoW))

## [v0.35.0](https://github.com/k1LoW/runn/compare/v0.34.0...v0.35.0) (2022-09-19)

* Append newLoop test. [#153](https://github.com/k1LoW/runn/pull/153) ([k2tzumi](https://github.com/k2tzumi))
* Revert "[BREAKING] Use gopkg.in/yaml.v2 instead" [#152](https://github.com/k1LoW/runn/pull/152) ([k1LoW](https://github.com/k1LoW))
* Relative paths to json reads [#150](https://github.com/k1LoW/runn/pull/150) ([k2tzumi](https://github.com/k2tzumi))
* Fix string escapes of capture.Runbook [#149](https://github.com/k1LoW/runn/pull/149) ([k1LoW](https://github.com/k1LoW))
* Add option `--capture` [#148](https://github.com/k1LoW/runn/pull/148) ([k1LoW](https://github.com/k1LoW))
* [BREAKING] Use gopkg.in/yaml.v2 instead [#147](https://github.com/k1LoW/runn/pull/147) ([k1LoW](https://github.com/k1LoW))
* Introduce built-in capturer `capture.Runbook` [#144](https://github.com/k1LoW/runn/pull/144) ([k1LoW](https://github.com/k1LoW))
* Create `testutil` pkg [#143](https://github.com/k1LoW/runn/pull/143) ([k1LoW](https://github.com/k1LoW))

## [v0.34.0](https://github.com/k1LoW/runn/compare/v0.33.0...v0.34.0) (2022-09-12)

* Revert 'Allow strings "true" and "false" as true/false' [#142](https://github.com/k1LoW/runn/pull/142) ([k1LoW](https://github.com/k1LoW))
* Expand vars of Include while preserving type information [#140](https://github.com/k1LoW/runn/pull/140) ([k2tzumi](https://github.com/k2tzumi))
* Fix debugger output of DB Runner [#138](https://github.com/k1LoW/runn/pull/138) ([k1LoW](https://github.com/k1LoW))
* Add CaptureStart() CaptureEnd() to `type Capturer interface` [#137](https://github.com/k1LoW/runn/pull/137) ([k1LoW](https://github.com/k1LoW))

## [v0.33.0](https://github.com/k1LoW/runn/compare/v0.32.2...v0.33.0) (2022-09-07)

* Introduce `type Capturer interface` [#136](https://github.com/k1LoW/runn/pull/136) ([k1LoW](https://github.com/k1LoW))
* Build vars parameters with templates [#132](https://github.com/k1LoW/runn/pull/132) ([k2tzumi](https://github.com/k2tzumi))
* Improve column type determination using MySQL integration testing environment [#134](https://github.com/k1LoW/runn/pull/134) ([k1LoW](https://github.com/k1LoW))
* Setup integration test using Docker [#133](https://github.com/k1LoW/runn/pull/133) ([k1LoW](https://github.com/k1LoW))
* Support for DATETIME columns in DBRunner [#129](https://github.com/k1LoW/runn/pull/129) ([k2tzumi](https://github.com/k2tzumi))
* Add badges [#130](https://github.com/k1LoW/runn/pull/130) ([k1LoW](https://github.com/k1LoW))

## [v0.32.2](https://github.com/k1LoW/runn/compare/v0.32.1...v0.32.2) (2022-08-31)

* More detailed error messages when retrieving columns [#128](https://github.com/k1LoW/runn/pull/128) ([k2tzumi](https://github.com/k2tzumi))
* Support uint64 expand [#127](https://github.com/k1LoW/runn/pull/127) ([k2tzumi](https://github.com/k2tzumi))
* Append current test. [#126](https://github.com/k1LoW/runn/pull/126) ([k2tzumi](https://github.com/k2tzumi))

## [v0.32.1](https://github.com/k1LoW/runn/compare/v0.32.0...v0.32.1) (2022-08-30)

* Should be able to access values using `current` in the `until:` section. [#122](https://github.com/k1LoW/runn/pull/122) ([k1LoW](https://github.com/k1LoW))

## [v0.32.0](https://github.com/k1LoW/runn/compare/v0.31.0...v0.32.0) (2022-08-28)

* Use parser.Parse() instead of own parser [#120](https://github.com/k1LoW/runn/pull/120) ([k1LoW](https://github.com/k1LoW))
* Support `current` variable [#119](https://github.com/k1LoW/runn/pull/119) ([k1LoW](https://github.com/k1LoW))
* Change type of store.stepMap [#118](https://github.com/k1LoW/runn/pull/118) ([k1LoW](https://github.com/k1LoW))
* Improved logic for determining `steps:` syntax [#115](https://github.com/k1LoW/runn/pull/115) ([k1LoW](https://github.com/k1LoW))

## [v0.31.0](https://github.com/k1LoW/runn/compare/v0.30.3...v0.31.0) (2022-08-25)

* [BREAKING] Fix miss spell [#114](https://github.com/k1LoW/runn/pull/114) ([k1LoW](https://github.com/k1LoW))
* Append compare function [#106](https://github.com/k1LoW/runn/pull/106) ([k2tzumi](https://github.com/k2tzumi))
* Normalize storing values [#111](https://github.com/k1LoW/runn/pull/111) ([k1LoW](https://github.com/k1LoW))

## [v0.30.3](https://github.com/k1LoW/runn/compare/v0.30.2...v0.30.3) (2022-08-23)

* Stop creating anonymous structures in functions [#110](https://github.com/k1LoW/runn/pull/110) ([k1LoW](https://github.com/k1LoW))

## [v0.30.2](https://github.com/k1LoW/runn/compare/v0.30.1...v0.30.2) (2022-08-22)

* Add functions to cast as built-in functions [#107](https://github.com/k1LoW/runn/pull/107) ([k1LoW](https://github.com/k1LoW))
* Error if duplicate step keys are found in the runbook in Map syntax. [#109](https://github.com/k1LoW/runn/pull/109) ([k1LoW](https://github.com/k1LoW))

## [v0.30.1](https://github.com/k1LoW/runn/compare/v0.30.0...v0.30.1) (2022-08-22)

* Fix recordToMap [#108](https://github.com/k1LoW/runn/pull/108) ([k1LoW](https://github.com/k1LoW))

## [v0.30.0](https://github.com/k1LoW/runn/compare/v0.29.0...v0.30.0) (2022-08-21)

* Bump stopw version to 0.7.0 [#105](https://github.com/k1LoW/runn/pull/105) ([k1LoW](https://github.com/k1LoW))
* Use go.uber.org/multierr instead of github.com/hashicorp/go-multierror [#104](https://github.com/k1LoW/runn/pull/104) ([k1LoW](https://github.com/k1LoW))
* Create extension points for built-in functions [#102](https://github.com/k1LoW/runn/pull/102) ([k2tzumi](https://github.com/k2tzumi))

## [v0.29.0](https://github.com/k1LoW/runn/compare/v0.28.0...v0.29.0) (2022-08-20)

* Support reading json in the vars section of includes [#101](https://github.com/k1LoW/runn/pull/101) ([k2tzumi](https://github.com/k2tzumi))
* [BREAKING] Create a new `loop:` section and integrate it with the features of the `retry:` section. [#97](https://github.com/k1LoW/runn/pull/97) ([k1LoW](https://github.com/k1LoW))

## [v0.28.0](https://github.com/k1LoW/runn/compare/v0.27.1...v0.28.0) (2022-08-18)

* [BREAKING] Allow strings "true" and "false" as true/false values when evaluating conditions such as `if:` sections. [#99](https://github.com/k1LoW/runn/pull/99) ([k1LoW](https://github.com/k1LoW))
* Add playbook path to sub test name [#98](https://github.com/k1LoW/runn/pull/98) ([k1LoW](https://github.com/k1LoW))
* Measure elapsed time as profile [#96](https://github.com/k1LoW/runn/pull/96) ([k1LoW](https://github.com/k1LoW))

## [v0.27.1](https://github.com/k1LoW/runn/compare/v0.27.0...v0.27.1) (2022-08-15)

* Convert number var to json compatible type [#95](https://github.com/k1LoW/runn/pull/95) ([k2tzumi](https://github.com/k2tzumi))

## [v0.27.0](https://github.com/k1LoW/runn/compare/v0.26.2...v0.27.0) (2022-08-12)

* Support json object expand [#94](https://github.com/k1LoW/runn/pull/94) ([k2tzumi](https://github.com/k2tzumi))
* Support for read json in var section [#93](https://github.com/k1LoW/runn/pull/93) ([k2tzumi](https://github.com/k2tzumi))
* Fix race condition on bidirectional streaming [#92](https://github.com/k1LoW/runn/pull/92) ([k1LoW](https://github.com/k1LoW))

## [v0.26.2](https://github.com/k1LoW/runn/compare/v0.26.1...v0.26.2) (2022-08-01)

* Improvement of unstable operation during OAS Spec verification [#91](https://github.com/k1LoW/runn/pull/91) ([k2tzumi](https://github.com/k2tzumi))

## [v0.26.1](https://github.com/k1LoW/runn/compare/v0.26.0...v0.26.1) (2022-07-27)

* Support bash like interpolate  [#89](https://github.com/k1LoW/runn/pull/89) ([k2tzumi](https://github.com/k2tzumi))

## [v0.26.0](https://github.com/k1LoW/runn/compare/v0.25.0...v0.26.0) (2022-07-15)

* Support gRPC over TLS [#88](https://github.com/k1LoW/runn/pull/88) ([k1LoW](https://github.com/k1LoW))
* Rename [#87](https://github.com/k1LoW/runn/pull/87) ([k1LoW](https://github.com/k1LoW))
* Bump up go and pkg version [#86](https://github.com/k1LoW/runn/pull/86) ([k1LoW](https://github.com/k1LoW))
* [BREAKING] Rename `exit` to `close` [#85](https://github.com/k1LoW/runn/pull/85) ([k1LoW](https://github.com/k1LoW))

## [v0.25.0](https://github.com/k1LoW/runn/compare/v0.24.0...v0.25.0) (2022-07-11)

* Refactor code [#84](https://github.com/k1LoW/runn/pull/84) ([k1LoW](https://github.com/k1LoW))
* Decrease number of loop [#83](https://github.com/k1LoW/runn/pull/83) ([k1LoW](https://github.com/k1LoW))
* Support bool expand [#82](https://github.com/k1LoW/runn/pull/82) ([k2tzumi](https://github.com/k2tzumi))
* Fix grpc.Dial() option [#80](https://github.com/k1LoW/runn/pull/80) ([k1LoW](https://github.com/k1LoW))

## [v0.24.0](https://github.com/k1LoW/runn/compare/v0.23.1...v0.24.0) (2022-07-09)

* Support gRPC [#78](https://github.com/k1LoW/runn/pull/78) ([k1LoW](https://github.com/k1LoW))
* Support gRPC (Step 3: Bidirectional streaming RPC) [#77](https://github.com/k1LoW/runn/pull/77) ([k1LoW](https://github.com/k1LoW))
* Support gRPC (Step 2: Client streaming RPC) [#76](https://github.com/k1LoW/runn/pull/76) ([k1LoW](https://github.com/k1LoW))
* Support gRPC (Step 1: Server streaming RPC) [#75](https://github.com/k1LoW/runn/pull/75) ([k1LoW](https://github.com/k1LoW))
* Support gRPC (Step 0: Unary RPC) [#74](https://github.com/k1LoW/runn/pull/74) ([k1LoW](https://github.com/k1LoW))

## [v0.23.1](https://github.com/k1LoW/runn/compare/v0.23.0...v0.23.1) (2022-07-08)

* Fix the order of applying options when running RunN [#79](https://github.com/k1LoW/runn/pull/79) ([k1LoW](https://github.com/k1LoW))

## [v0.23.0](https://github.com/k1LoW/runn/compare/v0.22.3...v0.23.0) (2022-07-03)

* Support OpenAPI3 validation at authentication request [#73](https://github.com/k1LoW/runn/pull/73) ([k2tzumi](https://github.com/k2tzumi))
* Rename option `RunShard()` to `RunPart()` and set alias. [#72](https://github.com/k1LoW/runn/pull/72) ([k1LoW](https://github.com/k1LoW))

## [v0.22.3](https://github.com/k1LoW/runn/compare/v0.22.2...v0.22.3) (2022-07-03)

* Fix support string expr [#71](https://github.com/k1LoW/runn/pull/71) ([k2tzumi](https://github.com/k2tzumi))

## [v0.22.2](https://github.com/k1LoW/runn/compare/v0.22.1...v0.22.2) (2022-06-22)

* The functions added with Func() do not work in the included runbook. [#70](https://github.com/k1LoW/runn/pull/70) ([k1LoW](https://github.com/k1LoW))

## [v0.22.1](https://github.com/k1LoW/runn/compare/v0.22.0...v0.22.1) (2022-06-21)

* Fix RunPart [#69](https://github.com/k1LoW/runn/pull/69) ([k1LoW](https://github.com/k1LoW))

## [v0.22.0](https://github.com/k1LoW/runn/compare/v0.21.0...v0.22.0) (2022-06-20)

* Add option RunPart [#68](https://github.com/k1LoW/runn/pull/68) ([k1LoW](https://github.com/k1LoW))
* Add option RunSample [#67](https://github.com/k1LoW/runn/pull/67) ([k1LoW](https://github.com/k1LoW))
* Add option RunMatch [#66](https://github.com/k1LoW/runn/pull/66) ([k1LoW](https://github.com/k1LoW))
* Filter runbooks to be executed by the environment variable `RUNN_RUN` [#65](https://github.com/k1LoW/runn/pull/65) ([k1LoW](https://github.com/k1LoW))

## [v0.21.0](https://github.com/k1LoW/runn/compare/v0.20.2...v0.21.0) (2022-06-17)

* [BREAKING] Fix binding of parent vars [#64](https://github.com/k1LoW/runn/pull/64) ([k1LoW](https://github.com/k1LoW))

## [v0.20.2](https://github.com/k1LoW/runn/compare/v0.20.1...v0.20.2) (2022-06-17)

* `include.vars` should be variable expanded [#63](https://github.com/k1LoW/runn/pull/63) ([k1LoW](https://github.com/k1LoW))

## [v0.20.1](https://github.com/k1LoW/runn/compare/v0.20.0...v0.20.1) (2022-06-03)

* Fix handling recording when skip test [#62](https://github.com/k1LoW/runn/pull/62) ([k1LoW](https://github.com/k1LoW))

## [v0.20.0](https://github.com/k1LoW/runn/compare/v0.19.1...v0.20.0) (2022-06-03)

* Add option BeforeFunc and AfterFunc [#61](https://github.com/k1LoW/runn/pull/61) ([k1LoW](https://github.com/k1LoW))
* Support skipTest: [#60](https://github.com/k1LoW/runn/pull/60) ([k1LoW](https://github.com/k1LoW))
* Support overriding `vars:` of included runbook. [#59](https://github.com/k1LoW/runn/pull/59) ([k1LoW](https://github.com/k1LoW))
* Fix tokenize [#58](https://github.com/k1LoW/runn/pull/58) ([k1LoW](https://github.com/k1LoW))

## [v0.19.1](https://github.com/k1LoW/runn/compare/v0.19.0...v0.19.1) (2022-06-02)

* Set t.Helper() [#57](https://github.com/k1LoW/runn/pull/57) ([k1LoW](https://github.com/k1LoW))

## [v0.19.0](https://github.com/k1LoW/runn/compare/v0.18.1...v0.19.0) (2022-05-31)

* Add option SkipIncluded [#56](https://github.com/k1LoW/runn/pull/56) ([k1LoW](https://github.com/k1LoW))

## [v0.18.1](https://github.com/k1LoW/runn/compare/v0.18.0...v0.18.1) (2022-05-30)

* Fix layout of func to be set in store [#55](https://github.com/k1LoW/runn/pull/55) ([k1LoW](https://github.com/k1LoW))

## [v0.18.0](https://github.com/k1LoW/runn/compare/v0.17.2...v0.18.0) (2022-05-30)

* Add Func(k,v) for custom function [#54](https://github.com/k1LoW/runn/pull/54) ([k1LoW](https://github.com/k1LoW))

## [v0.17.2](https://github.com/k1LoW/runn/compare/v0.17.1...v0.17.2) (2022-05-25)

* Run multiple queries in a single transaction. [#53](https://github.com/k1LoW/runn/pull/53) ([k1LoW](https://github.com/k1LoW))

## [v0.17.1](https://github.com/k1LoW/runn/compare/v0.17.0...v0.17.1) (2022-05-24)

* Improve tokenize of conditional statements [#52](https://github.com/k1LoW/runn/pull/52) ([k1LoW](https://github.com/k1LoW))
* Skip validate response for unsupported format [#51](https://github.com/k1LoW/runn/pull/51) ([k1LoW](https://github.com/k1LoW))

## [v0.17.0](https://github.com/k1LoW/runn/compare/v0.16.2...v0.17.0) (2022-05-23)

* Support retry step [#50](https://github.com/k1LoW/runn/pull/50) ([k1LoW](https://github.com/k1LoW))
* Fix validation error message of http runner [#49](https://github.com/k1LoW/runn/pull/49) ([k1LoW](https://github.com/k1LoW))

## [v0.16.2](https://github.com/k1LoW/runn/compare/v0.16.1...v0.16.2) (2022-05-09)

* Fix hundling runner result [#48](https://github.com/k1LoW/runn/pull/48) ([k1LoW](https://github.com/k1LoW))

## [v0.16.1](https://github.com/k1LoW/runn/compare/v0.16.0...v0.16.1) (2022-05-09)

* Fix handling for bind runner [#47](https://github.com/k1LoW/runn/pull/47) ([k1LoW](https://github.com/k1LoW))

## [v0.16.0](https://github.com/k1LoW/runn/compare/v0.15.0...v0.16.0) (2022-05-09)

* Support `steps[*].desc:` `steps.<key>.desc:` [#46](https://github.com/k1LoW/runn/pull/46) ([k1LoW](https://github.com/k1LoW))
* Support `steps[*].if:` `steps.<key>.if:` [#45](https://github.com/k1LoW/runn/pull/45) ([k1LoW](https://github.com/k1LoW))

## [v0.15.0](https://github.com/k1LoW/runn/compare/v0.14.0...v0.15.0) (2022-05-06)

* Reserve `if:` and `desc:` [#44](https://github.com/k1LoW/runn/pull/44) ([k1LoW](https://github.com/k1LoW))
* Allow test/dump/bind runner to run in the same step as other runners [#43](https://github.com/k1LoW/runn/pull/43) ([k1LoW](https://github.com/k1LoW))

## [v0.14.0](https://github.com/k1LoW/runn/compare/v0.13.2...v0.14.0) (2022-04-06)

* Introduce `if:` section to skip steps [#42](https://github.com/k1LoW/runn/pull/42) ([k1LoW](https://github.com/k1LoW))
* Trim delimiter of []interface{} [#41](https://github.com/k1LoW/runn/pull/41) ([k1LoW](https://github.com/k1LoW))

## [v0.13.2](https://github.com/k1LoW/runn/compare/v0.13.1...v0.13.2) (2022-04-02)

* Trim delimiter of string [#40](https://github.com/k1LoW/runn/pull/40) ([k1LoW](https://github.com/k1LoW))

## [v0.13.1](https://github.com/k1LoW/runn/compare/v0.13.0...v0.13.1) (2022-03-31)

* Support runner override settings by delaying runner setting fixes as long as possible. [#39](https://github.com/k1LoW/runn/pull/39) ([k1LoW](https://github.com/k1LoW))

## [v0.13.0](https://github.com/k1LoW/runn/compare/v0.12.2...v0.13.0) (2022-03-28)

* Fix option [#38](https://github.com/k1LoW/runn/pull/38) ([k1LoW](https://github.com/k1LoW))
* Skip scheme://host:port validation [#37](https://github.com/k1LoW/runn/pull/37) ([k1LoW](https://github.com/k1LoW))
* Validate HTTP request and HTTP response using OpenAPI Spec(v3) [#36](https://github.com/k1LoW/runn/pull/36) ([k1LoW](https://github.com/k1LoW))

## [v0.12.2](https://github.com/k1LoW/runn/compare/v0.12.1...v0.12.2) (2022-03-23)

* Allow empty body [#35](https://github.com/k1LoW/runn/pull/35) ([k1LoW](https://github.com/k1LoW))

## [v0.12.1](https://github.com/k1LoW/runn/compare/v0.12.0...v0.12.1) (2022-03-20)

* Colorize [#34](https://github.com/k1LoW/runn/pull/34) ([k1LoW](https://github.com/k1LoW))

## [v0.12.0](https://github.com/k1LoW/runn/compare/v0.11.1...v0.12.0) (2022-03-18)

* Support `application/x-www-form-urlencoded` [#33](https://github.com/k1LoW/runn/pull/33) ([k1LoW](https://github.com/k1LoW))
* Support `text/plain` [#32](https://github.com/k1LoW/runn/pull/32) ([k1LoW](https://github.com/k1LoW))

## [v0.11.1](https://github.com/k1LoW/runn/compare/v0.11.0...v0.11.1) (2022-03-17)

* Fix hundling nested variables [#31](https://github.com/k1LoW/runn/pull/31) ([k1LoW](https://github.com/k1LoW))

## [v0.11.0](https://github.com/k1LoW/runn/compare/v0.10.1...v0.11.0) (2022-03-17)

* Add bind runner for binding variables [#30](https://github.com/k1LoW/runn/pull/30) ([k1LoW](https://github.com/k1LoW))

## [v0.10.1](https://github.com/k1LoW/runn/compare/v0.10.0...v0.10.1) (2022-03-17)

* Add option `--fail-fast` to `runn run` [#29](https://github.com/k1LoW/runn/pull/29) ([k1LoW](https://github.com/k1LoW))
* Add option FailFast(bool) to disable running additional tests after any test fails ( for RunN(ctx) ) [#28](https://github.com/k1LoW/runn/pull/28) ([k1LoW](https://github.com/k1LoW))

## [v0.10.0](https://github.com/k1LoW/runn/compare/v0.9.0...v0.10.0) (2022-03-16)

* Fix behavior when runn acts as a test helper [#27](https://github.com/k1LoW/runn/pull/27) ([k1LoW](https://github.com/k1LoW))
* Fix casting numeric [#26](https://github.com/k1LoW/runn/pull/26) ([k1LoW](https://github.com/k1LoW))
* Enabling debug prevents override [#25](https://github.com/k1LoW/runn/pull/25) ([k1LoW](https://github.com/k1LoW))

## [v0.9.0](https://github.com/k1LoW/runn/compare/v0.8.2...v0.9.0) (2022-03-16)

* Introduce `interval:` section to configure the running interval between steps. [#24](https://github.com/k1LoW/runn/pull/24) ([k1LoW](https://github.com/k1LoW))
* Fix debug message [#23](https://github.com/k1LoW/runn/pull/23) ([k1LoW](https://github.com/k1LoW))

## [v0.8.2](https://github.com/k1LoW/runn/compare/v0.8.1...v0.8.2) (2022-03-15)

* Fix DB Runner [#22](https://github.com/k1LoW/runn/pull/22) ([k1LoW](https://github.com/k1LoW))

## [v0.8.1](https://github.com/k1LoW/runn/compare/v0.8.0...v0.8.1) (2022-03-15)

* Add `string()` to expr [#21](https://github.com/k1LoW/runn/pull/21) ([k1LoW](https://github.com/k1LoW))

## [v0.8.0](https://github.com/k1LoW/runn/compare/v0.7.0...v0.8.0) (2022-03-15)

* Support named ( ordered map ) steps [#20](https://github.com/k1LoW/runn/pull/20) ([k1LoW](https://github.com/k1LoW))
* Fix test runner out [#19](https://github.com/k1LoW/runn/pull/19) ([k1LoW](https://github.com/k1LoW))

## [v0.7.0](https://github.com/k1LoW/runn/compare/v0.6.0...v0.7.0) (2022-03-14)

* Add exec runner [#18](https://github.com/k1LoW/runn/pull/18) ([k1LoW](https://github.com/k1LoW))

## [v0.6.0](https://github.com/k1LoW/runn/compare/v0.5.2...v0.6.0) (2022-03-12)

* Add HTTPRunnerWithHandler for using http.Handler instead of http.Client [#17](https://github.com/k1LoW/runn/pull/17) ([k1LoW](https://github.com/k1LoW))
* Detect duplicate runner names [#16](https://github.com/k1LoW/runn/pull/16) ([k1LoW](https://github.com/k1LoW))

## [v0.5.2](https://github.com/k1LoW/runn/compare/v0.5.1...v0.5.2) (2022-03-11)

* Trim string as a workaround [#15](https://github.com/k1LoW/runn/pull/15) ([k1LoW](https://github.com/k1LoW))

## [v0.5.1](https://github.com/k1LoW/runn/compare/v0.5.0...v0.5.1) (2022-03-11)

* Make RunN() also behave as a test helper [#14](https://github.com/k1LoW/runn/pull/14) ([k1LoW](https://github.com/k1LoW))

## [v0.5.0](https://github.com/k1LoW/runn/compare/v0.4.0...v0.5.0) (2022-03-11)

* Dump runner also increments steps [#13](https://github.com/k1LoW/runn/pull/13) ([k1LoW](https://github.com/k1LoW))
* Support `include:` in steps [#12](https://github.com/k1LoW/runn/pull/12) ([k1LoW](https://github.com/k1LoW))
* DB Runner support multiple queries [#11](https://github.com/k1LoW/runn/pull/11) ([k1LoW](https://github.com/k1LoW))

## [v0.4.0](https://github.com/k1LoW/runn/compare/v0.3.0...v0.4.0) (2022-03-09)

* Add Var for setting vars [#10](https://github.com/k1LoW/runn/pull/10) ([k1LoW](https://github.com/k1LoW))
* Add dump runner [#9](https://github.com/k1LoW/runn/pull/9) ([k1LoW](https://github.com/k1LoW))
* Support path pattern [#8](https://github.com/k1LoW/runn/pull/8) ([k1LoW](https://github.com/k1LoW))

## [v0.3.0](https://github.com/k1LoW/runn/compare/v0.2.0...v0.3.0) (2022-03-08)

* Add debug: for printing debug output [#7](https://github.com/k1LoW/runn/pull/7) ([k1LoW](https://github.com/k1LoW))

## [v0.2.0](https://github.com/k1LoW/runn/compare/v0.1.0...v0.2.0) (2022-03-08)

* Add RunN [#6](https://github.com/k1LoW/runn/pull/6) ([k1LoW](https://github.com/k1LoW))

## [v0.1.0](https://github.com/k1LoW/runn/compare/v0.0.1...v0.1.0) (2022-03-08)

* Improve error handling for parsing [#5](https://github.com/k1LoW/runn/pull/5) ([k1LoW](https://github.com/k1LoW))
* Fix: panic: assignment to entry in nil map [#4](https://github.com/k1LoW/runn/pull/4) ([k1LoW](https://github.com/k1LoW))

## [v0.0.1](https://github.com/k1LoW/runn/compare/868753734db7...v0.0.1) (2022-03-07)

* Support for running as test helper [#3](https://github.com/k1LoW/runn/pull/3) ([k1LoW](https://github.com/k1LoW))
* Add command `runn run` [#2](https://github.com/k1LoW/runn/pull/2) ([k1LoW](https://github.com/k1LoW))
* Add command `runn list` [#1](https://github.com/k1LoW/runn/pull/1) ([k1LoW](https://github.com/k1LoW))
