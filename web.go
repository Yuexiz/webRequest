package main

import (
	"encoding/json"
	"fmt"
	"github.com/garyburd/redigo/redis"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
)

type User struct {
	Version         string `json:"version"`
	NowDownloadLink string `json:"nowDownloadLink"`
}

func main() {
	http.Handle("/", http.FileServer(http.Dir("./")))

	http.HandleFunc("/encode", func(w http.ResponseWriter, r *http.Request) {
		version := getInfoFromRedis("version")
		fileName := getInfoFromRedis("zip-name")
		xiaoyou := User{
			Version:         version,
			NowDownloadLink: "http://localhost:30002/" + fileName,
		}

		json.NewEncoder(w).Encode(xiaoyou)
	})

	http.HandleFunc("/uploadfile", index)
	http.HandleFunc("/upload", upload)
	http.HandleFunc("/uid", uidPage)
	http.HandleFunc("/uidpage", uid)
	http.ListenAndServe(":30002", nil)
}

func getUidInfoFromRedis(key int) ([]byte) {
	c, err := redis.Dial("tcp", "127.0.0.1:6379")
	if err != nil {
		fmt.Println("Connect to redis error", err)
		return []byte("")
	}
	defer c.Close()

	valueGet, e := redis.Bytes(c.Do("GET", key))
	if e != nil {
		return []byte("")
	}
	return valueGet
}

func getInfoFromRedis(key string) string {
	c, err := redis.Dial("tcp", "127.0.0.1:6379")
	if err != nil {
		fmt.Println("Connect to redis error", err)
		return "redis wrong, plz contact Yuexiz,hahahaha"
	}
	defer c.Close()
	strName, e := redis.String(c.Do("GET", key))
	if e != nil {
		return "never give this key a value"
	}
	return strName
}

func writeInfoToRedis(key string, value string) {

	c, err := redis.Dial("tcp", "127.0.0.1:6379")
	if err != nil {
		fmt.Println("Connect to redis error", err)
		return
	}
	defer c.Close()

	_, err2 := c.Do("SET", key, value)

	if err2 != nil {
		fmt.Println("存数据失败", err2)
	}

}

func upload(w http.ResponseWriter, r *http.Request) {
	r.ParseMultipartForm(32 << 20)
	file, handler, err := r.FormFile("uploadfile")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer file.Close()
	f, err := os.OpenFile(handler.Filename, os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer f.Close()
	io.Copy(f, file)
	version := r.PostFormValue("filename")
	writeInfoToRedis("version", version)
	writeInfoToRedis("zip-name", handler.Filename)
	fmt.Fprintln(w, "upload ok!")
}
func uid(w http.ResponseWriter, r *http.Request) {
	r.ParseMultipartForm(32 << 20)
	uid := r.PostFormValue("uid")

	int,err:=strconv.Atoi(uid)
	if err != nil {
		fmt.Fprintln(w, "参数肯定是错了,重新来吧,谢谢")
	}
	uidJson := string(getUidInfoFromRedis(int))
	uidJson = strings.Replace(uidJson, ",", ",\n  ", -1)
	if len(uidJson) > 0 {
		fmt.Fprintln(w, uidJson)
	} else {
		fmt.Fprintln(w, "没有这个uid的崩溃,感谢使用,再见!")
	}

}

func index(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte(tpl))
}

func uidPage(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte(checkUid))
}

const tpl = `<html>
<head>
<title>来啊,快活啊~</title>
</head>
<body>
<form enctype="multipart/form-data" action="/upload" method="post">
 <label for="a">请输入准确的版本号 :(例如1.0.5)</label>
 <input type="text" name="filename" /><br><br>
 <label for="b">请选择待上传文件 :</label>
 <input type="file" name="uploadfile" /><br><br>
 <label for="c">点击开始上传 :</label>
 <input type="hidden" name="token" value="{...{.}...}"/>
 <input type="submit" value="upload" />
</form>
</body>
</html>`

const checkUid = `<html>
<head>
<title>输入uid获取崩溃信息~</title>
</head>
<body>
<form enctype="multipart/form-data" action="/uidpage" method="post">
 <label for="a">输入uid以检索崩溃 :</label>
 <input type="text" name="uid" />
 <input type="submit" value="搜索" />
</form>
</body>
</html>`
