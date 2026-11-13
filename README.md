# Intro
<b>我是一只吃CPU和内存的怪兽🦕</b>

<b>I'm a monster that eats CPU and memory🦕</b>

开发者们经常会遇到需要快速占用 CPU 和内存的需求，我也是。所以我开发了一个名为 `eat` 的小工具来快速占用指定数量的 CPU 和内存。

Developer often need to quickly occupy CPU and memory, so do I. Therefore, I developed a small tool named `eat` to quickly occupy the specified number of CPU and memory.

# Usage

```shell
eat -c 4	# eating 4 CPU cores
eat -c 35%	# eating 35% CPU cores (CPU cores * 35%)
eat -c 100%	# eating all CPU cores
eat -m 4g	# eating 4GB memory
eat -m 20m	# eating 20MB memory
eat -m 35%	# eating 35% memory (total memory * 35%)
eat -m 100%	# eating all memory
eat -c 2.5 -m 1.5g	# eating 2.5 CPU cores and 1.5GB memory
eat -c 3 -m 200m	# eating 3 CPU cores and 200MB memory
eat -c 100% -m 100%	# eating all CPU cores and memory
```

## Dynamic mode / 动态模式

加上 `-d/--dynamic` 后，`-c`/`-m` 表示整机要维持的使用率，而不是 eat 自己占多少。其他进程用得多，eat 就少占；用得少，eat 就多占；其他进程已经超过目标时 eat 不再占用。

With `-d/--dynamic`, `-c`/`-m` set the usage level of the whole machine instead of a fixed amount for eat. eat adjusts itself every second as other processes come and go.

```shell
eat -c 70% -m 60% -d	# keep the whole machine at 70% CPU and 60% memory
eat -c 70% -d		# keep CPU at 70%, leave memory alone
eat -c 4 -m 8g -d	# absolute values are converted to shares of the machine
```

> - CPU 每秒采样一次，几秒内收敛到目标附近。
> - 内存以 32MB 为单位增减，稳定在目标值下方一个单位以内；始终保留至少 2%（不少于 64MB）的空闲内存，避免触发 OOM。
> - 在 Linux cgroup（容器）里有内存上限时，按 cgroup 的上限和用量计算。

> Tips:
> - Use `<Ctrl> + C` to stop program and release CPU and memory
> - While running, eat logs the CPU and memory usage of the whole system and of itself every 5 seconds

# Build

```shell
# Linux
make linux-amd64 linux-arm64
# macOs
make darwin-amd64 darwin-arm64
# Windows
make windows-amd64 windows-arm64
```

# Release

推送 `v` 开头的 tag 即触发 GitHub Actions 构建并发布 Release；也可以在 Actions 页面手动运行 Release 并填写要创建的 tag。

Push a tag starting with `v` to build and publish a GitHub Release, or run the Release workflow manually with the tag to create.

```shell
git tag v0.2.0
git push origin v0.2.0
```
