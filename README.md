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

> Tips:
> - Use `<Ctrl> + C` to stop program and release CPU and memory

# Build

```shell
# Linux
make linux-amd64 linux-arm64
# macOs
make darwin-amd64 darwin-arm64
# Windows
make windows-amd64 windows-arm64
```
