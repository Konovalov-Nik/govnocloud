package master

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/rusik69/govnocloud/pkg/types"
	"github.com/sirupsen/logrus"
)

// Serve starts the server.
func Serve() {
	r := gin.New()
	r.Use(cors.Default())
	r.POST("/api/v1/vms", CreateVMHandler)
	r.DELETE("/api/v1/vm/:name", DeleteVMHandler)
	r.GET("/api/v1/vms", ListVMHandler)
	r.GET("/api/v1/vmstart/:name", StartVMHandler)
	r.GET("/api/v1/vmstop/:name", StopVMHandler)
	r.GET("/api/v1/vm/:name", GetVMHandler)
	r.GET("/api/v1/container/:name", GetContainerHandler)
	r.POST("/api/v1/containers", CreateContainerHandler)
	r.GET("/api/v1/containerstart/:name", StartContainerHandler)
	r.GET("/api/v1/containerstop/:name", StopContainerHandler)
	r.DELETE("/api/v1/container/:name", DeleteContainerHandler)
	r.GET("/api/v1/containers", ListContainerHandler)
	r.POST("/api/v1/nodes", AddNodeHandler)
	r.GET("/api/v1/nodes", ListNodesHandler)
	r.GET("/api/v1/node/:name", GetNodeHandler)
	r.DELETE("/api/v1/node/:name", DeleteNodeHandler)
	r.POST("/api/v1/files", UploadFileHandler)
	r.GET("/api/v1/filecommit/:name", CommitFileHandler)
	r.DELETE("/api/v1/file/:name", DeleteFileHandler)
	r.GET("/api/v1/files", ListFilesHandler)
	r.GET("/api/v1/file/:name", GetFileHandler)
	logrus.Println("Master is listening on port " + string(types.MasterEnvInstance.ListenPort))
	r.Run(":" + types.MasterEnvInstance.ListenPort)
}
