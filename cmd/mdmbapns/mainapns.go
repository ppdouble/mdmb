package main

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/go-kit/kit/log"
	httptransport "github.com/go-kit/kit/transport/http"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/groob/finalizer/logutil"
	"github.com/jessepeterson/mdmb/internal/device"
	"github.com/micromdm/go4/httputil"
	"github.com/pkg/errors"
	bolt "go.etcd.io/bbolt"
	mathrand "math/rand"
	"net/http"
	"os"
	"strings"
	"time"
)

type Server struct {
	ConfigPath             string
	//PubClient              pubsub.PublishSubscriber
	//DB                     *bolt.DB
	ServerPublicURL        string
	CommandWebhookURL      string

	//ProfileDB              profile.Store
	//ConfigDB               config.Store
	//Queue                  string

	//APNSPushService apns.Service
	//CommandService  command.Service
	//MDMService      mdm.Service

	WebhooksHTTPClient *http.Client
}


func main() {

	var run func([]string) error
	run = serve

	if err := run(os.Args[2:]); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}

func NewRouter(logger log.Logger) (*mux.Router, []httptransport.ServerOption) {
	r := mux.NewRouter()
	options := []httptransport.ServerOption{
		httptransport.ServerErrorEncoder(ErrorEncoder),
		httptransport.ServerErrorLogger(logger),
		httptransport.ServerBefore(httptransport.PopulateRequestContext),
	}
	return r, options
}

