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
