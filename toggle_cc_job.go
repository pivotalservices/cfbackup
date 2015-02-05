package cfbackup

import (
	"errors"
	"fmt"
	"io"

	"github.com/pivotalservices/gtils/bosh"
	. "github.com/pivotalservices/gtils/http"
)

type CloudControllerJobs []string

type CloudController struct {
	deploymentName   string
	director         bosh.Bosh
	cloudControllers CloudControllerJobs
	manifest         io.Reader
}

var NewDirector = func(ip, username, password string, port int) bosh.Bosh {
	return bosh.NewBoshDirector(ip, username, password, port, NewHttpGateway())
}

func NewCloudController(ip, username, password, deploymentName string, cloudControllers CloudControllerJobs) *CloudController {
	director := NewDirector(ip, username, password, 25555)
	manifest, err := director.GetDeploymentManifest(deploymentName)
	if err != nil {
		panic(err)
	}
	return &CloudController{
		deploymentName:   deploymentName,
		director:         director,
		cloudControllers: cloudControllers,
		manifest:         manifest,
	}
}

func (c *CloudController) Start() error {
	return c.toggleController("started")
}

func (c *CloudController) Stop() error {
	return c.toggleController("stopped")
}

func (c *CloudController) toggleController(state string) error {
	for ccjobindex, ccjob := range c.cloudControllers {
		taskId, err := c.director.ChangeJobState(c.deploymentName, ccjob, state, ccjobindex, c.manifest)
		if err != nil {
			return err
		}
		err = c.waitUntilDone(taskId)
		if err != nil {
			return err
		}
	}
	return nil
}

func (c *CloudController) waitUntilDone(taskId int) (err error) {
	result, err := c.director.RetrieveTaskStatus(taskId)
	if err != nil {
		return
	}
	switch bosh.TASKRESULT[result.State] {
	case bosh.ERROR:
		err = errors.New(fmt.Sprintf("Task %d process failed", taskId))
		return
	case bosh.PROCESSING:
		err = c.waitUntilDone(taskId)
		return
	case bosh.DONE:
		return
	default:
		err = bosh.TASKRESULTUNKNOWN
		return
	}
}
