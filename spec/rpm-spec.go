package main
import (
	"fmt"
	"strconv"
	"path/filepath"
	"strings"
	"time"
	"io/ioutil"

)

type ChangeLog struct {
	Date time.Time
	Version string
	Dev string
}

func GetMonth(short string) time.Month {
	switch(short) {
		case "Jan":
			return time.January
		case "Feb":
			return time.February
		case "Mar":
			return time.March
		case "Apr":
			return time.April
		case "May":
			return time.May
		case "Jun":
			return time.June
		case "Jul":
			return time.July
		case "Aug":
			return time.August
		case "Sep":
			return time.September
		case "Oct":
			return time.October
		case "Nov":
			return time.November
		case "Dec":
			return time.December
	}
	return time.June 
}

func ChangeLogNew(contentin string) (ChangeLog, error){
	// * Tue Jun 21 2016 Tomáš Mráz <tmraz@redhat.com> 1.0.1e-58
	var c ChangeLog
	content := strings.Replace(contentin, "<", "", -1)
	content = strings.Replace(content, ">", "", -1)
	content = strings.Replace(content, "(", "", -1)
	content = strings.Replace(content, ")", "", -1)
	content = strings.Replace(content, "  ", " ", -1)
	infos := strings.Split(content, " ")
	length := len(infos)
	if length < 5 {
		return c, fmt.Errorf("invalid changelog")
	}
	month := GetMonth(infos[2])
	day, _   := strconv.Atoi(infos[3])
	year, _  := strconv.Atoi(infos[4])
	version := infos [length -1]

	c.Date = time.Date(year, month, day, 0, 0, 0, 0, time.UTC)
	c.Version = version

	for i := 5; i < length; i++ {
		if strings.Contains(infos[i], "@") {
			paras := strings.Split(infos[i], "@")
			c.Dev = paras[0]
		}
	}

	return c, nil
}

type RPMSpec struct {
	Name string
	Version string
	Group string
	ChangeLogs []ChangeLog

	Devs map[string]int 
	Effert float64
}


func (rs *RPMSpec) GetReleaseDate() (time.Time, error) {
	for _, c := range rs.ChangeLogs {
		if strings.Contains(c.Version, rs.Version) {
			return c.Date, nil
		}
	}
	return time.Time{}, fmt.Errorf("cannot find the release date of `%s:%s`", rs.Name, rs.Version)
}

func (rs *RPMSpec) AddDev(email string) {
	if !strings.Contains(email, "bug") {
		rs.Devs[email]  = 1 // fix me: should increase
	}
}

func (rs *RPMSpec) GetDevs() []string {
	var ret []string
	for cl := range rs.Devs {
		ret = append(ret, cl)
	}
	return ret
}

func (rs *RPMSpec) GetEffort(gDevs map[string]int) float64 {
	ret := 0.0
	for _, email := range rs.GetDevs() {
		if val, ok := gDevs[email]; !ok {
			panic("Cannot find email in global map")
		} else if val <= 0 {
			panic("Invalid developer participant")
		} else {
			ret = ret + float64(1)/float64(val)
		}
	}
	return ret 
}

func RPMSpecNew(filename string) (RPMSpec, error) {
	var rs RPMSpec
        data, err := ioutil.ReadFile(filename)
        if err != nil {
                return rs, err
        }

	changeBegin := false
        lines := strings.Split(string(data), "\n")

        for _, l := range lines {
		if strings.HasPrefix(l, "%changelog") {
			changeBegin = true
			continue
		} else if strings.HasPrefix(l, "Name:") || strings.HasPrefix(l, "Version:") || strings.HasPrefix(l, "Group:") {
                	infos := strings.Split(l, ":")
        	        field := strings.TrimSpace(infos[0])
	                value := strings.TrimSpace(infos[1])
        	        if field == "Name" {
				if len(rs.Name) < 2 {
		                        rs.Name = value
				}
        	        } else if field == "Version" {
				if len(rs.Version) < 2 {
	                	        rs.Version = value
				}
        	        } else if field == "Group" {
				if len(rs.Group) < 2 {
	                	        rs.Group = value
				}
	                }
		}

		if !changeBegin {
			continue
		}
		if strings.HasPrefix(l, "*") {
			c, err := ChangeLogNew(l)
			if err == nil {
				rs.ChangeLogs = append([]ChangeLog{c}, rs.ChangeLogs...)
			}
		}
        }

        if rs.Name == "" {
		return rs, fmt.Errorf("invalid spec file `%s`", filename)
        }

	rs.Devs = make(map[string]int)
	for _, cl := range rs.ChangeLogs {
		rs.AddDev(cl.Dev)
	}
        return rs, nil
}

var globalSpecs []RPMSpec 
var globalDevs map[string]int
var globalProjects map[string]string


