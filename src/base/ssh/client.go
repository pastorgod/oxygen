package ssh

import(
	"github.com/gnicod/goscplib"
	"code.google.com/p/go.crypto/ssh"
//	"crypto"
//	"crypto/rsa"
//	"crypto/x509"
//	"encoding/pem"
//	"io"
	"fmt"
	"net"
	"path/filepath"
	"io/ioutil"
	"os/user"
	"os/exec"
.	"logger"
)

func getKeyFile() (key ssh.Signer, err error) {

	usr, _ := user.Current()
	file := usr.HomeDir + "/.ssh/id_rsa"
	buf, err := ioutil.ReadFile(file)
	if err != nil {
		return
	}
	key, err = ssh.ParsePrivateKey(buf)
	if err != nil {
		return
	}
	return
}


type SshClient struct {
	client *ssh.Client
	user	string
	host	string
	port	string
}

// 免密码方式登录远程
// 192.168.1.2:22 username
func OpenSshClient( addr, user string ) (*SshClient, error) {

	host, port, err := net.SplitHostPort( addr )

	if err != nil {
		ERROR( "ssh: invalid addr, %s", err.Error() )
		return nil, err
	}

	key, err := getKeyFile()

	if err != nil {
		ERROR( "ssh: getKeyFile fail, %s", err.Error() )
		return nil, err
	}

	clientConfig := &ssh.ClientConfig {
		User : user,
		Auth : []ssh.AuthMethod {
			ssh.PublicKeys( key ),
		},
	}

	client, err := ssh.Dial("tcp", addr, clientConfig)

	if err != nil {
		ERROR( "ssh: dial fail, %s", err.Error() )
		return nil, err
	}

	return &SshClient {
		client	: client,
		user	: user,
		host	: host,
		port	: port,
	}, nil
}

// 使用特定用户名密码登录远程
// 192.168.1.2:22 username password
func OpenSshClientWithPwd( addr, user, pwd string ) (*SshClient, error) {

	host, port, err := net.SplitHostPort( addr )

	if err != nil {
		ERROR( "ssh: invalid addr, %s", err.Error() )
		return nil, err
	}

	clientConfig := &ssh.ClientConfig {
		User : user,
		Auth : []ssh.AuthMethod {
			ssh.Password( pwd ),
		},
	}

	client, err := ssh.Dial("tcp", addr, clientConfig)

	if err != nil {
		ERROR( "ssh: dial fail, %s", err.Error() )
		return nil, err
	}

	return &SshClient {
		client	: client,
		user	: user,
		host	: host,
		port	: port,
	}, nil
}

//------------------------------------------------------------------------------------------------
func(this *SshClient) Close() {
	this.client.Close()
}

// scp client.
func(this *SshClient) NewScpClient() *goscplib.Scp {
	return goscplib.NewScp( this.client )
}

// 通过scp推送文件
func(this *SshClient) ScpPushFile( src string, dest string ) error {

	client := this.NewScpClient()

	return client.PushFile( src, dest )
}

// 通过scp推送目录(缺陷: 目标有目录的时候无法传输)
func(this *SshClient) ScpPushDir( src string, dest string ) error {

	client := this.NewScpClient()

	return client.PushDir( src, dest )
}

// 通过rsync同步文件到远程主机用户主目录下
func(this *SshClient) RsyncPush( src string, dest string ) error {

	cmd := exec.Command( "rsync", "-avzqe", "ssh", src, fmt.Sprintf( "%s@%s:~/%s", this.user, this.host, filepath.Base(dest) ) )

	DEBUG( "ssh: %v", cmd.Args )

	return cmd.Run()
}

func(this *SshClient) NewSession() (*ssh.Session, error) {
	return this.client.NewSession()
}

func(this *SshClient) Output( cmd string )([]byte, error) {

	session, err := this.client.NewSession()

	if err != nil {
		ERROR( "ssh: new session fail, %s", err.Error() )
		return nil, err
	}

	defer session.Close()

	return session.Output( cmd )
}

func(this *SshClient) Run( cmd string ) error {

	session, err := this.client.NewSession()

	if err != nil {
		ERROR( "ssh: new session fail, %s", err.Error() )
		return err
	}

	defer session.Close()

	if err := session.Run( cmd ); err != nil {
		return err
	}

	return nil
}
