### 视频演示：
https://www.bilibili.com/video/BV1y24y1v7RX

### 系统简介：
本系统使用RSA算法生成密钥对， RSA私钥用于用户身份认证；用户发送的数据将存储于IPFS， IPFS返回的CID（IPFS Hash）使用用户的RSA公钥加密后存储于区块链; 区块链部分使用Hyperledger Fabric,并用Hyperledger Explorer追踪交易

### 包含功能
1. 基于Fabric v1.4.4 first-network，四个peer一个orderer节点，使用docker部署
2. IPFS使用的是ipfs/kubo镜像，负责用户数据文件的存储，IPFS返回的CID存储于Fabric
3. 项目包含了Hyperledger Explorer（区块链浏览器），默认跟随脚本启动
4. 项目包含了tape对链码压测
5. 使用RSA公私钥鉴别用户身份（1024位）
6. 链码对传输记录进行存储，包含：发送者公钥、接收者公钥、文件在IPFS的加密CID（由发送者或接收者的公钥加密）、文件名、时间戳、Fabric交易id
7. 后端使用gin框架实现，前端使用Vue和Element ui实现
   使用go fabric sdk调用智能合约；使用go-ipfs-api上传与下载用户文件；使用uuid对用户的文件名（下载时）进行加密

### 安装步骤(默认是在本地虚拟机)
1. 安装ubuntu 20.04(或其他Linux发行版),docker,docker-compose,go1.19
   docker,docker-compose,go1.19安装方法请参考此文章：https://blog.csdn.net/qq_41575489/article/details/129129086

2. 向/etc/hosts 写入：  
   ````bash
   127.0.0.1 orderer.example.com
   127.0.0.1 peer0.org1.example.com
   127.0.0.1 peer1.org1.example.com
   127.0.0.1 peer0.org2.example.com
   127.0.0.1 peer1.org2.example.com
    ````

3. 项目在服务器上运行需要操作，如果是虚拟机则省略这步。
   修改以下两个文件中127.0.0.1 为服务器公网IP：
   ````bash
   datashare/application/server/controller/controller.go
   datashare/application/web/index.html
   ````

   

4. 启动区块链部分
   ````bash
   cd blockchain
   ./start.sh
   ````
5. 启动前后端
   ````bash
   cd application/server
   go run main.go
   ````
6. 如果是云服务器
   在防火墙放行9090和8080TCP端口
7. 打开网页
   ip：9090/web

### tape测压命令：
在blockchain/tape中

./tape --config=config.yaml --number=100

# 注意：
1. 如果全部是在虚拟机内操作，不需要修改IP
2. 提示密钥不对、服务器错误请检查是否修改好hosts（步骤2）