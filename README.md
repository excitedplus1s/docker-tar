<h1 align="center">docker-tar</h1>

[English](./README.EN.md) | 简体中文

## 介绍
docker-tar 是一个用于从 Docker 仓库拉取镜像并自动打包为 Tar 包的工具。

## 特点

- **无需Docker环境**：二进制中包含了所以功能，开箱即用
- **多操作系统运行**： 提供相关二进制，`Linux`、`Windows`、`MacOS`、`Solaris`、`OpenBSD`等操作系统均可使用，更多支持操作系统见 Release
- **多处理器架构运行**：提供相关二进制，支持`amd64/x64`、`i386/x86`、`arm`、`arm64`、`mips`、`ppc64`等处理器架构，你可以使用 PC 甚至是路由器来运行此工具，更多支持架构见 Release
- **镜像模式下载**：镜像站点下载的镜像 tar 无需进行 tag，可保持和输入的镜像名称一致
- **登录下载**：支持需要认证的 Docker 仓库及其镜像站
- **选择镜像架构**：支持选择下载具有多架构的 Docker 镜像
- **更接近 Docker CLI 的下载结果**：此工具下载的 tar 镜像文件，与使用 `docker pull`和`docker save`保存的文件完全一致
- **实验室模式**：提供解决 DNS 污染和 SNI 阻断的实验室模式，帮助你直连某些无法访问的 Docker 仓库站点


## 命令行使用说明


```shell
-action action
        pull: 拉取镜像tar文件.
        list: 列出可以拉取的处理器架构镜像
  -arch architecture
        指定需要拉取的镜像架构 (默认值为 "amd64")
  -image name
        需要下载的镜像信息
        不携带版本默认为latest
        默认使用 DockerHub 的 Docker 仓库
        nginx
        home-assistant/home-assistant:stable
        ghcr.io/home-assistant/home-assistant:stable
  -output filename
        输出 tar 镜像的文件名，不指定此选项将随机生成文件名
   -username username
        Docker 仓库/镜像站点需要登录时，需要提供用户名
  -password password
        Docker 仓库/镜像站点需要登录时，需要提供密码
  -mirror string
        使用镜像站点下载镜像
        这种方式不会修改 -image 指定的名称
  -lab
        实验室模式，开启将使用默认参数来规避 DNS 污染和 SNI 阻断,pull和list模式下都可以使用。
        一般情况下，默认设置是足够的
  -dns-servers ip list
        实验室模式开启后，此选项生效。
        工具内置了列表，输入内容将覆盖内置列表。
        DNS服务器配置选项，必须为解析后的 IP 地址，以逗号分割
  -dns-timeout int
         实验室模式开启后，此选项生效。 
         过滤DNS抢答的超时时间配置. (默认值 2s)
  -network network
        实验室模式开启后，此选项生效。
        ip4: 仅使用 IPv4 下载镜像
        ip6: 仅使用 IPv6 下载镜像
        ip: 使用 IPv4以及IPv6 下载镜像 (默认值 "ip")
```

## 使用示例


#### 列出目标镜像支持的架构

查询 nginx 支持的架构
```shell
docker-tar -action list -image nginx
```
输出结果

```shell
Available Architecture:
amd64
armv5
armv7
arm64v8
386
mips64le
ppc64le
s390x
```

#### 下载镜像（直接下载）
使用这种方式需要配置 HTTPS_PROXY环境变量
下载 nginx armv7 架构的 nginx 
```shell
docker-tar -action pull -image nginx -arch armv7 -output nginx.tar
```
输出结果

```shell
Pulling from  nginx:latest
[1/7]3d83c6df5858: Downloading  100% |██████████████████████████████████████████████████| (24/24 MB, 1.9 MB/s)
[2/7]9260be83662e: Downloading  100% |██████████████████████████████████████████████████| (37/37 MB, 2.0 MB/s)
[3/7]3058d76256fc: Downloading  100% |██████████████████████████████████████████████████| (628/628 B, 986 kB/s)
[4/7]7888ef92f48d: Downloading  100% |██████████████████████████████████████████████████| (956/956 B, 994 kB/s)
[5/7]5c27d4148f08: Downloading  100% |██████████████████████████████████████████████████| (404/404 B, 400 kB/s)
[6/7]9c733a7e6553: Downloading  100% |█████████████████████████████████████████████████| (1.2/1.2 kB, 1.5 MB/s)
[7/7]931fb3c7333d: Downloading  100% |█████████████████████████████████████████████████| (1.4/1.4 kB, 2.5 MB/s)
Output File:  nginx.tar
```

