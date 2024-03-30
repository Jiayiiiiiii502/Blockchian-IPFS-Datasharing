package controller

import (
	"fmt"
	"net/http"
	"os"
	"server/api"
	"server/api/rsa"
	"server/model"

	"github.com/gin-gonic/gin"
	"github.com/gofrs/uuid"
)

func GenerateKeyPair(c *gin.Context) {
	sk, pk := api.GenerateKeyPair()
	//sk:private key pk:public key
	c.JSON(http.StatusOK, gin.H{"sk": sk, "pk": pk})

}
func RestoreKey(c *gin.Context) {
	//convert sk
	pk, err := rsa.SktoPub(c.PostForm("sk"))
	if err != nil {
		c.String(http.StatusOK, "no sk input")
		return
	}
	c.String(http.StatusOK, pk)

}
func Upload(c *gin.Context) {
	//save file
	file, _ := c.FormFile("file")
	fmt.Printf("file:%v", fmt.Sprintf("./files/uploadfiles/%v", file.Filename))
	err := c.SaveUploadedFile(file, fmt.Sprintf("./files/uploadfiles/%v", file.Filename))
	if err != nil {
		c.String(http.StatusBadRequest, fmt.Sprintf("file upload failed,err:%v", err))
		return
	}
	// upload to ipfs
	cid := api.IpfsAdd(file.Filename)
	// cid := file.Filename

	//assignment record
	var record model.Record
	c.ShouldBind(&record)
	serderpk, err := rsa.SktoPub(record.Sender)
	if err != nil {
		c.String(http.StatusBadRequest, "no sk input")
		return
	}
	record.SenderEncryptedCid = api.EncryptCid(cid, serderpk)
	record.RecevierEncryptedCid = api.EncryptCid(cid, record.Recevier)
	record.Filename = file.Filename

	//upload to fabric
	var args [][]byte
	args = append(args, []byte(serderpk))
	args = append(args, []byte(record.Recevier))
	args = append(args, []byte(record.SenderEncryptedCid))
	args = append(args, []byte(record.RecevierEncryptedCid))
	args = append(args, []byte(record.Filename))
	res, err := api.ChannelExecute("sendData", args)
	if err != nil {
		c.String(http.StatusBadRequest, fmt.Sprintf("execute chaincode failed err:%v", err))
		return
	}

	//return txid
	c.JSON(http.StatusOK, gin.H{
		"txid": res.TransactionID,
	})

}

func GetRecords(c *gin.Context) {
	//convert sk
	pk, err := rsa.SktoPub(c.PostForm("sk"))
	if err != nil {
		c.String(http.StatusBadRequest, "no sk input")
		return
	}
	//query chaincode
	var args [][]byte
	args = append(args, []byte(pk))
	res, err := api.ChannelQuery("queryRecord", args)
	if err != nil {
		c.String(http.StatusBadRequest, fmt.Sprintf("execute chaincode failed err:%v", err))
		return
	}
	//retun res
	c.String(http.StatusOK, string(res.Payload))
}

func GetFile(c *gin.Context) {
	//convert sk
	pk, err := rsa.SktoPub(c.PostForm("sk"))
	if err != nil {
		c.String(http.StatusBadRequest, "no sk input")
		return
	}
	//ecid:encrypted cid
	ecid := c.PostForm("ecid")
	filename := c.PostForm("filename")
	if pk == "" {
		c.String(http.StatusBadRequest, "no sk input")
		return
	}
	cid := api.DecryptCid(ecid, c.PostForm("sk"))
	//generate uuid
	u1, _ := uuid.NewV4()
	//encryptedfilename
	enFilename := fmt.Sprintf("%v%v", u1, filename)
	// fmt.Printf("sk:%v,ecid:%v,cid:%v,pk:%v", c.PostForm("sk"), ecid, cid, pk)
	//get file from ipfs
	err = api.IpfsGet(cid, enFilename)
	if err != nil {
		c.String(http.StatusBadGateway, fmt.Sprintf("%v", err))
		return
	}
	//return filepath
	c.String(http.StatusOK, fmt.Sprintf("http://127.0.0.1:9090/downloadfile?filepath=%v", enFilename))
}

func DownloadFile(c *gin.Context) {
	enFilename := c.Query("filepath")
	//uuid 36
	filename := enFilename[36:]
	c.Writer.Header().Add("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
	_, err := os.Stat(fmt.Sprintf("./files/downloadfiles/%v", enFilename))
	if err != nil {
		c.String(http.StatusBadRequest, "文件已删除！请重新获取链接！")
		return
	}
	//return file
	c.File(fmt.Sprintf("./files/downloadfiles/%v", enFilename))
	//delete file
	// os.Remove(fmt.Sprintf("./files/downloadfiles/%v", enFilename))
}
