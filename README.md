# ThLink

```
user1 --[udp/tcp]--> client --[tcp]--> broker --[udp/tcp]--> user2

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
