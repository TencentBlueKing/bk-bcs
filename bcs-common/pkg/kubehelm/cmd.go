package kubehelm

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"

	"k8s.io/klog"
)

type cmdHelm struct {}

func NewCmdHelm()KubeHelm{
	return &cmdHelm{}
}

//helm install --name xxxx chart-dir --set k1=v1 --set k2=v2 --kube-apiserver=xxxx --kube-token=xxxxx
func (h *cmdHelm) InstallChart(inf InstallFlags, glf GlobalFlags) error {
	parameters := inf.ParseParameters() + glf.ParseParameters()
	klog.Infof("helm install%s", parameters)
	file, err := os.OpenFile("install.sh", os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		return err
	}
	err = file.Truncate(0)
	if err != nil {
		file.Close()
		return err
	}
	_, err = file.Write([]byte(fmt.Sprintf("helm install%s", parameters)))
	file.Close()
	if err != nil {
		return err
	}
	cmd := exec.Command("/bin/bash","-c","./install.sh")
	buf := bytes.NewBuffer(make([]byte, 1024))
	cmd.Stderr = buf
	err = cmd.Run()
	if err != nil {
		klog.Errorf("helm install failed, stderr %s error %s", buf.String(), err.Error())
		return err
	}

	klog.Infof("helm install %s success", parameters)
	return nil
}
