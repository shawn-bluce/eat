# Introduce
<b>I'm a CPU and memory eating monster.</b>

Developer will encounter the need to quickly occupy CPU and memory, I am also deeply troubled, so I developed a tool named `eat` to help you quickly occupy a specified amount of CPU and memory.

# Usage

```shell
eat -c 4            # eating 4 CPU core
eat -c 100%         # eating all CPU core
eat -m 4g           # eating 4GB memory
eat -m 20m          # eating 20MB memory
eat -m 100%         # eating all memory
eat -c 2.5 -m 1.5g  # eating 2.5 CPU core and 1.5GB memory
eat -c 3 -m 200m    # eating 3 CPU core and 200MB memory
eat -c 100% -m 100% # eating all CPU core and memory
```

> Tips:
> - Using \<Ctrl\> + C to stop eating and release CPU and memory

# Build

```shell
go build -o eat
```

# 介绍
<b>我是一个吃CPU和内存的怪物。</b>

开发者在遇到需要快速占用CPU和内存的需求时，我也深受其扰，所以我开发了一个工具名为`eat`来帮助你快速占用指定数量的CPU和内存。

# 使用

```shell
eat -c 4            # 占用4个CPU核
eat -c 100%         # 占用所有CPU核
eat -m 4g           # 占用4GB内存
eat -m 20m          # 占用20MB内存
eat -m 100%         # 占用所有内存
eat -c 2.5 -m 1.5g  # 占用2.5个CPU核和1.5GB内存
eat -c 3 -m 200m    # 占用3个CPU核和200MB内存
eat -c 100% -m 100% # 占用所有CPU核和内存
```

> 提示：
> - 使用\<Ctrl\> + C来停止占用并释放CPU和内存

# 构建

```shell
go build -o eat
```