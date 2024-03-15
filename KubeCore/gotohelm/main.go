package main

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/cli"
)

var (
	namespace = "test"
)

func main() {
	installHelmChart()
}

func installHelmChart() {
	settings := cli.New()
	settings.KubeConfig = "/mnt/kubeconfig/config"
	settings.KubeContext = "p1-k1-cluster"
	settings.SetNamespace(namespace)
	settings.KubeInsecureSkipTLSVerify = true

	actionConfig := new(action.Configuration)
	if err := actionConfig.Init(settings.RESTClientGetter(), settings.Namespace(), os.Getenv("HELM_DRIVER"), log.Printf); err != nil {
		log.Printf("%+v", err)
		os.Exit(1)
	}

	client := action.NewInstall(actionConfig)
	client.ReleaseName = "gotohelm-" + strconv.FormatInt(time.Now().UTC().UnixNano(), 10)
	client.Namespace = namespace
	client.CreateNamespace = true
	client.InsecureSkipTLSverify = true
	// client.Version = "1.2.5"

	chrt_path, err := client.LocateChart("https://raw.githubusercontent.com/BumbleB-NL/gamekube/main/tgz_files/gamekube-dvwa.tgz", settings)
	if err != nil {
		panic(err)
	}

	myChart, err := loader.Load(chrt_path)
	if err != nil {
		panic(err)
	}

	releaseSjaak, err := client.Run(myChart, nil)
	if err != nil {
		panic(err)
	}

	log.Println(releaseSjaak)

	for {
		fmt.Println("Waiting...")
		time.Sleep(1 * time.Minute)
	}
}
