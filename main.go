package main

import (
	"context"
	"fmt"
	"log"
	"os"

	cloudflare "github.com/cloudflare/cloudflare-go"
)

func BoolPointer(b bool) *bool {
	return &b
}

func main() {
	// Construct a new API object using a global API key
	//api, err := cloudflare.New(os.Getenv("TF_VAR_CLOUDFLARE_API_KEY"), os.Getenv("TF_VAR_CLOUDFLARE_API_EMAIL"))
	// alternatively, you can use a scoped API token
	accountIdentifier := cloudflare.AccountIdentifier(os.Getenv("TF_VAR_CLOUDFLARE_ACCOUNT"))
	zoneIdentifier := cloudflare.ZoneIdentifier(os.Getenv("TF_VAR_CLOUDFLARE_ZONE"))
	api, err := cloudflare.NewWithAPIToken(os.Getenv("TF_VAR_CLOUDFLARE_TOKEN"))
	apiCode := "api302ac9c335f4"
	//api.Debug = true
	if err != nil {
		log.Fatal(err)
	}

	// Fetch user details on the account
	//ident := cloudflare.ZoneIdentifier(os.Getenv("TF_VAR_CLOUDFLARE_ZONE"))
	/*
		list, _, err := api.ListTunnels(ctx, cloudflare.AccountIdentifier(os.Getenv("TF_VAR_CLOUDFLARE_ACCOUNT")), cloudflare.TunnelListParams{ResultInfo: cloudflare.ResultInfo{PerPage: 1000}})
		if err != nil {
			log.Fatal(err)
		}
		for _, tunnel := range list {
			fmt.Printf("%s %s %s \n", tunnel.ID, tunnel.Name, tunnel.Status)
		}
	*/

	tunnelID := os.Getenv("CLOUDFLARE_TUNNEL_ID") // Format xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx
	testDNS := "tunnel-test-name.example.com"
	testInternalDNS := "tunel-test-name.internal.example.com"
	tunnel, err := api.GetTunnel(context.Background(), accountIdentifier, tunnelID)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Tunnel %s %s %s Connections:\n", tunnel.ID, tunnel.Name, tunnel.Status)
	for _, connection := range tunnel.Connections {
		fmt.Printf("- %s %s %s %s\n", connection.ID, connection.ClientID, connection.ClientVersion, connection.OriginIP)
	}
	tunnelConfiguration, err := api.GetTunnelConfiguration(context.Background(), accountIdentifier, tunnelID)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Configuration %s %d:\n", tunnelConfiguration.TunnelID, tunnelConfiguration.Version)

	found := false
	newIngress := []cloudflare.UnvalidatedIngressRule{{
		Hostname: testDNS,
		Service:  testInternalDNS},
	}
	for _, ingress := range tunnelConfiguration.Config.Ingress {
		fmt.Printf("- \"%s\" \"%s\" \"%s\"\n", ingress.Hostname, ingress.Path, ingress.Service)
		newIngress = append(newIngress, ingress)
		if ingress.Hostname == testDNS {
			found = true
		}
	}
	if !found {
		tunnelConfiguration.Config.Ingress = newIngress
		params := cloudflare.TunnelConfigurationParams{
			TunnelID: tunnelConfiguration.TunnelID,
			Config:   tunnelConfiguration.Config,
		}
		api.UpdateTunnelConfiguration(context.Background(), accountIdentifier, params)
	}

	searchRecord := cloudflare.ListDNSRecordsParams{
		Type:    "CNAME",
		Comment: apiCode,
	}

	fmt.Println("DNS Records :")
	dnsList, _, err := api.ListDNSRecords(context.Background(), zoneIdentifier, searchRecord)
	if err != nil {
		log.Fatal(err)
	}
	found = false
	for _, record := range dnsList {
		fmt.Printf("- %s %s %b %s %s\n", record.Content, record.Name, record.Proxied, record.Type, record.Comment)
		if record.Name == testDNS {
			found = true
		}
	}
	if !found {
		content := fmt.Sprintf("%s.cfargotunnel.com", tunnelID)
		record, err := api.CreateDNSRecord(context.Background(), zoneIdentifier, cloudflare.CreateDNSRecordParams{Content: content, Name: testDNS, Proxied: BoolPointer(true), Type: "CNAME", Comment: apiCode})
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("%s %s %s %s", record.Name, record.CreatedOn, record.ModifiedOn, record.ID)
	}
}
