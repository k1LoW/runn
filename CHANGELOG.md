## [v0.99.0](https://github.com/k1LoW/runn/compare/v0.98.4...v0.99.0) - 2024-02-20
### Breaking Changes 🛠
- Set up donegroup and timeout for waiting cleanup processes after context canceled by @k1LoW in https://github.com/k1LoW/runn/pull/789
- Remove deprecated code by @k1LoW in https://github.com/k1LoW/runn/pull/790
- Support Custom runner (using Include runner) by @k1LoW in https://github.com/k1LoW/runn/pull/805
- Unexport operator#appendStep by @k1LoW in https://github.com/k1LoW/runn/pull/806
### New Features 🎉
- Support `gist://` scheme by @k1LoW in https://github.com/k1LoW/runn/pull/787
### Fix bug 🐛
- Fix error handling by @k1LoW in https://github.com/k1LoW/runn/pull/793
### Other Changes
- Merge run() function context and runner (chrome) context by @k1LoW in https://github.com/k1LoW/runn/pull/791
- Enable Dependabot by @k1LoW in https://github.com/k1LoW/runn/pull/794
- chore(deps): bump the dependencies group with 5 updates by @dependabot in https://github.com/k1LoW/runn/pull/795
- Update httpstub by @k1LoW in https://github.com/k1LoW/runn/pull/797
- Update grpcstub by @k1LoW in https://github.com/k1LoW/runn/pull/799
- Update pkgs  by @k1LoW in https://github.com/k1LoW/runn/pull/801
- chore(deps): bump the dependencies group with 19 updates by @dependabot in https://github.com/k1LoW/runn/pull/800
- Add link of runn Tutorial by @k1LoW in https://github.com/k1LoW/runn/pull/802
- Update expr by @k1LoW in https://github.com/k1LoW/runn/pull/807
- Ignore github.com/tenntenn/golden by @k1LoW in https://github.com/k1LoW/runn/pull/809
- Use errors.Join by @k1LoW in https://github.com/k1LoW/runn/pull/810

## [v0.98.4](https://github.com/k1LoW/runn/compare/v0.98.3...v0.98.4) - 2024-02-08
### New Features 🎉
- Set default output for Dump runner by @k1LoW in https://github.com/k1LoW/runn/pull/782
### Fix bug 🐛
- Fix concurrency by @k1LoW in https://github.com/k1LoW/runn/pull/786

## [v0.98.3](https://github.com/k1LoW/runn/compare/v0.98.2...v0.98.3) - 2024-02-01
### Fix bug 🐛
- fix: Fix potential panic by @ikawaha in https://github.com/k1LoW/runn/pull/781
### Other Changes
- Bump github.com/opencontainers/runc from 1.1.5 to 1.1.12 by @dependabot in https://github.com/k1LoW/runn/pull/779

## [v0.98.2](https://github.com/k1LoW/runn/compare/v0.98.1...v0.98.2) - 2024-01-30
### Fix bug 🐛
- Fix verbose index by @k1LoW in https://github.com/k1LoW/runn/pull/778
### Other Changes
- Changed octocov benchmark locale to en for easier viewing of results by @k2tzumi in https://github.com/k1LoW/runn/pull/776

## [v0.98.1](https://github.com/k1LoW/runn/compare/v0.98.0...v0.98.1) - 2024-01-25
### Other Changes
- Use octocov-action@v1 by @k1LoW in https://github.com/k1LoW/runn/pull/773
- Fix CD pipeline by @k1LoW in https://github.com/k1LoW/runn/pull/775

## [v0.98.0](https://github.com/k1LoW/runn/compare/v0.97.0...v0.98.0) - 2024-01-22
### New Features 🎉
- Support multiple concurrency keys by @k1LoW in https://github.com/k1LoW/runn/pull/772

