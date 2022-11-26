# ThLink

```
user1 --[udp/tcp]--> client --[quic]--> broker --[udp/tcp]--> user2

New tunnel:
client --[transfer type]--> broker
broker --[client && server port] --> client

Establish tunnel:
client --[connect]--> broker
broker --[listen] --> user2

End tunnel:
client --[disconnect]--> broker
broker --[disconnect]--> user2
```

## 传输协议的尝试

+ v0.0.1 使用了 tcp ，虽然实时性不是很好但是在国内网络环境下比较稳定；
+ v0.0.3 使用了 [kcp-go](https://github.com/xtaci/kcp-go) ，试图提升一下性能，部署以后发现从服务端发往客户端的包都消失了……
+ v0.0.4 调整了部分 kcp 参数，但是没啥用，先放弃这个协议，也懒得改其他 bug 了；
+ v0.0.5 使用了 [quic-go](https://github.com/lucas-clemente/quic-go) ， 测试效果总体来说比 kcp 稳定得多。
