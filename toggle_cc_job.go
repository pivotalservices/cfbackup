package cfbackup

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"regexp"
	"strconv"
	"strings"
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
	manifest         string
}

type boshDirector struct {
	client *enamlbosh.Client
	ip     string
	port   int
}

func (s *boshDirector) GetInfo() (io.ReadCloser, error) {
	endpoint := fmt.Sprintf(ERDirectorInfoURL, s.ip)
	return s.get(endpoint)
}

func (s *boshDirector) GetCloudControllerVMSet(name string) (io.ReadCloser, error) {
	endpoint := fmt.Sprintf("%s:%d/deployments/%s/vms", s.ip, s.port, name)
	return s.get(endpoint)
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

func (s *boshDirector) GetDeploymentManifest(name string) (io.Reader, error) {
	endpoint := fmt.Sprintf("%s:%d/deployments/%s", s.ip, s.port, name)
	req, err := s.client.NewRequest("GET", endpoint, nil)
	if err != nil {
		return nil, errwrap.Wrap(err, "failed creating request")
	}
	req.Header.Set("content-type", "text/yaml")
	res, err := s.client.HTTPClient().Do(req)
	if err != nil {
		return nil, errwrap.Wrap(err, "failed calling http client")
	}
	if res.StatusCode != 200 {
		return nil, fmt.Errorf("unsuccessful status code in response: %v ", res.StatusCode)
	}
	return res.Body, nil
}

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
	return retrieveTaskId(res)
}

func (s *boshDirector) RetrieveTaskStatus(id int) (*Task, error) {
	endpoint := fmt.Sprintf("%s:%d/tasks/%d", s.ip, s.port, id)
	req, err := s.client.NewRequest("GET", endpoint, nil)
	if err != nil {
		return nil, errwrap.Wrap(err, "failed creating request")
	}
	req.Header.Set("content-type", "text/yaml")
	res, err := s.client.HTTPClient().Do(req)
	if err != nil {
		return nil, errwrap.Wrap(err, "failed calling http client")
	}
	return retrieveTaskStatus(res)
}

func retrieveTaskStatus(resp *http.Response) (task *Task, err error) {
	if resp.StatusCode != 200 {
		return nil, errors.New("unsuccessfull status code response")
	}
	task = &Task{}
	defer resp.Body.Close()
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, errwrap.Wrap(err, "unable to read response body")
	}

	err = json.Unmarshal(data, task)
	if err != nil {
		return nil, errwrap.Wrap(err, "unable to unmarshal response body")
	}
	return task, nil
}

func retrieveTaskId(resp *http.Response) (taskId int, err error) {
	if resp.StatusCode != 302 {
		return 0, errors.New("unsuccessfull status code response")
	}
	redirectUrls := resp.Header["Location"]
	if redirectUrls == nil || len(redirectUrls) < 1 {
		err = errors.New("Could not find redirect url for bosh tasks")
		return
	}
	regex := regexp.MustCompile(`^.*tasks/`)
	idString := regex.ReplaceAllString(redirectUrls[0], "")
	return strconv.Atoi(idString)
}

//NewDirector - a function representing a constructor for a director object
var NewDirector = func(ip, username, password string, port int) (Bosh, error) {
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
func NewCloudController(ip, username, password, deploymentName, manifest string, cloudControllers CloudControllerJobs) (*CloudController, error) {
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
		taskID, err := c.director.ChangeJobState(c.deploymentName, ccjob.Job, state, ccjob.Index, strings.NewReader(c.manifest))
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