## [v0.97.0](https://github.com/k1LoW/runn/compare/v0.96.0...v0.97.0) - 2024-01-21
### Breaking Changes 🛠
- Output results step by step when the `--verbose` option is enabled (Update `runn.Capturer` interface) by @k1LoW in https://github.com/k1LoW/runn/pull/766
### New Features 🎉
- Keep loaded OpenAPI documents by @k2tzumi in https://github.com/k1LoW/runn/pull/769
### Fix bug 🐛
- Reflect skipValidateRequest when http-openapi3 option is enabled by @k2tzumi in https://github.com/k1LoW/runn/pull/768

## [v0.96.0](https://github.com/k1LoW/runn/compare/v0.95.2...v0.96.0) - 2024-01-17
### Breaking Changes 🛠
- Delay connection of DB runner and SSH runner to target as long as possible. by @k1LoW in https://github.com/k1LoW/runn/pull/764
- Support host rules in SSH Runner by @k1LoW in https://github.com/k1LoW/runn/pull/763

## [v0.95.2](https://github.com/k1LoW/runn/compare/v0.95.1...v0.95.2) - 2024-01-11
### New Features 🎉
- Support host rules in DB Runner by @k1LoW in https://github.com/k1LoW/runn/pull/759

## [v0.95.1](https://github.com/k1LoW/runn/compare/v0.95.0...v0.95.1) - 2024-01-11
### Fix bug 🐛
- Use `--host-resolver-rules` instead of `--host-rules` in CDP Runner by @k1LoW in https://github.com/k1LoW/runn/pull/757

## [v0.95.0](https://github.com/k1LoW/runn/compare/v0.94.1...v0.95.0) - 2024-01-11
### New Features 🎉
- Support for `hostRules:` (control of hostname/IP mapping) in HTTP runner, gRPC runner and CDP runner by @k1LoW in https://github.com/k1LoW/runn/pull/754
- Add `--host-rules` option to loadt and run commands. by @k1LoW in https://github.com/k1LoW/runn/pull/756

## [v0.94.1](https://github.com/k1LoW/runn/compare/v0.94.0...v0.94.1) - 2024-01-10
### Fix bug 🐛
- uint64 that came across from the include runner can also be used for loop count. by @k1LoW in https://github.com/k1LoW/runn/pull/752
- Trim unnecessary CR and LF in queries by @k1LoW in https://github.com/k1LoW/runn/pull/753
### Other Changes
- Bump github.com/cloudflare/circl from 1.3.3 to 1.3.7 by @dependabot in https://github.com/k1LoW/runn/pull/749

## [v0.94.0](https://github.com/k1LoW/runn/compare/v0.93.0...v0.94.0) - 2024-01-05
### Breaking Changes 🛠
- RE Disable profile by default by @k1LoW in https://github.com/k1LoW/runn/pull/739

## [v0.93.0](https://github.com/k1LoW/runn/compare/v0.92.0...v0.93.0) - 2024-01-05
### Breaking Changes 🛠
- Update benchmarks by @k1LoW in https://github.com/k1LoW/runn/pull/734
### New Features 🎉
- Support using YAML's anchors and aliases in runbooks by @h6ah4i in https://github.com/k1LoW/runn/pull/722
- Specify the header name for the trace header by @k2tzumi in https://github.com/k1LoW/runn/pull/742
### Fix bug 🐛
- Fix detectRunbookAreas() failure when the runbook has only 1 step in maps format by @h6ah4i in https://github.com/k1LoW/runn/pull/724
- Revert inexplicable Unmarshal by @k1LoW in https://github.com/k1LoW/runn/pull/738
### Other Changes
- Stop unnecessary pointer passing by @k2tzumi in https://github.com/k1LoW/runn/pull/736
- Change indentation options for flattening Yaml aliases by @k2tzumi in https://github.com/k1LoW/runn/pull/737
- Update stopw by @k1LoW in https://github.com/k1LoW/runn/pull/740
- Reduce the number of execution of o.Result() by @k1LoW in https://github.com/k1LoW/runn/pull/741

