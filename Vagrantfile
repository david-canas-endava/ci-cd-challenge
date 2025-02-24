VAGRANTFILE_API_VERSION = "2"

Vagrant.configure(VAGRANTFILE_API_VERSION) do |config|
  
  #MARK: Master
  # Master Node
  config.vm.define "master" do |master|
    master.vm.box = "ubuntu/bionic64"
    master.vm.hostname = "master"
    master.vm.network "private_network", ip: "192.168.56.10"  # Static IP
    master.vm.network "forwarded_port", guest: 8080, host: 8080, host_ip: "127.0.0.1"
    master.vm.provider "virtualbox" do |vb|
      vb.memory = 4096  # 4GB RAM
      vb.cpus = 2       # 2 CPU cores
    end
    master.vm.provision "shell", path: "./script/installDocker.sh"    
    master.vm.provision "shell", inline: <<-SHELL
    mkdir -p /app
    SHELL
    master.vm.synced_folder "./script/master", "/app"
    master.vm.provision "shell", inline: <<-SHELL
    cd /app
    docker compose up -d
    SHELL
    master.vm.provision "shell", path: "./script/installJenkins.sh"
    master.vm.provision "shell", inline: <<-SHELL
    sudo cat /var/lib/jenkins/secrets/initialAdminPassword
    SHELL
  end

  #MARK: Agent
  # Jenkins Agent Node
  config.vm.define "agent" do |agent|
    agent.vm.box = "ubuntu/bionic64"
    agent.vm.hostname = "agent"
    agent.vm.network "private_network", ip: "192.168.56.11"  # Static IP
    # agent.vm.network "forwarded_port", guest: 8080, host: 8080, host_ip: "127.0.0.1"

    agent.vm.provision "shell", path: "./script/installDocker.sh"

    agent.vm.provision "shell", inline: <<-SHELL
      echo -e '{ \n\t "insecure-registries" : [ "192.168.56.10:5000" ]\n}' | sudo tee /etc/docker/daemon.json
      sudo systemctl daemon-reload
      sudo systemctl restart docker
    SHELL

    agent.vm.provision "shell", inline: <<-SHELL
      sudo apt update
      sudo apt install -y openjdk-17-jre
    SHELL


    agent.vm.provision "shell", inline: <<-SHELL
      sudo mkdir -p /home/jenkins_agent
      sudo chown -R vagrant:vagrant /home/jenkins_agent
      sudo chmod -R 777 /home/jenkins_agent 
      cd /home/jenkins_agent

      curl -sO http://192.168.56.10:8080/jnlpJars/agent.jar;java -jar agent.jar -url http://192.168.56.10:8080/ -secret 226828a6da74445c9a34609e8fbdf1559ecd7808e5afea4e2c69f4e784928b67 -name agent -webSocket -workDir "/home/jenkins_agent" & disown
    SHELL
    # agent.vm.synced_folder "./app", "/app"
    # agent.vm.provision "shell", inline: <<-SHELL
    #   cd /app      
    #   docker build -t app .
    #   docker tag app 192.168.56.10:5000/app
    #   docker push 192.168.56.10:5000/app
    # SHELL
  end

  #MARK: workers
  # Worker Nodes
  workers = {
    "prod1"   => "192.168.56.21",
    "prod2"   => "192.168.56.22",
    "dev"     => "192.168.56.23",
    "feature" => "192.168.56.24"
  }


  workers.each do |name, ip|
    config.vm.define name do |worker|
      worker.vm.box = "ubuntu/bionic64"
      worker.vm.hostname = name
      worker.vm.network "private_network", ip: ip  # Assign static IP
      worker.vm.provision "shell", path: "./script/installDocker.sh"
      worker.vm.provision "shell", inline: <<-SHELL
        echo -e '{ \n\t "insecure-registries" : [ "192.168.56.10:5000" ]\n}' | sudo tee /etc/docker/daemon.json
        sudo systemctl daemon-reload
        sudo systemctl restart docker
      SHELL

      worker.vm.provision "shell", inline: <<-SHELL
        docker run -d --network host --name app_#{name} --volume appData:/etc/todos 192.168.56.10:5000/app
      SHELL
    end
  end

end
