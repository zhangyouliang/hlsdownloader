package main

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
)
func AesEncrypt(encodeStr,iv string, key []byte) (string, error) {
	encodeBytes := []byte(encodeStr)
	//根据key 生成密文
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	blockSize := block.BlockSize()
	encodeBytes = PKCS5Padding(encodeBytes, blockSize)

	blockMode := cipher.NewCBCEncrypter(block, []byte(iv))
	crypted := make([]byte, len(encodeBytes))
	blockMode.CryptBlocks(crypted, encodeBytes)

	return base64.StdEncoding.EncodeToString(crypted), nil
}

func PKCS5Padding(ciphertext []byte, blockSize int) []byte {
	padding := blockSize - len(ciphertext)%blockSize
	//填充
	padtext := bytes.Repeat([]byte{byte(padding)}, padding)

	return append(ciphertext, padtext...)
}

func AesDecrypt(decodeStr,iv string, key []byte) ([]byte, error) {
	//先解密base64
	decodeBytes, err := base64.StdEncoding.DecodeString(decodeStr)
	if err != nil {
		return nil, err
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	fmt.Printf("iv:%+v,blockSize:%+v \n",len([]byte(iv)),block.BlockSize())
	blockMode := cipher.NewCBCDecrypter(block, []byte(iv))
	origData := make([]byte, len(decodeBytes))

	blockMode.CryptBlocks(origData, decodeBytes)
	origData = PKCS5UnPadding(origData)
	return origData, nil
}

func PKCS5UnPadding(origData []byte) []byte {
	length := len(origData)
	unpadding := int(origData[length-1])
	return origData[:(length - unpadding)]
}
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

//https://www.jianshu.com/p/7790ca1bc8f6
func main() {
	f, err := os.Open("1.key")
	if err!=nil {
		panic(err)
	}
	_key := make([]byte,16)
	f.Read(_key)
	key := hex.EncodeToString(_key)
	// []byte -> String
	fmt.Printf("key:%+v \n",key)
	// String -> []byte
	//test, _ := hex.DecodeString(encodedStr)
	//fmt.Println(bytes.Compare(test, src)) // 0
	iv:=fmt.Sprintf("%032x",4)
	fmt.Printf("iv:%+v \n",iv)
	es, err := ioutil.ReadFile("bbc-4-v1-a1.ts")
	if err != nil {
		fmt.Print(err,es)
	}
	// openssl aes-128-cbc -d -in bbc-4-v1-a1.ts -iv 00000000000000000000000000000004 -K fe61b09b14011c554e30d21f7995dc12 | openssl base64
	cmd := fmt.Sprintf("openssl aes-128-cbc -d -iv %s -K %s | openssl base64",iv,key)
	out, err:= exec_shell(cmd,*bytes.NewBuffer(es))
	decodeString, err := base64.StdEncoding.DecodeString(out)
	checkErr(err)

	outF, err := os.Create("2.ts")
	if err != nil {
		fmt.Print(err)
	}
	outF.Write(decodeString)
	outF.Close()

	//decrypt, err := AesDecrypt(base64.StdEncoding.EncodeToString(es), iv, key)
	//if err != nil {
	//	fmt.Print(err)
	//}
	//out, err := os.Create("2.ts")
	//if err != nil {
	//	fmt.Print(err)
	//}
	//out.Write(decrypt)
	//out.Close()
	//
	//ciphertext, err := ioutil.ReadFile("hello_en.txt")
	//if err != nil {
	//	fmt.Print(err)
	//}
	//block, err := aes.NewCipher(key)
	//if err != nil {
	//	panic(err)
	//}
	//// The IV needs to be unique, but not secure. Therefore it's common to
	//// include it at the beginning of the ciphertext.
	//if len(ciphertext) < aes.BlockSize {
	//	panic("ciphertext too short")
	//}
	////fmt.Println(iv)
	////fmt.Println(key)
	////fmt.Println(ciphertext)
	////ciphertext = ciphertext[aes.BlockSize:]
	////fmt.Println(ciphertext)
	//
	//// CBC mode always works in whole blocks.
	//if len(ciphertext)%aes.BlockSize != 0 {
	//	panic("ciphertext is not a multiple of the block size")
	//}
	//
	//mode := cipher.NewCBCDecrypter(block, []byte(iv))
	//// CryptBlocks can work in-place if the two arguments are the same.
	//mode.CryptBlocks(ciphertext, ciphertext)
	//
	//fmt.Println(mode)
	//fmt.Printf("%s\n", ciphertext)
	//
	////openssl base64 -in 1.ts -out ls.b64
	//fmt.Println(base64.StdEncoding.EncodeToString(ciphertext) == base64.StdEncoding.EncodeToString(decrypt))
	//fmt.Println(base64.StdEncoding.EncodeToString(decrypt))
}
//$ strkey=$(hexdump -v -e '16/1 "%02x"' 1.key)
//$ iv=$(printf '%032x' 1)
//$ openssl aes-128-cbc -d -in media_0.ts -out media_decryptd_0.ts -nosalt -iv $iv -K $strkey
// openssl aes-128-cbc -d -in hello_en.txt -out hello_de.txt  -nosalt -iv $iv -K $strkey
// openssl aes-128-cbc -d -in hello_en.txt -out hello_de.txt  -nosalt -iv $iv -K $strkey