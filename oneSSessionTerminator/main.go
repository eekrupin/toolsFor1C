//go:generate goversioninfo -icon=./icon.ico
package main

import (
	"bufio"
	"bytes"
	"golang.org/x/text/encoding/charmap"
	"golang.org/x/text/transform"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"syscall"
	"time"
)

func main() {

	if len(os.Args) != 3 {
		log.Fatal("Must be 2 params: Cluster (host) and Infobase (name)")
	}

	clusterHost := os.Args[1]
	infobaseName := os.Args[2]

	done := startRas(clusterHost)

	terminateSessions(done, clusterHost, infobaseName)

}
func startRas(clusterHost string) chan<- bool {
	done := make(chan bool)
	var inb, outb, errb bytes.Buffer
	cmd := exec.Command(`C:\Program Files\1cv8\8.3.13.1513\bin\ras.exe`, "cluster", clusterHost)
	//cmd := exec.Command(`cmd.exe`)
	cmd.Stdout = &outb
	cmd.Stderr = &errb
	cmd.Stdin = &inb
	cmd.Dir = `C:\Program Files\1cv8\8.3.13.1513\bin`
	cmd.Start()

	go func() {
		//log.Println("Waiting done")
		<-done
		log.Println("Sending interrupt")
		cmd.Process.Signal(syscall.SIGINT)
		time.Sleep(500 * time.Millisecond)
		cmd.Process.Kill()
		log.Println("Process killed")

	}()

	go func() {
		err := cmd.Wait()
		//err := cmd.Run()
		if err != nil {
			log.Println(err)
		}
		time.Sleep(1 * time.Second)
		log.Println(outb.String())
	}()

	for {
		if strings.Contains(outb.String(), "started") {
			time.Sleep(100 * time.Millisecond)
			break
		}
	}
	log.Println(outb.String())
	return done
}

func terminateSessions(done chan<- bool, clusterHost string, infobaseName string) {
	log.Println("cluster:", clusterHost)
	log.Println("infobase:", infobaseName)
	args := []string{"cluster", clusterHost, "list"}
	result := getResult(args)
	values := getMapResult(result.String())
	clusterId := values[0]["cluster"]
	log.Println("clusterId:", clusterId)
	//rac infobase --cluster=fc901f99-2223-4552-9f30-e915904ec33a summary list host

	args = []string{"infobase", "--cluster=" + clusterId, "summary", "list", clusterHost}
	result = getResult(args)
	values = getMapResult(result.String())
	var infobaseId string
	for _, line := range values {
		if line["name"] == infobaseName {
			infobaseId = line["infobase"]
			break
		}
	}
	log.Println("infobaseId:", infobaseId)
	///rac session --cluster=fc901f99-2223-4552-9f30-e915904ec33a list host --infobase=59c0d819-ca1b-4cb9-9507-986173a5df91
	args = []string{"session", "--cluster=" + clusterId, "list", clusterHost, "--infobase=" + infobaseId}
	result = getResult(args)
	values = getMapResult(result.String())
	curTime := getTime()
	for _, line := range values {
		startedAt, err := time.Parse(time.RFC3339, line["started-at"]+"+00:00")
		if err != nil {
			continue
		}
		lastActiveAt, err := time.Parse(time.RFC3339, line["last-active-at"]+"+00:00")
		if err != nil {
			continue
		}

		durationLast5Min, err := strconv.Atoi(line["duration-last-5min"])
		if err != nil {
			continue
		}

		if line["infobase"] == infobaseId &&
			line["user-name"] == "IIS-1csvc" &&
			curTime.Sub(startedAt).Seconds() > 3600 &&
			curTime.Sub(lastActiveAt).Seconds() > 3600 &&
			durationLast5Min == 0 {
			log.Println(line["session-id"] + "   " + line["started-at"] + "   " + line["last-active-at"] + "   " + line["duration-last-5min"])
			args = []string{"session", "--cluster=" + clusterId, "terminate", clusterHost, "--session=" + line["session"], "--error-message=killed hovering IIS session"}
			//result = getResult(args)
			log.Println("Terminate result: " + result.String())
		}
	}
	log.Println("Completed")
	done <- true
	log.Println("Waiting 5 second...")
	time.Sleep(5 * time.Second)
}

func getTime() time.Time {

	timeNow := time.Now()
	_, offset := timeNow.Zone()
	timeUTC := timeNow.UTC().Add(time.Duration(offset) * time.Second)

	return timeUTC
}

func getMapResult(result string) []map[string]string {
	var res []map[string]string
	data := make(map[string]string)
	scanner := bufio.NewScanner(strings.NewReader(result))
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			res = append(res, data)
			data = make(map[string]string)
			continue
		}
		ind := strings.Index(line, ":")
		key := strings.TrimSpace(line[:ind])
		val := strings.TrimSpace(line[ind+1:])
		data[key] = val
	}
	return res
}

func getResult(args []string) bytes.Buffer {
	var outb, errb bytes.Buffer
	cmd := exec.Command(`C:\Program Files\1cv8\8.3.13.1513\bin\rac.exe`, args...)
	cmd.Stdout = &outb
	cmd.Stderr = &errb
	cmd.Dir = `C:\Program Files\1cv8\8.3.13.1513\bin`
	err := cmd.Run()
	if err != nil {
		//for i:=1; i<=43; i++{
		//	log.Println(i, string(convert(i, errb.Bytes())))
		//}
		log.Fatal(string(convert(41, errb.Bytes())), err)
	}
	return outb
}

