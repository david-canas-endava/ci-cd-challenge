curl --output localstack-cli-4.1.1-linux-amd64-onefile.tar.gz \
    --location https://github.com/localstack/localstack-cli/releases/download/v4.1.1/localstack-cli-4.1.1-linux-amd64-onefile.tar.gz
sudo tar xvzf localstack-cli-4.1.1-linux-*-onefile.tar.gz -C /usr/local/bin

sudo apt install pipx
pipx install awscli-local[ver1]
source ~/.bashrc
