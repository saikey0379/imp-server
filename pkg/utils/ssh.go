package utils

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net"
	"os/exec"
	"strconv"
	"time"

	"golang.org/x/crypto/ssh"
)

// 连接的配置
type SSHClient struct {
	PrivateKey string `json:"private_key"`
	Address    string `json:"address"`
	Port       int    `json:"port"`
	User       string `json:"user"`
	Password   string `json:"password"`
	Client     *ssh.Client
}

func (sc *SSHClient) CreateClient() error {
	var (
		client *ssh.Client
		err    error
	)

	var signer ssh.Signer
	if sc.PrivateKey != "" {
		key, err := ioutil.ReadFile(sc.PrivateKey)
		if err != nil {
			return err
		}

		// Create the Signer for this private key.
		signer, err = ssh.ParsePrivateKey(key)
		if err != nil {
			return err
		}
	}

	//一般传入四个参数：user，[]ssh.AuthMethod{ssh.Password(password)}, HostKeyCallback，超时时间，
	config := ssh.ClientConfig{
		User: sc.User,
		Auth: []ssh.AuthMethod{ssh.Password(sc.Password), ssh.PublicKeys(signer)},
		HostKeyCallback: func(hostname string, remote net.Addr, key ssh.PublicKey) error {
			return nil
		},
		Timeout: 10 * time.Second,
	}
	addr := fmt.Sprintf("%s:%d", sc.Address, sc.Port)

	//获取client
	if client, err = ssh.Dial("tcp", addr, &config); err != nil {
		return err
	}

	sc.Client = client
	return nil
}

func (sc *SSHClient) RunShell(shell string) (output string, err error) {
	var session *ssh.Session

	//获取session，这个session是用来远程执行操作的
	if session, err = sc.Client.NewSession(); err != nil {
		return output, err
	}
	defer session.Close()

	//执行shell
	resp, err := session.CombinedOutput(shell)
	return string(resp), err
}

func (sc *SSHClient) FileSync(syncsrc, dstdir, dstfile string) (output string, err error) {
	var command string
	if dstfile == "" {
		command = fmt.Sprintf("/usr/bin/rsync -aP --rsync-path=\"mkdir -p %s && rsync\" -e \"ssh -p %s -o StrictHostKeyChecking=no -i %s\" %s %s@%s:%s", dstdir, strconv.Itoa(sc.Port), sc.PrivateKey, syncsrc, sc.User, sc.Address, dstdir)
	} else {
		command = fmt.Sprintf("/usr/bin/rsync -aP --rsync-path=\"mkdir -p %s && rsync\" -e \"ssh -p %s -o StrictHostKeyChecking=no -i %s\" %s %s@%s:%s/%s", dstdir, strconv.Itoa(sc.Port), sc.PrivateKey, syncsrc, sc.User, sc.Address, dstdir, dstfile)
	}
	cmd := exec.Command("/bin/bash", "-c", command)
	var out bytes.Buffer
	cmd.Stdout = &out
	err = cmd.Start()
	if err != nil {
		return string(out.Bytes()), err
	}
	err = cmd.Wait()
	if err != nil {
		return string(out.Bytes()), err
	}
	return string(out.Bytes()), nil
}
