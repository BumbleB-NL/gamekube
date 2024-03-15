package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/PimSanders/golang-zerotier-api/golangzerotierapi"
	"gopkg.in/yaml.v2"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/cli"
)

type apiNewServiceData struct {
	UserId        string `json:"userid"`
	ServerService string `json:"serverservice"`
}

type apiUpdateNetworkMember struct {
	NetworkId   string `json:"networkid"`
	UserId      string `json:"userid"`
	Hidden      bool   `json:"hidden"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Config      struct {
		ActiveBridge    bool     `json:"activeBridge"`
		Authorized      bool     `json:"authorized"`
		Capabilities    []int    `json:"capabilities"`
		IPAssignments   []string `json:"ipAssignments"`
		NoAutoAssignIps bool     `json:"noAutoAssignIps"`
		Tags            [][]int  `json:"tags"`
	}
}

var (
	insecure   = true
	ztApiToken = "A4gsvHsXQhKD8qgFikHnqzCNxyzNowrj"
	zt         *golangzerotierapi.Client
	serviceIp  = "10.200.0.0/24"
)

func main() {
	zt = golangzerotierapi.NewClient("https://api.zerotier.com/api/v1", ztApiToken, true)
	handleRequests()
}

func apiNewService(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusBadRequest)
		return
	}

	var requestBody apiNewServiceData

	err = json.Unmarshal(body, &requestBody)
	if err != nil {
		http.Error(w, "Failed to unmarshal JSON", http.StatusBadRequest)
		return
	}
	userId := requestBody.UserId
	serverService := requestBody.ServerService

	if userId == "" || serverService == "" {
		http.Error(w, "Empty field detected", http.StatusBadRequest)
		return
	}

	services := readYamlKeyValue()

	if _, ok := services[serverService]; !ok {
		http.Error(w, "Not a valid service option", http.StatusBadRequest)
		return
	}

	if !installHelmChart(userId, serverService) {
		http.Error(w, "Error installing HELM chart", http.StatusInternalServerError)
		return
	}
	if !networkExists(zt, userId) {
		createNetwork(zt, userId)
	}
	jsonResponse, _ := json.Marshal(requestBody)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(jsonResponse)
	fmt.Println("Endpoint Hit: NewService")
}

func authorizeUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusBadRequest)
		return
	}

	var requestBody apiUpdateNetworkMember
	userId := requestBody.UserId
	networkId := requestBody.NetworkId
	if userId == "" || networkId == "" {
		http.Error(w, "Empty field detected", http.StatusBadRequest)
		return
	}
	var userSettings golangzerotierapi.UpdateNetworkMember
	userSettings.Config.Authorized = true
	userSettings.Config.Tags = [][]int{{0, 0}}
	err = json.Unmarshal(body, &requestBody)
	if err != nil {
		http.Error(w, "Failed to unmarshal JSON", http.StatusBadRequest)
		return
	}
	if networkExists(zt, userId) {
		zt.UpdateNetworkMember(networkId, userId, userSettings)
	}

}

func handleRequests() {
	http.HandleFunc("/newserver", apiNewService)
	http.HandleFunc("/authuser", authorizeUser)
	log.Fatal(http.ListenAndServe(":80", nil))
}

func installHelmChart(username string, kubeService string) (success bool) {
	settings := cli.New()
	settings.KubeConfig = "/mnt/kubeconfig/config"
	settings.KubeContext = "p1-k1-cluster"
	settings.SetNamespace("kubeuser-" + username)
	settings.KubeInsecureSkipTLSVerify = insecure

	actionConfig := new(action.Configuration)
	if err := actionConfig.Init(settings.RESTClientGetter(), settings.Namespace(), os.Getenv("HELM_DRIVER"), log.Printf); err != nil {
		log.Printf("%+v", err)
		os.Exit(1)
	}

	client := action.NewInstall(actionConfig)
	client.ReleaseName = kubeService + strconv.FormatInt(time.Now().UTC().UnixNano(), 10)
	client.Namespace = "kubeuser-" + username
	client.CreateNamespace = true
	client.InsecureSkipTLSverify = insecure
	// client.Version = "1.2.5"

	services := readYamlKeyValue()

	if _, ok := services[kubeService]; !ok {
		return ok
	}

	chrt_path, err := client.LocateChart(services[kubeService], settings)
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
	return true
}

func readYamlKeyValue() map[string]string {
	// Read YAML file
	yamlFile, err := os.ReadFile("/mnt/kubeconfig/kubeservices")
	if err != nil {
		log.Fatalf("Failed to read YAML file: %v", err)
	}

	// Define a struct to unmarshal YAML data into
	var data map[string]string

	// Unmarshal YAML data into the struct
	if err := yaml.Unmarshal(yamlFile, &data); err != nil {
		log.Fatalf("Failed to unmarshal YAML: %v", err)
	}

	// Iterate over key-value pairs
	for key, value := range data {
		fmt.Printf("Key: %s, Value: %v\n", key, value)
	}

	return data
}

func networkExists(zt *golangzerotierapi.Client, userID string) bool {
	nets, err := zt.GetNetworkList()

	if err != nil {
		log.Panic(err)
	}

	for k := range nets {
		if nets[k].Config.Name == userID {
			return true
		}
	}

	return false
}

func createNetwork(zt *golangzerotierapi.Client, userID string) {
	//userNetExists := false

	network, err := zt.CreateNetwork()
	if err != nil {
		log.Panic(err)
	}

	fmt.Println(network.ID, network.Description)
	//userNetExists = true

	var update golangzerotierapi.UpdateNetwork
	var updateConfig golangzerotierapi.NetworkConfig

	route1 := struct {
		Target string `json:"target"`
		Via    string `json:"via,omitempty"`
	}{Target: serviceIp,
		Via: "10.0.0.1/24"}

	route2 := struct {
		Target string `json:"target"`
		Via    string `json:"via,omitempty"`
	}{Target: "10.0.0.0/24"}

	networkassignmentpool := struct {
		IPRangeStart string `json:"ipRangeStart"`
		IPRangeEnd   string `json:"ipRangeEnd"`
	}{
		IPRangeStart: "10.0.0.2",
		IPRangeEnd:   "10.0.0.254",
	}

	//update.Description = userID
	updateConfig.Routes = append(update.Config.Routes, route1, route2)
	updateConfig.IPAssignmentPools = append(update.Config.IPAssignmentPools, networkassignmentpool)
	updateConfig.Private = true
	updateConfig.MulticastLimit = 32
	updateConfig.Name = userID
	updateConfig.V4AssignMode = struct {
		Zt bool `json:"zt"`
	}{
		Zt: true,
	}

	update.Config = updateConfig

	updateNetwork, err := zt.UpdateNetwork(network.ID, update)
	if err != nil {
		log.Panic(err)
	}

	fmt.Println(updateNetwork)
}
