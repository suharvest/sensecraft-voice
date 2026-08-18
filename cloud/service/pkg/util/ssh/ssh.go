package ssh

import (
	"fmt"
	"time"

	"golang.org/x/crypto/ssh"

	"github.com/YOUR-ORG/sensecraft-voice/cloud/service/pkg/types"
)

func NewSSHClient(sshConfig *types.WebSSHRequest) (*ssh.Client, error) {
	// TODO：利用 gin 的解析，直接设置默认值
	port := sshConfig.Port
	if port == 0 {
		port = 22
	}

	// TODO 补充支持 PrivateKey 场景
	return ssh.Dial("tcp", fmt.Sprintf("%s:%d", sshConfig.Host, port), &ssh.ClientConfig{
		Timeout:         time.Second * 5,
		User:            sshConfig.User,
		Auth:            []ssh.AuthMethod{ssh.Password(sshConfig.Password)},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(), // 忽略 know_hosts 检查
	})
}
