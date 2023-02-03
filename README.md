# 白玉楼製作所 ThLink

[![License](https://img.shields.io/github/license/weilinfox/youmu-thlink)](https://github.com/weilinfox/youmu-thlink/blob/master/LICENSE)
[![Release](https://img.shields.io/github/v/release/weilinfox/youmu-thlink)](https://github.com/weilinfox/youmu-thlink/releases)
[![Go Report Card](https://goreportcard.com/badge/github.com/weilinfox/youmu-thlink)](https://goreportcard.com/report/github.com/weilinfox/youmu-thlink)

[![FOSSA Status](https://app.fossa.com/api/projects/git%2Bgithub.com%2Fweilinfox%2Fyoumu-thlink.svg?type=shield)](https://app.fossa.com/projects/git%2Bgithub.com%2Fweilinfox%2Fyoumu-thlink?ref=badge_shield)
[![FOSSA Status](https://app.fossa.com/api/projects/git%2Bgithub.com%2Fweilinfox%2Fyoumu-thlink.svg?type=shield&issueType=security)](https://app.fossa.com/projects/git%2Bgithub.com%2Fweilinfox%2Fyoumu-thlink?ref=badge_shield)

通用的方便自搭建的东方联机器。

本质上是个支持 UDP 的端口转发器， thlink 客户端和 thlink 服务端之间使用可选的 QUIC 和 TCP 传输。

thlink 客户端以非想天则/凭依华插件的形式实现独立于对战双方的观战， thlink 客户端将从对战双方预取观战数据，然后拦截并回应来自观战客户端的请求。

为了使客户端能够在连接任意一个服务端的情况下获知所有存在于该网络的服务端，推荐将所有服务端连接成树状， ``broker -u hostname:port`` 
将指定本服务端连接到另一个服务端的地址。这样一来命令行客户端可以传入 ``-a`` 来自动选择延迟最低的服务端， 
gtk 客户端则可以在菜单的 ``Network Discovery`` 自主选择客户端。不过要注意客户端和服务端之间的 ``ping`` 延迟并不能完整展现对战双方的网络延迟情况，
而在打开非想天则插件 ``th123`` 的情况下，客户端会在信息栏显示对战双方交换数据的单程延迟。

动机来源于自己构建并搭建的 [shitama](https://github.com/u-u-z/shitama) 很不稳定，~~为了逃避繁重的复习而给自己一点事情做做，~~
因为不知道 shitama 就是用的 kcp 协议，一开始用 kcp 写结果发出去的包一个都回不来。

初入东方对 0 萌新一只，感谢**飞翔君**带我一起玩！

![screenshot-v0.0.10-amd64-windows](screenshot/screenshot-v0.0.10-amd64-windows.png)

感谢 [JetBrains](https://www.jetbrains.com/) 为本项目提供的开源开发许可证！ [![JetBrains Main Logo](https://resources.jetbrains.com/storage/products/company/brand/logos/jb_beam.svg)](https://www.jetbrains.com/)

## 特性

1. 使用 [QUIC](https://en.wikipedia.org/wiki/QUIC)/TCP 作为传输协议
2. 可选的 QUIC 和 TCP 传输
3. 支持使用 UDP 进行联机的东方作品
4. 可配置的监听端口和服务器地址，方便自搭建
5. 支持去中心化的多服务器结构
6. 支持非想天则观战，观战支持的原理见 [hisoutensoku-spectacle](https://github.com/weilinfox/youmu-hisoutensoku-spectacle)
7. 支持凭依华观战，观战支持的原理见 [hyouibana-spectacle](https://github.com/weilinfox/youmu-hyouibana-spectacle)
8. 使用 [LZW](https://en.wikipedia.org/wiki/Lempel%E2%80%93Ziv%E2%80%93Welch) 压缩，节约少量带宽
9. 符合习惯的命令行客户端和还算易用的 gtk3 图形客户端
10. Linux 下以 [AppImage](https://appimage.org/) 格式发布图形客户端
11. 代码乱七八糟的，就是说，这个东西，被我写得很糟糕
12. 我的英文很差很差，注释将就看吧别来打我（缩）

## TODO

1. 测试更多作品
2. 支持 TCP 转发

## 预编译的二进制

+ [github release](https://github.com/weilinfox/youmu-thlink/releases)
+ [gitee release](https://gitee.com/weilinfox/youmu-thlink/releases) （镜像，如果有文件缺失那就是上传失败了）
+ [pling](https://www.pling.com/p/1963595/) （仅 AppImage）

### Archlinux

从 AUR 安装：

```shell
$ yay -S thlink-client-gtk
$ yay -S thlink-client
$ yay -S thlink-broker
```

## 客户端使用指导

### TH09

作品名称：

|日文|中文|
|:-|:-|
|東方花映塚　～ Phantasmagoria of Flower View.|东方花映塚　～ Phantasmagoria of Flower View.|

东方花映冢（东方花映塚）使用 DirectPlay 实现联机，故需要 adonis2 配合才能使用 thlink 联机。

在 [Maribel Hearn's Touhou Portal](https://maribelhearn.com/pofv) 的相关页面可以找到需要的工具 [Adonis2](https://maribelhearn.com/mirror/PoFV-Adonis-VPatch-Goodies.zip) 。

将其 ``files`` 目录下的所有文件直接放到缩到花映塚目录。

联机时对战双方都不要运行花映塚，而是直接运行 ``adonis2.exe`` （日文）或 ``adonis2e.exe`` （英文），按照提示操作。

主机端 adonis2 会提示监听端口，默认为 ``10800`` 。配置好 adonis2 后启动 thlink ，在端口输入部分输入 adonis2 提示的端口（或保持默认端口），其他默认或按需配置， thlink 会提示对端 IP 。

客机端使用 adonis2 连接 thlink 返回的 IP 。

联机成功后 adomis2 会自动启动花映塚。

### TH10.5 TH12.3 TH13.5 TH14.5

作品名称：

|日文|中文|
|:-|:-|
|東方緋想天　～ Scarlet Weather Rhapsody.|东方绯想天　～ Scarlet Weather Rhapsody.|
|東方非想天則　～ 超弩級ギニョルの謎を追え|东方非想天则　～ 追寻特大型人偶之谜|
|東方心綺楼　～ Hopeless Masquerade.|东方心绮楼　～ Hopeless Masquerade.|
|東方深秘録　～ Urban Legend in Limbo.|东方深秘录　～ Urban Legend in Limbo.|

游戏内文字：

|日文|中文|
|:-|:-|
|対戦サーバーを立てる|建立对战服务器（主机端选择）|
|IPとポートを指定してサーバーに接続|连接到指定IP和端口的服务器（客机选择）|
|使用するポート|连接使用的端口号|

直接在游戏内联机即可，主机默认端口为 ``10800`` 。将 thlink 设置成一样的配置，客机输入 thlink 返回的 IP 。

### TH15.5

作品名称：

|日文|中文|
|:-|:-|
|東方憑依華　～ Antinomy of Common Flowers.|东方凭依华　～ Antinomy of Common Flowers.|

游戏内文字：

|日文|中文|
|:-|:-|
|対戦相手の接続を待つ|等待对方连接（主机端选择）|
|接続先を指定して対戦相手接続に|指定对端以连接到对方（客机选择）|
|観戦する|观战|
|使用するポート番号|连接使用的端口号|

直接在游戏内联机即可，主机默认端口为 ``10800`` 。将 thlink 设置成一样的配置，客机输入 thlink 返回的 IP 。

若在 Wine 环境下，终端运行 ``th155.exe`` 在输出下面的内容一次后退出：

```
Allocator::Info[system] total 134217728 / free 134282760 / use 504
Allocator::Info[stl] total 33554432 / free 33619464 / use 504
```

注意这不是报错，只是单纯的症状，不知道为啥 ``th155.exe`` 没有产生任何错误信息。此时只要先 ``cd`` 到游戏所在目录，再重新尝试运行。

## 构建和部署

Go >= 1.18

注意 loong64 架构从 Go 1.19 开始被支持。

### 本机构建

```shell
$ make
```

可用选项：

+ ``static`` 构建静态链接的二进制
+ ``loong64`` 构建 loong64 架构的二进制
+ ``windows`` 构建 windows amd64 可执行文件
+ ``gui`` Linux 下构建本机动态链接的图形界面客户端二进制

构建得到的二进制在 build 目录下。

### 离线构建

下载全部依赖的包：

```shell
$ git mod vendor
```

将生成的 ``./vendor`` 目录打包拷贝到别处，再解包到项目根目录运行构建：

```shell
$ go mod=vendor build -o build/thlink-client-gtk ./client-gtk3
```

### 部署

broker 为服务端， client 为客户端。

若想将自己的 broker 连入其他 broker 的网络，只要连接这个网络中的任一 broker 即可，它们的地位是平行的。

broker 在服务器运行即可， ``broker -h`` 查看选项； client 在本地运行， ``client -h`` 查看选项。

client-gtk 没有提供特殊的命令行界面。

### Linux GTK3 GUI

go1.18 需要自行下载或构建。

安装依赖（以 Debian 为例，水平有限，可能不全）

```shell
$ sudo apt-get install libgtk-3-dev libcairo2-dev glib2.0-dev
```

构建：

```shell
$ make gui
```

制作 AppImage （参见 [linuxdeploy-plugin-gtk](https://github.com/linuxdeploy/linuxdeploy-plugin-gtk) ）：

```shell
$ cd ./build

$ wget -c "https://raw.githubusercontent.com/linuxdeploy/linuxdeploy-plugin-gtk/master/linuxdeploy-plugin-gtk.sh"
$ wget -c "https://github.com/linuxdeploy/linuxdeploy/releases/download/continuous/linuxdeploy-x86_64.AppImage"
$ chmod +x linuxdeploy-x86_64.AppImage linuxdeploy-plugin-gtk.sh

$ mkdir -p thlink-client-gtk.AppDir/usr/bin/
$ install -Dm775 ./thlink-client-gtk thlink-client-gtk.AppDir/usr/bin/
$ ./linuxdeploy-x86_64.AppImage --appdir thlink-client-gtk.AppDir/ --plugin gtk --output appimage --icon-file thlink-client-gtk.png --desktop-file thlink-client-gtk.desktop
$ mv ThLink_Client_Gtk-x86_64.AppImage thlink-client-gtk-amd64-linux.AppImage
```

### Windows GTK3 GUI

图形界面客户端在 Windows 上使用 [MSYS2](https://www.msys2.org/) 构建，水平有限，只能提供一个不全的指南。

也可参考 gotk3 的 [Wiki](https://github.com/gotk3/gotk3/wiki/Installing-on-Windows#chocolatey) 使用 Chocolatey 搭建环境。

注意 go 应该使用 Windows 版本而不是 MSYS2 软件源中提供的版本 ，否则可能会有构建失败的情况（待考证，我有出现 ``ld`` 找不到 ``-lmingwex`` 和 ``-lmingw32`` 的报错）。

更新到最新并安装依赖：

```shell
$ pacman -Syuu
$ pacman -S mingw-w64-x86_64-gtk3 mingw-w64-x86_64-toolchain base-devel glib2-devel
```

配置环境变量（根据实际情况修改），其中 ``/c/msys64/mingw64/bin`` 代表的是 Mingw64 gcc 所在目录， ``/c/Go/bin`` 则代表的是 Windows 的 go 所在目录：

```shell
$ echo 'export PATH=/c/msys64/mingw64/bin:/c/Go/bin:$PATH' >> ~/.bashrc
$ source ~/.bashrc
```

修复 gdk-3.0.pc 中的一个 bug （坑了我好久，其实 gotk3 的 Wiki 有写到）：

```shell
$ sed -i -e 's/-Wl,-luuid/-luuid/g' /mingw64/lib/pkgconfig/gdk-3.0.pc
```

构建图标和本体， ``-H windowsgui`` 使其运行时没有黑色终端：

```shell
$ windres -o ./client-gtk3/icon.syso ./client-gtk3/icon.rc
$ go build -ldflags "-H windowsgui" -o ./build/client-gtk3-windows/thlink-client-gtk.exe ./client-gtk3
```

复制依赖库：

```shell
$ ldd ./build/client-gtk3-windows/thlink-client-gtk.exe | grep -o '/mingw64/bin/[^ ]*' | xargs --replace=R -t cp R ./build/client-gtk3-windows/
```

GTK icons：

```shell
$ mkdir -p ./build/client-gtk3-windows/lib/gdk-pixbuf-2.0/2.10.0/loaders
$ cp /mingw64/lib/gdk-pixbuf-2.0/2.10.0/loaders/libpixbufloader-png.dll ./build/client-gtk3-windows/lib/gdk-pixbuf-2.0/2.10.0/loaders/
$ cp /mingw64/lib/gdk-pixbuf-2.0/2.10.0/loaders/libpixbufloader-xpm.dll ./build/client-gtk3-windows/lib/gdk-pixbuf-2.0/2.10.0/loaders/
$ cp /mingw64/lib/gdk-pixbuf-2.0/2.10.0/loaders.cache ./build/client-gtk3-windows/lib/gdk-pixbuf-2.0/2.10.0/
```

打包整个 ``./build/client-gtk3-windows`` 目录即可。

## 使用的端口

broker （服务端）使用 TCP 端口 4646 作为和所有 client （客户端）交互的固定端口， client 通过这个端口向 broker 请求转发通道。

```
      (dynamic)        4646
+------------+          +---------------+
| 主机  Host | quic/tcp | 服务端 Server |
|   client   | <------> |     broker    |
+------------+          +---------------+
```

client 请求转发通道成功后获得一个端口对（ ``port1`` 和 ``port2 ``）。其中一个建立 client 和 broker 之间的连接，用作之后所有数据的交换；另一个用于客户机的连接。

```
          10800         (dynamic)               port1         port2         (dynamic)
+------------+           +------------+          +---------------+           +------------+
| 主机  Host |  tcp/udp  | 主机  Host | quic/tcp | 服务端 Server |  tcp/udp  | 客机 Guest |
|  TH Game   | <-------> |   client   | <------> |     broker    | <-------> |  TH Game   |
+------------+           +------------+          +---------------+           +------------+
```

通常 ``port1`` 、 ``port2`` 和其他动态端口均在 32768-65535 。

## 关于传输协议

+ v0.0.1 使用了 tcp ，虽然实时性不是很好（？）但是在国内网络环境下比较稳定
+ v0.0.3 使用了 [kcp](https://github.com/skywind3000/kcp) ，试图提升一下性能，部署以后发现从 broker 发往 client 的包都消失了……
+ v0.0.5 开始使用 [quic](https://en.wikipedia.org/wiki/QUIC) ，测试效果总体来说比 kcp 稳定得多
+ V0.0.6 开始可以在客户端自主选择使用 quic 或 tcp 传输，增强在复杂网络环境的适应性
