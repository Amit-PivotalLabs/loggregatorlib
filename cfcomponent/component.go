package cfcomponent

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/cloudfoundry/gosteno"
	"github.com/cloudfoundry/loggregatorlib/cfcomponent/instrumentation"
	"github.com/cloudfoundry/loggregatorlib/cfcomponent/localip"
	"net/http"
)

type Component struct {
	*gosteno.Logger
	IpAddress         string
	HealthMonitor     HealthMonitor
	Type              string //Used by the collector to find data processing class
	Index             uint
	UUID              string
	StatusPort        uint32
	StatusCredentials []string
	Instrumentables   []instrumentation.Instrumentable
}

func NewComponent(logger *gosteno.Logger, componentType string, index uint, heathMonitor HealthMonitor, statusPort uint32, statusCreds []string, instrumentables []instrumentation.Instrumentable) (Component, error) {
	ip, port, err := localip.LocalIPAndPort()
	if err != nil {
		return Component{}, err
	}
	if statusPort == 0 {
		statusPort = port
	}
	if len(statusCreds) == 0 || statusCreds[0] == "" || statusCreds[1] == "" {
		randUser := make([]byte, 42)
		randPass := make([]byte, 42)
		rand.Read(randUser)
		rand.Read(randPass)
		en := base64.URLEncoding
		user := en.EncodeToString(randUser)
		pass := en.EncodeToString(randPass)

		statusCreds = []string{user, pass}
	}

	return Component{
		Logger:            logger,
		IpAddress:         ip,
		Type:              componentType,
		Index:             index,
		HealthMonitor:     heathMonitor,
		StatusPort:        statusPort,
		StatusCredentials: statusCreds,
		Instrumentables:   instrumentables,
	}, nil
}

func (c Component) StartMonitoringEndpoints() error {
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", healthzHandlerFor(c))
	mux.HandleFunc("/varz", varzHandlerFor(c))

	c.Debugf("Starting endpoints for component %s with collect at ip: %s, port: %d, username: %s, password %s", c.UUID, c.IpAddress, c.StatusPort, c.StatusCredentials[0], c.StatusCredentials[1])
	err := http.ListenAndServe(fmt.Sprintf("%s:%d", c.IpAddress, c.StatusPort), mux)
	return err
}

func healthzHandlerFor(c Component) func(w http.ResponseWriter, req *http.Request) {
	return func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusOK)
		if c.HealthMonitor.Ok() {
			fmt.Fprintf(w, "ok")
		} else {
			fmt.Fprintf(w, "bad")
		}
	}
}

func varzHandlerFor(c Component) func(w http.ResponseWriter, req *http.Request) {
	return func(w http.ResponseWriter, req *http.Request) {
		message, err := instrumentation.NewVarzMessage(c.Type, c.Instrumentables)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, err.Error())
			return
		}

		json, err := json.Marshal(message)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, err.Error())
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		w.Write(json)
	}
}
