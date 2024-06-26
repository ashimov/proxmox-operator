/*
Copyright 2024 The CRASH-Tech.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package proxmox

import (
	"encoding/json"
	"fmt"

	log "github.com/sirupsen/logrus"
)

const (
	STATUS_RUNNING = "running"
	STATUS_STOPPED = "stopped"
)

type Qemu struct {
	node *Node
}

type QemuConfigResp struct {
	Data QemuConfig `json:"data"`
}

type (
	QemuConfig map[string]interface{}
)

type QemuPendingConfigResp struct {
	Data []QemuPendingConfig `json:"data"`
}

type QemuPendingConfig struct {
	Key     string      `json:"key"`
	Value   interface{} `json:"value"`
	Pending interface{} `json:"pending"`
}

type QemuStatus struct {
	Data struct {
		Maxdisk    int64   `json:"maxdisk"`
		Diskread   int     `json:"diskread"`
		Maxmem     int64   `json:"maxmem"`
		CPU        float64 `json:"cpu"`
		BalloonMin int64   `json:"balloon_min"`
		Disk       int     `json:"disk"`
		Qmpstatus  string  `json:"qmpstatus"`
		Uptime     int     `json:"uptime"`
		Netin      int     `json:"netin"`
		Shares     int     `json:"shares"`
		Ha         struct {
			Managed int `json:"managed"`
		} `json:"ha"`
		Diskwrite int    `json:"diskwrite"`
		Vmid      int    `json:"vmid"`
		Mem       int    `json:"mem"`
		Status    string `json:"status"`
		Netout    int    `json:"netout"`
		Cpus      int    `json:"cpus"`
		Name      string `json:"name"`
	} `json:"data"`
}

func checkQemuConfig(qemuConfig QemuConfig) error {
	if _, isExist := qemuConfig["node"]; !isExist {
		return fmt.Errorf("no node name in qemu config")
	}
	if _, isExist := qemuConfig["vmid"]; !isExist {
		return fmt.Errorf("no vmid in qemu config")
	}

	return nil
}

func (qemu *Qemu) Create(qemuConfig QemuConfig) error {
	log.Infof("Creating qemu VM, cluster: %s, node: %s config: %+v", qemu.node.cluster.name, qemu.node.name, qemuConfig)
	err := checkQemuConfig(qemuConfig)
	if err != nil {
		return err
	}

	apiPath := fmt.Sprintf("/nodes/%s/qemu", qemu.node.name)
	err = qemu.node.cluster.PostReq(apiPath, qemuConfig)
	if err != nil {
		return err
	}

	return nil
}

func (qemu *Qemu) Clone(name string, templatePlace, targetPlace QemuPlace) error {
	log.Infof("Cloning qemu VM, cluster: %s, node: %s vmid: %d -> cluster: %s, node: %s vmid: %d name: %s",
		templatePlace.Cluster,
		templatePlace.Node,
		templatePlace.VmId,
		targetPlace.Cluster,
		targetPlace.Node,
		targetPlace.VmId,
		name,
	)

	data := fmt.Sprintf(`{"vmid":"%d", "newid":"%d", "name":"%s", "target":"%s", "full":true}`, templatePlace.VmId, targetPlace.VmId, name, targetPlace.Node)
	apiPath := fmt.Sprintf("/nodes/%s/qemu/%d/clone", templatePlace.Node, templatePlace.VmId)
	err := qemu.node.cluster.PostReq(apiPath, data)
	if err != nil {
		return err
	}

	return nil
}

func (qemu *Qemu) SetConfig(qemuConfig QemuConfig) error {
	log.Infof("Set qemu VM config, cluster: %s, node: %s config: %+v", qemu.node.cluster.name, qemu.node.name, qemuConfig)
	err := checkQemuConfig(qemuConfig)
	if err != nil {
		return err
	}

	apiPath := fmt.Sprintf("/nodes/%s/qemu/%d/config", qemu.node.name, qemuConfig["vmid"])
	err = qemu.node.cluster.PostReq(apiPath, qemuConfig)
	if err != nil {
		return err
	}

	return nil
}

func (qemu *Qemu) GetConfig(vmId int) (QemuConfig, error) {
	log.Debugf("Get qemu config, cluster: %s node: %s vmid: %d", qemu.node.cluster.name, qemu.node.name, vmId)
	apiPath := fmt.Sprintf("/nodes/%s/qemu/%d/config", qemu.node.name, vmId)
	data := fmt.Sprintf(`{"node":"%s", "vmid":"%d"}`, qemu.node.name, vmId)

	qemuConfigData, err := qemu.node.cluster.GetReq(apiPath, data)
	if err != nil {
		return nil, err
	}

	qemuConfig := QemuConfigResp{}
	err = json.Unmarshal(qemuConfigData, &qemuConfig)
	if err != nil {
		return nil, err
	}

	log.Debug(qemuConfig.Data)
	return qemuConfig.Data, nil
}

func (qemu *Qemu) GetPendingConfig(vmId int) ([]QemuPendingConfig, error) {
	log.Debugf("Get qemu pending config, cluster: %s node: %s vmid: %d", qemu.node.cluster.name, qemu.node.name, vmId)
	apiPath := fmt.Sprintf("/nodes/%s/qemu/%d/pending", qemu.node.name, vmId)
	data := fmt.Sprintf(`{"node":"%s", "vmid":"%d"}`, qemu.node.name, vmId)

	qemuConfigData, err := qemu.node.cluster.GetReq(apiPath, data)
	if err != nil {
		return []QemuPendingConfig{}, err
	}

	pendingConfig := QemuPendingConfigResp{}
	err = json.Unmarshal(qemuConfigData, &pendingConfig)
	if err != nil {
		return []QemuPendingConfig{}, err
	}

	log.Debug(pendingConfig.Data)
	return pendingConfig.Data, nil
}

func (qemu *Qemu) Delete(vmId int) error {
	log.Infof("Deleting qemu VM, cluster: %s node: %s vmid: %d", qemu.node.cluster.name, qemu.node.name, vmId)
	apiPath := fmt.Sprintf("/nodes/%s/qemu/%d", qemu.node.name, vmId)
	err := qemu.node.cluster.DeleteReq(apiPath)
	if err != nil {
		return err
	}

	return nil
}

func (qemu *Qemu) Start(vmId int) error {
	log.Infof("Starting qemu, cluster: %s node: %s vmid: %d", qemu.node.cluster.name, qemu.node.name, vmId)
	data := fmt.Sprintf(`{"node":"%s", "vmid":"%d"}`, qemu.node.name, vmId)
	apiPath := fmt.Sprintf("/nodes/%s/qemu/%d/status/start", qemu.node.name, vmId)
	err := qemu.node.cluster.PostReq(apiPath, data)
	if err != nil {
		return err
	}

	return nil
}

func (qemu *Qemu) Stop(vmId int) error {
	log.Infof("Stopping qemu, cluster: %s node: %s vmid: %d", qemu.node.cluster.name, qemu.node.name, vmId)
	data := fmt.Sprintf(`{"node":"%s", "vmid":"%d"}`, qemu.node.name, vmId)
	apiPath := fmt.Sprintf("/nodes/%s/qemu/%d/status/stop", qemu.node.name, vmId)
	err := qemu.node.cluster.PostReq(apiPath, data)
	if err != nil {
		return err
	}

	return nil
}

func (qemu *Qemu) GetStatus(vmId int) (QemuStatus, error) {
	log.Debugf("Get qemu status, cluster: %s node: %s vmid: %d", qemu.node.cluster.name, qemu.node.name, vmId)

	apiPath := fmt.Sprintf("/nodes/%s/qemu/%d/status/current", qemu.node.name, vmId)
	data := fmt.Sprintf(`{"node":"%s", "vmid":"%d"}`, qemu.node.name, vmId)

	qemuStatusData, err := qemu.node.cluster.GetReq(apiPath, data)
	if err != nil {
		return QemuStatus{}, err
	}

	qemuStatus := QemuStatus{}
	err = json.Unmarshal(qemuStatusData, &qemuStatus)
	if err != nil {
		return QemuStatus{}, err
	}

	log.Debug(qemuStatus)
	return qemuStatus, nil
}

func (qemu *Qemu) Resize(vmId int, disk, size string) error {
	log.Infof("Resize qemu disk, cluster: %s node: %s vmid: %d disk: %s size: %s", qemu.node.cluster.name, qemu.node.name, vmId, disk, size)
	data := fmt.Sprintf(`{"node":"%s", "vmid":"%d", "disk":"%s", "size":"%s"}`, qemu.node.name, vmId, disk, size)
	apiPath := fmt.Sprintf("/nodes/%s/qemu/%d/resize", qemu.node.name, vmId)
	err := qemu.node.cluster.PutReq(apiPath, data)
	if err != nil {
		return err
	}

	return nil
}
