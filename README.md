# GinChat 说明文档 



## 一、数据库表设计

use_basic 用户表（记录用户基础信息）

contact 联系表（记录用户间的好友关系、群聊关系等）

communities 群聊表（记录群聊信息——群名、群图像、群主）



## 二、功能实现



### （一）用户消息发送

**步骤 1：WebSocket 连接建立（`Chat`函数，客户端接入的入口） ** 

​	主要功能： “**HTTP 升级为 WS 连接** + **用户 - 连接绑定** + **协程初始化**”

​	**1、参数校验与 WS 升级：**

​	从 URL 参数获取userId，并且校验 token，通过 `websocket.Upgrader` 将 `HTTP` 请求升级为 `WebSocket` 连接。

​	**2、创建 Node 实例：**

​	初始化 `Node`： **“WS 连接、客户端地址、心跳时间、登录时间、消息队列（`DataQueue`，缓冲待发送消息）”** 

​	**3、绑定用户 - 连接：**

​	通过读写锁 `rwLocker` 将 `userId` 和 `Node` 绑定到 `clientMap（全局连接映射）`，保证并发安全。

​	**4、启动核心协程：**

- `go sendProc(node)`：负责从`DataQueue`取消息，推送给客户端；
- `go recvProc(node)`：负责监听客户端发送的消息，处理接收逻辑。

​	**5、缓存在线状态**

​	将用户在线信息存入 Redis，并设置过期时间（适配离线检测）。



**步骤 2：客户端消息接收（`recvProc`协程）**

​	每个 WS 连接对应一个`recvProc`协程，**持续监听客户端发送的消息**，是消息流入的入口：

​	**1、读取 WS 消息**

​	循环调用 `conn.ReadMessage()` 读取客户端发送的二进制消息。

​	**2、解析消息结构体**

​	将消息 JSON 反序列化为 `Message` 结构体，若解析失败则打印错误。

​	**3、消息类型分流**

- **心跳消息（Type=3）**：调用`node.Heartbeat()`更新心跳时间（防止连接被清理）；
- 业务消息（私聊 / 群聊）
  - 调用 `dispatch(data)` ：后端核心调度逻辑，分发消息；
  - 调用 `broadMsg(data)` ：将消息放入 UDP 通道，实现局域网多节点广播（适配多服务实例）。



**步骤 3：UDP 广播（多节点同步，`broadMsg/udpSendProc/udpRecvProc`）**

​	为适配**多服务节点部署**（比如多台服务器运行聊天服务），通过 UDP 广播同步消息：

​	**1、`broadMsg(data)`**：将消息写入`udpSendChan`通道；

​	**2、`udpSendProc`（初始化时启动）**：从`udpSendChan`取消息，通过 UDP 广播到局域网（目标 IP：10.70.0.255，端口从配置读取）；

​	**3、`udpRecvProc`（初始化时启动）**：监听 UDP 端口，接收其他节点广播的消息，拿到消息后再次调用`dispatch(data)`，保证多节点消息一致。



**步骤 4：消息调度（`dispatch`函数）**

​	核心路由逻辑，根据`Message.Type`（消息类型）分发消息 —— 私聊 or 群聊

- 群聊补充（sendGroupMsg）：通过 `SearchUserByGroupId` 获取 `群 ID` 下所有用户 ID，循环调用消息发送接口（`sendMsg`）实现群消息群发（排除发送者自己）。





### （二）图片上传以及发送

1. 前端调用Upload服务，将图片上传到服务器，后端获取图片内容，并构建唯一的文件名（时间+随机数+文件后缀 或者 uuid+文件后缀），
2. 将文件保存在服务器路径后，返回图片发送成功的相应（包含图片存储路径）
3. 前端通过向后端请求上述存储路径，来获取图片内容





### （三）群聊流程及方案

1. 群聊功能前端发送消息时将群聊类型设置为对应的 **”群聊消息类型“** 
2. 后端通过群信息找到对应的群，再找到群聊里面的人，并将消息发送给群聊中的所有人



群聊设计方案：

**方案1：**

> 群聊存储方案：map \<userId> \<groupId, groupId, ......> 以用户为基准
>
> 优点：锁的频率低
>
> 缺点：需要轮询全部的map （用户量大，轮询代价太大，不合理）
>
> 伪代码：
> 	type Node struct {
> 		Conn * websocket.Conn
> 		DataQueue chan[]byte
> 		GroupSets set.Interface
> 	}
> 	var clientMap map[int64]*Node = make()....

**√ 方案2：**

> 群聊存储方案：map \<groupId> \<userId, userId, ......> 以群聊为基准
>
> 优点：查询效率更快
>
> 缺点：发送消息需要根据用户ID获取Node，锁的频率较高
>
> 代码落地：
> 	1. 新建群聊（初始化 groupSet）
> 	1. 加入群聊（刷新 groupSet）
> 	1. 分发消息（群里面的人都要收到消息）





## 项目配置

swagger ip：	http://localhost:8082/swagger/index.html

application：	http://localhost:8082/



优化：Golang项目的网关、拦截器、ThreadLocal是如何实现的？





