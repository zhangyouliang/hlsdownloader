package test

import (
	"github.com/containous/traefik/log"
	"io/ioutil"
	"os/exec"
)

func main() {

	cmd := exec.Command("/bin/bash", "-c", " which openssl")
	stdout,err := cmd.StdoutPipe()
	if err != nil {
		log.Fatal(err)
	}
	defer stdout.Close()
	if err:= cmd.Start();err!=nil{
		log.Fatal(err)
	}
	// 读取输出结果
	opBytes, err := ioutil.ReadAll(stdout)
	if err != nil {
		log.Fatal(err)
	}
	if len(opBytes) == 0 {
		log.Println("not found openssl")
	}
	log.Println("found openssl")

}