## [v0.92.0](https://github.com/k1LoW/runn/compare/v0.91.4...v0.92.0) - 2023-12-30
### Breaking Changes 🛠
- Profile is default enabled and provides a way to disable it by @k1LoW in https://github.com/k1LoW/runn/pull/716
### New Features 🎉
- Introduce pick() expr built-in function by @h6ah4i in https://github.com/k1LoW/runn/pull/714
- Introduce omit() expr built-in function by @h6ah4i in https://github.com/k1LoW/runn/pull/719
### Other Changes
- Fix disable profile by @k2tzumi in https://github.com/k1LoW/runn/pull/713
- Set up benchmark by @k1LoW in https://github.com/k1LoW/runn/pull/718
- Update expr (change org) by @k1LoW in https://github.com/k1LoW/runn/pull/721
- Introduce merge() expr built-in function by @h6ah4i in https://github.com/k1LoW/runn/pull/720
- Tuning slice/map allocations by @k2tzumi in https://github.com/k1LoW/runn/pull/723

## [v0.91.4](https://github.com/k1LoW/runn/compare/v0.91.3...v0.91.4) - 2023-12-24
### Fix bug 🐛
- Add workaround for go-sql-spanner internal connection sharing by @h6ah4i in https://github.com/k1LoW/runn/pull/711

## [v0.91.3](https://github.com/k1LoW/runn/compare/v0.91.2...v0.91.3) - 2023-12-19
### Other Changes
- Bump golang.org/x/crypto from 0.14.0 to 0.17.0 by @dependabot in https://github.com/k1LoW/runn/pull/707

## [v0.91.2](https://github.com/k1LoW/runn/compare/v0.91.1...v0.91.2) - 2023-12-11
### Fix bug 🐛
- Fixe a bug that caused other values to change when bind. by @k1LoW in https://github.com/k1LoW/runn/pull/706

## [v0.91.1](https://github.com/k1LoW/runn/compare/v0.91.0...v0.91.1) - 2023-12-01
### Fix bug 🐛
- Implicitly enable scope `run:parent` if `--and-run` is enabled by @k1LoW in https://github.com/k1LoW/runn/pull/702

## [v0.91.0](https://github.com/k1LoW/runn/compare/v0.90.4...v0.91.0) - 2023-11-28
### Breaking Changes 🛠
- Support to bind to slice values by @k1LoW in https://github.com/k1LoW/runn/pull/700
- Support for bind to slice/map values by @k1LoW in https://github.com/k1LoW/runn/pull/699
### Fix bug 🐛
- Fix checking reverved store keys by @k1LoW in https://github.com/k1LoW/runn/pull/698

## [v0.90.4](https://github.com/k1LoW/runn/compare/v0.90.3...v0.90.4) - 2023-11-22
### Fix bug 🐛
- If it is not a file, do not raise an error even if the value is long. by @k1LoW in https://github.com/k1LoW/runn/pull/695

## [v0.90.3](https://github.com/k1LoW/runn/compare/v0.90.2...v0.90.3) - 2023-11-20
### Fix bug 🐛
- Fix execution timing of functions that retrieve results. by @k1LoW in https://github.com/k1LoW/runn/pull/693
### Other Changes
- More validate `labels:` by @k1LoW in https://github.com/k1LoW/runn/pull/691

## [v0.90.2](https://github.com/k1LoW/runn/compare/v0.90.1...v0.90.2) - 2023-11-13
### New Features 🎉
- Validate runbook `labels:` by @k1LoW in https://github.com/k1LoW/runn/pull/688
### Fix bug 🐛
- Allow hyphened labels by @k1LoW in https://github.com/k1LoW/runn/pull/690

## [v0.90.1](https://github.com/k1LoW/runn/compare/v0.90.0...v0.90.1) - 2023-11-12
### Fix bug 🐛
- Fix `--label` by @k1LoW in https://github.com/k1LoW/runn/pull/686

