package main

import (
	//"errors"
	"fmt"
	"log"
	"math"
	"os"
	"os/user"
	"strconv"
	"strings"
	"encoding/json"
	"os/exec"

	"github.com/charmbracelet/bubbles/table"
)

// type Partition struct {
//     GBfree, GBused, MBfree, MBused float64
//     Label, FStype, MountPoint string
// }

type Partition struct {
     Path       string
     Fssize     interface{}
     Fsavail    interface{}
     Fstype     string
     Size       interface{}
     Fsused     interface{}
     Label      string
     Mountpoint string
     Type       string
}

var commands map[string]string
var TablePartitionArr []table.Row
var TableMaxStrLenArr [5]int
var devArr            []int

func IntMax(a, b int) int {
    if a > b {
        return a
    }
    return b
}

func IntMin(a, b int) int {
    if a < b {
        return a
    }
    return b
}

func (p Partition) makeTableRow() table.Row {
    return table.Row{p.Path, fmt.Sprint(p.Fssize), fmt.Sprint(p.Fsavail), fmt.Sprint(p.Fstype), fmt.Sprint(p.Size)}
}

func isRoot() bool {
    currentUser, err := user.Current()
    if err != nil {
        log.Fatalf("[isRoot] Unable to get current user: %s", err)
    }
    return currentUser.Username == "root"
}

func isMounted(dev string) bool {
    _, err := RunCmdOutput(fmt.Sprintf(commands["checkMount"], dev))

    if err != nil {
        return false
    }
    return true
}

func makeCmdMap() {
    commands = map[string]string {
        "checkMount" : "findmnt %s",
        "getFree" : `tune2fs -l %s | grep -E 'Free blocks|Reserved block count|Block size' | awk '{print $NF}'`,
        "getTotal" : `tune2fs -l %s | grep -E 'Block count|Reserved block count|Block size' | awk '{print $NF}'`,
        "lsblkTable" : `lsblk --bytes --json --noheadings --paths -o PATH,SIZE,FSSIZE,FSAVAIL,FSTYPE,FSUSED,LABEL,TYPE,MOUNTPOINT | sed 's/\"blockdevices\"\://g;1d;$d'`,
        "lvmCreateVg" : "vgcreate -f %s %s",
        "lvmCreatePv" : "pvcreate -f %s",
        "lvmCreateLv" : "lvcreate -L %s -n %s %s",
        "lvmGetInfo0"  : `lvdisplay | grep -P -o '(?<=LV Path).*|(?<=LV Size).*' | tr -d ' '`,
        "lvmGetInfo" : `lvdisplay --units G -C -o "lv_dm_path,lv_size" --noheadings --separator ',' | tr -d ' '`,
        "fsFormatExt4" : "mkfs.ext4 %s",
        "luksFormat" : "printf %s | sudo cryptsetup -q luksFormat phyos",
        "luksAddPass" : "printf %s | sudo cryptsetup -q luksOpen phyos %s",
    }

}

func GetPartFreeSpace(dev string) float64 {
    tmp, err := RunCmdOutput(fmt.Sprintf(commands["getFree"], dev)); if err != nil {
        return -1
    }
    calcArr := strings.Split(string(tmp), "\n")
    if len(calcArr) >= 3 {
        var rbc, fb, bs float64
        rbc, _ = strconv.ParseFloat(calcArr[0], 64)
        fb, _  = strconv.ParseFloat(calcArr[1], 64)
        bs, _  = strconv.ParseFloat(calcArr[2], 64)
        return (math.Abs(fb - rbc) * bs / math.Pow(2, 30))
    }
    return -1
}

func PrettyPrint(data interface{}) {
    var p []byte
    //    var err := error
    p, err := json.MarshalIndent(data, "", "\t")
    if err != nil {
        fmt.Println(err)
        return
    }
    fmt.Printf("%s \n", p)
}

func RunCmd(cmd string) {
    exec.Command("/bin/bash", "-c", cmd).Start()
}

func RunCmdOutput(cmd string) ([]byte, error) {
    return exec.Command("/bin/bash", "-c", cmd).Output()
}

func MakePartitionTable() {
    tableStr, _ := RunCmdOutput(commands["lsblkTable"])

    var partitions []Partition
    json.Unmarshal(tableStr, &partitions)
    lvmStr, err := RunCmdOutput(commands["lvmGetInfo"]); if err != nil {
        fmt.Fprintf(os.Stderr, err.Error())
    }
    lvMap := make(map[string]float64)
    tmp := strings.Split(string(lvmStr), "\n")

    for _, i := range tmp {
        tmpLv := strings.Split(i, ",")
        if len(tmpLv) > 1 {
            lvMap[tmpLv[0]], _ = strconv.ParseFloat(tmpLv[1][0:len(tmpLv[1])-1], 64)
        }
    }

    for i, part := range partitions {
        _, ok := lvMap[part.Path]; if ok && !isMounted(part.Path) {
            part.Fssize = lvMap[part.Path]
            part.Fsavail = GetPartFreeSpace(part.Path)
        } else {
            if !isMounted(part.Path) {
                tmp  := commands["getFree"]
                part.Fsavail = GetPartFreeSpace(part.Path)
                commands["getFree"] = commands["getTotal"]
                part.Fssize = GetPartFreeSpace(part.Path)
                commands["getFree"] = tmp
                part.Fsused = part.Fssize.(float64) - part.Fsavail.(float64)
            } else {
                    f, ok := part.Fssize.(float64); if ok {
                        part.Fssize = float64(float64(f) / math.Pow(2, 30))
                    }

                    f, ok = part.Fsavail.(float64); if ok {
                        part.Fsavail = float64(float64(f) / math.Pow(2, 30))
                    }

                    f, ok = part.Fsused.(float64); if ok {
                        part.Fsused = float64(float64(f) / math.Pow(2, 30))
                    }
                }
            }

            f, ok := part.Size.(float64); if ok {
                part.Size = float64(float64(f) / math.Pow(2, 30))
            }

        if part.Type != "disk" {
            TablePartitionArr = append(TablePartitionArr, part.makeTableRow())
            TableMaxStrLenArr[0] = IntMax(TableMaxStrLenArr[0], len(part.Path))
            TableMaxStrLenArr[1] = IntMax(TableMaxStrLenArr[1], len(fmt.Sprint(part.Fssize)))
            TableMaxStrLenArr[2] = IntMax(TableMaxStrLenArr[2], len(fmt.Sprint(part.Fsavail)))
            TableMaxStrLenArr[3] = IntMax(TableMaxStrLenArr[3], len(fmt.Sprint(part.Fstype)))
            TableMaxStrLenArr[4] = IntMax(TableMaxStrLenArr[4], len(fmt.Sprint(part.Size)))
        } else {
            devArr = append(devArr, i)
        }
    }
}

func main() {
    if !isRoot() {
        panic("This program must be runned as superuser.")
    }
    makeCmdMap()
    MakePartitionTable()
    CreatePartitionTable()
}
