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
	var config Config
	var ConfigFile = flag.String("config_file", "config/dev.yaml", "Path to the YAML file containing the configuration.")
	flag.Parse()
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
		if cfg.Secret != "" && !isRequestValid(cfg, body, ctx.GetHeader("X-HUB-SIGNATURE")) {
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

		cmd := exec.Command("/bin/sh", "-c", fmt.Sprintf(
			"cd %[1]s; git checkout %[3]s; git pull %[2]s %[3]s",
			cfg.Path,
			cfg.Remote,
			cfg.Branch,
		))
		err = cmd.Start()
		if err != nil {
			log.Print(err)
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
	sha := hex.EncodeToString(h.Sum(nil))

	return hmac.Equal([]byte(sha), []byte(header[5:]))
}
