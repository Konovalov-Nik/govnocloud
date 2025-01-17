package master

import (
	"encoding/json"
	"math/rand"

	"github.com/gin-gonic/gin"
	"github.com/rusik69/govnocloud/pkg/client"
	"github.com/rusik69/govnocloud/pkg/types"
	"github.com/sirupsen/logrus"
)

// CreateContainerHandler handles the create container request.
func CreateContainerHandler(c *gin.Context) {
	body := c.Request.Body
	defer body.Close()
	var tempContainer types.Container
	if err := c.ShouldBindJSON(&tempContainer); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		logrus.Error(err.Error())
		return
	}
	if tempContainer.Name == "" || tempContainer.Image == "" {
		c.JSON(400, gin.H{"error": "name or image is empty"})
		logrus.Error("name or image is empty")
		return
	}
	logrus.Println("Creating container", tempContainer)
	containerInfoString, err := ETCDGet("/containers/" + tempContainer.Name)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		logrus.Error(err.Error())
		return
	}
	if containerInfoString != "" {
		c.JSON(400, gin.H{"error": "container with this id already exists"})
		logrus.Error("container with this id already exists")
		return
	}
	var newContainerID string
	created := false
	var newContainer types.Container
	rand.Shuffle(len(types.MasterEnvInstance.Nodes), func(i, j int) {
		types.MasterEnvInstance.Nodes[i], types.MasterEnvInstance.Nodes[j] = types.MasterEnvInstance.Nodes[j], types.MasterEnvInstance.Nodes[i]
	})
	for _, node := range types.MasterEnvInstance.Nodes {
		newContainerID, err = client.CreateContainer(node.Host, node.Port, tempContainer.Name, tempContainer.Image)
		if err != nil {
			logrus.Error(node.Host, node.Port, err.Error())
			continue
		}
		newContainer.ID = newContainerID
		newContainer.Host = node.Host
		created = true
		break
	}
	if !created {
		c.JSON(500, gin.H{"error": "can't create container"})
		logrus.Error("can't create container", tempContainer.Name, tempContainer.Image)
		return
	}
	newContainer.Committed = true
	newContainer.Image = tempContainer.Image
	newContainer.Name = tempContainer.Name
	newContainer.State = "running"
	newContainerString, err := json.Marshal(newContainer)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		logrus.Error(err.Error())
		return
	}
	err = ETCDPut("/containers/"+tempContainer.Name, string(newContainerString))
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		logrus.Error(err.Error())
		return
	}
	c.JSON(200, newContainer)
}

// DeleteContainerHandler handles the delete container request.
func DeleteContainerHandler(c *gin.Context) {
	name := c.Param("name")
	if name == "" {
		c.JSON(400, gin.H{"error": "name is empty"})
		return
	}
	logrus.Printf("Deleting container with name %s\n", name)
	vmInfoString, err := ETCDGet("/containers/" + name)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		logrus.Error(err.Error())
		return
	}
	if vmInfoString == "" {
		c.JSON(400, gin.H{"error": "container with this name doesn't exist"})
		logrus.Error("container with this name doesn't exist")
		return
	}
	var tempContainer types.Container
	err = json.Unmarshal([]byte(vmInfoString), &tempContainer)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		logrus.Error(err.Error())
		return
	}
	deleted := false
	for _, node := range types.MasterEnvInstance.Nodes {
		if node.Host == tempContainer.Host {
			err = client.DeleteContainer(node.Host, node.Port, tempContainer.ID)
			if err != nil {
				logrus.Error(err.Error())
				break
			}
			deleted = true
		}
	}
	if !deleted {
		c.JSON(500, gin.H{"error": "can't delete container"})
		logrus.Error("can't delete container")
		return
	}
	err = ETCDDelete("/containers/" + name)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		logrus.Error(err.Error())
		return
	}
	c.JSON(200, gin.H{"message": "container deleted"})
}