## [v0.90.0](https://github.com/k1LoW/runn/compare/v0.89.1...v0.90.0) - 2023-11-11
### Breaking Changes 🛠
- Record trail for each loop. by @k1LoW in https://github.com/k1LoW/runn/pull/676
- Profile is turned on by default. by @k1LoW in https://github.com/k1LoW/runn/pull/677
- Add runbook ID (Full) and elapsed time to result.json by @k1LoW in https://github.com/k1LoW/runn/pull/678
### New Features 🎉
- Support multiple ids specification ( Update `--id` and `RUNN_ID` ) by @k1LoW in https://github.com/k1LoW/runn/pull/679
- Support acceptance of runbook IDs from STDIN by @k1LoW in https://github.com/k1LoW/runn/pull/681
- Support `labels:` section in runbooks by @k1LoW in https://github.com/k1LoW/runn/pull/683
### Fix bug 🐛
- Fix bug that remote files are not readable in `include.path:` by @k1LoW in https://github.com/k1LoW/runn/pull/675
- Fix sorting by ids by @k1LoW in https://github.com/k1LoW/runn/pull/684
### Other Changes
- Fixed a bug with subdomain enabled cookies by @k2tzumi in https://github.com/k1LoW/runn/pull/673
- Integrate runbookID and runbookIDFull by @k1LoW in https://github.com/k1LoW/runn/pull/680
- Add design doc for runbook by @k1LoW in https://github.com/k1LoW/runn/pull/682

## [v0.89.1](https://github.com/k1LoW/runn/compare/v0.89.0...v0.89.1) - 2023-11-03
### Other Changes
- Support gzip response by @k2tzumi in https://github.com/k1LoW/runn/pull/671

## [v0.89.0](https://github.com/k1LoW/runn/compare/v0.88.0...v0.89.0) - 2023-11-01
### Breaking Changes 🛠
- Error if the file does not exist. by @k1LoW in https://github.com/k1LoW/runn/pull/662
- Fix single value expand by @k1LoW in https://github.com/k1LoW/runn/pull/664
- Add the concept of scope. by @k1LoW in https://github.com/k1LoW/runn/pull/663
- Add `run:exec` scope. and deny by default. by @k1LoW in https://github.com/k1LoW/runn/pull/667
### New Features 🎉
- Support wildcards in `json://` and `yaml://` so that values can be retrieved with slice by @k1LoW in https://github.com/k1LoW/runn/pull/545
### Other Changes
- Listing variable expansion patterns in the Include runner. by @k1LoW in https://github.com/k1LoW/runn/pull/661
- Bump github.com/docker/docker from 20.10.24+incompatible to 24.0.7+incompatible by @dependabot in https://github.com/k1LoW/runn/pull/665
- Refactor generating runbook ID by @k1LoW in https://github.com/k1LoW/runn/pull/666
- Add prefix to random ID to make it clear that it is randomly generated by @k1LoW in https://github.com/k1LoW/runn/pull/668

## [v0.88.0](https://github.com/k1LoW/runn/compare/v0.87.0...v0.88.0) - 2023-10-26
### Breaking Changes 🛠
- Support comment for tracing query by @k1LoW in https://github.com/k1LoW/runn/pull/654
### New Features 🎉
- Support header for tracing gRPC requests by @k1LoW in https://github.com/k1LoW/runn/pull/656
- Support `trace:` for tracing by @k1LoW in https://github.com/k1LoW/runn/pull/658
### Fix bug 🐛
- Fix handling of gRPC headers by @k1LoW in https://github.com/k1LoW/runn/pull/655
- Fix handling of http headers by @k1LoW in https://github.com/k1LoW/runn/pull/657

## [v0.87.0](https://github.com/k1LoW/runn/compare/v0.86.0...v0.87.0) - 2023-10-25
### Breaking Changes 🛠
- Add runbookID() for returning the runbook ID of the root runbook. by @k1LoW in https://github.com/k1LoW/runn/pull/646
- Pass `*step` at the time of running the runner. by @k1LoW in https://github.com/k1LoW/runn/pull/651
### New Features 🎉
- Add header for trace by @k2tzumi in https://github.com/k1LoW/runn/pull/645
### Other Changes
- Remove test using external site by @k1LoW in https://github.com/k1LoW/runn/pull/652

