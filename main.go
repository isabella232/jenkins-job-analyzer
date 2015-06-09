package main

import (
	"bytes"
	"encoding/xml"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
)

type JobType int

const (
	Maven JobType = iota
	Freestyle
	Unknown
)

type Job struct {
	XMLName      xml.Name `xml:"maven2-moduleset"`
	AssignedNode string   `xml:"assignedNode"`
}

// Quick and dirty Jenkins job analyzer.  It works by reading in all job config.xml files and parsing out useful information in the underlying XML document.
func main() {
	jobDir := flag.String("job-dir", "", "Job directory")
	flag.Parse()

	jobs := make([]string, 0)
	err := filepath.Walk(*jobDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.Name() == "config.xml" && info.Mode().IsRegular() {
			jobs = append(jobs, path)
		}
		return nil
	})
	if err != nil {
		log.Fatalf("%v\n", err)
	}
	log.Printf("Found %d jobs\n", len(jobs))

	for _, file := range jobs {
		data, err := ioutil.ReadFile(file)
		if err != nil {
			log.Fatalf("%v\n", err)
		}
		jobType, err := getJobType(data)
		if err != nil {
			log.Fatalf("%v\n", err)
		}
		if jobType == Maven {
			reader := bytes.NewBuffer(data)
			var aJob Job
			if err := xml.NewDecoder(reader).Decode(&aJob); err != nil {
				log.Fatalf("%v\n", err)
			}
			var label string
			if aJob.AssignedNode == "" {
				label = "<none>"
			} else {
				label = aJob.AssignedNode
			}
			fmt.Printf("%s %s\n", label, file)
		}
	}

}

func getJobType(xmlDocument []byte) (JobType, error) {
	decoder := xml.NewDecoder(bytes.NewBuffer(xmlDocument))

	var t string
	for {
		token, err := decoder.Token()
		if err != nil {
			return Unknown, err
		}
		if v, ok := token.(xml.StartElement); ok {
			t = v.Name.Local
			break
		}
	}

	switch t {
	case "maven2-moduleset":
		return Maven, nil
	case "project":
		return Freestyle, nil
	}
	return Unknown, nil
}