// ListContainerHandler handles the list container request.
func ListContainerHandler(c *gin.Context) {
	logrus.Println("Listing containers")
	containers, err := ETCDList("/containers/")
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		logrus.Error(err.Error())
		return
	}
	var res []types.Container
	for _, container := range containers {
		var tempContainer types.Container
		c, err := ETCDGet(container)
		if err != nil {
			logrus.Error(err.Error())
			continue
		}
		err = json.Unmarshal([]byte(c), &tempContainer)
		if err != nil {
			logrus.Error(err.Error())
			continue
		}
		res = append(res, tempContainer)
	}
	logrus.Println(res)
	c.JSON(200, res)
}

// GetContainerHandler handles the get container request.
func GetContainerHandler(c *gin.Context) {
	name := c.Param("name")
	if name == "" {
		c.JSON(400, gin.H{"error": "name is empty"})
		return
	}
	logrus.Printf("Getting container with name %s\n", name)
	containerInfoString, err := ETCDGet("/containers/" + name)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		logrus.Error(err.Error())
		return
	}
	if containerInfoString == "" {
		c.JSON(400, gin.H{"error": "container with this name doesn't exist"})
		logrus.Error("container with this name doesn't exist")
		return
	}
	var container types.Container
	err = json.Unmarshal([]byte(containerInfoString), &container)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		logrus.Error(err.Error())
		return
	}
	c.JSON(200, container)
}

// StartContainerHandler handles the start container request.
func StartContainerHandler(c *gin.Context) {
	name := c.Param("name")
	if name == "" {
		c.JSON(400, gin.H{"error": "name is empty"})
		return
	}
	logrus.Printf("Starting container with name %s\n", name)
	containerInfoString, err := ETCDGet("/containers/" + name)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		logrus.Error(err.Error())
		return
	}
	if containerInfoString == "" {
		c.JSON(400, gin.H{"error": "container with this name doesn't exist"})
		logrus.Error("container with this name doesn't exist")
		return
	}
	var container types.Container
	err = json.Unmarshal([]byte(containerInfoString), &container)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		logrus.Error(err.Error())
		return
	}
	started := false
	for _, node := range types.MasterEnvInstance.Nodes {
		if node.Host == container.Host {
			err = client.StartContainer(node.Host, node.Port, container.ID)
			if err != nil {
				logrus.Error(err.Error())
				break
			}
			started = true
		}
	}
	if !started {
		c.JSON(500, gin.H{"error": "can't start container"})
		logrus.Error("can't start container")
		return
	}
	container.State = "running"
	containerString, err := json.Marshal(container)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		logrus.Error(err.Error())
		return
	}
	err = ETCDPut("/containers/"+name, string(containerString))
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		logrus.Error(err.Error())
		return
	}
	c.JSON(200, container)
}

// StopContainerHandler handles the stop container request.
func StopContainerHandler(c *gin.Context) {
	name := c.Param("name")
	if name == "" {
		c.JSON(400, gin.H{"error": "name is empty"})
		return
	}
	logrus.Printf("Stopping container with name %s\n", name)
	containerInfoString, err := ETCDGet("/containers/" + name)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		logrus.Error(err.Error())
		return
	}
	if containerInfoString == "" {
		c.JSON(400, gin.H{"error": "container with this name doesn't exist"})
		logrus.Error("container with this name doesn't exist")
		return
	}
	var container types.Container
	err = json.Unmarshal([]byte(containerInfoString), &container)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		logrus.Error(err.Error())
		return
	}
	stopped := false
	for _, node := range types.MasterEnvInstance.Nodes {
		if node.Host == container.Host {
			err = client.StopContainer(node.Host, node.Port, container.ID)
			if err != nil {
				logrus.Error(err.Error())
				break
			}
			stopped = true
		}
	}
	if !stopped {
		c.JSON(500, gin.H{"error": "can't stop container"})
		logrus.Error("can't stop container")
		return
	}
	container.State = "stopped"
	containerString, err := json.Marshal(container)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		logrus.Error(err.Error())
		return
	}
	err = ETCDPut("/containers/"+name, string(containerString))
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		logrus.Error(err.Error())
		return
	}
	c.JSON(200, container)
}
