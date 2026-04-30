package cmd

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/spf13/cobra"

	"github.com/cars24/seo-automation/internal/server"
)

var (
	flagServerPort    int
	flagServerBaseDir string
	flagUIDir         string
)

var serverCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start the SEO Audit HTTP API server",
	Long: `Starts an HTTP server that exposes the SEO audit engine as a REST API.

Endpoints:
  POST   /api/audits                    start a new audit
  GET    /api/audits                    list all past audits
  GET    /api/audits/defaults           get backend-owned audit defaults and UI controls
  GET    /api/audits/{id}               get audit metadata
  DELETE /api/audits/{id}               delete audit + reports
  POST   /api/audits/{id}/cancel        cancel a running audit
  GET    /api/audits/{id}/events        SSE stream of crawl progress
  GET    /api/audits/{id}/report.html   serve the interactive HTML report
  GET    /api/audits/{id}/report.json   serve the raw JSON report
  GET    /api/audits/diff?a={id}&b={id} compare two audits`,
	Example: `  seo-audit serve
  seo-audit serve --port 9090
  seo-audit serve --reports-dir ~/my-seo-reports`,
	RunE: runServer,
}

func init() {
	serverCmd.Flags().IntVar(&flagServerPort, "port", 8080, "HTTP server port")
	serverCmd.Flags().StringVar(&flagServerBaseDir, "reports-dir", "", "Root directory for audit reports (default: ~/.seo-reports)")
	serverCmd.Flags().StringVar(&flagUIDir, "ui-dir", "", "Path to built frontend assets (e.g. ui/dist)")
}

func runServer(cmd *cobra.Command, args []string) error {
	srv, err := server.New(flagServerBaseDir, flagUIDir)
	if err != nil {
		return fmt.Errorf("initialise server: %w", err)
	}

	addr := fmt.Sprintf(":%d", flagServerPort)
	httpSrv := &http.Server{
		Addr:         addr,
		Handler:      srv.Handler(),
		ReadTimeout:  5 * time.Minute, // generous for long crawls
		WriteTimeout: 0,               // 0 = no timeout (SSE streams are long-lived)
		IdleTimeout:  60 * time.Second,
	}

	// Graceful shutdown on SIGINT / SIGTERM
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-quit
		fmt.Fprintln(os.Stderr, "\nShutting down…")
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		_ = httpSrv.Shutdown(ctx)
	}()

	fmt.Printf("SEO Audit Server  →  http://localhost%s\n", addr)
	if err := httpSrv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("server error: %w", err)
	}
	return nil
}
