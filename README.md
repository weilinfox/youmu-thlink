# 白玉楼製作所 ThLink

通用的方便自搭建的东方联机器。

本质上是个支持 UDP 的端口转发器。所以其实理论上其他联机器也都是通用的（？）

服务端和客户端使用可选的 QUIC 或 TCP 传输，具有较高的网络适应能力；支持 LZW 压缩，在一定程度上节约带宽。

感谢 飞翔君 的一起测试。

## 特性

1. 使用 [QUIC](https://en.wikipedia.org/wiki/QUIC)/TCP 作为传输协议
2. 可选的 QUIC 和 TCP 传输
3. 支持使用 UDP 进行联机的东方作品
4. 可配置的监听端口和服务器地址，方便自搭建
5. 使用 [LZW](https://en.wikipedia.org/wiki/Lempel%E2%80%93Ziv%E2%80%93Welch) 压缩，节约少量带宽
6. 显示延迟，字符界面 log 直接打印，方便 debug
7. 代码乱七八糟

## TODO

1. 观战支持
2. 测试更多作品
3. 支持 TCP 转发

## 使用方法

按 client 界面提示操作即可。

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

构建得到的二进制在 build 目录下。

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
