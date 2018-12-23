package main

import (
	"bytes"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
)
//错误处理函数
func checkErr(err error) {
	if err != nil {
		fmt.Println(err)
		panic(err)
	}
}
//阻塞式的执行外部shell命令的函数,等待执行完毕并返回标准输出
func exec_shell(s string,in bytes.Buffer) (string, error){
	//函数返回一个*Cmd，用于使用给出的参数执行name指定的程序
	cmd := exec.Command("/bin/bash", "-c", s)

	//读取io.Writer类型的cmd.Stdout，再通过bytes.Buffer(缓冲byte类型的缓冲器)将byte类型转化为string类型(out.String():这是bytes类型提供的接口)
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stdin = &in

	//Run执行c包含的命令，并阻塞直到完成。  这里stdout被取出，cmd.Wait()无法正确获取stdin,stdout,stderr，则阻塞在那了
	err := cmd.Run()
	checkErr(err)

	return out.String(), err
}

func main() {
	f, err := os.Open("1.key")
	checkErr(err)
	_key := make([]byte,16)
	f.Read(_key)
	key := hex.EncodeToString(_key)

	iv:=fmt.Sprintf("%032x",1)
	log.Printf("key:%+v \n",key)
	log.Printf("iv:%+v \n",iv)
	es, err := ioutil.ReadFile("1.ts")
	checkErr(err)
	// openssl aes-128-cbc -d -in 1.ts -iv 00000000000000000000000000000001 -K fe61b09b14011c554e30d21f7995dc12 | openssl base64
	cmd := fmt.Sprintf("openssl aes-128-cbc -d -iv %s -K %s | openssl base64",iv,key)
	out, err:= exec_shell(cmd,*bytes.NewBuffer(es))
	decodeString, err := base64.StdEncoding.DecodeString(out)
	checkErr(err)

	outF, err := os.Create("out_01.ts")
	checkErr(err)
	outF.Write(decodeString)
	outF.Close()
}
