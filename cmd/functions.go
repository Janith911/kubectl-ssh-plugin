package cmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"

	"golang.org/x/crypto/ssh"
	"golang.org/x/term"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

func configSSHClient(user string, ip string, pvtkey string) ssh.ClientConfig {
	key, _ := os.ReadFile(pvtkey)
	signer, _ := ssh.ParsePrivateKey(key)

	config := ssh.ClientConfig{
		User: user,
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(signer),
			ssh.PasswordCallback(func() (secret string, err error) {
				fmt.Print(user + "@" + ip + ":" + "'s password:")
				originalState, _ := term.GetState(int(os.Stdin.Fd()))
				defer term.Restore(int(os.Stdin.Fd()), originalState)
				sigch := make(chan os.Signal, 1)
				signal.Notify(sigch, os.Interrupt)
				go func() {
					for _ = range sigch {
						term.Restore(int(os.Stdin.Fd()), originalState)
						os.Exit(1)
					}
				}()
				bs, e := term.ReadPassword(int(os.Stdin.Fd()))
				fmt.Println("")
				if e != nil {
					fmt.Println(e)
				}
				return string(bs), nil
			}),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	return config

}

func sshConnect(ip string, port string, config ssh.ClientConfig) {
	client, err := ssh.Dial("tcp", ip+":"+port, &config)

	defer func() {
		if r := recover(); r != nil {
			fmt.Println("An ERROR Occurred : ", err)
			os.Exit(1)
		}
	}()

	session, err := client.NewSession()

	defer client.Close()

	modes := ssh.TerminalModes{
		ssh.ECHO:          1,
		ssh.TTY_OP_ISPEED: 14400,
		ssh.TTY_OP_OSPEED: 14400,
	}

	session.Stdin = os.Stdin
	session.Stdout = os.Stdout
	session.Stderr = os.Stderr

	if term.IsTerminal(int(os.Stdin.Fd())) {

		originalState, err := term.MakeRaw(int(os.Stdin.Fd()))
		if err != nil {
			fmt.Println("An ERROR Occurred : ", err)
		}

		defer term.Restore(int(os.Stdin.Fd()), originalState)

		termWidth, termHeight, err := term.GetSize(int(os.Stdin.Fd()))
		if err != nil {
			fmt.Println("An ERROR Occurred : ", err)
		}

		e := session.RequestPty("xterm-256color", termHeight, termWidth, modes)
		if e != nil {
			fmt.Println("An ERROR Occurred : ", e)
		}
	}
	session.Shell()
	session.Wait()
}

func getNodeIP(kubeconfigPath string, Kubecontext string, nodeName string) string {
	cfg, err := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		&clientcmd.ClientConfigLoadingRules{
			ExplicitPath: kubeconfigPath,
		},
		&clientcmd.ConfigOverrides{
			CurrentContext: Kubecontext,
		}).ClientConfig()
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("An ERROR Occurred : ", err)
			os.Exit(1)
		}
	}()

	clientset, err := kubernetes.NewForConfig(cfg)
	nodeInterface := clientset.CoreV1().Nodes()
	node, err := nodeInterface.Get(context.TODO(), nodeName, v1.GetOptions{})
	return node.Status.Addresses[0].Address
}
