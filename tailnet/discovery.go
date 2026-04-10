package tailnet

import (
	"context"
	"log"
	"time"

	"hiclaw-server/core"

	"tailscale.com/client/local"
)

type Discovery struct {
	client    local.Client
	deviceSvc core.DeviceService
	agentName string
}

func NewDiscovery(deviceSvc core.DeviceService, agentName string) *Discovery {
	return &Discovery{deviceSvc: deviceSvc, agentName: agentName}
}

func (d *Discovery) SyncOnce(ctx context.Context) error {
	status, err := d.client.Status(ctx)
	if err != nil {
		return err
	}
	for _, peer := range status.Peer {
		if len(peer.TailscaleIPs) == 0 {
			continue
		}
		device := &core.Device{
			IP:      peer.TailscaleIPs[0].String(),
			Name:    peer.HostName,
			IsAgent: peer.HostName == d.agentName,
		}
		if err := d.deviceSvc.Register(device); err != nil {
			log.Printf("register %s failed: %v", device.Name, err)
		}
	}
	return nil
}

func (d *Discovery) RunLoop(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := d.SyncOnce(ctx); err != nil {
				log.Printf("tailnet sync: %v", err)
			}
		}
	}
}
