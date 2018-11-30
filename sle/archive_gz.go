package main

import (
	"strings"
	"fmt"
	"log"
	"io/ioutil"
)

/*
<metadata xmlns="http://linux.duke.edu/metadata/common" xmlns:rpm="http://linux.duke.edu/metadata/rpm" packages="2481">
<package type="rpm">
  <name>389-ds-base</name>
  <arch>src</arch>
  <version epoch="0" ver="1.3.1.6" rel="25.el7"/>
*/

type MetaData struct {
	Package []Package 
}

type UpdateState int

const (
	ReleaseSame UpdateState = iota
	VersionSame 
	PackageSame
	PackageMissing
)

// used accrose all the release
var pkgMap map[string]int
var mapLock bool
func init() {
	pkgMap = make(map[string]int)
	mapLock = false
}

func (m *MetaData) HasPackage(name string) bool {
	for _, p := range m.Package {
		if p.Name == name {
			return true
		}
	}
	return false
}


func (m *MetaData) PackageUpdated(pkg Package)  UpdateState {
	for _, p := range m.Package {
		if p.Name != pkg.Name {
			continue
		}
		if p.Version.Ver != pkg.Version.Ver {
			if !mapLock {
				if _, ok := pkgMap[p.Name] ; ok {
					count := pkgMap[p.Name] + 1
					pkgMap[p.Name] = count
				} else {
					pkgMap [p.Name] = 1
				}
			}
			return PackageSame 
		}
		if p.Version.Rel != pkg.Version.Rel {
			return VersionSame
		}
		return ReleaseSame
	}
	return PackageMissing
}

type Package struct {
	Name string 
	Version Version
}

type Version struct {
	Ver string
	Rel string
}

func LoadMeta(filename string) MetaData {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		log.Fatal(err)
	}

	var result MetaData
	var pkg Package
	gotPkg := false

	lines := strings.Split(string(data), "\n")


	for _, l := range lines {
		infos := strings.Split(l, ": ")
		if len(infos) < 3 {
			continue
		}
		field := strings.TrimSpace(infos[1])
		value := infos[2]
		if field == "Name" {
			if gotPkg {
				result.Package = append(result.Package, pkg)
			} else {
				gotPkg = true
			}
			pkg.Name = value
		} else if field == "Version" {
			pkg.Version.Ver = value
		} else if field == "Release" {
			pkg.Version.Rel = value
		}
	}

	if gotPkg {
		result.Package = append(result.Package, pkg)
	}

	return result
}

func Diff(f1 string, f2 string) {
	m1 := LoadMeta(f1)
	m2 := LoadMeta(f2)
	fmt.Printf("%s has %d packages, %s has %d packages\n", f1, len(m1.Package), f2, len(m2.Package))
	sameMap := make(map[UpdateState]int)
	sameMap[ReleaseSame] = 0
        sameMap[VersionSame] = 0
        sameMap[PackageSame] = 0
        sameMap[PackageMissing] = 0

	for _, p1 := range m1.Package {
		result := m2.PackageUpdated(p1)
		sameMap[result] = sameMap[result] + 1
	}

	fmt.Printf("%d is release same(no rebuilt), %d is version same(rebuilt), %d is version updated, %d is missing\n",
		sameMap[ReleaseSame], sameMap[VersionSame], sameMap[PackageSame], sameMap[PackageMissing])
	fmt.Println("\n")
}

func  main() {
	full()
	return
}

func full() {
	f0 := "./data/SLE12_ARCHIVES"
	f1 := "./data/SLE12_SP1_ARCHIVES"
	f2 := "./data/SLE12_SP2_ARCHIVES"
	f3 := "./data/SLE12_SP3_ARCHIVES"

	Diff(f0, f1)
	Diff(f1, f2)
	Diff(f2, f3)

	mapLock = true
	Diff(f0, f3)

	cs := make(map[int]int)
	cs[1] = 0
	cs[2] = 0
	cs[3] = 0

	for _, v := range pkgMap {
		cs[v] ++ 
	}
	for k, v := range cs {
		fmt.Printf("%d packages been updated %d times\n", v, k)
		
	}

	for k, v := range pkgMap {
		fmt.Println(k, v)
	}
}

