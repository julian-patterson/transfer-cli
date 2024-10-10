# TransferCLI

TransferCLI is a command-line tool for performing SSH operations, such as listing and transferring files between a remote directory and a local directory.

## Installation

To install TransferCLI, follow these steps:

1. Clone the repository:

   ```sh
   git clone https://github.com/yourusername/transfercli.git
   cd transfercli
   ```

2. Run the installation script:

   ```sh
   sudo chmod +x install.sh
   ./install.sh
   ```

   The installation script will:

   - Check if Go is installed, and if not, install it.
   - Install all project dependencies.
   - Create a [config.yaml](http://_vscodecontentref_/#%7B%22uri%22%3A%7B%22%24mid%22%3A1%2C%22fsPath%22%3A%22%2Fhome%2Fjulian%2FDeveloper%2Fgo-cli%2Fconfig.yaml%22%2C%22path%22%3A%22%2Fhome%2Fjulian%2FDeveloper%2Fgo-cli%2Fconfig.yaml%22%2C%22scheme%22%3A%22file%22%7D%7D) file.
   - Prompt you to enter the SSH configuration details.

## Usage

TransferCLI provides several commands for SSH operations. Below are the available commands:

### Set Configuration

To set or update the SSH configuration, use the `set` command:

```sh
./transfercli set --user your_user --password your_password --host your_host --port 22 --remoteDir ./remote/dir --localDir ./local/dir
```

### List Files

To list files in a remote directory, use the list command:

```sh
./transfercli list
```

### Transfer Files

To transfer files from a remote directory to a local directory, use the `transfer` command:

```sh
./transfercli transfer
```

### Configuration

The SSH configuration is stored in the `config.yaml` file. The configuration file contains the following fields:

```yaml
user: your_user
password: your_password
host: your_host
port: 22
remoteDir: ./remote/dir
localDir: ./local/dir
```

You can manually edit this file or use the `set` command to update the configuration.

Development
To contribute to TransferCLI, follow these steps:

1. Fork the repository.
2. Create a new branch:

```sh
git checkout -b feature-branch
```

3. Make your changes.

4. Commit your changes:

```sh
git commit -m "Description of changes
```

5. Push to the branch:

```sh
git push origin feature-branch
```

6. Create a pull request.

### License

This project is licensed under the MIT License. See the LICENSE file for details.

### Acknowledgements

Cobra for the CLI framework.
Go for the programming language.
YAML for the configuration format.
Contact
For any questions or feedback, please open an issue on the GitHub repository.
