package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	"github.com/spf13/viper"
	"golang.org/x/crypto/ssh"
	kh "golang.org/x/crypto/ssh/knownhosts"
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
		Timeout:           7 * time.Second,
	}

	var wg sync.WaitGroup

	for _, host := range cowrie.Hosts {
		remote := host
		wg.Add(1)

		go func() {
			defer wg.Done()
			ok, err := connect(remote, cowrie.Port, config)
			if err != nil {
				log.Printf("ERROR | couldn't connect to host %v:%v | %v", remote, cowrie.Port, err)
			}
			honeypots[remote] = ok
		}()
	}
	wg.Wait()

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

type Result struct {
	Active int
	Total  int
	Info   map[string]string
}

func generateResponse(m *map[string]bool) Result {

	nHoneypots := len(*m)
	var activeHoneypots int
	for _, v := range *m {
		if v {
			activeHoneypots++
		}
	}

	info := make(map[string]string, nHoneypots)

	for k, v := range *m {
		switch v {
		case true:
			info[k] = "cowrie active"
		case false:
			info[k] = "cowrie not active"
		}
	}
	result := Result{
		Active: activeHoneypots,
		Total:  nHoneypots,
		Info:   info,
	}
	return result

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

	result := generateResponse(&h)
	data, err := json.Marshal(result)
	if err != nil {
		log.Printf("error while marshaling result: %v", err)
		panic(err)
	}
	log.SetFlags(0)
	log.Println(string(data))

}
