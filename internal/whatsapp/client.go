package whatsapp

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/mdp/qrterminal/v3"
	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/store/sqlstore"

	_ "modernc.org/sqlite"
)

type Client struct {
	WA      *whatsmeow.Client
	handler *Handler
}

func NewClient(ctx context.Context, dbPath string) (*Client, error) {

	if err := os.MkdirAll(filepath.Dir(dbPath), 0750); err != nil {
		return nil, fmt.Errorf("create whatsapp db dir: %w", err)
	}

	dbLog := newWaLog("WAStore")

	container, err := sqlstore.New(ctx, "sqlite",
		"file:"+dbPath+"?_pragma=foreign_keys(1)&_pragma=busy_timeout(10000)", dbLog)
	if err != nil {
		return nil, fmt.Errorf("open whatsapp device store: %w", err)
	}

	deviceStore, err := container.GetFirstDevice(ctx)
	if err != nil {
		return nil, fmt.Errorf("get whatsapp device: %w", err)
	}

	waClient := whatsmeow.NewClient(deviceStore, newWaLog("WAClient"))
	waClient.EnableAutoReconnect = true
	waClient.AutoTrustIdentity = true

	c := &Client{WA: waClient}
	return c, nil
}

func (c *Client) SetHandler(h *Handler) {
	c.handler = h
	c.WA.AddEventHandler(h.HandleEvent)
}

func (c *Client) Connect(ctx context.Context) error {
	if c.WA.Store.ID == nil {
		return c.connectWithQR(ctx)
	}
	return c.connectExisting()
}

func (c *Client) connectWithQR(ctx context.Context) error {
	slog.Info("whatsapp session not found, starting QR pairing")

	qrChan, err := c.WA.GetQRChannel(ctx)
	if err != nil {
		return fmt.Errorf("get qr channel: %w", err)
	}

	if err := c.WA.Connect(); err != nil {
		return fmt.Errorf("connect: %w", err)
	}

	go func() {
		for evt := range qrChan {
			switch evt.Event {
			case "code":

				printQR(evt.Code)
			case "success":
				slog.Info("whatsapp qr pairing successful")
			case "timeout":
				slog.Warn("whatsapp qr pairing timeout")
			case "err":
				slog.Error("whatsapp qr pairing error", "error", evt.Error)
			}
		}
	}()

	return nil
}

func (c *Client) connectExisting() error {
	slog.Info("whatsapp session found, connecting")
	if err := c.WA.Connect(); err != nil {
		return fmt.Errorf("connect: %w", err)
	}
	return nil
}

func (c *Client) Disconnect() {
	slog.Info("disconnecting from whatsapp")
	c.WA.Disconnect()
}

func printQR(code string) {
	fmt.Println("\n╔════════════════════════════════╗")
	fmt.Println("║   Scan QR Code di WhatsApp     ║")
	fmt.Println("║   WhatsApp > Linked Devices    ║")
	fmt.Println("╚════════════════════════════════╝")
	fmt.Println()
	qrterminal.GenerateHalfBlock(code, qrterminal.L, os.Stdout)
	fmt.Println()
}
