package cfbackup

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	. "github.com/pivotalservices/gtils/http"
)

type CloudControllerJobs []string

type RestAdapter func(method, connectionURL, username, password string, isYaml bool) (*http.Response, error)

type JobTogglerAdapter func(serverUrl, username, password string) (res string, err error)

type EvenTaskCreaterAdapter func(method, url, username, password string, isYaml bool) (task EventTasker)

type EventTasker interface {
	WaitForEventStateDone(contents bytes.Buffer, eventObject *EventObject) (err error)
}

type EventObject struct {
	Id          int    `json:"id"`
	State       string `json:"state"`
	Description string `json:"description"`
	Result      string `json:"result"`
}

type CloudController struct {
	ip                  string
	username            string
	password            string
	deploymentName      string
	state               string
	JobToggler          JobTogglerAdapter
	NewEventTaskCreater EvenTaskCreaterAdapter
	httpGateway         HttpGateway
}

type Task struct {
	Method     string
	Url        string
	Username   string
	Password   string
	IsYaml     bool
	RestRunner RestAdapter
}

func ToggleCCHandler(response *http.Response) (redirectUrl interface{}, err error) {
	if response.StatusCode != 301 {
		err = errors.New("The response code from toggle request should return 301")
		return
	}
	redirectUrls := response.Header["Location"]
	if redirectUrls == nil || len(redirectUrls) < 1 {
		err = errors.New("Could not find redirect url for bosh tasks")
		return
	}
	return redirectUrls[0], nil
}

var NewToggleGateway = func(serverUrl, username, password string) HttpGateway {
	return NewHttpGateway(serverUrl, username, password, "Content-Type:text/yaml", ToggleCCHandler)
}

func (restAdapter RestAdapter) Run(method, connectionURL, username, password string, isYaml bool) (statusCode int, body io.Reader, err error) {
	res, err := restAdapter(method, connectionURL, username, password, isYaml)
	defer res.Body.Close()
	body = res.Body
	statusCode = res.StatusCode
	return
}

func ToggleCCJobRunner(serverUrl, username, password string) (redirectUrl string, err error) {
	httpGateway := NewToggleGateway(serverUrl, username, password)
	ret, err := httpGateway.Execute("PUT")
	return ret.(string), err
}

func NewCloudController(ip, username, password, deploymentName, state string) *CloudController {
	return &CloudController{
		ip:                  ip,
		username:            username,
		password:            password,
		deploymentName:      deploymentName,
		state:               state,
		JobToggler:          ToggleCCJobRunner,
		NewEventTaskCreater: EvenTaskCreaterAdapter(NewTask),
	}
}

func (s *CloudController) ToggleJobs(ccjobs CloudControllerJobs) (err error) {
	serverURL := serverUrlFromIp(s.ip)

	for ccjobindex, ccjob := range ccjobs {
		err = s.ToggleJob(ccjob, serverURL, ccjobindex)
	}
	return
}

func (s *CloudController) ToggleJob(ccjob, serverURL string, ccjobindex int) (err error) {
	var (
		contents      bytes.Buffer
		eventObject   EventObject
		connectionURL string = newConnectionURL(serverURL, s.deploymentName, ccjob, s.state, ccjobindex)
	)

	if originalUrl, err := s.JobToggler(connectionURL, s.username, s.password); err == nil {
		task := s.NewEventTaskCreater("GET", modifyUrl(s.ip, serverURL, originalUrl), s.username, s.password, false)
		err = task.WaitForEventStateDone(contents, &eventObject)
	}
	return
}

func NewTask(method, url, username, password string, isYaml bool) (task EventTasker) {
	task = &Task{
		Method:     method,
		Url:        url,
		Username:   username,
		Password:   password,
		IsYaml:     isYaml,
		RestRunner: RestAdapter(invoke),
	}
	return
}

func (s *Task) getEvents(dest io.Writer) (err error) {
	statusCode, body, err := s.RestRunner.Run(s.Method, s.Url, s.Username, s.Password, s.IsYaml)

	if statusCode == 200 {
		io.Copy(dest, body)

	} else {
		err = fmt.Errorf("Invalid Bosh Director Credentials")
	}
	return
}

func (s *Task) WaitForEventStateDone(contents bytes.Buffer, eventObject *EventObject) (err error) {

	if err = json.Unmarshal(contents.Bytes(), eventObject); err == nil && eventObject.State != "done" {
		contents.Reset()

		if err = s.getEvents(&contents); err == nil {
			s.WaitForEventStateDone(contents, eventObject)
		}
	}
	return
}

func modifyUrl(ip, serverURL, originalUrl string) (newUrl string) {
	newUrl = strings.Replace(originalUrl, "https://"+ip+"/", serverURL, 1)
	return
}

func serverUrlFromIp(ip string) (serverUrl string) {
	serverUrl = "https://" + ip + ":25555/"
	return
}

func newConnectionURL(serverURL, deploymentName, ccjob, state string, ccjobindex int) (connectionURL string) {
	connectionURL = serverURL + "deployments/" + deploymentName + "/jobs/" + ccjob + "/" + strconv.Itoa(ccjobindex) + "?state=" + state
	return
}