## [v0.86.0](https://github.com/k1LoW/runn/compare/v0.85.1...v0.86.0) - 2023-10-19
### Breaking Changes 🛠
- Enable to specify OpenAPI3 documents for each HTTP runner. by @k1LoW in https://github.com/k1LoW/runn/pull/647
- Enable to specify protos for each gRPC runner. by @k1LoW in https://github.com/k1LoW/runn/pull/649

## [v0.85.1](https://github.com/k1LoW/runn/compare/v0.85.0...v0.85.1) - 2023-10-13
### Fix bug 🐛
- If coverage cannot be collected, report an error. by @k1LoW in https://github.com/k1LoW/runn/pull/642
- Sort specs of coverage by @k1LoW in https://github.com/k1LoW/runn/pull/643
### Other Changes
- Update README with interval option by @IzumiSy in https://github.com/k1LoW/runn/pull/640
- Update toolchain version by @k1LoW in https://github.com/k1LoW/runn/pull/644

## [v0.85.0](https://github.com/k1LoW/runn/compare/v0.84.2...v0.85.0) - 2023-10-12
### Breaking Changes 🛠
- Negative numbers are now passed as int type as they are. by @k1LoW in https://github.com/k1LoW/runn/pull/637
### Other Changes
- Update pkgs by @k1LoW in https://github.com/k1LoW/runn/pull/639

## [v0.84.2](https://github.com/k1LoW/runn/compare/v0.84.1...v0.84.2) - 2023-10-12
### Fix bug 🐛
- Fix buildTree (support negative numbers) by @k1LoW in https://github.com/k1LoW/runn/pull/635
### Other Changes
- docs: add the installation guide with aqua by @suzuki-shunsuke in https://github.com/k1LoW/runn/pull/632
- Fix gostyle repetition by @k2tzumi in https://github.com/k1LoW/runn/pull/634
- Bump golang.org/x/net from 0.14.0 to 0.17.0 by @dependabot in https://github.com/k1LoW/runn/pull/636

## [v0.84.1](https://github.com/k1LoW/runn/compare/v0.84.0...v0.84.1) - 2023-10-08
### New Features 🎉
- Add `--http-openapi3` opiton to set the path to the OpenAPI v3 document for all HTTP runners by @k1LoW in https://github.com/k1LoW/runn/pull/626
- `coverage` command support JSON output by @k1LoW in https://github.com/k1LoW/runn/pull/629
- Show total coverage by @k1LoW in https://github.com/k1LoW/runn/pull/630
### Fix bug 🐛
- Consider the path of `servers:` in OpenAPI Spec and endpoint of HTTP runners for coverage path resolution. by @k1LoW in https://github.com/k1LoW/runn/pull/627

## [v0.84.0](https://github.com/k1LoW/runn/compare/v0.83.1...v0.84.0) - 2023-10-08
### Breaking Changes 🛠
- Support `exec.shell:` for specifying the shell to use by @k1LoW in https://github.com/k1LoW/runn/pull/622
### New Features 🎉
- Add `coverage` command for showing coverage for paths/operations of OpenAPI spec and methods of protocol buffers. by @k1LoW in https://github.com/k1LoW/runn/pull/625

## [v0.83.1](https://github.com/k1LoW/runn/compare/v0.83.0...v0.83.1) - 2023-09-23
### New Features 🎉
- Support application/octet-stream via http by @DaichiUeura in https://github.com/k1LoW/runn/pull/617

## [v0.83.0](https://github.com/k1LoW/runn/compare/v0.82.0...v0.83.0) - 2023-09-21
### Breaking Changes 🛠
- Add gostyle-action by @k1LoW in https://github.com/k1LoW/runn/pull/613
- Fix resolvePaths() to handle relative paths correctly by @k1LoW in https://github.com/k1LoW/runn/pull/616
### Other Changes
- Update grpcstub to v0.13.0 by @k1LoW in https://github.com/k1LoW/runn/pull/615

