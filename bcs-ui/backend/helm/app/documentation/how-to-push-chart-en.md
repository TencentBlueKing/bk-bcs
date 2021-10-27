## Push helm chart to repository
Before you begin, this article assumes that you have changed your deployment scheme to Helm Chart format.

In order to use helm，this article will take [Blueking
Game Chart (rumpetroll)]({{rumpetroll_demo_url}}) as an example to explain how to push Chart to the repository.

Note: The content of the article is generated according to the current project, and the account information in it is the real account of the project, please keep it properly.


### 1. Install Helm Flow

  - Install Helm Tools
    + method one: package management tool

    ```
    # package manager for Mac
    brew install helm

    # package manager for Windows
    choco install kubernetes-helm

    # cross-platform systems package manager
    gofish install helm
    ```

    + method two: download binary
        + [Helm](https://github.com/helm/helm/releases/tag/v3.5.4)

  - Initialize Helm Environment

    ```
    helm init --client-only --skip-refresh
    ```

  - Install Push Tools

    + method one: cli install
    ```
    helm plugin install https://github.com/chartmuseum/helm-push
    ```

    + method two: download binary
        +[Helm Push](https://github.com/chartmuseum/helm-push/releases)

### 2. Add Helm Chart Repository
  + node: account and password are private for current project, please keep it safely

    ```
    helm repo add {{ project_code }} {{ repo_url }} --username={{ username }} --password={{ password }}
    ```

### 3. Push Helm Chart
- Prepare Chart

The following will be supplemented by the chart of the Blueking Rumpetroll Game, explaining how to draw the chart. If the project already has a chart, you can use the project's chart directly.

```
wget {{ rumpetroll_demo_url }}
tar -xf rumpetroll-1.0.0.tgz
# edit chart version to 1.0.1
sed -E -i.bak s/version\:\ .+/version\:\ 1\.0\.1/g rumpetroll/Chart.yaml
```

- Push Chart

if the version of `push` plugin is gte 0.10.0, command line is the following:

```
helm cm-push rumpetroll/ {{ project_code }}
```

else, command line is the following:

```
helm push rumpetroll/ {{ project_code }}
```

- After successfully pushing the Chart, you can see output similar to the following:
```
Pushing rumpetroll-1.0.1.tgz to {{ project_code }}...
Done.
```

### 4. Synchronize Project Chart
There are two methods:

- method one: refresh the page manually. In the "Blueking Container Service" -> "Helm" -> "Chart" page, click the "Synchronous Repo" button

- method two: sync every ten minutes automatically

### FAQ
- CASE 1: If adding the repository is successful, pushing Chart fails with error code 411, returning to a Html page, please close the agent first and try again

```
Error: 411: could not properly parse response JSON:
<!DOCTYPE HTML PUBLIC "-//W3C//DTD HTML 4.01 Transitional//EN" "http://www.w3.org/TR/html4/loose.dtd">
<HTML><HEAD>
<META HTTP-EQUIV="Content-Type" CONTENT="text/html; charset=gb2312">
...
```

- CASE 2: The Chart version already exists. If the following 409 error message appears, please modify the Chart version number and try again. To protect your data security, it is forbidden to push repeatedly to the same version.

```
Pushing rumpetroll-0.1.22.tgz to {{ project_code }}...
Error: 409: rumpetroll-0.1.22.tgz already exists
Error: plugin "push" exited with error
```