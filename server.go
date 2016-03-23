// Copyright © 2016 National Data Service
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/ant0ine/go-json-rest/rest"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
)

type Federation struct {
	IcatHost       string `json:"icat_host"`
	ZoneName       string `json:"zone_name"`
	NegotiationKey string `json:"negotiation_key"`
	ZoneKey        string `json:"zone_key"`
}

type FederationRequest struct {
	User        string     `json:"user"`
	IcatAddress string     `json:"icat_address"`
	Federation  Federation `json:"federation"`
}

var zone, host string

func main() {

	var port, adminPassword string
	flag.StringVar(&host, "host", "localhost", "Server listening address")
	flag.StringVar(&port, "port", "8080", "Server listening port")
	flag.StringVar(&adminPassword, "password", "admin", "Server password")
	flag.StringVar(&zone, "zone", "tempZone", "Server password")
	flag.Parse()

	fmt.Printf("Host %s\n", host)
	fmt.Printf("Port %s\n", port)
	fmt.Printf("Zone %s\n", zone)

	api := rest.NewApi()
	api.Use(&rest.AuthBasicMiddleware{
		Realm: "nds-irods-demo",
		Authenticator: func(userId string, password string) bool {
			if userId == "admin" && password == adminPassword {
				return true
			}
			return false
		},
	})

	router, err := rest.MakeRouter(
		rest.Get("/version", Version),
		rest.Post("/federation", PostFederation),
		rest.Get("/federation", GetFederation),
		//rest.Put("/federate/:zone", PutFederation),
		//rest.Delete("/federate/:zone", DeleteFederation),
	)

	if err != nil {
		fmt.Println(err)
		return
	}
	api.SetApp(router)

	fmt.Printf("Listening on :%s\n", port)
	http.ListenAndServe(":"+port, api.MakeHandler())
}

func Version(w rest.ResponseWriter, r *rest.Request) {
	w.WriteJson(fmt.Sprintf("%s %s", VERSION, BUILD_DATE))
}

func PostFederation(w rest.ResponseWriter, r *rest.Request) {

	req := FederationRequest{}
	err := r.DecodeJsonPayload(&req)
	if err != nil {
		fmt.Errorf("Error decoding json payload: %s\n", err)
		rest.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	fed := req.Federation

	data, err := ioutil.ReadFile("/etc/irods/server_config.json")
	if err != nil {
		fmt.Errorf("Error reading server config: %s\n", err)
		rest.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	config := make(map[string]interface{})
	err = json.Unmarshal(data, &config)
	if err != nil {
		fmt.Errorf("Error unmarshaling server config: %s\n", err)
		rest.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	update := false
	federations := config["federation"]
	for _, federation := range federations.([]interface{}) {
		icatHost := (federation.(map[string]interface{})["icat_host"]).(string)
		if icatHost == fed.IcatHost {
			federation.(map[string]interface{})["zone_key"] = fed.ZoneKey
			federation.(map[string]interface{})["zone_name"] = fed.ZoneName
			federation.(map[string]interface{})["negotiation_key"] = fed.NegotiationKey
			update = true
		}
	}

	if !update {
		federations = append(federations.([]interface{}), fed)
		config["federation"] = federations
		data, err = json.MarshalIndent(config, "", "   ")
		if err != nil {
			fmt.Errorf("Error marshaling server config: %s\n", err)
			rest.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		err = ioutil.WriteFile("/etc/irods/server_config.json", data, 0700)
		if err != nil {
			fmt.Errorf("Error writing server config: %s\n", err)
			rest.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		user := req.User
		icatAddress := req.IcatAddress
		icatHost := fed.IcatHost
		zoneName := fed.ZoneName

		fmt.Printf("Updating /etc/hosts %s %s\n", icatAddress, icatHost)
		f, err := os.OpenFile("/etc/hosts", os.O_APPEND|os.O_WRONLY, 0600)
		if err != nil {
			fmt.Errorf("Error opening hosts: %s\n", err)
			rest.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		defer f.Close()

		if _, err = f.WriteString(fmt.Sprintf("%s\t%s\n", icatAddress, icatHost)); err != nil {
			fmt.Errorf("Error updating hosts: %s\n", err)
			rest.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Create the remote zone
		fmt.Printf("Creating zone %s %s:1247\n", zoneName, icatAddress)
		mkzone := exec.Command("iadmin", "mkzone", zoneName, "remote", fmt.Sprintf("%s:1247", icatAddress))
		err = mkzone.Run()
		if err != nil {
			fmt.Errorf("Error in mkzone: %s\n", err)
			rest.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Add the zone directory
		zoneDir := fmt.Sprintf("/%s/%s", zone, zoneName)
		fmt.Printf("Creating collection %s\n", zoneDir)
		mkdir := exec.Command("imkdir", zoneDir)
		err = mkdir.Run()
		if err != nil {
			fmt.Errorf("Error in imkdir: %s\n", err)
			rest.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Add the remote user
		zoneUser := fmt.Sprintf("%s#%s", user, zoneName)
		fmt.Printf("Creating remote user %s\n", zoneUser)
		mkuser := exec.Command("iadmin", "mkuser", zoneUser, "rodsuser")
		err = mkuser.Run()
		if err != nil {
			fmt.Errorf("Error in mkuser: %s\n", err)
			rest.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Grant permissions to remote user
		fmt.Printf("Granting write permissions on %s to %s\n", zoneDir, zoneUser)
		ichmod := exec.Command("ichmod", "-r", "write", zoneUser, zoneDir)
		err = ichmod.Run()
		if err != nil {
			fmt.Errorf("Error in ichmod: %s\n", err)
			rest.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
	w.WriteHeader(http.StatusOK)
}

func GetFederation(w rest.ResponseWriter, r *rest.Request) {

	data, err := ioutil.ReadFile("/etc/irods/server_config.json")
	if err != nil {
		fmt.Errorf("%s\n", err)
		return
	}
	config := make(map[string]interface{})
	err = json.Unmarshal(data, &config)
	if err != nil {
		fmt.Errorf("%s\n", err)
		return
	}

	fed := Federation{}
	fed.IcatHost = host
	fed.ZoneKey = config["zone_key"].(string)
	fed.ZoneName = config["zone_name"].(string)
	fed.NegotiationKey = config["negotiation_key"].(string)
	w.WriteJson(&fed)

	return
}
