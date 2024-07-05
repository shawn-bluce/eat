# Introduce
<b>I'm a CPU and memory eating monster🦕</b>

Developer will encounter the need to quickly occupy CPU and memory, I am also deeply troubled, so I developed a tool named `eat` to help you quickly occupy a specified amount of CPU and memory.

# Todo

- [x] Support `eat -c 35%` and `eat -m 35%`
- [x] support gracefully exit: capture process signal SIGINT(2), SIGTERM(15)
- [x] support deadline: `-t` specify the duration eat progress. such as "300ms", "1.5h", "2h45m". (unit: "ns", "us" (or "µs"), "ms", "s", "m", "h")
- [] CPU Affinity
- [] Memory read/write, prevent memory from being swapped out
- [] Dynamic adjustment of CPU and memory usage
- [] Eat GPU

# Usage

```shell
eat -c 4            # eating 4 CPU core
eat -c 35%          # eating 35% CPU core (CPU count * 35%)
eat -c 100%         # eating all CPU core
eat -m 4g           # eating 4GB memory
eat -m 20m          # eating 20MB memory
eat -m 35%          # eating 35% memory (total memory * 35%)
eat -m 100%         # eating all memory
eat -c 2.5 -m 1.5g  # eating 2.5 CPU core and 1.5GB memory
eat -c 3 -m 200m    # eating 3 CPU core and 200MB memory
eat -c 100% -m 100% # eating all CPU core and memory
eat -c 100% -t 1h # eating all CPU core and quit after 1hour
```

> Tips:
> - Using \<Ctrl\> + C to stop eating and release CPU and memory

# Build

```shell
go build -o eat
```

# 介绍
<b>我是一个吃CPU和内存的怪兽🦕</b>

开发者们经常会遇到需要快速占用 CPU 和内存的需求，我也是。所以我开发了一个名为 `eat` 的小工具来快速占用指定数量的 CPU 和内存。

# 待办

- [x] 支持`eat -c 35%`和`eat -m 35%`
- [x] 支持优雅退出: 捕捉进程 SIGINT, SIGTERM 信号实现有序退出
- [x] 支持时限: `-t` 限制吃资源的时间，示例 "300ms", "1.5h", "2h45m". (单位: "ns", "us" (or "µs"), "ms", "s", "m", "h")
- [] CPU亲和性
- [] 内存读写，防止内存被交换出去
- [] 动态调整CPU和内存使用
- [] 吃GPU

# 使用

```shell
eat -c 4            # 占用4个CPU核
eat -c 35%          # 占用35%CPU核（CPU核数 * 35%）
eat -c 100%         # 占用所有CPU核
eat -m 4g           # 占用4GB内存
eat -m 20m          # 占用20MB内存
eat -m 35%          # 占用35%内存（总内存 * 35%）
eat -m 100%         # 占用所有内存
eat -c 2.5 -m 1.5g  # 占用2.5个CPU核和1.5GB内存
eat -c 3 -m 200m    # 占用3个CPU核和200MB内存
eat -c 100% -m 100% # 占用所有CPU核和内存
eat -c 100% -t 1h   # 占用所有CPU核并在一小时后退出
```

> 提示：
> - 使用\<Ctrl\> + C来停止占用并释放CPU和内存

# 构建

```shell
go build -o eat
```