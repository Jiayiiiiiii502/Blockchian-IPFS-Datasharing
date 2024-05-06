package controller

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"server/api"
	"server/api/rsa"
	"server/model"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	//"github.com/gofrs/uuid"
	"github.com/google/uuid"
)

type ranges struct {
	start  int64
	end    int64
	length int64
}

func GenerateKeyPair(c *gin.Context) {
	sk, pk := api.GenerateKeyPair()
	//sk:private key pk:public key
	// id := uuid.New()
	// c.SetCookie("sk_"+id.String(), sk, 30*24*60*60, "/", "", false, true)
	// c.SetCookie("pk_"+id.String(), pk, 30*24*60*60, "/", "", false, true)

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

// func Upload(c *gin.Context) {
// 	//save file
// 	file, _ := c.FormFile("file")
// 	fmt.Printf("file:%v", fmt.Sprintf("./files/uploadfiles/%v", file.Filename))
// 	err := c.SaveUploadedFile(file, fmt.Sprintf("./files/uploadfiles/%v", file.Filename))
// 	if err != nil {
// 		c.String(http.StatusBadRequest, fmt.Sprintf("file upload failed,err:%v", err))
// 		return
// 	}
// 	// upload to ipfs
// 	cid := api.IpfsAdd(file.Filename)
// 	// cid := file.Filename

// 	//assignment record
// 	var record model.Record
// 	c.ShouldBind(&record)
// 	serderpk, err := rsa.SktoPub(record.Sender)
// 	if err != nil {
// 		c.String(http.StatusBadRequest, "no sk input")
// 		return
// 	}
// 	record.SenderEncryptedCid = api.EncryptCid(cid, serderpk)
// 	record.RecevierEncryptedCid = api.EncryptCid(cid, record.Recevier)
// 	record.Filename = file.Filename
// 	record.Message = c.PostForm("message")

// 	//upload to fabric
// 	var args [][]byte
// 	args = append(args, []byte(serderpk))
// 	args = append(args, []byte(record.Recevier))
// 	args = append(args, []byte(record.SenderEncryptedCid))
// 	args = append(args, []byte(record.RecevierEncryptedCid))
// 	args = append(args, []byte(record.Filename))
// 	args = append(args, []byte(record.Message))
// 	res, err := api.ChannelExecute("sendData", args)
// 	if err != nil {
// 		c.String(http.StatusBadRequest, fmt.Sprintf("execute chaincode failed err:%v", err))
// 		return
// 	}

// 	//return txid
// 	c.JSON(http.StatusOK, gin.H{
// 		"txid": res.TransactionID,
// 	})

// }

func Upload(c *gin.Context) {
	//save file
	file, _ := c.FormFile("file")
	fmt.Printf("file:%v", fmt.Sprintf("./files/uploadfiles/%v", file.Filename))
	err := c.SaveUploadedFile(file, fmt.Sprintf("./files/uploadfiles/%v", file.Filename))
	if err != nil {
		c.String(http.StatusBadRequest, fmt.Sprintf("file upload failed,err:%v", err))
		return
	}

	// 检查是否是续传
	rangeHeader := c.GetHeader("Range")
	var offset int64
	if rangeHeader != "" {
		// 解析 Range 头
		offset, err = parseRangeHeader(rangeHeader)
		if err != nil {
			c.String(http.StatusBadRequest, "无效的 Range 头")
			return
		}
	}

	// 打开文件
	var dstFile *os.File
	if offset == 0 {
		// 如果不是续传，则创建新文件
		dstFile, err = os.Create(fmt.Sprintf("./files/uploadfiles/%v", file.Filename))
	} else {
		// 如果是续传，则以追加模式打开已存在的文件
		dstFile, err = os.OpenFile(fmt.Sprintf("./files/uploadfiles/%v", file.Filename), os.O_APPEND|os.O_WRONLY, 0644)
	}
	if err != nil {
		c.String(http.StatusInternalServerError, fmt.Sprintf("无法创建或打开文件，错误：%v", err))
		return
	}
	defer dstFile.Close()

	// 保存上传文件到本地
	srcFile, err := file.Open()
	if err != nil {
		c.String(http.StatusInternalServerError, fmt.Sprintf("无法打开上传文件，错误：%v", err))
		return
	}
	defer srcFile.Close()

	// 定位到续传的起始位置
	if offset > 0 {
		_, err = srcFile.Seek(offset, io.SeekStart)
		if err != nil {
			c.String(http.StatusInternalServerError, fmt.Sprintf("无法定位到续传的起始位置，错误：%v", err))
			return
		}
	}

	// 拷贝文件内容
	_, err = io.Copy(dstFile, srcFile)
	if err != nil {
		c.String(http.StatusInternalServerError, fmt.Sprintf("文件拷贝失败，错误：%v", err))
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
	record.Message = c.PostForm("message")

	//upload to fabric
	var args [][]byte
	args = append(args, []byte(serderpk))
	args = append(args, []byte(record.Recevier))
	args = append(args, []byte(record.SenderEncryptedCid))
	args = append(args, []byte(record.RecevierEncryptedCid))
	args = append(args, []byte(record.Filename))
	args = append(args, []byte(record.Message))
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

// 解析 Range 头
func parseRangeHeader(rangeHeader string) (int64, error) {
	parts := strings.SplitN(rangeHeader, "=", 2)
	if len(parts) != 2 || parts[0] != "bytes" {
		return 0, errors.New("无效的 Range 头")
	}

	offset, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil {
		return 0, errors.New("无效的 Range 值")
	}

	return offset, nil
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

func GetAllSKs(c *gin.Context) {
	cookies := c.Request.Cookies()
	//chooce cookies
	var skCookies []*http.Cookie
	for _, cookie := range cookies {
		if strings.Contains(cookie.Name, "sk_") {
			skCookies = append(skCookies, cookie)
		}
	}
	//return list
	c.JSON(http.StatusOK, gin.H{"skCookies": skCookies})
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
	//u1, _ := uuid.NewV4()
	u1, _ := uuid.NewRandom()
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
	c.String(http.StatusOK, fmt.Sprintf("http://1.14.74.234:9090/downloadfile?filepath=%v", enFilename))
}

func DownloadFile(c *gin.Context) {
	enFilename := c.Query("filepath")
	//uuid 36
	filename := enFilename[36:]
	c.Writer.Header().Add("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
	_, err := os.Stat(fmt.Sprintf("./files/downloadfiles/%v", enFilename))
	if err != nil {
		c.String(http.StatusBadRequest, "File has been deleted! Please reobtain the file link.")
		return
	}
	//return file
	c.File(fmt.Sprintf("./files/downloadfiles/%v", enFilename))
	//delete file
	// os.Remove(fmt.Sprintf("./files/downloadfiles/%v", enFilename))
}
