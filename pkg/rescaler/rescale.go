package rescaler

import (
	"context"
	"log"
	"time"

	"github.com/hetznercloud/hcloud-go/hcloud"
)

/* Default logger but more fancy */
var sublogger = log.New(
	log.Writer(),
	"▶ ",
	log.Ldate|log.Ltime,
)

/* Rescale the provided server to the target machine type */
func Rescale(client *hcloud.Client, server *hcloud.Server, targetServerName string) error {
	// Server is already of the target server type
	if server.ServerType.Name == targetServerName {
		sublogger.Printf("Server is already of type %s, rescale skipped.\n", targetServerName)
		return nil
	}

	if server.Status == hcloud.ServerStatusRunning {
		sublogger.Println("Shutting down the server...")
		// Shutdown the server
		action, _, err := client.Server.Shutdown(context.Background(), server)
		if err != nil {
			return err
		}

		// Wait for the server to shut down
		if err := pollAction(client, action); err != nil {
			return err
		}

		// Wait for the hetzner provisioner to be updated
		time.Sleep(time.Second * 30)
		sublogger.Println("done.")
	}

	// Rescale to top server type
	sublogger.Printf("Rescaling server to type %s...\n", targetServerName)
	action, _, err := client.Server.ChangeType(context.Background(), server, hcloud.ServerChangeTypeOpts{
		UpgradeDisk: false,
		ServerType: &hcloud.ServerType{
			Name: targetServerName,
		},
	})
	if err != nil {
		return err
	}

	// Wait for the server to be rescaled
	if err := pollAction(client, action); err != nil {
		return err
	}
	sublogger.Println("done.")

	// Start the server
	sublogger.Println("Starting the server...")
	action, _, err = client.Server.Poweron(context.Background(), server)
	if err != nil {
		return err
	}

	// Wait for the server to be started
	if err := pollAction(client, action); err != nil {
		return err
	}
	sublogger.Println("done.")

	return nil
}

/* Fetch the status of the action until it's completed */
func pollAction(client *hcloud.Client, action *hcloud.Action) error {
	for {
		_action, _, err := client.Action.GetByID(context.Background(), action.ID)
		if err != nil {
			return err
		}

		switch _action.Status {
		case hcloud.ActionStatusError:
			return _action.Error()
		case hcloud.ActionStatusSuccess:
			return nil
		default:
			time.Sleep(time.Second * 5)
			continue
		}
	}
}
