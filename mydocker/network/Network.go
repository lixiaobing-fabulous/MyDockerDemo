package network

import (
	"encoding/json"
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/vishvananda/netlink"
	"net"
	"os"
	"path"
	"path/filepath"
	"strings"
	"text/tabwriter"
)

var (
	defaultNetworkPath = "./network/"
	drivers            = map[string]NetworkDriver{}
	networks           = map[string]*NetWork{}
)

type Endpoint struct {
	ID          string           `json:"id"`
	Device      netlink.Veth     `json:"device"`
	IpAddress   net.IP           `json:"ipAddress"`
	MacAddress  net.HardwareAddr `json:"macAddress"`
	Network     *NetWork         `json:"network"`
	PortMapping []string
}

type NetWork struct {
	Name      string
	IpRange   *net.IPNet
	Driver    string
	GatewayIP net.IP
	Subnet    string
}

type NetworkDriver interface {
	Name() string
	Create(subnet string, name string) (*NetWork, error)
	Delete(network NetWork) error
	Connect(network *NetWork, endpoint *Endpoint) error
	Disconnect(network NetWork, endpoint *Endpoint) error
}

func ListNetwork() error {
	w := tabwriter.NewWriter(os.Stdout, 12, 1, 3, ' ', 0)
	_, _ = fmt.Fprint(w, "NAME\tIpRange\tDriver\n")
	for _, nw := range networks {
		fmt.Fprintf(w, "%s\t%s\t\t%s\n",
			nw.Name,
			nw.IpRange.String(),
			nw.Driver,
		)
	}
	if err := w.Flush(); err != nil {
		return fmt.Errorf("flush error: %w", err)
	}
	return nil
}
func DeleteNetwork(networkName string) error {
	nw, ok := networks[networkName]
	if !ok {
		return fmt.Errorf("no such network: %s", networkName)
	}

	_, ipNet, _ := net.ParseCIDR(nw.Subnet)
	if err := ipAllocator.Release(ipNet, &nw.GatewayIP); err != nil {
		return fmt.Errorf("remove network gateway ip err: %w", err)
	}

	if err := drivers[nw.Driver].Delete(*nw); err != nil {
		return fmt.Errorf("remove network driver err: %w", err)
	}

	return nw.remove(defaultNetworkPath)
}
func (nw *NetWork) remove(dumpPath string) error {
	if _, err := os.Stat(path.Join(dumpPath, nw.Name)); err != nil {
		if os.IsNotExist(err) {
			return nil
		} else {
			return fmt.Errorf("remvove path err: %w", err)
		}
	} else {
		return os.Remove(path.Join(dumpPath, nw.Name))
	}
}

func CreateNetwork(driver, subnet, name string) error {
	nw, err := drivers[driver].Create(subnet, name)
	if err != nil {
		return err
	}
	log.Infof("create network success")
	return nw.dump(defaultNetworkPath)
}
func (nw *NetWork) dump(dumpPath string) error {
	if _, err := os.Stat(dumpPath); err != nil {
		if os.IsNotExist(err) {
			_ = os.MkdirAll(dumpPath, 0644)
		} else {
			return fmt.Errorf("dump path err: %w", err)
		}
	}

	nwPath := path.Join(dumpPath, nw.Name)
	nwFile, err := os.OpenFile(nwPath, os.O_TRUNC|os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		return fmt.Errorf("open file %s, err: %w", nwPath, err)
	}
	defer nwFile.Close()

	nwJson, err := json.Marshal(nw)
	if err != nil {
		return fmt.Errorf("%s file json marshal err: %w", nwPath, err)
	}

	_, err = nwFile.Write(nwJson)
	if err != nil {
		return fmt.Errorf("save network config json err: %w", err)
	}
	return nil
}

func Init() error {
	var bridgeDriver = BridgeNetworkDriver{}
	drivers[bridgeDriver.Name()] = &bridgeDriver
	if _, err := os.Stat(defaultNetworkPath); err != nil {
		if os.IsNotExist(err) {
			_ = os.MkdirAll(defaultNetworkPath, 0644)
		} else {
			return err
		}
	}

	_ = filepath.Walk(defaultNetworkPath, func(nwPath string, info os.FileInfo, err error) error {
		if strings.HasSuffix(nwPath, "/") {
			return nil
		}
		_, nwName := path.Split(nwPath)
		nw := &NetWork{
			Name: nwName,
		}

		if err := nw.load(nwPath); err != nil {
			log.Errorf("error load network: %v", err)
		}

		networks[nwName] = nw
		return nil
	})
	return nil
}
func (nw *NetWork) load(dumpPath string) error {
	nwConfigFile, err := os.Open(dumpPath)
	defer nwConfigFile.Close()
	if err != nil {
		return fmt.Errorf("open file err: %w", err)
	}

	nwJson := make([]byte, 2000)
	n, err := nwConfigFile.Read(nwJson)
	if err != nil {
		return fmt.Errorf("read file err： %w", err)
	}

	err = json.Unmarshal(nwJson[:n], nw)
	if err != nil {
		return fmt.Errorf("load nw info err: %w", err)
	}
	return nil
}