func init() {
	globalDevs = make(map[string]int)
	globalProjects = make(map[string]string)

	// this is deeper diving, we check every project to collect their commits
	globalProjects["kernel"] = "kernel.txt"
}


type CVEInfo struct {
	Name string
	Total int
	High  int
}

func CVEInfoNew(line string) (cveInfo CVEInfo, err error) {
	params := strings.Split(line, "\t")
	if len(params) < 3 {
		err = fmt.Errorf("invalid cve line %s", line)
		return 
	}
	var i int
	cveInfo.Name = params[0]
	if i, err = strconv.Atoi(params[1]); err == nil {
		cveInfo.Total = i
	} else {
		return
	}

	if i, err = strconv.Atoi(params[2]); err == nil {
		cveInfo.High = i
	} else {
		return
	}

	return
}

func ReadCVEFile(filename string) []CVEInfo {
	var infos []CVEInfo
        data, err := ioutil.ReadFile(filename)
        if err != nil {
		return infos
        }
        lines := strings.Split(string(data), "\n")
	for _, l := range lines {
		cveInfo, err :=	CVEInfoNew(l)
		if err == nil {
			infos = append(infos, cveInfo)
		}
	}

	return infos
}

func ReadPatchFile(filename string) string {
        data, err := ioutil.ReadFile(filename)
        if err != nil {
		return "bug"
        }
        lines := strings.Split(string(data), "\n")

	for _, l := range lines {
		if strings.HasPrefix(l, "From:") && strings.Contains(l, "@") {
			l := strings.Replace(l, "<", " ", -1)
			l = strings.Replace(l, ">", " ", -1)
			l = strings.Replace(l, "  ", " ", -1)
                	infos := strings.Split(l, " ")
			for _, in := range infos {
				if strings.Contains(in, "@") && strings.Contains(in, "redhat") {
					emails := strings.Split(in, "@")
					return emails[0]
				}
			}
		} 
	}
	return "bug"
}

func ReadSpecFile(filename string) RPMSpec {
	rs, err := RPMSpecNew(filename)
	if err != nil {
		return rs
	}
	return rs
}

func ReadCommitFile(filename string) []string {
	var lines []string
        data, err := ioutil.ReadFile(filename)
        if err != nil {
		fmt.Println(err)
		return lines
        }
        lines = strings.Split(string(data), "\n")
	return lines
}


func ReadPackageDir(pkgDir string) {
	var rs RPMSpec
	var patchDevs []string
	files, _ := ioutil.ReadDir(pkgDir) 
	for _, file := range files {
		if file.IsDir() {
			continue
		}
		if strings.HasSuffix(file.Name(), ".spec") {
			rs = ReadSpecFile(filepath.Join(pkgDir, file.Name()))
		} else if strings.HasSuffix(file.Name(), ".patch") || strings.HasSuffix(file.Name(), ".diff") {
			patchDevs = append(patchDevs, ReadPatchFile(filepath.Join(pkgDir, file.Name())))
		}
	}


	for _, dev := range patchDevs {
		rs.AddDev(dev)
	}

	if val, ok := globalProjects[rs.Name]; ok {
		commitDevs := ReadCommitFile(val)

		for _, commitDev := range commitDevs {
			rs.AddDev(commitDev)
		}
	}

	globalSpecs = append(globalSpecs, rs)
	devs := rs.GetDevs()

	for _, dev := range devs {
		if value, ok := globalDevs[dev]; ok {
			globalDevs[dev] = value + 1
		} else {
			globalDevs[dev] = 1
		}
	}
}

func DataMineGroupEfforts() {
	fmt.Println("Group efforts")

	groupEfforts := make(map[string]float64)
	for _, rs := range globalSpecs {
		effort := rs.GetEffort(globalDevs)
		if val, ok := groupEfforts[rs.Group]; ok {
			groupEfforts[rs.Group] = val + effort
		} else {
			groupEfforts[rs.Group] = effort
		}
	}
		
	for key, val := range groupEfforts {
		fmt.Println(key, "\t", val)
	}
}

func DataMineGroupPackage() {
	for _, rs := range globalSpecs {
		if len(rs.Group) < 2 {
			fmt.Println("Unknown", "\t", rs.Name)
		} else {
			fmt.Println(rs.Group, "\t", rs.Name)
		}
	}
}

func DataMineGroup() {
	fmt.Println("Total package: ", len(globalSpecs))

	groupCount := make(map[string]int)
	for _, rs := range globalSpecs {
		fmt.Println(" group name is <", rs.Group, "> ---", rs.Name)
		if val, ok := groupCount[rs.Group]; ok {
			groupCount[rs.Group] = val + 1
		} else {
			groupCount[rs.Group] = 1
		}
	}
		
	num := 0
	for key, val := range groupCount {
		fmt.Println(key, "\t", val)
		num = num + val
	}
	fmt.Println(num)
}

