// Configstand.go
package main

import (
	"encoding/xml"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
)

var xmlstr = `<?xml version="1.0" encoding="utf-8" ?>
<configs>
  <profiles version="1.0">
    <profile name="DoD" orgsupport="" bytes="0x00;C;0xFF;C;*;C" pass="1" description="Wipe device using US DoD 5220.22-M method (3 passes)"/>
    <profile name="DoD7" orgsupport="" bytes="*;*;~*;*;*;~*;*" pass="1" description=" Wipe device using US DoD 5200.28-STD method (7 passes)"/>
    <profile name="DoE" orgsupport="" bytes="*;*;0x00" pass="1" description=" Wipe device using US DoE M 205.1-2 method (3 passes)"/>
    <profile name="NSA" orgsupport="" bytes="0x00;C;0xFF;C;*;C" pass="1" description=" Wipe device using US NSA NCSC-TG-025 method (3 passes)"/>
    <profile name="USAirForce" orgsupport="" bytes="0x00;0xFF;*;C" pass="1" description="Wipe device using US Air Force AFSSI-5020 method (3 passes)"/>
    <profile name="USArmy" orgsupport="" bytes="*;*;~*;C" pass="1" description="Wipe device using US Army AR 380-19 method (3 passes)"/>
    <profile name="USNavy" orgsupport="" bytes="0x00;0xFF;*;C" pass="1" description="Wipe device using US Navy NAVSO P-5239-26 method (3 passes)"/>
    <profile name="Canada_RCMP" orgsupport="" bytes="0x00;0xFF;0x00;0xFF;0x00;0xFF;*;C" pass="1" description="Wipe device using Canada RCMP TSSIT OPS-II method (7 passes)"/>
    <profile name="Canada_CSEC" orgsupport="" bytes="0x00;0xFF;*;C" pass="1" description="Wipe device using Canada CSEC ITSG-06 method (3 passes)"/>
    <profile name="UK" orgsupport="" bytes="0x00;0xFF;*;C" pass="1" description="Wipe device using UK HMG IS5 method (3 passes)"/>
    <profile name="German" orgsupport="" bytes="0x00;0xFF;0x00;0xFF;0x00;0xFF;*" pass="1" description="Wipe device using German VSITR method (7 passes)"/>
    <profile name="Australia" orgsupport="" bytes="*;C" pass="1" description="Wipe device using Australia ISM 6.2.92 method (1 pass)"/>
    <profile name="Australia_15" orgsupport="" bytes="*;*;*;C" pass="1" description="Wipe device using Australia ISM 6.2.92 method (3 passes)"/>
    <profile name="NewZealand" orgsupport="" bytes="*;C" pass="1" description="Wipe device using New Zealand NZSIT 402 method (1 pass)"/>
    <profile name="Russia" orgsupport="" bytes="0x00;*" pass="1" description="Wipe device using Russia GOST R 50739-95 method (2 passes)"/>
    <profile name="BSchneier" orgsupport="-b" bytes="" pass="1" description="Wipe device using Bruce Schneier's method (7 passes)"/>   
    <profile name="PGutmann" orgsupport="-g" bytes="" pass="1" description="Wipe device using Peter Gutmann's method (35 passes)"/>
    <profile name="Pfitzner" orgsupport="" bytes="*" pass="33" description="Wipe device using Roy Pfitzner's method (33 passes)"/>   
    <profile name="OneTime_0" orgsupport="" bytes="0x00" pass="1" description="Wipe device once using 0 (1 pass)"/>
    <profile name="OneTime_1" orgsupport="" bytes="0xFF" pass="1" description="Wipe device once using 1 (1 pass)"/>   
    <profile name="OneTime_Rdm" orgsupport="" bytes="*" pass="1" description="Wipe device once using a random character (1 pass)"/>
    <profile name="NTime_Rdm" orgsupport="" bytes="*" pass="10" description="Wipe device once using a random character (N pass)"/>   
    <profile name="LLFormat" orgsupport="" bytes="0x00" pass="1" description="Wipe device using Low Level Format method (1 Pass)"/>
    <profile name="SecureErase" orgsupport="" bytes="" pass="1" description="Wipe device using Secure Wipe method (1 pass)"/>      
  </profiles>
</configs>`

type configs struct {
	Configs Recurlyprofiles `xml:"profiles"`
}

// Recurlyprofiles define profiles
type Recurlyprofiles struct {
	Version string    `xml:"version,attr"`
	Profis  []profile `xml:"profile"`
}

type profile struct {
	Name        string `xml:"name,attr"`
	Orgsupport  string `xml:"orgsupport,attr"`
	Bytes       string `xml:"bytes,attr"`
	Pass        int    `xml:"pass,attr"`
	Description string `xml:"description,attr"`
}

func (e profile) String() string {
	return fmt.Sprintf("%s - %d - %s", e.Name, e.Pass, e.Description)
}

// Split string
func Split(r rune) bool {
	return r == ',' || r == ';' || r == ':'
}

func (e profile) CreatePatten() string {
	patts := strings.FieldsFunc(e.Bytes, Split)
	random := 0
	s := make([]string, len(patts))
	for i, patt := range patts {
		if strings.EqualFold("*", patt) {
			random = rand.Intn(256)
			s[i] = strconv.Itoa(random)
		} else if strings.EqualFold("~*", patt) {
			random = ^random
			s[i] = strconv.Itoa(random)
		} else if strings.EqualFold("C", patt) {
			s[i] = patt
		} else if strings.EqualFold("R", patt) {
			s[i] = patt
		} else {
			s[i] = patt
		}
	}
	return strings.Join(s, " ")
}

func (cc configs) FindProfileByName(name string) (profile, error) {
	for _, item := range cc.Configs.Profis {
		if strings.EqualFold(item.Name, name) {
			return item, nil
		}
	}
	var pp profile
	return pp, errors.New("not found item")
}

// LoadConfigXML from config xml
func LoadConfigXML() {
	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		log.Fatal(err)
		return
	}
	configxmlpath := path.Join(dir, "hdsesconfig.xml")

	xmlFile, err := os.Open(configxmlpath)
	if err != nil {
		fmt.Println("Error opening file:", err)
		//xmlFile = []byte(xmlstr)
	}
	defer xmlFile.Close()
	b, err := ioutil.ReadAll(xmlFile)
	if err != nil {
		b = []byte(xmlstr)
	}

	configxmldata = &configs{}
	err = xml.Unmarshal(b, configxmldata)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(configxmldata)

	for _, item := range configxmldata.Configs.Profis {
		fmt.Printf("%s\n", item)
	}

}
