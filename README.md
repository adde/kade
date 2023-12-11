# KADE

__KADE__ or __Kubernetes Application Deployment Engine__ is a simple CLI tool that helps streamline the process of setting up and configuring web application environments in Kubernetes.

## Installation

You can download the latest version of the app from the [release section](https://github.com/adde/kade/releases/latest) of this repo.

Below follows an example to download and install the latest version on Mac(Apple Silicon). For other platforms, replace the binary name in the curl command with the relevant binary for your system:

### Download the latest release:

```sh
curl -LO https://github.com/adde/kade/releases/latest/download/kade-darwin-arm64
```

### Make the kade binary executable:

```sh
chmod +x ./kade-darwin-arm64
```

### Move the kade binary to a file location on your system `PATH`:

```sh
sudo mv ./kade-darwin-arm64 /usr/local/bin/kade
```

### Test to ensure the version you installed is up-to-date:

```sh
kade --version
```
The command above should output the version number of the app.    
(since the app is unsigned, the first time you try to run the app you have to allow the app in security settings)

## Configuration

Before running the app, make sure that you have a valid kubeconfig for the cluster that you intend to work with in the following path:

```sh
~/.kube/config
```

## Usage

In a terminal, simply run the following command and follow the steps presented:

```sh
kade
```

## Disclaimer

Do not, I repeat, DO NOT use this tool to deploy applications to a production cluster. This tool is for testing purposes only.

## Acknowledgements

This library is built on top of many great libraries, especially the following:

* https://github.com/erikgeiser/promptkit/

## License

MIT License © 2023-Present [Andreas Jönsson](https://github.com/adde)
