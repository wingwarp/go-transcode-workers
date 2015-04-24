package main


import (
	"fmt"
	"sync"
	"os/exec"
	"bytes"
	"strings"
	"time"
	"encoding/json"
)

const ffmpegCommand string = "~/Downloads/FFMPEG/ffmpeg"
const masterVideo string = "master.mp4"

var commands = map[string]string {
	"1280": ffmpegCommand + " -i " + masterVideo + " -y -vf \"scale=1280:trunc(ow/a/2)*2\" -c:v libx264 -acodec aac -strict experimental -c:d copy -f mp4 result-1280.mp4",
	"854": ffmpegCommand + " -i " + masterVideo + "  -y -vf \"scale=854:trunc(ow/a/2)*2\" -c:v libx264 -acodec aac -strict experimental -c:d copy -f mp4 result-854.mp4",
	"640": ffmpegCommand + " -i " + masterVideo + " -y -vf \"scale=640:trunc(ow/a/2)*2\" -c:v libx264 -acodec aac -strict experimental -c:d copy -f mp4 result-640.mp4",
	"426": ffmpegCommand + " -i " + masterVideo + " -y -vf \"scale=426:trunc(ow/a/2)*2\" -c:v libx264 -acodec aac -strict experimental -c:d copy -f mp4 result-426.mp4",
	"all": ffmpegCommand + " -i " + masterVideo + " -y -vf \"scale=1280:trunc(ow/a/2)*2\" -c:v libx264 -acodec aac -strict experimental -c:d copy -f mp4 result-1280.mp4 -vf \"scale=854:trunc(ow/a/2)*2\" -c:v libx264 -acodec aac -strict experimental -c:d copy -f mp4 result-854.mp4 -vf \"scale=640:trunc(ow/a/2)*2\" -c:v libx264 -acodec aac -strict experimental -c:d copy -f mp4 result-640.mp4 -vf \"scale=426:trunc(ow/a/2)*2\" -c:v libx264 -acodec aac -strict experimental -c:d copy -f mp4 result-426.mp4",
}


type Result struct {
   Test string
   Real int16
   User int16
   Sys int16
}


func executeCommand(cmd string) (r Result, err error) {
	r = Result{}

	app := "bash"
        arg0 := "-c"
        arg1 := "time " + "sleep 2"

        c := exec.Command(app, arg0, arg1)
	var stderr bytes.Buffer
	c.Stderr = &stderr
	err = c.Run()

	if err != nil {
		return
	}

        s := strings.Trim(string(stderr.Bytes()), "\n") 

	s = s[strings.Index(s, "real\t"):]

	var rm, rs, rus, um, us, uus, sm, ss, sus int16

	fmt.Sscanf(s, "real\t%dm%d.%ds\nuser\t%dm%d.%ds\nsys\t%dm%d.%ds",
                   &rm, &rs, &rus, &um, &us, &uus, &sm, &ss, &sus)
	r.Real = rm * 60 + rs
	r.User = um * 60 + us
	r.Sys = sm * 60 + ss

	return
}


func worker(id int, jobs <-chan string, results chan<- Result) {
   for {
        j, more := <-jobs 
        if more {
		fmt.Println("worker", id, "processing job", j)
		fmt.Println(commands[j])
       		r, err := executeCommand(commands[j])
		r.Test = j
		if err != nil {
			fmt.Println("Error executing", j, "Error: ", err)
		}
	        results <- r
	} else { break }
   }
}


func runTest(cmds []string, parallel bool) ([]Result, int16) {
    var workers int
    if parallel {
	workers = len(cmds)
    } else {
	workers = 1
    }

    jobs := make(chan string, 100)
    results := make(chan Result, 100)
    var wg sync.WaitGroup

    for w := 1; w <= workers; w++ {
        wg.Add(1)
        fmt.Println("Starting worker", w)
        go func (id int) { worker(id, jobs, results); wg.Done() } (w)
    }
    go func() {
        wg.Wait()
        close(results)
    }()

    start := time.Now()
    for _, j := range cmds {
        jobs <- j
    }
    close(jobs)

    testResults := make([]Result, 0, len(cmds))

    for {
        r, more := <-results
        if more {
		testResults = append(testResults, r)
                fmt.Println("result", r)
        } else { break }
    }
    end := time.Now()

    return testResults, int16(end.Sub(start)/time.Second)
}

type Results struct {
	Name string
	Duration int16
	CmdResults []Result
}

func main() {
    cmds := []string{"1280", "854", "640", "426"}
    all := []string{"all"}
    testResults := []Results{}
    
    r, d := runTest(cmds, false)
    fmt.Println("Sequential ffmpeg:", cmds, d)
    fmt.Println(r)
    testResults = append(testResults, Results{Name: "Sequential ffmpeg", Duration: d, CmdResults: r})

    r, d = runTest(cmds, false)
    fmt.Println("Sequential ffmpeg:", cmds, d)
    fmt.Println(r)
    testResults = append(testResults, Results{Name: "Sequential ffmpeg", Duration: d, CmdResults: r})

    r, d = runTest(cmds, true)
    fmt.Println("Parallel ffmpeg:", cmds, d)
    fmt.Println(r)
    testResults = append(testResults, Results{Name: "Parallel ffmpeg", Duration: d, CmdResults: r})
    r, d = runTest(cmds, true)
    fmt.Println("Parallel ffmpeg:", cmds, d)
    fmt.Println(r)
    testResults = append(testResults, Results{Name: "Parallel ffmpeg", Duration: d, CmdResults: r})

    r, d = runTest(all, false)
    fmt.Println("Single ffmpeg: yields all", d)
    fmt.Println(r)
    testResults = append(testResults, Results{Name: "1 all ffmpeg", Duration: d, CmdResults: r})
    r, d = runTest(all, false)
    fmt.Println("Single ffmpeg: yields all", d)
    fmt.Println(r)
    testResults = append(testResults, Results{Name: "1 all ffmpeg", Duration: d, CmdResults: r})

    fmt.Println(testResults)
    j, err := json.MarshalIndent(testResults, "", "   ")
    if err != nil {
	fmt.Println(err)
    } 
    fmt.Println("json:", string(j))
}
