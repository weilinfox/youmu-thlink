# 白玉楼製作所 ThLink

通用的方便自搭建的东方联机器。

同时支持 TCP 和 UDP 的端口转发器。所以其实理论上其他联机器也都是通用的（？）

## 特性

1. 使用 [QUIC](https://en.wikipedia.org/wiki/QUIC) 作为传输协议
2. 支持使用 UDP 进行联机的东方作品
3. 可配置的监听端口和服务器地址，方便自搭建
4. 使用 [LZW](https://en.wikipedia.org/wiki/Lempel%E2%80%93Ziv%E2%80%93Welch) 压缩，节约少量带宽
5. 显示延迟，字符界面 log 直接打印，方便 debug
6. 代码乱七八糟

## TODO

1. 观战支持
2. 可选的 QUIC 和 TCP 传输
3. 测试更多作品

## 使用方法

与 [shitama](https://github.com/u-u-z/shitama) 和 shitama 的[原作者](https://github.com/evshiron)新作 [swarm](https://github.com/evshiron/swarm-ng-build) 一样。

## TH09

花映塚使用 DirectPlay 实现联机，故需要 adonis2 配合才能使用 thlink 联机。

在 [Maribel Hearn's Touhou Portal](https://maribelhearn.com/pofv) 的相关页面可以找到需要的工具 [Adonis2](https://maribelhearn.com/mirror/PoFV-Adonis-VPatch-Goodies.zip) 。

将其 ``files`` 目录下的所有文件直接放到缩到花映塚目录。

联机时对战双方都不要运行花映塚，而是直接运行 ``adonis2.exe`` （日文）或 ``adonis2e.exe`` （英文），按照提示操作。

主机端 adonis2 会提示监听端口，默认为 ``10800`` 。配置好 adonis2 后启动 thlink ，在端口输入部分输入 adonis2 提示的端口（或保持默认端口），其他默认或按需配置， thlink 会提示对端 IP 。

客机端使用 adonis2 连接 thlink 返回的 IP 。

联机成功后 adomis2 会自动启动花映塚。

## TH10.5

绯想天直接在游戏内联机即可，主机默认端口为 ``10800`` 。将 thlink 设置成一样的配置，客机输入 thlink 返回的 IP 。

## TH12.3

非想天则同样直接在游戏内联机即可，主机默认端口为 ``10800`` 。将 thlink 设置成一样的配置，客机输入 thlink 返回的 IP 。

## 关于传输协议

+ v0.0.1 使用了 tcp ，虽然实时性不是很好（？）但是在国内网络环境下比较稳定
+ v0.0.3 使用了 [kcp](https://github.com/skywind3000/kcp) ，试图提升一下性能，部署以后发现从 broker 发往 client 的包都消失了……
+ v0.0.5 开始使用 [quic](https://en.wikipedia.org/wiki/QUIC) ，测试效果总体来说比 kcp 稳定得多
