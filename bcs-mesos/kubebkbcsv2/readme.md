

```shell
kubebuilder create api --group bkbcs --version v2 --kind AdmissionWebhookConfiguration --resource true --controller false
kubebuilder create api --group bkbcs --version v2 --kind Agent --resource true --controller false
kubebuilder create api --group bkbcs --version v2 --kind AgentSchedInfo --resource true --controller false
kubebuilder create api --group bkbcs --version v2 --kind Application --resource true --controller false
kubebuilder create api --group bkbcs --version v2 --kind BcsClusterAgentSetting --resource true --controller false
kubebuilder create api --group bkbcs --version v2 --kind BcsCommandInfo --resource true --controller false
kubebuilder create api --group bkbcs --version v2 --kind BcsConfigMap --resource true --controller false
kubebuilder create api --group bkbcs --version v2 --kind BcsDaemonset --resource true --controller false
kubebuilder create api --group bkbcs --version v2 --kind BcsEndpoint --resource true --controller false
kubebuilder create api --group bkbcs --version v2 --kind BcsSecret --resource true --controller false
kubebuilder create api --group bkbcs --version v2 --kind BcsService --resource true --controller false
kubebuilder create api --group bkbcs --version v2 --kind Transaction --resource true --controller false
kubebuilder create api --group bkbcs --version v2 --kind Crd --resource true --controller false
kubebuilder create api --group bkbcs --version v2 --kind Crr --resource true --controller false
kubebuilder create api --group bkbcs --version v2 --kind Deployment --resource true --controller false
kubebuilder create api --group bkbcs --version v2 --kind Framework --resource true --controller false
kubebuilder create api --group bkbcs --version v2 --kind Task --resource true --controller false
kubebuilder create api --group bkbcs --version v2 --kind TaskGroup --resource true --controller false
kubebuilder create api --group bkbcs --version v2 --kind Version --resource true --controller false
kubebuilder create api --group bkbcs --version v2 --kind BcsTransaction --resource true --controller false
kubebuilder create api --group monitor --version v1 --kind ServiceMonitor --resource true --controller false
```