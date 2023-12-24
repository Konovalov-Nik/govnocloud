package master

import (
	"encoding/json"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/rusik69/govnocloud/pkg/client"
	"github.com/rusik69/govnocloud/pkg/types"
	"github.com/sirupsen/logrus"
)

// CreateVMHandler handles the create vm request.
func CreateVMHandler(c *gin.Context) {
	body := c.Request.Body
	defer body.Close()
	var tempVM types.VM
	if err := c.ShouldBindJSON(&tempVM); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		logrus.Error(err.Error())
		return
	}
	if tempVM.Name == "" || tempVM.Image == "" || tempVM.Flavor == "" {
		c.JSON(400, gin.H{"error": "name, image or flavor is empty"})
		logrus.Error("name, image or flavor is empty")
		return
	}
	logrus.Println("Creating VM", tempVM)
	vmInfoString, err := ETCDGet("/vms/" + tempVM.Name)
	if err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		logrus.Error(err.Error())
		return
	}
	if vmInfoString != "" {
		c.JSON(400, gin.H{"error": "vm with this id already exists"})
		logrus.Error("vm with this id already exists")
		return
	}
	newVMID := 0
	created := false
	var newVM types.VM
	for _, node := range types.MasterEnvInstance.Nodes {
		newVMID, err = client.CreateVM(node.Host, node.Port, tempVM.Name, tempVM.Image, tempVM.Flavor)
		if err != nil {
			logrus.Error(err.Error())
			continue
		}
		newVM.ID = newVMID
		newVM.Host = node.Host
		created = true
	}
	if !created {
		c.JSON(500, gin.H{"error": "vm was not created"})
		logrus.Error("vm was not created")
		return
	}
	newVM.Committed = true
	newVmstring, err := json.Marshal(newVM)
	if err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		logrus.Error(err.Error())
		return
	}
	err = ETCDPut("/vms/"+newVM.Name, string(newVmstring))
	if err != nil {

		c.JSON(400, gin.H{"error": err.Error()})
		logrus.Error(err.Error())
		return
	}
	return
}

// DeleteVMHandler handles the delete request.
func DeleteVMHandler(c *gin.Context) {
	body := c.Request.Body
	defer body.Close()
	var tempVM types.VM
	if err := c.ShouldBindJSON(&tempVM); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		logrus.Error(err.Error())
		return
	}
	if tempVM.ID == 0 {
		c.JSON(400, gin.H{"error": "id is empty"})
		logrus.Error("id is empty")
		return
	}
	logrus.Printf("Deleting VM %d\n", tempVM.ID)
	tempVMIDString := strconv.Itoa(tempVM.ID)
	vmInfoString, err := ETCDGet("/vms/" + tempVMIDString)
	if err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		logrus.Error(err.Error())
		return
	}
	if vmInfoString == "" {
		c.JSON(400, gin.H{"error": "vm with this id does not exist"})
		logrus.Error("vm with this id does not exist")
		return
	}
	var vmInfo types.VM
	err = json.Unmarshal([]byte(vmInfoString), &vmInfo)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		logrus.Error(err.Error())
		return
	}
	for _, node := range types.MasterEnvInstance.Nodes {
		if node.Host == vmInfo.Host {
			err = client.DeleteVM(node.Host, node.Port, tempVM.ID)
			if err != nil {
				logrus.Error(err.Error())
				c.JSON(500, gin.H{"error": err.Error()})
				break
			}
		}
	}
	err = ETCDDelete("/vms/" + tempVMIDString)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		logrus.Error(err.Error())
		return
	}
	c.JSON(200, gin.H{"status": "ok"})
	return
}

// GetVMHandler handles the get vm info request.
func GetVMHandler(c *gin.Context) {
	body := c.Request.Body
	defer body.Close()
	var tempVM types.VM
	if err := c.ShouldBindJSON(&tempVM); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		logrus.Error(err.Error())
		return
	}
	if tempVM.ID == 0 {
		c.JSON(400, gin.H{"error": "id is empty"})
		logrus.Error("id is empty")
		return
	}
	logrus.Printf("Getting VM %d\n", tempVM.ID)
	vmIDString := strconv.Itoa(tempVM.ID)
	vmInfoString, err := ETCDGet("/vms/" + vmIDString)
	if err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		logrus.Error(err.Error())
		return
	}
	if vmInfoString == "" {
		c.JSON(400, gin.H{"error": "vm with this id does not exist"})
		logrus.Error("vm with this id does not exist")
		return
	}
	var vmInfo types.VM
	err = json.Unmarshal([]byte(vmInfoString), &vmInfo)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		logrus.Error(err.Error())
		return
	}
	c.JSON(200, vmInfo)
	return
}

// ListVMHandler handles the list vm request.
func ListVMHandler(c *gin.Context) {
	logrus.Println("Listing VMs")
	vms, err := ETCDList("/vms/")
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		logrus.Error(err.Error())
		return
	}
	c.JSON(200, vms)
	return
}