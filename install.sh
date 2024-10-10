#!/bin/bash

# Function to check if a command exists
command_exists() {
    command -v "$1" >/dev/null 2>&1
}

# Check if Go is installed
if command_exists go; then
    echo -e "\033[32mGo is already installed.\033[0m"
else
    echo -e "\033[31mGo is not installed. Installing Go...\033[0m"
    wget https://golang.org/dl/go1.17.6.linux-amd64.tar.gz
    sudo tar -C /usr/local -xzf go1.17.6.linux-amd64.tar.gz
    echo "export PATH=$PATH:/usr/local/go/bin" >> ~/.profile
    source ~/.profile
    echo -e "\033[32mGo installed successfully.\033[0m"
fi

# Install project dependencies
echo -e "\033[32mInstalling project dependencies...\033[0m"
go mod tidy

# Create config.yaml file
CONFIG_FILE="config.yaml"
if [ -f "$CONFIG_FILE" ]; then
    echo -e "\033[32m$CONFIG_FILE already exists.\033[0m"
else
    echo -e "\033[31m$CONFIG_FILE does not exist. Creating $CONFIG_FILE...\033[0m"
    touch $CONFIG_FILE
    echo -e "\033[32m$CONFIG_FILE created successfully.\033[0m"
fi

# Prompt user for configuration values
read -p "Enter SSH username: " user
read -p "Enter SSH password: " password
read -p "Enter SSH host: " host
read -p "Enter SSH port (default is 22): " port
read -p "Enter remote directory (relative to user login; ex: ./remote/Dir): " remoteDir
read -p "Enter local directory (relative path to source file; ex: ./local/Dir): " localDir


# Update config.yaml with user input
cat <<EOL > $CONFIG_FILE
user: $user
password: $password
host: $host
port: $port
remoteDir: $remoteDir
localDir: $localDir
EOL

echo -e "\033[32mConfiguration saved to $CONFIG_FILE\033[0m"

echo -e "\033[32mInstallation and setup completed successfully.\033[0m"