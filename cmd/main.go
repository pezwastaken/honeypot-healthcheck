package main

import (
	"fmt"
	"github.com/spf13/viper"
	"golang.org/x/crypto/ssh"
	kh "golang.org/x/crypto/ssh/knownhosts"
	"log"
	"os"
)

type HealthCheckConf struct {
	KHPath  string     `mapstructure:"known_hosts_path"`
	LogFile string     `mapstructure:"log_file"`
	Cowrie  CowrieConf `mapstructure:"cowrie"`
}

type CowrieConf struct {
	Hosts    []string `mapstructure:"hosts"`
	Port     int      `mapstructure:"port"`
	User     string   `mapstructure:"user"`
	Password string   `mapstructure:"password"`
}

// read config file into the given struct
func readConfig(conf *HealthCheckConf) error {

	viper.SetConfigName("conf")
	viper.SetConfigType("toml")

	viper.AddConfigPath("$HOME/honeypot-healthcheck")

	err := viper.ReadInConfig()
	if err != nil {
		log.Printf("ERROR | error while reading config file: %v", err)
		return err
	}
	log.Printf("DEBUG | successfully read config file")

	if err := viper.Unmarshal(conf); err != nil {
		log.Printf("ERROR | error while unmarshaling config: %v", err)
		return err
	}
	log.Printf("DEBUG | successfully unmarshaled config")
	return nil

}

func testHoneypots(conf *HealthCheckConf) (map[string]bool, error) {

	cowrie := conf.Cowrie

	var honeypots map[string]bool = make(map[string]bool)

	hostKeyCallback, err := kh.New(conf.KHPath)
	if err != nil {
		return nil, err
	}

	config := &ssh.ClientConfig{
		User: cowrie.User,
		Auth: []ssh.AuthMethod{
			ssh.Password(cowrie.Password),
		},
		// HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		HostKeyCallback:   hostKeyCallback,
		HostKeyAlgorithms: []string{ssh.KeyAlgoED25519},
	}

	for _, host := range cowrie.Hosts {

		ok, err := connect(host, cowrie.Port, config)
		if err != nil {
			log.Printf("ERROR | couldn't connect to host %v:%v | %v", host, cowrie.Port, err)
		}
		honeypots[host] = ok
	}

	return honeypots, nil
}

func connect(host string, port int, config *ssh.ClientConfig) (bool, error) {

	remote := fmt.Sprintf("%v:%v", host, port)
	client, err := ssh.Dial("tcp", remote, config)
	if err != nil {
		return false, err
	}
	defer client.Close()

	return true, nil

}

func main() {

	var conf HealthCheckConf
	if err := readConfig(&conf); err != nil {
		panic(err)
	}

	f, err := os.OpenFile(conf.LogFile, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0660)
	if err != nil {
		log.Fatalf("Error opening log file: %v", err)
	}
	defer f.Close()

	log.SetOutput(f)
	log.Print("INFO | healthcheck started")

	h, err := testHoneypots(&conf)
	if err != nil {
		panic(err)
	}

	log.Printf("honeypots: %v", h)

}