func convert(i int, s []byte) []byte {
	var reader *transform.Reader
	switch i {
	case 1:
		reader = transform.NewReader(bytes.NewReader(s), charmap.ISO8859_1.NewDecoder())
	case 2:
		reader = transform.NewReader(bytes.NewReader(s), charmap.ISO8859_2.NewDecoder())
	case 3:
		reader = transform.NewReader(bytes.NewReader(s), charmap.ISO8859_3.NewDecoder())
	case 4:
		reader = transform.NewReader(bytes.NewReader(s), charmap.ISO8859_4.NewDecoder())
	case 5:
		reader = transform.NewReader(bytes.NewReader(s), charmap.ISO8859_5.NewDecoder())
	case 6:
		reader = transform.NewReader(bytes.NewReader(s), charmap.ISO8859_6.NewDecoder())
	case 7:
		reader = transform.NewReader(bytes.NewReader(s), charmap.ISO8859_7.NewDecoder())
	case 8:
		reader = transform.NewReader(bytes.NewReader(s), charmap.ISO8859_8.NewDecoder())
	case 9:
		reader = transform.NewReader(bytes.NewReader(s), charmap.ISO8859_9.NewDecoder())
	case 10:
		reader = transform.NewReader(bytes.NewReader(s), charmap.ISO8859_10.NewDecoder())
	case 11:
		reader = transform.NewReader(bytes.NewReader(s), charmap.ISO8859_13.NewDecoder())
	case 12:
		reader = transform.NewReader(bytes.NewReader(s), charmap.ISO8859_14.NewDecoder())
	case 13:
		reader = transform.NewReader(bytes.NewReader(s), charmap.ISO8859_15.NewDecoder())
	case 14:
		reader = transform.NewReader(bytes.NewReader(s), charmap.ISO8859_16.NewDecoder())
	case 15:
		reader = transform.NewReader(bytes.NewReader(s), charmap.KOI8R.NewDecoder())
	case 16:
		reader = transform.NewReader(bytes.NewReader(s), charmap.KOI8U.NewDecoder())
	case 17:
		reader = transform.NewReader(bytes.NewReader(s), charmap.Macintosh.NewDecoder())
	case 18:
		reader = transform.NewReader(bytes.NewReader(s), charmap.MacintoshCyrillic.NewDecoder())
	case 19:
		reader = transform.NewReader(bytes.NewReader(s), charmap.Windows874.NewDecoder())
	case 20:
		reader = transform.NewReader(bytes.NewReader(s), charmap.Windows1250.NewDecoder())
	case 21:
		reader = transform.NewReader(bytes.NewReader(s), charmap.Windows1251.NewDecoder())
	case 22:
		reader = transform.NewReader(bytes.NewReader(s), charmap.Windows1252.NewDecoder())
	case 23:
		reader = transform.NewReader(bytes.NewReader(s), charmap.Windows1253.NewDecoder())
	case 24:
		reader = transform.NewReader(bytes.NewReader(s), charmap.Windows1254.NewDecoder())
	case 25:
		reader = transform.NewReader(bytes.NewReader(s), charmap.Windows1255.NewDecoder())
	case 26:
		reader = transform.NewReader(bytes.NewReader(s), charmap.Windows1256.NewDecoder())
	case 27:
		reader = transform.NewReader(bytes.NewReader(s), charmap.Windows1257.NewDecoder())
	case 28:
		reader = transform.NewReader(bytes.NewReader(s), charmap.Windows1258.NewDecoder())
	case 29:
		reader = transform.NewReader(bytes.NewReader(s), charmap.XUserDefined.NewDecoder())
	case 30:
		reader = transform.NewReader(bytes.NewReader(s), charmap.CodePage037.NewDecoder())
	case 31:
		reader = transform.NewReader(bytes.NewReader(s), charmap.CodePage437.NewDecoder())
	case 32:
		reader = transform.NewReader(bytes.NewReader(s), charmap.CodePage850.NewDecoder())
	case 33:
		reader = transform.NewReader(bytes.NewReader(s), charmap.CodePage852.NewDecoder())
	case 34:
		reader = transform.NewReader(bytes.NewReader(s), charmap.CodePage855.NewDecoder())
	case 35:
		reader = transform.NewReader(bytes.NewReader(s), charmap.CodePage858.NewDecoder())
	case 36:
		reader = transform.NewReader(bytes.NewReader(s), charmap.CodePage860.NewDecoder())
	case 37:
		reader = transform.NewReader(bytes.NewReader(s), charmap.CodePage862.NewDecoder())
	case 38:
		reader = transform.NewReader(bytes.NewReader(s), charmap.CodePage863.NewDecoder())
	case 39:
		reader = transform.NewReader(bytes.NewReader(s), charmap.CodePage865.NewDecoder())
	case 40:
		reader = transform.NewReader(bytes.NewReader(s), charmap.CodePage865.NewDecoder())
	case 41:
		reader = transform.NewReader(bytes.NewReader(s), charmap.CodePage866.NewDecoder())
	case 42:
		reader = transform.NewReader(bytes.NewReader(s), charmap.CodePage1047.NewDecoder())
	case 43:
		reader = transform.NewReader(bytes.NewReader(s), charmap.CodePage1140.NewDecoder())
	}

	d, _ := ioutil.ReadAll(reader)

	return d
}