func ErrorEncoder(_ context.Context, err error, w http.ResponseWriter) {
	errMap := map[string]interface{}{"error": err.Error()}
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")

	if headerer, ok := err.(httptransport.Headerer); ok {
		for k, values := range headerer.Headers() {
			for _, v := range values {
				w.Header().Add(k, v)
			}
		}
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	code := http.StatusInternalServerError
	if sc, ok := err.(httptransport.StatusCoder); ok {
		code = sc.StatusCode()
	}
	w.WriteHeader(code)

	enc.Encode(errMap)
}

func serve(args []string) error {
	flagset := flag.NewFlagSet("serve", flag.ExitOnError)
	var (
		//flConfigPath             = flagset.String("config-path", "/var/db/micromdm/micormdm.db", "Path to configuration directory")
		//flServerURL              = flagset.String("server-url", "https://127.0.0.1:8999", "Public HTTPS url of your server")
		flTLS                    = flagset.Bool("tls", false, "Use https")
		flHTTPAddr               = flagset.String("http-addr", "127.0.0.1:8989", "http(s) listen address of mdm server. defaults to :8090 if tls is false")
		//flCommandWebhookURL      = flagset.String("command-webhook-url", "/webhook", "URL to send command responses")

	)

	//if *flServerURL == "" {
	//	return errors.New("must supply -server-url")
	//}
	//if !strings.HasPrefix(*flServerURL, "https://") {
	//	return errors.New("-server-url must begin with https://")
	//}
	logger := log.NewLogfmtLogger(os.Stderr)
	mainLogger := log.With(logger, "component", "main")
	mainLogger.Log("msg", "started")

	//if err := os.MkdirAll(*flConfigPath, 0755); err != nil {
	//	return errors.Wrapf(err, "creating config directory %s", *flConfigPath)
	//}
	//sm := &Server{
	//	//ConfigPath:             *flConfigPath,
	//	ServerPublicURL:        strings.TrimRight(*flServerURL, "/"),
	//	CommandWebhookURL:      *flCommandWebhookURL,
	//	WebhooksHTTPClient: &http.Client{Timeout: time.Second * 30},
	//}


	httpLogger := log.With(logger, "transport", "http")

	r, _ := NewRouter(logger)
	r.Handle("/mock/apns/{devicetoken}", mockAPNSHandler())
	r.Handle("/3/device/{devicetoken}", mockAPNSHandler())

	var handler http.Handler

	handler = r

	handler = logutil.NewHTTPLogger(httpLogger).Middleware(handler)


	serveOpts := serveOptions(
		handler,
		*flHTTPAddr,
		logger,
		//sm.ConfigPath,
		*flTLS,
	)
	err := httputil.ListenAndServe(serveOpts...)
	return errors.Wrap(err, "calling ListenAndServe")
}

// serveOptions configures the []httputil.Options for ListenAndServe
func serveOptions( handler http.Handler, addr string, logger log.Logger,
	//configPath string,
	tls bool, ) []httputil.Option {

	serveOpts := []httputil.Option{
		httputil.WithLogger(logger),
		httputil.WithHTTPHandler(handler),
	}
	if !tls && addr == ":https" {
		serveOpts = append(serveOpts, httputil.WithAddress(":8090"))
	}
	//if tls {
	//	serveOpts = append(serveOpts, httputil.WithAutocertCache(autocert.DirCache(filepath.Join(configPath, "le-certificates"))))
	//}
	if addr != ":https" {
		serveOpts = append(serveOpts, httputil.WithAddress(addr))
	}
	return serveOpts
}

type ApnsResInfo struct {
	Status             string `json:"status"`
}

// Handler provides an HTTP Handler which returns JSON formatted version info.
func mockAPNSHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		var (
			dbPath = "mdmb.db"
		)

		deviceTokenHex := mux.Vars(r)["devicetoken"]
		//deviceTokenByte := []byte(deviceTokenHex)
		//deviceToken := make([]byte, len(deviceTokenByte))
		//hex.Decode(deviceToken,[]byte(deviceTokenHex))
		fmt.Printf("devicetokenHex = %s\n", deviceTokenHex)
		//fmt.Printf("devicetoken = %s\n", deviceToken)
		deviceToken, _:=hex.DecodeString(deviceTokenHex)
		deviceTokenStr := fmt.Sprintf("%s", deviceToken)
		fmt.Printf("devicetokenstr = %s\n", deviceTokenStr)



		push_notification_id := r.Header.Get("apns-id")
		fmt.Printf("udidfrommicromdm header: apns-id:,  = %s\n", push_notification_id)
		if "" == strings.TrimSpace(push_notification_id) {
			push_notification_id = strings.ToUpper(uuid.NewString())
			fmt.Printf("create response push_notification_id: apns-id:,  = %s\n", push_notification_id)
		}

		// TODO Query the device ID according to devicetoken given by micromdm

		db, err := bolt.Open(dbPath, 0644, nil)
		if err != nil {
			fmt.Errorf("%s", err)
		}
		fmt.Printf("db status: %s,\n db Info %s\n", db.Stats(), db.Info())
		defer db.Close()

		mathrand.Seed(time.Now().UnixNano())


		rctx := RunContext{
			DB: db,
			UUIDs: []string{deviceTokenStr},
		}
		fmt.Printf("rctx.UUIDs[0] %s", rctx.UUIDs[0])

		// TODO mdm.Connect

		devicesConnect(rctx)

		// TODO response apns
		apnsRes := &ApnsResInfo{
			Status: "success",
		}

		w.Header().Set("apns-id", push_notification_id)
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		enc := json.NewEncoder(w)
		enc.SetIndent("", "  ")
		enc.Encode(apnsRes)
		fmt.Printf("%s - apns status %s\n", apnsRes, apnsRes.Status)
	})
}


func devicesConnect(rctx RunContext) {
	var (
		workers    = 1
		iterations = 1
	)

	//err := checkDeviceUUIDs(rctx, false, name)
	//if err != nil {
	//	fmt.Errorf("%s", err)
	//}

	workerData := []*ConnectWorkerData{}

	for _, u := range rctx.UUIDs {
		dev, err := device.Load(u, rctx.DB)
		if err != nil {
			fmt.Println(err)
			continue
		}

		client, err := dev.MDMClient()
		if err != nil {
			fmt.Println(err)
			continue
		}

		workerData = append(workerData, &ConnectWorkerData{
			Device:    dev,
			MDMClient: client,
		})
	}

	startConnectWorkers(workerData, workers, iterations)
}



func checkDeviceUUIDs(rctx RunContext, requireEmpty bool, subCmdName string) error {
	if requireEmpty && len(rctx.UUIDs) != 0 {
		return errors.New("cannot supply UUIDs for " + subCmdName)
	} else if !requireEmpty && len(rctx.UUIDs) < 1 {
		return errors.New("no device UUIDs supplied, use -uuids argument for " + subCmdName)
	}
	return nil
}

// RunContext contains "global" runtime environment settings
type RunContext struct {
	DB    *bolt.DB
	UUIDs []string
}

func printExamples() {
	const exampleText = `
		Quickstart:
		mdmb serve
		`
	fmt.Println(exampleText)
}
