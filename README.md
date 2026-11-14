# eat 🦕

[![CI](https://github.com/shawn-bluce/eat/actions/workflows/ci.yml/badge.svg)](https://github.com/shawn-bluce/eat/actions/workflows/ci.yml)
[![Release](https://img.shields.io/github/v/release/shawn-bluce/eat?sort=semver)](https://github.com/shawn-bluce/eat/releases/latest)

**我是一只吃 CPU 和内存的怪兽。I'm a monster that eats CPU and memory.**

开发者经常需要快速占用一定量的 CPU 和内存（压测、验证告警和扩缩容、模拟资源紧张），`eat` 就是做这件事的小工具：一个没有依赖的二进制，支持 Linux、macOS 和 Windows，在容器里会按容器的 CPU 和内存限制计算。

Developers often need to occupy some CPU and memory quickly — load tests, checking alerts and autoscaling, simulating a busy machine. `eat` does just that: a single dependency-free binary for Linux, macOS and Windows that respects container CPU and memory limits.

## Install

从 [Releases](https://github.com/shawn-bluce/eat/releases/latest) 下载对应平台的二进制，或者：

Download a binary from [Releases](https://github.com/shawn-bluce/eat/releases/latest), or:

```shell
# os: linux | darwin, arch: amd64 | arm64
curl -fL -o eat https://github.com/shawn-bluce/eat/releases/latest/download/eat-linux-amd64
chmod +x eat
```

## Usage

```shell
eat -c 4                # eat 4 CPU cores
eat -c 35%              # eat 35% of all CPU cores
eat -m 4g               # eat 4GB memory (units: b, k, m, g)
eat -m 35%              # eat 35% of all memory
eat -c 2.5 -m 1.5g      # eat 2.5 CPU cores and 1.5GB memory
eat -c 100%             # eat all CPU cores
```

按 `Ctrl + C` 退出并释放资源。运行期间每 5 秒打印一次整机和 eat 自身的 CPU、内存占用。

Press `Ctrl + C` to stop and release everything. While running, eat logs the CPU and memory usage of the system and of itself every 5 seconds.

### Dynamic mode

加上 `-d/--dynamic` 后，`-c`/`-m` 表示**整机**要维持的使用率，而不是 eat 自己占多少：其他进程用得多，eat 就少占；用得少，eat 就多占。

With `-d/--dynamic`, `-c`/`-m` set the usage level of the **whole system** instead of a fixed amount for eat, and eat adjusts itself every second as other processes come and go.

```shell
eat -c 70% -m 60% -d    # keep the system at 70% CPU and 60% memory
eat -c 70% -d           # keep CPU at 70%, leave memory alone
eat -c 4 -m 8g -d       # absolute values are shares of the system too
```

- CPU 几秒内收敛到目标附近；其他进程已经超过目标时，eat 不再占用。
- 内存以 32MB 为单位增减，稳定在目标下方 32MB 以内，并始终保留 2%（至少 64MB）的空闲内存，避免触发 OOM。

### Containers

在有 CPU 或内存限制的容器里（`docker run --cpus 2 --memory 1g`、Kubernetes limits），百分比和动态模式都以容器的限制为准，支持 cgroup v1 和 v2。

Inside a container with CPU or memory limits, percentages and dynamic mode are relative to the container's limits (cgroup v1 and v2).

### Memory locking

在 Linux 和 macOS 上，eat 会尝试锁定内存，防止被系统换出到 swap。普通用户的锁定额度通常很小，锁定失败时 eat 会提示一次并继续运行；需要锁定时用 root 运行或调大 `ulimit -l`。Windows 不锁定。

On Linux and macOS, eat locks its memory so it is not swapped out. If the limit for normal users is too small, eat warns once and keeps running; run it as root or raise `ulimit -l` to lock. Windows does not lock memory.

## Build

```shell
make            # dist/eat for this system
make all        # dist/eat-<os>-<arch> for every platform
make test       # go vet + go test
```

## Release

每次推送到 `master` 都会在 Linux、macOS、Windows 上跑测试，通过后自动发布新版本：版本号在最新的 `vX.Y.Z` 上加一个补丁号，发布说明列出这次的提交，并附上所有平台的二进制和 `checksums.txt`。只改文档的提交不发布。要发 minor 或 major 版本，在 Actions 里手动运行 CI 并填写版本号（例如 `v1.0.0`）。

Every push to `master` is tested on Linux, macOS and Windows and then released automatically: the patch number of the latest `vX.Y.Z` is bumped, the notes list the new commits, and binaries for every platform plus `checksums.txt` are attached. Documentation-only pushes are not released. For a minor or major version, run the CI workflow manually with the version (e.g. `v1.0.0`).