## [v0.82.0](https://github.com/k1LoW/runn/compare/v0.81.1...v0.82.0) - 2023-08-31
### Breaking Changes 🛠
- Fix DB Runner connections are not closed properly by @k1LoW in https://github.com/k1LoW/runn/pull/607
- Disconnect from gRPC server for each scenario only for gRCP Runner with target. by @k1LoW in https://github.com/k1LoW/runn/pull/608
### Other Changes
- Update bufbuild/protocompile by @k1LoW in https://github.com/k1LoW/runn/pull/604

## [v0.81.1](https://github.com/k1LoW/runn/compare/v0.81.0...v0.81.1) - 2023-08-23
### New Features 🎉
- Enable to get error output for each result by @k1LoW in https://github.com/k1LoW/runn/pull/602

## [v0.81.0](https://github.com/k1LoW/runn/compare/v0.80.3...v0.81.0) - 2023-08-22
### Breaking Changes 🛠
- Update ryo-yamaoka/otchkiss and fix `runn loadt` by @k1LoW in https://github.com/k1LoW/runn/pull/595
- Use bufbuild/protocompile instead of jhump/protoreflect by @k1LoW in https://github.com/k1LoW/runn/pull/597
- Use github.com/jhump/protoreflect/v2/grpcreflect by @k1LoW in https://github.com/k1LoW/runn/pull/598
- Bump up expr version to v1.14.0 by @k1LoW in https://github.com/k1LoW/runn/pull/599
- Migrate to built-in functions of expr by @k1LoW in https://github.com/k1LoW/runn/pull/600
- Update pkgs by @k1LoW in https://github.com/k1LoW/runn/pull/601

## [v0.80.3](https://github.com/k1LoW/runn/compare/v0.80.2...v0.80.3) - 2023-08-14
### Fix bug 🐛
- Fix escape failure if application/json response data value contains a single backslash by @k1LoW in https://github.com/k1LoW/runn/pull/593

## [v0.80.2](https://github.com/k1LoW/runn/compare/v0.80.1...v0.80.2) - 2023-08-10
### Fix bug 🐛
- Fix pickStepYAML logic by @k1LoW in https://github.com/k1LoW/runn/pull/591

## [v0.80.1](https://github.com/k1LoW/runn/compare/v0.80.0...v0.80.1) - 2023-08-09
### New Features 🎉
- Add `--run` option for selecting runbooks by @k1LoW in https://github.com/k1LoW/runn/pull/589
- Add `--run` and `--id` option to `runn loadt` and `runn list` command. by @k1LoW in https://github.com/k1LoW/runn/pull/590
### Fix bug 🐛
- Fix gRPC request message being pre-expanded values by @k1LoW in https://github.com/k1LoW/runn/pull/587

## [v0.80.0](https://github.com/k1LoW/runn/compare/v0.79.0...v0.80.0) - 2023-08-08
### Breaking Changes 🛠
- Bind runner support any type value ( not only string ). by @k1LoW in https://github.com/k1LoW/runn/pull/586
### Other Changes
- Fix condition tree to make it easier to read. by @k1LoW in https://github.com/k1LoW/runn/pull/583
- Improve `--verbose` option output by @k1LoW in https://github.com/k1LoW/runn/pull/585

## [v0.79.0](https://github.com/k1LoW/runn/compare/v0.78.1...v0.79.0) - 2023-08-06
### Breaking Changes 🛠
- Include in RunResult the execution results of the runbooks loaded by the include runner. by @k1LoW in https://github.com/k1LoW/runn/pull/580
- Improve error message when `runn run` fails by @k1LoW in https://github.com/k1LoW/runn/pull/581
### Other Changes
- Fix TestSSHPortFowarding flaky test by @k1LoW in https://github.com/k1LoW/runn/pull/578
- Aggregate function to colorize by @k1LoW in https://github.com/k1LoW/runn/pull/582

