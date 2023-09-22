# Changelog

## [1.6.0](https://github.com/soerenschneider/conditional-reboot/compare/v1.5.1...v1.6.0) (2023-09-22)


### Features

* add safeguards to not get stuck in reboot loop ([ee4ac7d](https://github.com/soerenschneider/conditional-reboot/commit/ee4ac7dd5a918aaa320b029f830d2dd919f2a7be))
* make needrestart checker configurable ([ae38700](https://github.com/soerenschneider/conditional-reboot/commit/ae38700b870992e9469d6fd6eb6e10b0443bf100))


### Bug Fixes

* actually honor 'safe for reboot' condition ([2858d90](https://github.com/soerenschneider/conditional-reboot/commit/2858d900d9fbe3933dd61f75234e20f9f84c044d))
* fix linux code ([7fec818](https://github.com/soerenschneider/conditional-reboot/commit/7fec81823399ce603e9ad0bd70fe52003c04d988))

## [1.5.1](https://github.com/soerenschneider/conditional-reboot/compare/v1.5.0...v1.5.1) (2023-09-06)


### Bug Fixes

* fix memory aliasing bug ([8483897](https://github.com/soerenschneider/conditional-reboot/commit/8483897ed536a862e28398c77eb9e59566cbc30a))
* make sure to always run needrestart as root ([fc55bce](https://github.com/soerenschneider/conditional-reboot/commit/fc55bce8455a62293b0d4aacada38bba16887983))

## [1.5.0](https://github.com/soerenschneider/conditional-reboot/compare/v1.4.2...v1.5.0) (2023-09-05)


### Features

* use 'HH:MM' timestamps for precondition 'time window' ([286f2c5](https://github.com/soerenschneider/conditional-reboot/commit/286f2c5589acf4b9d0a8f28af59782abcc00fa24))

## [1.4.2](https://github.com/soerenschneider/conditional-reboot/compare/v1.4.1...v1.4.2) (2023-08-05)


### Bug Fixes

* fix potential deadlock ([2e05765](https://github.com/soerenschneider/conditional-reboot/commit/2e057657b251158ad3c1b28a9f8713ef0e466bbf))

## [1.4.1](https://github.com/soerenschneider/conditional-reboot/compare/v1.4.0...v1.4.1) (2023-08-02)


### Bug Fixes

* fix all configs sharing the same reference ([2bda2c5](https://github.com/soerenschneider/conditional-reboot/commit/2bda2c5cbcb7267cb40dbef749ff5a8b176edd06))
* respect streak ([1b97738](https://github.com/soerenschneider/conditional-reboot/commit/1b97738e9e24b53ec92a8172749bf6830222a08b))

## [1.4.0](https://github.com/soerenschneider/conditional-reboot/compare/v1.3.0...v1.4.0) (2023-06-23)


### Features

* add heartbeat metric ([c2ac1fb](https://github.com/soerenschneider/conditional-reboot/commit/c2ac1fb5bc3dd4c836ca108bfdd159bc06b445a9))

## [1.3.0](https://github.com/soerenschneider/conditional-reboot/compare/v1.2.2...v1.3.0) (2023-06-22)


### Features

* add version metric ([e152a66](https://github.com/soerenschneider/conditional-reboot/commit/e152a6656c10706ac5a00f5cf891985562fc0e6c))

## [1.2.2](https://github.com/soerenschneider/conditional-reboot/compare/v1.2.1...v1.2.2) (2023-06-22)


### Bug Fixes

* set metric when actually performing check ([4753755](https://github.com/soerenschneider/conditional-reboot/commit/47537558d0ce881982a0d58b3eae4a8d227a497b))

## [1.2.1](https://github.com/soerenschneider/conditional-reboot/compare/v1.2.0...v1.2.1) (2023-06-21)


### Bug Fixes

* run metrics dumper in own goroutine ([651e612](https://github.com/soerenschneider/conditional-reboot/commit/651e612538031ef61def7168fd96c34d6ec0b0c1))

## [1.2.0](https://github.com/soerenschneider/conditional-reboot/compare/v1.1.0...v1.2.0) (2023-06-21)


### Features

* allow dumping of metrics to be picked up by node_exporter ([c9e3b64](https://github.com/soerenschneider/conditional-reboot/commit/c9e3b64ff65d59acc68475baaea4a182507df073))


### Bug Fixes

* invoke metric on startup ([10d7067](https://github.com/soerenschneider/conditional-reboot/commit/10d7067bddfa2d1e000e4e8f04b9433f84669e8d))

## [1.1.0](https://github.com/soerenschneider/conditional-reboot/compare/v1.0.0...v1.1.0) (2023-06-21)


### Features

* add icmp checker ([22aac53](https://github.com/soerenschneider/conditional-reboot/commit/22aac5324ae03f46f54cb553cb5b4198c515f62a))
* automatically decide how to reboot ([36b4f7e](https://github.com/soerenschneider/conditional-reboot/commit/36b4f7e358f0bc388f15bc986c39abfd2a77d117))
* log errors in state ([c3ad779](https://github.com/soerenschneider/conditional-reboot/commit/c3ad7795397eece41e808ddc56efb6abb253de6f))

## 1.0.0 (2023-06-15)


### Miscellaneous Chores

* release 1.0.0 ([572aa5a](https://github.com/soerenschneider/conditional-reboot/commit/572aa5a183c5f25eedaa85ab1c267ff27e101ead))
