VAGRANTFILE_API_VERSION = "2"

Vagrant.configure(VAGRANTFILE_API_VERSION) do |config|
  
  # Master Node
  config.vm.define "master" do |master|
    master.vm.box = "ubuntu/bionic64"
    master.vm.hostname = "master"
    master.vm.network "private_network", ip: "192.168.56.10"  # Static IP
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

  # Jenkins Agent Node
  config.vm.define "agent" do |agent|
    agent.vm.box = "ubuntu/bionic64"
    agent.vm.hostname = "agent"
    agent.vm.network "private_network", ip: "192.168.56.11"  # Static IP
    # agent.vm.network "forwarded_port", guest: 8080, host: 8080, host_ip: "127.0.0.1"
    # agent.vm.network "forwarded_port", guest: 80, host: 80, host_ip: "127.0.0.1"

    agent.vm.provision "shell", path: "./script/installDocker.sh"
    # agent.vm.provision "shell", path: "./script/installJenkins.sh"
    # agent.vm.provision "shell", inline: <<-SHELL
    # sudo cat /var/lib/jenkins/secrets/initialAdminPassword
    # SHELL
  end

  # # Worker Nodes
  # workers = {
  #   "prod1"   => "192.168.56.21",
  #   "prod2"   => "192.168.56.22",
  #   "dev"     => "192.168.56.23",
  #   "feature" => "192.168.56.24"
  # }

  # workers.each do |name, ip|
  #   config.vm.define name do |worker|
  #     worker.vm.box = "ubuntu/bionic64"
  #     worker.vm.hostname = name
  #     worker.vm.network "private_network", ip: ip  # Assign static IP
  #     worker.vm.provision "shell", path: "./script/installDocker.sh"
  #     worker.vm.provision "shell", inline: <<-SHELL
  #     SHELL
  #   end
  # end

end