## [v0.78.1](https://github.com/k1LoW/runn/compare/v0.78.0...v0.78.1) - 2023-08-03
### Breaking Changes 🛠
- In runn as Go package, change the test name for each of the runbooks when they are run. by @k1LoW in https://github.com/k1LoW/runn/pull/576
### Fix bug 🐛
- Fix a problem that functions with uint type arguments can't be used in runbook by @n3xem in https://github.com/k1LoW/runn/pull/575

## [v0.78.0](https://github.com/k1LoW/runn/compare/v0.77.0...v0.78.0) - 2023-07-27
### Breaking Changes 🛠
- Add design doc for runbook ID by @k1LoW in https://github.com/k1LoW/runn/pull/558
- Rename ID to Trail by @k1LoW in https://github.com/k1LoW/runn/pull/570
- Generate IDs using file path of runbooks. by @k1LoW in https://github.com/k1LoW/runn/pull/571
- Set runnbook ID to result (STDOUT/JSON) by @k1LoW in https://github.com/k1LoW/runn/pull/574
### New Features 🎉
- Add methods to specify runbook ID by @k1LoW in https://github.com/k1LoW/runn/pull/572
- Add show ids in `runn list` command by @k1LoW in https://github.com/k1LoW/runn/pull/573

## [v0.77.0](https://github.com/k1LoW/runn/compare/v0.76.2...v0.77.0) - 2023-07-23
### New Features 🎉
- Append use cookie option by @k2tzumi in https://github.com/k1LoW/runn/pull/559
### Other Changes
- fix typo by @okazaki-kk in https://github.com/k1LoW/runn/pull/566

## [v0.76.2](https://github.com/k1LoW/runn/compare/v0.76.1...v0.76.2) - 2023-07-21
### Fix bug 🐛
- Only for string, expand gRPC messages by @k1LoW in https://github.com/k1LoW/runn/pull/564

## [v0.76.1](https://github.com/k1LoW/runn/compare/v0.76.0...v0.76.1) - 2023-07-18
### Fix bug 🐛
- Fix values of stepMapKeys that were not being deleted in the loop. by @k1LoW in https://github.com/k1LoW/runn/pull/560
### Other Changes
- Guarantee sequential run for each operator by @k1LoW in https://github.com/k1LoW/runn/pull/562

## [v0.76.0](https://github.com/k1LoW/runn/compare/v0.75.3...v0.76.0) - 2023-07-17
### Breaking Changes 🛠
- Allow `.yml` in YAML extensions and add extension checking by @k1LoW in https://github.com/k1LoW/runn/pull/544
- Fix path.go by @k1LoW in https://github.com/k1LoW/runn/pull/550
- Fix resolving protos and importpaths by @k1LoW in https://github.com/k1LoW/runn/pull/551
### New Features 🎉
- Enable configurable http runner timeout by @k2tzumi in https://github.com/k1LoW/runn/pull/547
- Keep cookies in store by @k2tzumi in https://github.com/k1LoW/runn/pull/556
### Fix bug 🐛
- Fix a bug that prevented requests with multipart/form-data when the value is numeric by @k1LoW in https://github.com/k1LoW/runn/pull/553
### Other Changes
- Fix build settings for release by @k1LoW in https://github.com/k1LoW/runn/pull/541
- Change `interface{}` to `any` by @k1LoW in https://github.com/k1LoW/runn/pull/543
- Run govulncheck by @k2tzumi in https://github.com/k1LoW/runn/pull/554

## [v0.75.3](https://github.com/k1LoW/runn/compare/v0.75.2...v0.75.3) - 2023-06-11
### New Features 🎉
- Add built-in functions for JSON. by @k1LoW in https://github.com/k1LoW/runn/pull/539

