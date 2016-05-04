package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
)

var ipRangesURL = "https://ip-ranges.amazonaws.com/ip-ranges.json"

type amazonIPRanges struct {
	SyncToken  string `json:"syncToken"`
	CreateDate string `json:"createDate"`
	Prefixes   []struct {
		IPPrefix string `json:"ip_prefix"`
		Region   string `json:"region"`
		Service  string `json:"service"`
	} `json:"prefixes"`
}

func main() {

	var region = flag.String("region", "", "AWS Region")
	var service = flag.String("service", "", "AWS Service")
	flag.Parse()

	// If service or region flag is specified, both must be specified
	if (len(*service) == 0 && len(*region) > 0) ||
		(len(*service) > 0 && len(*region) == 0) {
		fmt.Println("Region and Service must be specified together")
		os.Exit(1)
	}

	// Get the Amazon IP Ranges
	amazonIPRanges, err := getAmazonIPRanges()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// If service and region were not specified, print out the Amazon
	// IP Range Service and Region list
	if len(*service) == 0 && len(*region) == 0 {
		printRegionServiceMap(amazonIPRanges)
		os.Exit(0)
	}

	// Get the IP Ranges in a map
	// Service [ Region ] [] IP Ranges
	ranges := getRanges(amazonIPRanges)
	var rang []string
	// Check if the given Service and Region exist
	if s, ok := ranges[*service]; ok {
		if r, ok := s[*region]; ok {
			rang = r
		} else {
			fmt.Printf("Region %s not found\n", *region)
			os.Exit(1)
		}
	} else {
		fmt.Printf("Service %s not found\n", *service)
		os.Exit(1)
	}
	fmt.Printf("Service: %s\n", *service)
	fmt.Printf("Region: %s\n", *region)
	fmt.Println("IP Ranges:")
	for _, val := range rang {
		fmt.Println(val)
	}

}

// getAmazonIPRanges Returns an amazonIPRanges structure populated with the data from ipRangesURL.
func getAmazonIPRanges() (*amazonIPRanges, error) {
	r, _ := http.Get(ipRangesURL)
	response, _ := ioutil.ReadAll(r.Body)
	amazonIPRanges := &amazonIPRanges{}
	err := json.Unmarshal(response, amazonIPRanges)
	if err != nil {
		return nil, err
	}
	return amazonIPRanges, nil
}

// printRegionServiceMap Print out a list of Amazon Services and available Regions
func printRegionServiceMap(amazonIPRanges *amazonIPRanges) {
	serviceRegionMap := make(map[string][]string)
	for _, prefix := range amazonIPRanges.Prefixes {
		if val, ok := serviceRegionMap[prefix.Service]; ok {
			// Check if the region is already in the list
			inList := false
			for _, region := range val {
				if region == prefix.Region {
					inList = true
					break
				}
			}
			if !inList {
				val = append(val, prefix.Region)
				serviceRegionMap[prefix.Service] = val
			}
		} else {
			val := []string{prefix.Region}
			serviceRegionMap[prefix.Service] = val
		}
	}
	fmt.Print("Please run the application specifying -service and -region with a combination listed below:\n\n")
	fmt.Println("-----------------------")
	for service, regions := range serviceRegionMap {
		fmt.Printf("Service: %s\n", service)
		fmt.Printf("Regions: %s\n", regions)
		fmt.Println("-----------------------")
	}
}

// getRanges Create a map of services to regions with a list of ip ranges
func getRanges(amazonIPRanges *amazonIPRanges) map[string]map[string][]string {
	ranges := make(map[string]map[string][]string)
	for _, prefixe := range amazonIPRanges.Prefixes {
		if service, ok := ranges[prefixe.Service]; ok {
			if region, ok := service[prefixe.Region]; ok {
				// Append the IP to the range for the given region and service
				service[prefixe.Region] = append(region, prefixe.IPPrefix)
				ranges[prefixe.Service] = service
			} else {
				// First time seeing the region in this service, create it and set the IP
				service[prefixe.Region] = []string{prefixe.IPPrefix}
				ranges[prefixe.Service] = service
			}
		} else {
			// First time seeing the service, create it and set the region to the IP
			service = make(map[string][]string)
			service[prefixe.Region] = []string{prefixe.IPPrefix}
			ranges[prefixe.Service] = service
		}
	}
	return ranges
}