#### 下载镜像（通过镜像站点）
下载 nginx armv7 架构的 nginx 
```shell
docker-tar -action pull -image nginx -arch armv7 -output nginx.tar -mirror docker.xuanyuan.me
```
输出结果

```shell
Pulling from  nginx:latest
[1/7]9c733a7e6553: Downloading  100% |█████████████████████████████████████████████████| (1.2/1.2 kB, 1.0 MB/s)
[2/7]931fb3c7333d: Downloading  100% |█████████████████████████████████████████████████| (1.4/1.4 kB, 2.8 MB/s)
[3/7]3d83c6df5858: Downloading  100% |███████████████████████████████████████████████████| (24/24 MB, 3.6 MB/s)
[4/7]9260be83662e: Downloading  100% |███████████████████████████████████████████████████| (37/37 MB, 5.2 MB/s)
[5/7]3058d76256fc: Downloading  100% |███████████████████████████████████████████████████| (628/628 B, 68 kB/s)
[6/7]7888ef92f48d: Downloading  100% |██████████████████████████████████████████████████| (956/956 B, 107 kB/s)
[7/7]5c27d4148f08: Downloading  100% |███████████████████████████████████████████████████| (404/404 B, 35 kB/s)
Output File:  nginx.tar
```

#### 下载镜像（使用实验室模式直接下载）

下载 nginx armv7 架构的 nginx 
```shell
docker-tar -action pull -image nginx -arch armv7 -output nginx.tar -lab
```
输出结果

```shell
Pulling from  nginx:latest
[1/7]3d83c6df5858: Downloading  100% |███████████████████████████████████████████████████| (24/24 MB, 6.9 MB/s)
[2/7]9260be83662e: Downloading  100% |████████████████████████████████████████████████████| (37/37 MB, 12 MB/s)
[3/7]3058d76256fc: Downloading  100% |██████████████████████████████████████████████████| (628/628 B, 489 kB/s)
[4/7]7888ef92f48d: Downloading  100% |██████████████████████████████████████████████████| (956/956 B, 671 kB/s)
[5/7]5c27d4148f08: Downloading  100% |██████████████████████████████████████████████████| (404/404 B, 353 kB/s)
[6/7]9c733a7e6553: Downloading  100% |█████████████████████████████████████████████████| (1.2/1.2 kB, 1.1 MB/s)
[7/7]931fb3c7333d: Downloading  100% |█████████████████████████████████████████████████| (1.4/1.4 kB, 2.8 MB/s)
Output File:  nginx.tar
```
## TODO

- [ ] 多Layer并发下载
- [ ] 分片下载支持，对支持 Range 的响应实现分片下载
- [ ] Docker API的兼容（ContentType）
- [ ] 代码结构、命名调整
- [ ] 更友好的错误输出

## 兼容情况（目前已测试）
- [x] DockerHub 
- [x] DockerHub 镜像站
- [x] ghcr.io
- [x] ghcr.io 镜像站
- [ ] quay.io (目前未适配 application/vnd.docker.distribution.manifest.v1+prettyjws）
- [ ] quay.io 镜像站(目前未适配 application/vnd.docker.distribution.manifest.v1+prettyjws）

## 一些可能有用的信息

| 请求阶段 | 支持的 ContentType | 其他 |
| ----- | ----- | ----- |
| 获取Token | 只处理响应体，忽略类型 | Access Token:Bearer,用户名密码(如需要）: Basic|
| 获取所有架构镜像清单索引 | application/vnd.oci.image.index.v1+json | 响应非此类型会直接终止 |
| 获取指定架构的镜像清单 | application/vnd.oci.image.manifest.v1+json | 兼容类型也可 |
| 获取镜像配置 | application/vnd.oci.image.config.v1+json | 兼容类型也可 |
| 下载Layer | application/vnd.oci.image.layer.v1.tar<br>application/vnd.oci.image.layer.v1.tar+gzip<br>application/vnd.oci.image.layer.v1.tar+zstd<br>application/vnd.oci.image.layer.nondistributable.v1.tar<br>application/vnd.oci.image.layer.nondistributable.v1.tar+gzip<br>application/vnd.oci.image.layer.nondistributable.v1.tar+zstd<br>application/vnd.docker.image.rootfs.diff.tar<br>application/vnd.docker.image.rootfs.diff.tar.gzip<br>application/vnd.docker.image.rootfs.diff.tar.zstd | 加密镜像不支持 |
