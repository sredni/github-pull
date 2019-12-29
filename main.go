package main

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/gin-gonic/gin"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
	"net/http"
	"os/exec"
	"strings"
)

type Config struct {
	Secret string
	Port   string
	Path   string
	Remote string
	Branch string
}

type GithubRequest struct {
	Ref string
}

func main() {
	flag.Parse()
	var ConfigFile = flag.String("config_file", "config/dev.yaml", "Path to the YAML file containing the configuration.")
	var config Config
	err := getConfigFromPath(*ConfigFile, &config)
	if err != nil {
		log.Fatalf("Failed to read the conf file %s: %v", *ConfigFile, err)
	}

	engine := gin.Default()
	engine.POST("/pull", handlePull(config))

	err = engine.Run(fmt.Sprintf(":%s", config.Port))
	log.Fatal("Gin HTTP server exited.", err)
}

func handlePull(cfg Config) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		body, err := ctx.GetRawData()
		if err != nil {
			ctx.Status(http.StatusBadRequest)
			return
		}
		if cfg.Secret != "" && !isRequestValid(cfg, body, ctx.GetHeader("HTTP_X_HUB_SIGNATURE")) {
			ctx.Status(http.StatusBadRequest)
			return
		}

		var data GithubRequest
		err = json.Unmarshal(body, &data)
		if err != nil {
			ctx.Status(http.StatusBadRequest)
			return
		}

		parts := strings.Split(data.Ref, "/")
		if parts[len(parts)-1] != cfg.Branch {
			ctx.Status(http.StatusBadRequest)
			return
		}

		cmd := exec.Command("/bin/sh", "-ctx", fmt.Sprintf(
			"cd %s; git checkout %s; git pull %s %s",
			cfg.Path,
			cfg.Branch,
			cfg.Remote,
			cfg.Branch,
		))
		err = cmd.Start()
		if err != nil {
			ctx.Status(http.StatusInternalServerError)
			return
		}

		ctx.Status(http.StatusOK)
	}
}

func getConfigFromPath(configFilePath string, config interface{}) error {
	configRaw, err := ioutil.ReadFile(configFilePath)
	if err != nil {
		return err
	}
	if err = yaml.UnmarshalStrict(configRaw, config); err != nil {
		return err
	}
	return nil
}

func isRequestValid(cfg Config, body []byte, header string) bool {
	h := hmac.New(sha1.New, []byte(cfg.Secret))
	h.Write(body)
	sha := "sha1=" + hex.EncodeToString(h.Sum(nil))

	return sha == header
}
