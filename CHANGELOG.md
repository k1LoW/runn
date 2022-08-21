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
