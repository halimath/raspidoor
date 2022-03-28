package main

import (
	"embed"
	"errors"
	"flag"
	"fmt"
	"html/template"
	"io"
	"io/fs"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/halimath/raspidoor/controller"
	"github.com/halimath/raspidoor/systemd/notify"
	"github.com/halimath/raspidoor/webapp/internal/config"
	"google.golang.org/grpc"
)

var (
	Version        = "0.1.0"
	Revision       = "local"
	BuildTimestamp = "0000-00-00T00:00:00"

	configFile = flag.String("config-file", "", "The config file to read instead of the default configuration files")

	//go:embed public/*.css
	public embed.FS

	//go:embed templates/index.html
	indexTemplateString string

	indexTemplate *template.Template
)

func init() {
	indexTemplate = template.Must(template.New("index").Parse(indexTemplateString))
}

func main() {
	flag.Parse()

	var c *config.Config
	var err error

	if *configFile == "" {
		c, err = config.ReadConfig()
	} else {
		c, err = config.ReadConfigFromFile(*configFile)
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "%s: failed to read config file %s: %s\n", os.Args[0], *configFile, err)
		os.Exit(1)
	}

	logger, err := c.NewLogger()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s: failed to create logger: %s\n", os.Args[0], err)
		os.Exit(2)
	}

	logger.Info("Connecting to daemon on %s", c.Socket)

	dialer := func(addr string, t time.Duration) (net.Conn, error) {
		return net.Dial("unix", addr)
	}

	con, err := grpc.Dial(c.Socket, grpc.WithInsecure(), grpc.WithDialer(dialer))
	if err != nil {
		logger.Error("Connecting to daemon on %s failed: %s", c.Socket, err)
		fmt.Fprintf(os.Stderr, "%s: Failed to open socket %s: %s\n", os.Args[0], c.Socket, err)
		os.Exit(2)
	}
	defer con.Close()

	ctrl := controller.NewControllerClient(con)

	mux := http.NewServeMux()

	staticAssets, err := fs.Sub(public, "public")
	if err != nil {
		panic(err)
	}
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.FS(staticAssets))))

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		info, err := ctrl.Info(r.Context(), &controller.Empty{})
		if err != nil {
			logger.Error("Failed to load info: %s", err)
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusOK)

		indexTemplate.Execute(w, info)
	})

	mux.HandleFunc("/update", func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		r.Body.Close()
		if err != nil {
			logger.Err(err)
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}

		form, err := url.ParseQuery(string(body))
		if err != nil {
			logger.Err(err)
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}

		target := form.Get("target")
		state, err := strconv.ParseBool(form.Get("state"))
		if err != nil {
			logger.Err(err)
			http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
			return
		}
		index, err := strconv.ParseInt(form.Get("index"), 10, 32)
		if err != nil {
			logger.Err(err)
			http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
			return
		}

		req := &controller.EnabledState{
			State: state,
			Index: int32(index),
		}

		if target == "bell" {
			req.Target = controller.Target_BELL
		} else {
			req.Target = controller.Target_BELL_PUSH
		}

		if _, err := ctrl.SetState(r.Context(), req); err != nil {
			logger.Err(err)
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
	})

	notifier := notify.Detect(logger)

	httpServer := &http.Server{
		Addr:    c.Address,
		Handler: mux,
	}
	httpCloseChan := make(chan error)

	go func() {
		logger.Info("Listening on %s", c.Address)
		httpCloseChan <- httpServer.ListenAndServe()
	}()

	if err := notifier.Notify(notify.NotificationReady); err != nil {
		logger.Err(err)
	}

	signalChan := make(chan os.Signal)
	signal.Notify(signalChan, os.Interrupt, syscall.SIGTERM, syscall.SIGQUIT)

	select {
	case <-signalChan:
		logger.Info("Shutting down")
		httpServer.Close()
	case err := <-httpCloseChan:
		if !errors.Is(err, http.ErrServerClosed) {
			logger.Err(err)
		}
	}

	if err := notifier.Notify(notify.NotificationStopping); err != nil {
		logger.Err(err)
	}

}