## [v0.75.2](https://github.com/k1LoW/runn/compare/v0.75.1...v0.75.2) - 2023-06-11
### Fix bug 🐛
- Revert "Fix logic to skip validation using `servers:` section in OpenAPI v3 spec" by @k1LoW in https://github.com/k1LoW/runn/pull/538
### Other Changes
- Add runbook for local port forwarding with OpenAPI3 by @k1LoW in https://github.com/k1LoW/runn/pull/535
- Friendly find route error message. by @k1LoW in https://github.com/k1LoW/runn/pull/537

## [v0.75.1](https://github.com/k1LoW/runn/compare/v0.75.0...v0.75.1) - 2023-06-09
### Other Changes
- Fix logic to skip validation using `servers:` section in OpenAPI v3 spec by @k1LoW in https://github.com/k1LoW/runn/pull/533

## [v0.75.0](https://github.com/k1LoW/runn/compare/v0.74.1...v0.75.0) - 2023-06-08
### New Features 🎉
- Support running gRPC runner without Server reflection using .proto files by @k1LoW in https://github.com/k1LoW/runn/pull/529
- Set environment variables to be extracted at each step in a variable as `env`. by @k1LoW in https://github.com/k1LoW/runn/pull/530

## [v0.74.1](https://github.com/k1LoW/runn/compare/v0.74.0...v0.74.1) - 2023-06-07
### Other Changes
- Update pkgs by @k1LoW in https://github.com/k1LoW/runn/pull/524
- Add `skipVerify:` section / HTTPSkipVerify option for HTTP Runner by @k1LoW in https://github.com/k1LoW/runn/pull/527

## [v0.74.0](https://github.com/k1LoW/runn/compare/v0.73.0...v0.74.0) - 2023-06-01
### Breaking Changes 🛠
- Fix for marshaling gRPC responses. by @k1LoW in https://github.com/k1LoW/runn/pull/521

## [v0.73.0](https://github.com/k1LoW/runn/compare/v0.72.0...v0.73.0) - 2023-05-30
### Breaking Changes 🛠
- Set gRPC error message to store by @k1LoW in https://github.com/k1LoW/runn/pull/517
- Change CaptureGRPCResponseStatus signature by @k1LoW in https://github.com/k1LoW/runn/pull/519
### New Features 🎉
- Add `faker.*` to built-in functions for generating random data by @k1LoW in https://github.com/k1LoW/runn/pull/516

## [v0.72.0](https://github.com/k1LoW/runn/compare/v0.71.0...v0.72.0) - 2023-05-17
### New Features 🎉
- Support yaml:// scheme for vars by @IzumiSy in https://github.com/k1LoW/runn/pull/508
- Add `timeout:` section for gRPC Runner by @k1LoW in https://github.com/k1LoW/runn/pull/511
### Other Changes
- Enable errorlint by @k1LoW in https://github.com/k1LoW/runn/pull/510

## [v0.71.0](https://github.com/k1LoW/runn/compare/v0.70.1...v0.71.0) - 2023-05-16
### Breaking Changes 🛠
- Use google.golang.org/protobuf/reflect/protoreflect instead of github.com/jhump/protoreflect by @k1LoW in https://github.com/k1LoW/runn/pull/506

## [v0.70.1](https://github.com/k1LoW/runn/compare/v0.70.0...v0.70.1) - 2023-05-16
### Other Changes
- Fix integration test by @k1LoW in https://github.com/k1LoW/runn/pull/503
- First try the lightweight service descripter acquisition process by @k1LoW in https://github.com/k1LoW/runn/pull/505

## [v0.70.0](https://github.com/k1LoW/runn/compare/v0.69.1...v0.70.0) - 2023-05-10
### Breaking Changes 🛠
- Support for JSON type columns by @xande0812 in https://github.com/k1LoW/runn/pull/499

## [v0.70.0](https://github.com/k1LoW/runn/compare/v0.69.1...v0.70.0) - 2023-05-10
### Breaking Changes 🛠
- Support for JSON type columns by @xande0812 in https://github.com/k1LoW/runn/pull/499

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
