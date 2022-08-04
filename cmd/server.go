package main

import (
	"encoding/json"
	"flag"
	"net/http"
	"os"

	"github.com/Azure/brigade/pkg/brigade"
	"github.com/Azure/brigade/pkg/storage"
	"github.com/Azure/brigade/pkg/storage/kube"
	"github.com/radu-matei/brigade-vsts-gateway/pkg/vsts"

	log "github.com/Sirupsen/logrus"
	"github.com/gin-gonic/gin"
)

var (
	debug     bool
	namespace string
)

func init() {
	flag.BoolVar(&debug, "debug", true, "enable verbose output")
	flag.StringVar(&namespace, "namespace", "default", "Kubernetes namespace")

	flag.Parse()
	if debug {
		log.SetLevel(log.DebugLevel)
	}
}

func main() {
	client, err := kube.GetClient("", os.Getenv("KUBECONFIG"))
	if err != nil {
		log.Fatalf("cannot get Kubernetes client: %v", err)
	}
	store := kube.New(client, namespace)

	router := setupRouter(store)
	router.Run()
}

func setupRouter(s storage.Store) *gin.Engine {
	router := gin.New()
	router.Use(gin.Recovery())
	router.GET("/healthz", healthz)

	e := router.Group("/vsts")
	e.Use(storeMiddleware(s))
	e.POST("/:project/:token", vstsFn)

	return router
}

// storeMiddleware passes a Brigade storage to the handler func
func storeMiddleware(s storage.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("store", s)
		c.Next()
	}
}

func healthz(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "ok"})
}

func vstsFn(c *gin.Context) {
	s := c.MustGet("store").(storage.Store)
	defer c.Request.Body.Close()

	ev, err := vsts.NewFromRequestBody(c.Request.Body)
	if err != nil {
		log.Debugf("cannot get event from request: %v", err)
	}
	log.Debugf("received event: %v", ev)

	pid := c.Param("project")
	project, err := s.GetProject(pid)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"status": "Resource Not Found"})
		log.Debugf("cannot get project ID: %v", err)
		return
	}
	log.Debugf("found project: %v", project)

	if realToken := project.Secrets["vstsToken"]; realToken != "" {
		tok := c.Param("token")
		if realToken != tok {
			c.JSON(http.StatusForbidden, gin.H{"status": "Forbidden"})
			log.Debugf("token does not match project's version: %v", err)
			return
		}
	}

	payload, err := json.Marshal(ev)
	if err != nil {
		log.Debugf("failed to marshal event: %v", err)
	}

	build := &brigade.Build{
		ProjectID: pid,
		Type:      ev.EventType,
		Provider:  "vsts",
		Payload:   payload,
		Revision: &brigade.Revision{
			Ref:    ev.Resource.RefUpdates[0].Name, // it seems every payload only has a single refUpdate, so hardcoding zero is OK
			Commit: ev.Resource.RefUpdates[0].NewObjectId,
		},
	}

	log.Debugf("created build: %v", build)

	err = s.CreateBuild(build)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "Failed to invoke hook"})
		log.Debugf("failed to create build: %v", err)
		return
	}

	c.JSON(http.StatusOK, ev)
}
