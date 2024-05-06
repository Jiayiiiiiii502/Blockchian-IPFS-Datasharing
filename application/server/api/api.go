package api

import (
	"fmt"
	"os"
	"server/api/rsa"

	"github.com/hyperledger/fabric-sdk-go/pkg/client/channel"
	"github.com/hyperledger/fabric-sdk-go/pkg/core/config"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk"
	shell "github.com/ipfs/go-ipfs-api"
)

var (
	sdk           *fabsdk.FabricSDK                                              // Fabric SDK
	configPath    = "./api/config.yaml"                                          // 配置文件路径
	channelName   = "mychannel"                                                  // 通道名称
	user          = "Admin"                                                      // 用户
	chainCodeName = "datashare"                                                  // 链码名称
	endpoints     = []string{"peer0.org1.example.com", "peer0.org2.example.com"} // 要发送交易的节点
	sh            = shell.NewShell("127.0.0.1:6001")                             //ipfs api
)

// Initiation
func Init() {

	var err error
	//Initiation SDK by config file
	sdk, err = fabsdk.New(config.FromFile(configPath))
	if err != nil {
		panic(err)
	}

}

func GenerateKeyPair() (string, string) {
	rsakey, _ := rsa.GenerateRsaKeyBase64(1024)
	return rsakey.PrivateKey, rsakey.PublicKey
}

// use public key to encrypt IPFS CID and return the encrypted strings
func EncryptCid(cid string, publicKey string) string {
	//ecid:encrypted cid
	ecid, err := rsa.RsaEncryptToBase64([]byte(cid), publicKey)
	if err != nil {
		return err.Error()
	}
	return ecid
}

// use private key to decript IPFS CID and return the decrypted strings
func DecryptCid(ecid string, privatekey string) string {
	// fmt.Printf("================ecid:%v,sk:%v", ecid, privatekey)
	cid, err := rsa.RsaDecryptByBase64(ecid, privatekey)
	if err != nil {
		return err.Error()
	}
	return string(cid)
}

// upload file to IPFS network and return the IPFS CID
func IpfsAdd(filename string) string {
	ipfsfile, _ := os.Open(fmt.Sprintf("./files/uploadfiles/%v", filename))
	defer ipfsfile.Close()
	cid, err := sh.Add(ipfsfile)
	if err != nil {
		return fmt.Sprintf("file upload failed,err:%v", err)
	}
	return cid
}

// Get file by IPFS CID in IPFS network
func IpfsGet(cid string, filename string) error {
	// fmt.Printf("cid:%v,filename:%v", cid, filename)
	err := sh.Get(cid, fmt.Sprintf("./files/downloadfiles/%v", filename))
	if err != nil {
		fmt.Printf("err:%v", err)
		return fmt.Errorf("ipfs get file faile,err:%v", err)
	}
	return nil

}

// ChannelExecute interaction of blockchain
func ChannelExecute(fcn string, args [][]byte) (channel.Response, error) {
	// create client end and add to the channel
	ctx := sdk.ChannelContext(channelName, fabsdk.WithUser(user))
	cli, err := channel.New(ctx)
	if err != nil {
		return channel.Response{}, err
	}
	//write to blockchain digital ledger, using Invoke function in chaincode
	resp, err := cli.Execute(channel.Request{
		ChaincodeID: chainCodeName,
		Fcn:         fcn,
		Args:        args,
	}, channel.WithTargetEndpoints(endpoints...))
	if err != nil {
		return channel.Response{}, err
	}
	// only return results from chaincode
	return resp, nil
}

// ChannelQuery Query on blockchain
func ChannelQuery(fcn string, args [][]byte) (channel.Response, error) {
	// create client end and add to the channel
	ctx := sdk.ChannelContext(channelName, fabsdk.WithUser(user))
	cli, err := channel.New(ctx)
	if err != nil {
		return channel.Response{}, err
	}
	// Query operation on blockchain ledger, using Invoke function in chaincode
	resp, err := cli.Query(channel.Request{
		ChaincodeID: chainCodeName,
		Fcn:         fcn,
		Args:        args,
	}, channel.WithTargetEndpoints(endpoints...))
	if err != nil {
		return channel.Response{}, err
	}
	// only return results from chaincode
	return resp, nil
}
