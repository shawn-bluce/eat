# Introduce
<b>I'm a CPU and memory eating monster🦕</b>

Developer will encounter the need to quickly occupy CPU and memory, I am also deeply troubled, so I developed a tool named `eat` to help you quickly occupy a specified amount of CPU and memory.

# Todo

- [x] Support `eat -c 35%` and `eat -m 35%`
- [x] support gracefully exit: capture process signal SIGINT(2), SIGTERM(15)
- [x] support deadline: `-t` specify the duration of eat progress. such as "300ms", "1.5h", "2h45m". (unit: "ns", "us" (or "µs"), "ms", "s", "m", "h")
- [x] CPU Affinity
  - [x] Linux
  - [ ] macOs
  - [ ] Windows
- [x] Memory read/write periodically , prevent memory from being swapped out
- [ ] Dynamic adjustment of CPU and memory usage
- [ ] Eat GPU

# Usage

```shell
$ ./eat.out --help
A monster that eats cpu and memory 🦕

Usage:
    eat [flags]

Flags:
  --cpu-affinities ints                  Which cpu core(s) would you want to eat? multiple cores separate by ','
  -c, --cpu-usage string                 How many cpu would you want eat (default "0")
  -h, --help                             help for eat
  -r, --memory-refresh-interval string   How often to trigger a refresh to prevent the ate memory from being swapped out (default "5m")
  -m, --memory-usage string              How many memory would you want eat(GB) (default "0m")
  -t, --time-deadline string             Deadline to quit eat process (default "0")
```

```shell
eat -c 4		# eating 4 CPU core
eat -c 35%		# eating 35% CPU core (CPU count * 35%)
eat -c 100%		# eating all CPU core
eat -m 4g		# eating 4GB memory
eat -m 20m		# eating 20MB memory
eat -m 35%		# eating 35% memory (total memory * 35%)
eat -m 100%		# eating all memory
eat -c 2.5 -m 1.5g	# eating 2.5 CPU core and 1.5GB memory
eat -c 3 -m 200m	# eating 3 CPU core and 200MB memory
eat -c 100% -m 100%	# eating all CPU core and memory
eat -c 100% -t 1h	# eating all CPU core and quit after 1hour

eat --cpu-affinities 0 -c 1	# only run eat in core #0 (first core)
eat --cpu-affinities 0,1 -c 2	# run eat in core #0,1 (first and second core)
eat --cpu-affinities 0,1,2,3 -c 100% # error case: in-enough cpu affinities
# Have 8C15G.
# Error: failed to parse cpu affinities, reason: each request cpu cores need specify its affinity, aff 4 < req 8
eat --cpu-affinities 0,1,2,3 -c 50%	# run eat in core #0,1,2,3 (first to fourth core)
eat --cpu-affinities 0,1,2,3,4,5,6,7 -c 92%	# run eat in all core(full of 7 cores, part of last core)

```

> Tips:
> - Using \<Ctrl\> + C to stop eating and release CPU and memory

# Build

```shell
# Linux
make linux-amd64 linux-arm64
# macOs
make darwin-amd64 darwin-arm64
# Windows
make windows-amd64 windows-arm64
```

# 介绍
<b>我是一只吃CPU和内存的怪兽🦕</b>

开发者们经常会遇到需要快速占用 CPU 和内存的需求，我也是。所以我开发了一个名为 `eat` 的小工具来快速占用指定数量的 CPU 和内存。

# 待办

- [x] 支持`eat -c 35%`和`eat -m 35%`
- [x] 支持优雅退出: 捕捉进程 SIGINT, SIGTERM 信号实现有序退出
- [x] 支持时限: `-t` 限制吃资源的时间，示例 "300ms", "1.5h", "2h45m". (单位: "ns", "us" (or "µs"), "ms", "s", "m", "h")
- [x] CPU亲和性(绑定 CPU 核心)
  - [X] Linux 
  - [ ] macOS
  - [ ] Windows
- [x] 定期内存读写，防止内存被交换出去
- [ ] 动态调整CPU和内存使用
- [ ] 吃GPU

# 使用


```shell
$ ./eat.out --help
我是一只吃CPU和内存的怪兽🦕

使用方法
    eat [flags］

Flags：
  --cpu-affinities 			整数	指定在几个核心上运行 Eat，多个核心索引之间用 ',' 分隔，索引从 0 开始。
  -c, --cpu-usage 			字符串	你想吃掉多少个 CPU（默认为 '0'）？
  -h，--help				输出 eat 的帮助
  -r, --memory-refresh-interval 字符串	每隔多长时间触发一次刷新，以防止被吃掉的内存被交换出去（默认值为 '5m'）
  -m, --memory-usage 字符串		你希望吃掉多少内存（GB）（默认值 '0m'）
  -t，--time-deadline 字符串		退出 eat 进程的截止日期（默认为 "0'）。
```

```shell
eat -c 4	# 占用4个CPU核
eat -c 35%	# 占用35%CPU核（CPU核数 * 35%）
eat -c 100%	# 占用所有CPU核
eat -m 4g	# 占用4GB内存
eat -m 20m	# 占用20MB内存
eat -m 35%	# 占用35%内存（总内存 * 35%）
eat -m 100%	# 占用所有内存
eat -c 2.5 -m 1.5g	# 占用2.5个CPU核和1.5GB内存
eat -c 3 -m 200m	# 占用3个CPU核和200MB内存
eat -c 100% -m 100%	# 占用所有CPU核和内存
eat -c 100% -t 1h	# 占用所有CPU核并在一小时后退出

eat --cpu-affinities 0 -c 1	# 只占用 #0 第一个核心
eat --cpu-affinities 0,1 -c 2	# 占用 #0,1 前两个个核心
eat --cpu-affinities 0,1,2,3 -c 100%	# 错误参数: 每个请求核都要指定对应的亲和性核心
# Have 8C15G.
# Error: failed to parse cpu affinities, reason: each request cpu cores need specify its affinity, aff 4 < req 8
# 出错: 无法解析 CPU 亲和性, 原因: 每个请求核都要指定对应的亲和性核心, 亲和核  4 < 请求核 8
eat --cpu-affinities 0,1,2,3 -c 50%	# 占用前4个核心
eat --cpu-affinities 0,1,2,3,4,5,6,7 -c 92%	# 占用前8个核心 (全部7个核心，部分的最后一个核心)
```

> 提示：
> - 使用\<Ctrl\> + C来停止占用并释放CPU和内存

# 构建

```shell
# Linux
make linux-amd64 linux-arm64
# macOs
make darwin-amd64 darwin-arm64
# Windows
make windows-amd64 windows-arm64
```
