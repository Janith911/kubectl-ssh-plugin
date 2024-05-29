package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "ssh",
	Short: "SSH into cluster nodes",
	Args:  cobra.ExactArgs(1),
	Long:  `Can login to cluster nodes via SSH. Both public key authentication and password authentication are supported`,

	Run: func(cmd *cobra.Command, args []string) {
		user, e := cmd.Flags().GetString("user")
		kubeconfig, e := cmd.Flags().GetString("kubeconfig")
		kubecontext, e := cmd.Flags().GetString("context")
		nodeIp := getNodeIP(kubeconfig, kubecontext, args[0])
		if e != nil {
			fmt.Println("An ERROR Occurred : ", e)
		}
		port, e := cmd.Flags().GetString("port")
		if e != nil {
			fmt.Println("An ERROR Occurred : ", e)
		}
		homeDir, err := os.UserHomeDir()
		if err != nil {
			fmt.Println("An ERROR Occurred : ", err)
		}
		config := configSSHClient(user, nodeIp, homeDir+"/.ssh/id_rsa")
		sshConnect(nodeIp, port, config)
	},
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		fmt.Println("An ERROR Occurred : ", err)
	}
	rootCmd.Flags().StringP("user", "u", "ubuntu", "Specify user name")
	rootCmd.Flags().StringP("port", "p", "22", "Specify SSH port")
	rootCmd.Flags().String("kubeconfig", homeDir+"/.kube/config", "Specify Kubeconfig file")
	rootCmd.Flags().String("context", "", "Specify context")
}