func getRange(effort float64) float64 {
	devRangeList := []float64 {20.0, 5.0, 1.0, 0.5, 0.3, 0.1, 0.0}
	for _, val := range devRangeList {
		if effort >= val  {
			return val
		}
	}
	return 0.0
}

func DataMineDev() {
	fmt.Println("Total dev: ", len(globalDevs))

	devRangeMap:= make(map[float64]int)
	num := 0.0
	for _, rs := range globalSpecs {
		effort := rs.GetEffort(globalDevs)
		fmt.Println(rs.Name, "\t", effort)
		num += effort

		er := getRange(effort)
		if val, ok := devRangeMap[er]; ok {
			devRangeMap[er] = val + 1
		} else {
			devRangeMap[er] = 1
		}
	}
		
	fmt.Println("double check total devs: ", num)

	fmt.Println("Packages sum of  Devs effort range")
	for i, val := range devRangeMap {
		fmt.Println(i,  "\t", val)
	}
}

func DataMineTest() {
	fmt.Println(globalDevs)
	fmt.Println("----- ", len(globalDevs))

	for _, rs := range globalSpecs {
		effort := rs.GetEffort(globalDevs)
		fmt.Println(effort, rs.Name)
	}
}

func DataMineGroupCVE(infos []CVEInfo) {
	groupCVETotal := make(map[string]int)
	groupCVEHigh := make(map[string]int)

	specMap := make(map[string]RPMSpec)
	for _, rs := range globalSpecs {
		specMap[rs.Name] = rs
	}

	for _, cveInfo := range infos {
		rs, ok := specMap[cveInfo.Name]
		if !ok {
//			fmt.Println("Cannot find the matched ", cveInfo.Name)
			continue
		}
		total := cveInfo.Total
		high := cveInfo.High
		if val, ok := groupCVETotal[rs.Group]; ok {
			groupCVETotal[rs.Group] = val + total
		} else {
			groupCVETotal[rs.Group] = total
		}
		if val, ok := groupCVEHigh[rs.Group]; ok {
			groupCVEHigh[rs.Group] = val + high
		} else {
			groupCVEHigh[rs.Group] = high
		}
	}

	for key, value := range groupCVETotal {
		fmt.Println(key, value, groupCVEHigh[key])
	}
}

func main() {
	// 1. loop dir	/tmp/centos7.5

	// in each dir we do
	// read spec 
	// read patch
	dir :=  "/tmp/centos7.5"
	files , _ := ioutil.ReadDir(dir)
	for _, file := range files {
		if file.IsDir() {
			ReadPackageDir(filepath.Join(dir, file.Name()))
		}
	}

//	DataMineGroup()
//	DataMineDev()
//	DataMineGroupEfforts()
	DataMineGroupPackage()


//	infos := ReadCVEFile("cve.txt")	
//	DataMineGroupCVE(infos)
}


func main1() {
/*
	filedir70 := "/root/centos/7.0/"
	date70 := time.Date(2014, time.June, 15, 0, 0, 0, 0, time.UTC)
	filedir73 := "/root/centos/vault.centos.org/7.3.1611/os/Source/SPackages/"
	date73 := time.Date(2016, time.November, 15, 0, 0, 0, 0, time.UTC)
*/
	filedir75 := "/root/centos/7.5/vault.centos.org/7.5.1804/os/Source/SPackages/"
//	date75 := time.Date(2018, time.April, 15, 0, 0, 0, 0, time.UTC)

	filedirCur := filedir75
//	dateCur := date75

        data, err := ioutil.ReadFile(filedirCur + "all.spec")
        if err != nil {
		fmt.Println(err)
		return
        }
        lines := strings.Split(string(data), "\n")
/*
	var count float64
	var diffAll float64
	count = 0.0
	diffAll = 0.0
*/
	globalDevs := make(map[string]int) 
	var specs []RPMSpec
        for _, l := range lines {
		filename := filedirCur + l
		rs, err := RPMSpecNew(filename)
		if err != nil {
			continue
		}

		specs = append(specs, rs)
/*
		date, err := rs.GetReleaseDate()
		if err == nil {
			count = count + 1.0
			diff := dateCur.Sub(date)
			diffAll += diff.Hours()/24.0/365
			fmt.Printf("%s\t%s\t%s\t%f\n", rs.Name, rs.Version, rs.Group, diff.Hours()/24.0/365)
		}
*/
		emails := rs.GetDevs()
		for _, email := range emails {
			if value, ok := globalDevs[email]; ok {
				globalDevs[email] = value + 1
			} else {
				globalDevs[email] = 1
			}
		}
	}
//	fmt.Println(count, 1.0*diffAll/count)

	fmt.Println(globalDevs)

	for _, rs := range specs {
		effort := rs.GetEffort(globalDevs)
		fmt.Println(effort, rs.Name)
	}
}	
