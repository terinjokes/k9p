package main

import (
	"context"
	"flag"
	"net"
	"os"

	p9p "github.com/docker/go-p9p"
	"github.com/oklog/run"
	"github.com/rs/zerolog"
	"go.terinstock.com/k9p/pkg/k9p"
	"go.terinstock.com/k9p/pkg/k9p/logger"
	"k8s.io/client-go/kubernetes"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog"
)

func main() {
	fs := flag.NewFlagSet("k9p", flag.ExitOnError)
	klog.InitFlags(fs)

	fs.Set("logtostderr", "false")
	fs.Set("alsologtostderr", "false")

	var (
		prettyLog  = fs.Bool("pretty-log", false, "output human-friendly logs")
		master     = fs.String("master", "", "The address of the Kubernetes API server (overrides any value in kubeconfig).")
		kubeconfig = fs.String("kubeconfig", "", "Path to kubeconfig file with authorization and master location information.")
		bind9p     = fs.String("bind-9p", ":564", "The address the 9P server should bind and listen on")
	)
	fs.Parse(os.Args[1:])

	ctx := context.Background()

	var log zerolog.Logger
	if *prettyLog {
		log = zerolog.New(zerolog.ConsoleWriter{Out: os.Stderr}).With().Timestamp().Logger()
	} else {
		log = zerolog.New(os.Stderr)
	}

	client, err := createClient(*master, *kubeconfig)
	if err != nil {
		log.Fatal().Err(err).Send()
	}

	klog.SetOutput(log.With().Str("component", "klog").Logger())

	var g run.Group
	{
		ln, err := net.Listen("tcp", *bind9p)
		if err != nil {
			log.Fatal().Err(err).Msg("error listening")
		}

		g.Add(func() error {
			for {
				c, err := ln.Accept()
				if err != nil {
					log.Warn().Err(err).Msg("error accepting")
					continue
				}

				go func(conn net.Conn) {
					defer conn.Close()

					ctx, cancel := context.WithCancel(context.WithValue(ctx, "conn", conn))
					defer cancel()

					log.Info().Str("remote", conn.RemoteAddr().String()).Msg("connected")

					var session p9p.Session
					{
						ksession := k9p.New(ctx, client)
						ksession.WaitForCacheSync(ctx.Done())
						session = logger.New(
							log.With().Str("component", "9p").Logger(),
							ksession,
						)
					}
					if err := p9p.ServeConn(ctx, conn, p9p.Dispatch(session)); err != nil {
						log.Warn().Err(err).Msg("ServeConn")
					}
				}(c)
			}
		}, func(error) {
			ln.Close()
		})
	}

	log.Info().Err(g.Run())
}

func createClient(master string, kubeconfig string) (kubernetes.Interface, error) {
	config, err := clientcmd.BuildConfigFromFlags(master, kubeconfig)
	if err != nil {
		return nil, err
	}
	return kubernetes.NewForConfig(config)
}
