package cfbackup

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"regexp"
	"strconv"
	"time"

	"github.com/enaml-ops/enaml/enamlbosh"
	errwrap "github.com/pkg/errors"
)

// Not ping server so frequently and exausted the resources
var TaskPingFreq = 1000 * time.Millisecond

//CloudControllerJobs - array storing a list of CCJobs
type CloudControllerJobs []CCJob

//CloudController - a struct representing a cloud controller
type CloudController struct {
	deploymentName   string
	director         Bosh
	cloudControllers CloudControllerJobs
	manifest         []byte
}

type boshDirector struct {
	client *enamlbosh.Client
	ip     string
	port   int
}

//GetInfo -- calls info endpoint on targetted bosh director
func (s *boshDirector) GetInfo() (io.ReadCloser, error) {
	endpoint := fmt.Sprintf(ERDirectorInfoURL, s.ip)
	return s.get(endpoint)
}

//GetCloudControllerVMSet - returns a list of vm objects from your targetted
//bosh director and given deployment
func (s *boshDirector) GetCloudControllerVMSet(name string) (io.ReadCloser, error) {
	endpoint := fmt.Sprintf("%s:%d/deployments/%s/vms", s.ip, s.port, name)
	return s.get(endpoint)
}

//GetDeploymentManifest -- returns the deployment manifest for the given
//deployment on the targetted bosh director
func (s *boshDirector) GetDeploymentManifest(name string) (io.ReadCloser, error) {
	endpoint := fmt.Sprintf("%s:%d/deployments/%s", s.ip, s.port, name)
	return s.get(endpoint)
}

//ChangeJobState -- will alter the state of the given job on the given
//deployment. this can be used to start or stop a vm
func (s *boshDirector) ChangeJobState(deployment, job, state string, index int, manifest io.Reader) (int, error) {
	endpoint := fmt.Sprintf("%s:%d/deployments/%s/jobs/%s/%d?state=%s", s.ip, s.port, deployment, job, index, state)
	req, err := s.client.NewRequest("PUT", endpoint, manifest)
	if err != nil {
		return 0, errwrap.Wrap(err, "failed creating request")
	}
	req.Header.Set("content-type", "text/yaml")
	res, err := s.client.HTTPClient().Do(req)
	if err != nil {
		return 0, errwrap.Wrap(err, "failed calling http client")
	}
	taskID, err := retrieveTaskID(res)
	if err != nil {
		return 0, errwrap.Wrap(err, "failed retrieving taskid from response body")
	}

	if taskID == 0 {
		return 0, fmt.Errorf("invalid taskid returned. value is: %v", taskID)
	}
	return taskID, nil
}

//RetrieveTaskStatus - returns a task object containing the status for a given
//task id
func (s *boshDirector) RetrieveTaskStatus(id int) (*Task, error) {
	endpoint := fmt.Sprintf("%s:%d/tasks/%d", s.ip, s.port, id)
	data, err := s.get(endpoint)
	if err != nil {
		return nil, errwrap.Wrap(err, "failed on get request")
	}

	return retrieveTaskStatus(data)
}

func (s *boshDirector) get(endpoint string) (io.ReadCloser, error) {
	req, err := s.client.NewRequest("GET", endpoint, nil)
	if err != nil {
		return nil, errwrap.Wrap(err, "failed creating request")
	}
	req.Header.Set("content-type", "application/json")
	res, err := s.client.HTTPClient().Do(req)
	if err != nil {
		return nil, errwrap.Wrap(err, "failed calling http client")
	}
	if res.StatusCode != 200 {
		return nil, fmt.Errorf("unsuccessful status code in response: %v ", res.StatusCode)
	}
	return res.Body, nil
}

func retrieveTaskID(resp *http.Response) (taskId int, err error) {
	if resp.StatusCode != 302 && resp.StatusCode != 200 {
		return 0, fmt.Errorf("unsuccessfull status code response: %v", resp.StatusCode)
	}
	redirectUrl := resp.Request.URL.String()
	if redirectUrl == "" {
		return 0, fmt.Errorf("Could not find redirect url for bosh tasks: %v", redirectUrl)
	}
	regex := regexp.MustCompile(`^.*tasks/`)
	idString := regex.ReplaceAllString(redirectUrl, "")
	return strconv.Atoi(idString)
}

func retrieveTaskStatus(data io.ReadCloser) (task *Task, err error) {
	defer data.Close()
	task = &Task{}
	dbytes, err := ioutil.ReadAll(data)
	if err != nil {
		return nil, errwrap.Wrap(err, "unable to read from data")
	}
	if len(dbytes) <= 0 {
		return nil, fmt.Errorf("empty dataset returned from read")
	}

	err = json.Unmarshal(dbytes, task)
	if err != nil {
		return nil, errwrap.Wrap(err, "unable to unmarshal response body")
	}
	return task, nil
}

//NewDirector - a function representing a constructor for a director object
//go:generate counterfeiter -o fakes/fake_director_creator.go . DirectorCreator
type DirectorCreator func(ip, username, password string, port int) (Bosh, error)

var NewDirector DirectorCreator = func(ip, username, password string, port int) (Bosh, error) {
	// Check if a scheme is present (RFC 3986, section 3.1).
	// If not, prepend "//" to use the network-path reference format (section 4.2).
	if !regexp.MustCompile(`^([a-zA-Z][-+.a-zA-Z0-9]+:)?//`).MatchString(ip) {
		ip = "https://" + ip
	}
	return newBoshDirector(ip, username, password, port)
}

func newBoshDirector(ip, username, password string, port int) (Bosh, error) {
	client, err := enamlbosh.NewClient(username, password, ip, port, true)
	if err != nil {
		return nil, errwrap.Wrap(err, "failed creating bosh client")
	}
	return &boshDirector{
		client: client,
		ip:     ip,
		port:   port,
	}, nil
}

//NewCloudController - a function representing a constructor for a cloud controller
func NewCloudController(ip, username, password, deploymentName string, manifest []byte, cloudControllers CloudControllerJobs) (*CloudController, error) {
	director, err := NewDirector(ip, username, password, 25555)
	if err != nil {
		return nil, errwrap.Wrap(err, "failed creating new director")
	}

	return &CloudController{
		deploymentName:   deploymentName,
		director:         director,
		cloudControllers: cloudControllers,
		manifest:         manifest,
	}, nil
}

//Start - a method to execute a start event on a cloud controller
func (c *CloudController) Start() error {
	return c.toggleController("started")
}

//Stop - a method which executes a stop against a cloud controller
func (c *CloudController) Stop() error {
	return c.toggleController("stopped")
}

func (c *CloudController) toggleController(state string) error {
	for _, ccjob := range c.cloudControllers {
		reqBody := bytes.NewReader(c.manifest)
		taskID, err := c.director.ChangeJobState(c.deploymentName, ccjob.Job, state, ccjob.Index, reqBody)
		if err != nil {
			return err
		}
		err = c.waitUntilDone(taskID)
		if err != nil {
			return err
		}
	}
	return nil
}

func (c *CloudController) waitUntilDone(taskID int) (err error) {
	time.Sleep(TaskPingFreq)
	result, err := c.director.RetrieveTaskStatus(taskID)
	if err != nil {
		return
	}

	switch Taskresult[result.State] {
	case BOSHError:
		err = fmt.Errorf("Task %d process failed", taskID)
		return
	case BOSHQueued:
		err = c.waitUntilDone(taskID)
		return
	case BOSHProcessing:
		err = c.waitUntilDone(taskID)
		return
	case BOSHDone:
		return
	default:
		return errors.New("unkown task result error")
	}
}
